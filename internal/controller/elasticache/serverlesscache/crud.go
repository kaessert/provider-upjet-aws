// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package serverlesscache implements the shared CRUD logic for ServerlessCacheRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the ServerlessCacheCR interface.
package serverlesscache

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	awselasticache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	ectypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta1/native"
	nativehelper "github.com/upbound/provider-aws/v2/internal/native"
)

const (
	errDescribe = "cannot describe ElastiCache Serverless Cache"
	errCreate   = "cannot create ElastiCache Serverless Cache"
	errUpdate   = "cannot update ElastiCache Serverless Cache"
	errDelete   = "cannot delete ElastiCache Serverless Cache"
	errListTags = "cannot list tags for ElastiCache Serverless Cache"
	errAddTags  = "cannot add tags to ElastiCache Serverless Cache"
	errDelTags  = "cannot remove tags from ElastiCache Serverless Cache"
)

// ElastiCacheServerlessCacheClient is the interface for AWS ElastiCache operations
// required by the serverless cache controller. Defined as an interface to enable
// mocking in unit tests; *awselasticache.Client satisfies it.
type ElastiCacheServerlessCacheClient interface {
	CreateServerlessCache(ctx context.Context, params *awselasticache.CreateServerlessCacheInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateServerlessCacheOutput, error)
	DescribeServerlessCaches(ctx context.Context, params *awselasticache.DescribeServerlessCachesInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeServerlessCachesOutput, error)
	ModifyServerlessCache(ctx context.Context, params *awselasticache.ModifyServerlessCacheInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyServerlessCacheOutput, error)
	DeleteServerlessCache(ctx context.Context, params *awselasticache.DeleteServerlessCacheInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteServerlessCacheOutput, error)
	ListTagsForResource(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error)
	AddTagsToResource(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error)
	RemoveTagsFromResource(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error)
}

// ServerlessCacheCR abstracts over cluster-scoped and namespaced ServerlessCacheRAW types.
// Both scope types implement this interface so that the shared ExternalClient
// can operate on either without scope-specific logic.
type ServerlessCacheCR interface {
	resource.Managed
	GetForProvider() *clusternative.ServerlessCacheRAWParameters
	GetInitProvider() *clusternative.ServerlessCacheRAWInitParameters
	GetAtProvider() clusternative.ServerlessCacheRAWObservation
	SetAtProvider(clusternative.ServerlessCacheRAWObservation)
}

// ExternalClient implements the shared CRUD logic for ServerlessCacheRAW resources.
// It is exported so that scope-specific wrapper packages can reference it.
type ExternalClient struct {
	// Client is the AWS ElastiCache SDK client (interface for testability).
	Client ElastiCacheServerlessCacheClient
	// Kube is the Kubernetes client for reading referenced resources.
	Kube client.Client
}

// Observe checks whether the external ServerlessCacheRAW resource exists and is up-to-date.
func (e *ExternalClient) Observe(ctx context.Context, cr ServerlessCacheCR) (managed.ExternalObservation, error) { //nolint:gocyclo
	extName := nativehelper.GetExternalName(cr)
	if extName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	asyncState := nativehelper.GetAsyncState(cr)

	resp, err := e.Client.DescribeServerlessCaches(ctx, &awselasticache.DescribeServerlessCachesInput{
		ServerlessCacheName: aws.String(extName),
	})
	if err != nil {
		if nativehelper.IsNotFound(err) {
			if asyncState != nil && asyncState.Operation == "deleting" {
				// Delete has completed — resource is gone.
				nativehelper.ClearAsyncState(cr)
				return managed.ExternalObservation{ResourceExists: false}, nil
			}
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, nativehelper.Wrap(err, errDescribe)
	}

	if len(resp.ServerlessCaches) == 0 {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	sc := resp.ServerlessCaches[0]
	status := aws.ToString(sc.Status)

	// Populate observation — always, regardless of status, so atProvider stays fresh.
	cr.SetAtProvider(observationFromSDK(sc))

	// Handle async in-progress and terminal states.
	// Note: AWS ElastiCache API returns lowercase status values for ServerlessCache
	// (e.g., "creating", "available", "create-failed"), not uppercase.
	switch status {
	case "creating":
		cr.SetConditions(xpv1.Unavailable())
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	case "modifying":
		cr.SetConditions(xpv1.Unavailable())
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	case "deleting":
		cr.SetConditions(xpv1.Deleting())
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	case "create-failed":
		// Auto-recovery: delete the failed resource so the controller can recreate it.
		// AWS allows deleting create-failed resources. After deletion, the next
		// Observe call will see NotFound → ResourceExists=false → Create is triggered.
		nativehelper.ClearAsyncState(cr)
		cr.SetConditions(xpv1.Unavailable())
		_, delErr := e.Client.DeleteServerlessCache(ctx, &awselasticache.DeleteServerlessCacheInput{
			ServerlessCacheName: aws.String(extName),
		})
		if delErr != nil && !nativehelper.IsNotFound(delErr) {
			// Delete failed; surface the error so it appears in the CR status.
			return managed.ExternalObservation{ResourceExists: true},
				fmt.Errorf("serverless cache is in create-failed state and could not be deleted for recovery: %w", delErr)
		}
		// Delete initiated successfully (or resource already gone).
		// Return ResourceExists=true, ResourceUpToDate=true to let the deletion
		// complete on the next poll (when status becomes "deleting" or NotFound).
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	case "available":
		// Clear any stale async annotation from a previous create/modify operation.
		nativehelper.ClearAsyncState(cr)
		cr.SetConditions(xpv1.Available())
	default:
		// Treat unknown statuses as transitional.
		cr.SetConditions(xpv1.Unavailable())
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	}

	// Resource is AVAILABLE — fetch tags and check isUpToDate.
	var observedTags []ectypes.Tag
	if sc.ARN != nil {
		tagsResp, err := e.Client.ListTagsForResource(ctx, &awselasticache.ListTagsForResourceInput{
			ResourceName: sc.ARN,
		})
		if err == nil && tagsResp != nil {
			observedTags = tagsResp.TagList
		}
		// Non-fatal: tags errors are ignored to keep Observe idempotent.
	}

	// Update tags in atProvider
	if len(observedTags) > 0 {
		obs := cr.GetAtProvider()
		obs.Tags = tagsToMap(observedTags)
		cr.SetAtProvider(obs)
	}

	// Build connection details.
	connDetails := buildConnectionDetails(sc)

	// Check whether the resource matches the desired spec.
	upToDate := isUpToDate(cr, sc, observedTags)

	// Set the Test condition for uptest compatibility (mirrors upjet behaviour).
	nativehelper.SetTestConditionIfAnnotated(cr, upToDate)

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: connDetails,
	}, nil
}

// Create creates the external ServerlessCacheRAW resource.
func (e *ExternalClient) Create(ctx context.Context, cr ServerlessCacheCR) (managed.ExternalCreation, error) {
	input := buildCreateInput(cr)
	resp, err := e.Client.CreateServerlessCache(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, nativehelper.Wrap(err, errCreate)
	}

	// NameAsIdentifier: the serverless cache name is the external name.
	// SetExternalName is idempotent if it's already set.
	if resp.ServerlessCache != nil && resp.ServerlessCache.ServerlessCacheName != nil {
		nativehelper.SetExternalName(cr, *resp.ServerlessCache.ServerlessCacheName)
	}

	// Record the in-flight async operation.
	rid, _ := awsmiddleware.GetRequestIDMetadata(resp.ResultMetadata)
	nativehelper.SetAsyncState(cr, nativehelper.AsyncState{
		Operation: "creating",
		StartedAt: time.Now(),
		RequestID: rid,
	})

	return managed.ExternalCreation{}, nil
}

// Update updates the external ServerlessCacheRAW resource.
func (e *ExternalClient) Update(ctx context.Context, cr ServerlessCacheCR) (managed.ExternalUpdate, error) {
	input := buildModifyInput(cr)
	resp, err := e.Client.ModifyServerlessCache(ctx, input)
	if err != nil {
		return managed.ExternalUpdate{}, nativehelper.Wrap(err, errUpdate)
	}

	// Update tags separately (ModifyServerlessCache does not accept Tags).
	if err := e.syncTags(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, err
	}

	// Record the in-flight async operation.
	rid, _ := awsmiddleware.GetRequestIDMetadata(resp.ResultMetadata)
	nativehelper.SetAsyncState(cr, nativehelper.AsyncState{
		Operation: "updating",
		StartedAt: time.Now(),
		RequestID: rid,
	})

	return managed.ExternalUpdate{}, nil
}

// Delete deletes the external ServerlessCacheRAW resource.
func (e *ExternalClient) Delete(ctx context.Context, cr ServerlessCacheCR) (managed.ExternalDelete, error) {
	cr.SetConditions(xpv1.Deleting())

	extName := nativehelper.GetExternalName(cr)
	resp, err := e.Client.DeleteServerlessCache(ctx, &awselasticache.DeleteServerlessCacheInput{
		ServerlessCacheName: aws.String(extName),
	})
	if err != nil {
		if nativehelper.IsNotFound(err) {
			// Already deleted — idempotent.
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, nativehelper.Wrap(err, errDelete)
	}

	// Record the in-flight async delete operation.
	rid, _ := awsmiddleware.GetRequestIDMetadata(resp.ResultMetadata)
	nativehelper.SetAsyncState(cr, nativehelper.AsyncState{
		Operation: "deleting",
		StartedAt: time.Now(),
		RequestID: rid,
	})

	return managed.ExternalDelete{}, nil
}

// ── Private helpers ────────────────────────────────────────────────────────────

// buildCreateInput constructs the CreateServerlessCacheInput from the CR spec.
func buildCreateInput(cr ServerlessCacheCR) *awselasticache.CreateServerlessCacheInput {
	spec := cr.GetForProvider()
	extName := nativehelper.GetExternalName(cr)

	input := &awselasticache.CreateServerlessCacheInput{
		ServerlessCacheName:   aws.String(extName),
		Engine:                spec.Engine,
		Description:           spec.Description,
		DailySnapshotTime:     spec.DailySnapshotTime,
		KmsKeyId:              spec.KMSKeyID,
		MajorEngineVersion:    spec.MajorEngineVersion,
		UserGroupId:           spec.UserGroupID,
		SecurityGroupIds:      derefStringSlice(spec.SecurityGroupIds),
		SubnetIds:             derefStringSlice(spec.SubnetIds),
		SnapshotArnsToRestore: derefStringSlice(spec.SnapshotArnsToRestore),
	}

	if spec.SnapshotRetentionLimit != nil {
		input.SnapshotRetentionLimit = aws.Int32(int32(*spec.SnapshotRetentionLimit))
	}

	if len(spec.CacheUsageLimits) > 0 {
		input.CacheUsageLimits = buildCacheUsageLimitsInput(spec.CacheUsageLimits[0])
	}

	// Build tag list for creation.
	for k, v := range spec.Tags {
		tagKey := k
		input.Tags = append(input.Tags, ectypes.Tag{Key: aws.String(tagKey), Value: v})
	}

	return input
}

// buildModifyInput constructs the ModifyServerlessCacheInput from the CR spec.
// Note: Engine, SubnetIds, KmsKeyId, SnapshotArnsToRestore are immutable — excluded.
func buildModifyInput(cr ServerlessCacheCR) *awselasticache.ModifyServerlessCacheInput {
	spec := cr.GetForProvider()
	extName := nativehelper.GetExternalName(cr)

	input := &awselasticache.ModifyServerlessCacheInput{
		ServerlessCacheName: aws.String(extName),
		Description:         spec.Description,
		DailySnapshotTime:   spec.DailySnapshotTime,
		MajorEngineVersion:  spec.MajorEngineVersion,
		UserGroupId:         spec.UserGroupID,
		SecurityGroupIds:    derefStringSlice(spec.SecurityGroupIds),
	}

	if spec.SnapshotRetentionLimit != nil {
		input.SnapshotRetentionLimit = aws.Int32(int32(*spec.SnapshotRetentionLimit))
	}

	if len(spec.CacheUsageLimits) > 0 {
		input.CacheUsageLimits = buildCacheUsageLimitsInput(spec.CacheUsageLimits[0])
	}

	return input
}

// syncTags reconciles the desired tags on the resource using its ARN.
// It is a no-op when the ARN is not yet known.
func (e *ExternalClient) syncTags(ctx context.Context, cr ServerlessCacheCR) error {
	obs := cr.GetAtProvider()
	if obs.Arn == nil || *obs.Arn == "" {
		return nil // ARN not yet available; skip tag sync.
	}
	arn := *obs.Arn
	spec := cr.GetForProvider()

	// Get current observed tags.
	tagsResp, err := e.Client.ListTagsForResource(ctx, &awselasticache.ListTagsForResourceInput{
		ResourceName: aws.String(arn),
	})
	if err != nil {
		return nativehelper.Wrap(err, errListTags)
	}
	observed := tagsResp.TagList

	// Compute tags to add and remove.
	desired := mapToTags(spec.Tags)
	toAdd, toRemove := diffTags(desired, observed)

	if len(toAdd) > 0 {
		if _, err := e.Client.AddTagsToResource(ctx, &awselasticache.AddTagsToResourceInput{
			ResourceName: aws.String(arn),
			Tags:         toAdd,
		}); err != nil {
			return nativehelper.Wrap(err, errAddTags)
		}
	}

	if len(toRemove) > 0 {
		keys := make([]string, 0, len(toRemove))
		for _, t := range toRemove {
			if t.Key != nil {
				keys = append(keys, *t.Key)
			}
		}
		if _, err := e.Client.RemoveTagsFromResource(ctx, &awselasticache.RemoveTagsFromResourceInput{
			ResourceName: aws.String(arn),
			TagKeys:      keys,
		}); err != nil {
			return nativehelper.Wrap(err, errDelTags)
		}
	}

	return nil
}

// buildCacheUsageLimitsInput converts the spec type to the SDK type.
func buildCacheUsageLimitsInput(params clusternative.CacheUsageLimitsRAWParameters) *ectypes.CacheUsageLimits {
	ul := &ectypes.CacheUsageLimits{}

	if len(params.DataStorage) > 0 {
		ds := params.DataStorage[0]
		sdkDS := &ectypes.DataStorage{
			Unit: ectypes.DataStorageUnitGb, // Only "GB" is valid.
		}
		if ds.Maximum != nil {
			sdkDS.Maximum = aws.Int32(int32(*ds.Maximum))
		}
		if ds.Minimum != nil {
			sdkDS.Minimum = aws.Int32(int32(*ds.Minimum))
		}
		ul.DataStorage = sdkDS
	}

	if len(params.EcpuPerSecond) > 0 {
		ecpu := params.EcpuPerSecond[0]
		sdkECPU := &ectypes.ECPUPerSecond{}
		if ecpu.Maximum != nil {
			sdkECPU.Maximum = aws.Int32(int32(*ecpu.Maximum))
		}
		if ecpu.Minimum != nil {
			sdkECPU.Minimum = aws.Int32(int32(*ecpu.Minimum))
		}
		ul.ECPUPerSecond = sdkECPU
	}

	return ul
}

// observationFromSDK converts the SDK ServerlessCache struct to the CR observation type.
func observationFromSDK(sc ectypes.ServerlessCache) clusternative.ServerlessCacheRAWObservation { //nolint:gocyclo
	obs := clusternative.ServerlessCacheRAWObservation{
		Arn:                sc.ARN,
		ID:                 sc.ServerlessCacheName,
		Engine:             sc.Engine,
		FullEngineVersion:  sc.FullEngineVersion,
		KMSKeyID:           sc.KmsKeyId,
		MajorEngineVersion: sc.MajorEngineVersion,
		Status:             sc.Status,
		DailySnapshotTime:  sc.DailySnapshotTime,
		Description:        sc.Description,
		UserGroupID:        sc.UserGroupId,
	}

	if sc.CreateTime != nil {
		obs.CreateTime = aws.String(sc.CreateTime.UTC().Format(time.RFC3339))
	}

	if sc.SnapshotRetentionLimit != nil {
		v := float64(*sc.SnapshotRetentionLimit)
		obs.SnapshotRetentionLimit = &v
	}

	// Endpoint (nil-guarded).
	if sc.Endpoint != nil {
		ep := clusternative.ServerlessCacheEndpointRAWObservation{
			Address: sc.Endpoint.Address,
		}
		if sc.Endpoint.Port != nil {
			v := float64(*sc.Endpoint.Port)
			ep.Port = &v
		}
		obs.Endpoint = []clusternative.ServerlessCacheEndpointRAWObservation{ep}
	}

	// ReaderEndpoint (nil-guarded).
	if sc.ReaderEndpoint != nil {
		ep := clusternative.ServerlessCacheEndpointRAWObservation{
			Address: sc.ReaderEndpoint.Address,
		}
		if sc.ReaderEndpoint.Port != nil {
			v := float64(*sc.ReaderEndpoint.Port)
			ep.Port = &v
		}
		obs.ReaderEndpoint = []clusternative.ServerlessCacheEndpointRAWObservation{ep}
	}

	// SecurityGroupIds.
	for _, sg := range sc.SecurityGroupIds {
		sgCopy := sg
		obs.SecurityGroupIds = append(obs.SecurityGroupIds, &sgCopy)
	}

	// SubnetIds.
	for _, sn := range sc.SubnetIds {
		snCopy := sn
		obs.SubnetIds = append(obs.SubnetIds, &snCopy)
	}

	// CacheUsageLimits.
	if sc.CacheUsageLimits != nil {
		ul := clusternative.CacheUsageLimitsRAWObservation{}
		if sc.CacheUsageLimits.DataStorage != nil {
			ds := clusternative.DataStorageRAWObservation{
				Unit: aws.String(string(sc.CacheUsageLimits.DataStorage.Unit)),
			}
			if sc.CacheUsageLimits.DataStorage.Maximum != nil {
				v := float64(*sc.CacheUsageLimits.DataStorage.Maximum)
				ds.Maximum = &v
			}
			if sc.CacheUsageLimits.DataStorage.Minimum != nil {
				v := float64(*sc.CacheUsageLimits.DataStorage.Minimum)
				ds.Minimum = &v
			}
			ul.DataStorage = []clusternative.DataStorageRAWObservation{ds}
		}
		if sc.CacheUsageLimits.ECPUPerSecond != nil {
			ecpu := clusternative.EcpuPerSecondRAWObservation{}
			if sc.CacheUsageLimits.ECPUPerSecond.Maximum != nil {
				v := float64(*sc.CacheUsageLimits.ECPUPerSecond.Maximum)
				ecpu.Maximum = &v
			}
			if sc.CacheUsageLimits.ECPUPerSecond.Minimum != nil {
				v := float64(*sc.CacheUsageLimits.ECPUPerSecond.Minimum)
				ecpu.Minimum = &v
			}
			ul.EcpuPerSecond = []clusternative.EcpuPerSecondRAWObservation{ecpu}
		}
		obs.CacheUsageLimits = []clusternative.CacheUsageLimitsRAWObservation{ul}
	}

	return obs
}

// buildConnectionDetails extracts connection detail key-values from the SDK type.
// Uses indexed keys (endpoint_0_*, reader_endpoint_0_*) for backward compatibility
// with existing connection secret consumers.
func buildConnectionDetails(sc ectypes.ServerlessCache) managed.ConnectionDetails {
	cd := managed.ConnectionDetails{}

	// Nil-guard: Endpoint may be nil during CREATING or other transitional states.
	if sc.Endpoint != nil {
		if sc.Endpoint.Address != nil {
			cd["endpoint_0_address"] = []byte(*sc.Endpoint.Address)
		}
		if sc.Endpoint.Port != nil {
			cd["endpoint_0_port"] = []byte(strconv.Itoa(int(*sc.Endpoint.Port)))
		}
	}

	// Nil-guard: ReaderEndpoint may be nil for single-node or transitional states.
	if sc.ReaderEndpoint != nil {
		if sc.ReaderEndpoint.Address != nil {
			cd["reader_endpoint_0_address"] = []byte(*sc.ReaderEndpoint.Address)
		}
		if sc.ReaderEndpoint.Port != nil {
			cd["reader_endpoint_0_port"] = []byte(strconv.Itoa(int(*sc.ReaderEndpoint.Port)))
		}
	}

	return cd
}

// isUpToDate returns true when the CR spec matches the observed AWS state.
// Called only when the resource is in AVAILABLE status.
func isUpToDate(cr ServerlessCacheCR, sc ectypes.ServerlessCache, observedTags []ectypes.Tag) bool { //nolint:gocyclo
	spec := cr.GetForProvider()

	// Description.
	if aws.ToString(spec.Description) != aws.ToString(sc.Description) {
		return false
	}

	// DailySnapshotTime.
	if aws.ToString(spec.DailySnapshotTime) != aws.ToString(sc.DailySnapshotTime) {
		return false
	}

	// MajorEngineVersion.
	if aws.ToString(spec.MajorEngineVersion) != aws.ToString(sc.MajorEngineVersion) {
		return false
	}

	// SnapshotRetentionLimit.
	specRetention := int32(0)
	if spec.SnapshotRetentionLimit != nil {
		specRetention = int32(*spec.SnapshotRetentionLimit)
	}
	observedRetention := int32(0)
	if sc.SnapshotRetentionLimit != nil {
		observedRetention = *sc.SnapshotRetentionLimit
	}
	if specRetention != observedRetention {
		return false
	}

	// UserGroupID.
	if aws.ToString(spec.UserGroupID) != aws.ToString(sc.UserGroupId) {
		return false
	}

	// SecurityGroupIds (order-independent — both are +listType=set).
	if !sortedStringSliceEqual(derefStringSlice(spec.SecurityGroupIds), sc.SecurityGroupIds) {
		return false
	}

	// CacheUsageLimits.
	if !cacheUsageLimitsUpToDate(spec.CacheUsageLimits, sc.CacheUsageLimits) {
		return false
	}

	// Tags.
	if !tagsUpToDate(spec.Tags, observedTags) {
		return false
	}

	return true
}

// cacheUsageLimitsUpToDate returns true when the spec CacheUsageLimits match the observed state.
func cacheUsageLimitsUpToDate(spec []clusternative.CacheUsageLimitsRAWParameters, observed *ectypes.CacheUsageLimits) bool { //nolint:gocyclo
	if len(spec) == 0 && observed == nil {
		return true
	}
	if len(spec) == 0 {
		return true // spec doesn't constrain limits; whatever AWS set is fine.
	}
	if observed == nil {
		return false
	}

	s := spec[0]

	// DataStorage.
	if len(s.DataStorage) > 0 {
		ds := s.DataStorage[0]
		if observed.DataStorage == nil {
			return false
		}
		obsMax := int32(0)
		if observed.DataStorage.Maximum != nil {
			obsMax = *observed.DataStorage.Maximum
		}
		obsMin := int32(0)
		if observed.DataStorage.Minimum != nil {
			obsMin = *observed.DataStorage.Minimum
		}
		specMax := int32(0)
		if ds.Maximum != nil {
			specMax = int32(*ds.Maximum)
		}
		specMin := int32(0)
		if ds.Minimum != nil {
			specMin = int32(*ds.Minimum)
		}
		if specMax != obsMax || specMin != obsMin {
			return false
		}
	}

	// ECPUPerSecond.
	if len(s.EcpuPerSecond) > 0 {
		ecpu := s.EcpuPerSecond[0]
		if observed.ECPUPerSecond == nil {
			return false
		}
		obsMax := int32(0)
		if observed.ECPUPerSecond.Maximum != nil {
			obsMax = *observed.ECPUPerSecond.Maximum
		}
		obsMin := int32(0)
		if observed.ECPUPerSecond.Minimum != nil {
			obsMin = *observed.ECPUPerSecond.Minimum
		}
		specMax := int32(0)
		if ecpu.Maximum != nil {
			specMax = int32(*ecpu.Maximum)
		}
		specMin := int32(0)
		if ecpu.Minimum != nil {
			specMin = int32(*ecpu.Minimum)
		}
		if specMax != obsMax || specMin != obsMin {
			return false
		}
	}

	return true
}

// tagsUpToDate returns true when the spec tags match the observed AWS tags.
func tagsUpToDate(specTags map[string]*string, observed []ectypes.Tag) bool {
	observedMap := make(map[string]string, len(observed))
	for _, t := range observed {
		if t.Key != nil {
			observedMap[*t.Key] = aws.ToString(t.Value)
		}
	}

	// Every spec tag must exist in observed with the same value.
	for k, v := range specTags {
		obsVal, exists := observedMap[k]
		if !exists || obsVal != aws.ToString(v) {
			return false
		}
	}

	// Every observed tag must exist in spec (no extra tags).
	for k := range observedMap {
		if _, exists := specTags[k]; !exists {
			return false
		}
	}

	return true
}

// sortedStringSliceEqual compares two string slices in a set-equal manner (order-independent).
func sortedStringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	aSorted := make([]string, len(a))
	copy(aSorted, a)
	sort.Strings(aSorted)

	bSorted := make([]string, len(b))
	copy(bSorted, b)
	sort.Strings(bSorted)

	for i := range aSorted {
		if aSorted[i] != bSorted[i] {
			return false
		}
	}
	return true
}

// derefStringSlice dereferences a []*string slice to []string, skipping nil pointers.
func derefStringSlice(s []*string) []string {
	result := make([]string, 0, len(s))
	for _, v := range s {
		if v != nil {
			result = append(result, *v)
		}
	}
	return result
}

// mapToTags converts a map[string]*string to a []ectypes.Tag.
func mapToTags(m map[string]*string) []ectypes.Tag {
	tags := make([]ectypes.Tag, 0, len(m))
	for k, v := range m {
		tagKey := k
		tags = append(tags, ectypes.Tag{Key: aws.String(tagKey), Value: v})
	}
	return tags
}

// tagsToMap converts a []ectypes.Tag to a map[string]*string.
func tagsToMap(tags []ectypes.Tag) map[string]*string {
	m := make(map[string]*string, len(tags))
	for _, t := range tags {
		if t.Key != nil {
			m[*t.Key] = t.Value
		}
	}
	return m
}

// diffTags computes the sets of tags to add and remove.
func diffTags(desired, observed []ectypes.Tag) (toAdd, toRemove []ectypes.Tag) { //nolint:gocyclo
	observedMap := make(map[string]string, len(observed))
	for _, t := range observed {
		if t.Key != nil {
			observedMap[*t.Key] = aws.ToString(t.Value)
		}
	}
	desiredMap := make(map[string]string, len(desired))
	for _, t := range desired {
		if t.Key != nil {
			desiredMap[*t.Key] = aws.ToString(t.Value)
		}
	}

	for _, t := range desired {
		if t.Key == nil {
			continue
		}
		if v, exists := observedMap[*t.Key]; !exists || v != aws.ToString(t.Value) {
			toAdd = append(toAdd, t)
		}
	}

	for _, t := range observed {
		if t.Key == nil {
			continue
		}
		if _, exists := desiredMap[*t.Key]; !exists {
			toRemove = append(toRemove, t)
		}
	}

	return toAdd, toRemove
}
