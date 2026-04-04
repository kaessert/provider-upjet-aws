// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package replicationgroup implements the shared CRUD logic for ReplicationGroupRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the ReplicationGroupCR interface.
package replicationgroup

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
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/password"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	xpresource "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusternativev2 "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta2/native"
	nativehelper "github.com/upbound/provider-aws/v2/internal/native"
)

const (
	errDescribe      = "cannot describe ElastiCache Replication Group"
	errCreate        = "cannot create ElastiCache Replication Group"
	errUpdate        = "cannot update ElastiCache Replication Group"
	errUpdateShard   = "cannot update ElastiCache Replication Group shard configuration"
	errDelete        = "cannot delete ElastiCache Replication Group"
	errListTags      = "cannot list tags for ElastiCache Replication Group"
	errAddTags       = "cannot add tags to ElastiCache Replication Group"
	errDelTags       = "cannot remove tags from ElastiCache Replication Group"
	errGenerateToken = "cannot generate auth token"
	errWriteToken    = "cannot write auth token to secret"
	errReadToken     = "cannot read auth token from secret"
	errGetSecret     = "cannot get auth token secret"
)

// ElastiCacheRGClient is the interface for AWS ElastiCache operations
// required by the replication group controller. Defined as an interface to enable
// mocking in unit tests; *awselasticache.Client satisfies it.
type ElastiCacheRGClient interface {
	CreateReplicationGroup(ctx context.Context, params *awselasticache.CreateReplicationGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateReplicationGroupOutput, error)
	DescribeReplicationGroups(ctx context.Context, params *awselasticache.DescribeReplicationGroupsInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeReplicationGroupsOutput, error)
	ModifyReplicationGroup(ctx context.Context, params *awselasticache.ModifyReplicationGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyReplicationGroupOutput, error)
	ModifyReplicationGroupShardConfiguration(ctx context.Context, params *awselasticache.ModifyReplicationGroupShardConfigurationInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyReplicationGroupShardConfigurationOutput, error)
	DeleteReplicationGroup(ctx context.Context, params *awselasticache.DeleteReplicationGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteReplicationGroupOutput, error)
	ListTagsForResource(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error)
	AddTagsToResource(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error)
	RemoveTagsFromResource(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error)
}

// ReplicationGroupCR abstracts over cluster-scoped and namespaced ReplicationGroupRAW types.
type ReplicationGroupCR interface {
	xpresource.Managed
	GetForProvider() *clusternativev2.ReplicationGroupRAWParameters
	GetInitProvider() *clusternativev2.ReplicationGroupRAWInitParameters
	GetAtProvider() clusternativev2.ReplicationGroupRAWObservation
	SetAtProvider(clusternativev2.ReplicationGroupRAWObservation)
}

// ExternalClient implements the shared CRUD logic for ReplicationGroupRAW resources.
type ExternalClient struct {
	// Client is the AWS ElastiCache SDK client (interface for testability).
	Client ElastiCacheRGClient
	// Kube is the Kubernetes client for reading/writing secrets.
	Kube client.Client
}

// Observe checks whether the external ReplicationGroupRAW resource exists and is up-to-date.
//
//nolint:gocyclo
func (e *ExternalClient) Observe(ctx context.Context, cr ReplicationGroupCR) (managed.ExternalObservation, error) {
	extName := nativehelper.GetExternalName(cr)
	if extName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	asyncState := nativehelper.GetAsyncState(cr)

	resp, err := e.Client.DescribeReplicationGroups(ctx, &awselasticache.DescribeReplicationGroupsInput{
		ReplicationGroupId: aws.String(extName),
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

	if len(resp.ReplicationGroups) == 0 {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	rg := resp.ReplicationGroups[0]
	status := aws.ToString(rg.Status)

	// Populate observation — always, regardless of status, preserving internally-tracked fields.
	existing := cr.GetAtProvider()
	obs := observationFromSDK(rg)
	// Preserve auth_token_update_strategy: AWS doesn't return it; we track it internally
	// to detect when auth token rotation has been applied (idempotency).
	obs.AuthTokenUpdateStrategy = existing.AuthTokenUpdateStrategy
	cr.SetAtProvider(obs)

	// Handle transitional and terminal states (lowercase for ReplicationGroup).
	switch status {
	case "creating", "modifying", "snapshotting":
		cr.SetConditions(xpv1.Unavailable())
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	case "deleting":
		cr.SetConditions(xpv1.Deleting())
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	case "create-failed":
		nativehelper.ClearAsyncState(cr)
		return managed.ExternalObservation{ResourceExists: true},
			fmt.Errorf("replication group %q creation failed", extName)
	case "available":
		nativehelper.ClearAsyncState(cr)
		cr.SetConditions(xpv1.Available())
	default:
		// Treat unknown statuses as transitional.
		cr.SetConditions(xpv1.Unavailable())
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	}

	// Resource is available — fetch tags and check isUpToDate.
	var observedTags []ectypes.Tag
	if rg.ARN != nil {
		tagsResp, err := e.Client.ListTagsForResource(ctx, &awselasticache.ListTagsForResourceInput{
			ResourceName: rg.ARN,
		})
		if err == nil && tagsResp != nil {
			observedTags = tagsResp.TagList
		}
		// Non-fatal: tags errors are ignored to keep Observe idempotent.
	}

	// Update tags in atProvider.
	if len(observedTags) > 0 {
		o := cr.GetAtProvider()
		o.Tags = tagsToMap(observedTags)
		cr.SetAtProvider(o)
	}

	// Late-initialize AWS-defaulted fields (spec §8).
	// Returns early with ResourceLateInitialized=true so the reconciler saves the
	// spec before calling isUpToDate. The next Observe will find all fields set and
	// isUpToDate=true, breaking any potential infinite-update loop.
	if lateInitializeRG(cr.GetForProvider(), rg) {
		connDetails := e.buildConnectionDetails(ctx, cr, rg)
		return managed.ExternalObservation{
			ResourceExists:          true,
			ResourceUpToDate:        false,
			ResourceLateInitialized: true,
			ConnectionDetails:       connDetails,
		}, nil
	}

	// Build connection details — includes auth_token re-read from Secret.
	connDetails := e.buildConnectionDetails(ctx, cr, rg)

	// Check whether the resource matches the desired spec.
	upToDate := isUpToDate(cr, rg, observedTags)

	// Set the Test condition for uptest compatibility (mirrors upjet behaviour).
	nativehelper.SetTestConditionIfAnnotated(cr, upToDate)

	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  upToDate,
		ConnectionDetails: connDetails,
	}, nil
}

// Create creates the external ReplicationGroupRAW resource.
//
//nolint:gocyclo
func (e *ExternalClient) Create(ctx context.Context, cr ReplicationGroupCR) (managed.ExternalCreation, error) {
	spec := cr.GetForProvider()
	extName := nativehelper.GetExternalName(cr)

	// Auth token handling (spec §1):
	// Write BEFORE the AWS call so it is recoverable on pod restart.
	var authToken string
	if spec.AutoGenerateAuthToken != nil && *spec.AutoGenerateAuthToken {
		// Idempotent: check if Secret already has a value (retry safety).
		existing, _ := e.readAuthTokenFromRef(ctx, spec.AuthTokenSecretRef)
		if existing != "" {
			authToken = existing
		} else {
			// Generate a new secure random token.
			token, err := password.Generate()
			if err != nil {
				return managed.ExternalCreation{}, nativehelper.Wrap(err, errGenerateToken)
			}
			// Write to K8s Secret with OwnerReference BEFORE calling AWS.
			if err := e.writeAuthTokenToSecret(ctx, cr, token); err != nil {
				return managed.ExternalCreation{}, nativehelper.Wrap(err, errWriteToken)
			}
			authToken = token
		}
	} else if spec.AuthTokenSecretRef != nil {
		// User provided an explicit auth token via secretRef — read it.
		token, err := e.readAuthTokenFromRef(ctx, spec.AuthTokenSecretRef)
		if err != nil {
			return managed.ExternalCreation{}, nativehelper.Wrap(err, errReadToken)
		}
		authToken = token
	}

	input := buildCreateInput(spec, extName, authToken)
	resp, err := e.Client.CreateReplicationGroup(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, nativehelper.Wrap(err, errCreate)
	}

	// ParameterAsIdentifier: set external name from response.
	if resp.ReplicationGroup != nil && resp.ReplicationGroup.ReplicationGroupId != nil {
		nativehelper.SetExternalName(cr, *resp.ReplicationGroup.ReplicationGroupId)
	}

	// Record the in-flight async create operation.
	rid, _ := awsmiddleware.GetRequestIDMetadata(resp.ResultMetadata)
	nativehelper.SetAsyncState(cr, nativehelper.AsyncState{
		Operation: "creating",
		StartedAt: time.Now(),
		RequestID: rid,
	})

	// Publish auth_token as connection detail if applicable.
	connDetails := managed.ConnectionDetails{}
	if authToken != "" {
		connDetails["auth_token"] = []byte(authToken)
	}
	return managed.ExternalCreation{ConnectionDetails: connDetails}, nil
}

// Update updates the external ReplicationGroupRAW resource.
// Implements the update decomposition from spec §8: one change per reconcile cycle.
//
//nolint:gocyclo
func (e *ExternalClient) Update(ctx context.Context, cr ReplicationGroupCR) (managed.ExternalUpdate, error) {
	spec := cr.GetForProvider()
	extName := nativehelper.GetExternalName(cr)

	// Describe current state to determine what changed.
	descResp, err := e.Client.DescribeReplicationGroups(ctx, &awselasticache.DescribeReplicationGroupsInput{
		ReplicationGroupId: aws.String(extName),
	})
	if err != nil {
		return managed.ExternalUpdate{}, nativehelper.Wrap(err, errDescribe)
	}
	if len(descResp.ReplicationGroups) == 0 {
		return managed.ExternalUpdate{}, fmt.Errorf("replication group %q not found during update", extName)
	}
	rg := descResp.ReplicationGroups[0]

	// Step 1: NumNodeGroups change → ModifyReplicationGroupShardConfiguration.
	// (spec §8: shard configuration change must come first)
	if spec.NumNodeGroups != nil {
		specNumNG := int32(*spec.NumNodeGroups)
		obsNumNG := int32(len(rg.NodeGroups))
		if specNumNG != obsNumNG {
			shardResp, err := e.Client.ModifyReplicationGroupShardConfiguration(ctx, &awselasticache.ModifyReplicationGroupShardConfigurationInput{
				ReplicationGroupId: aws.String(extName),
				NodeGroupCount:     aws.Int32(specNumNG),
				ApplyImmediately:   aws.Bool(true),
			})
			if err != nil {
				return managed.ExternalUpdate{}, nativehelper.Wrap(err, errUpdateShard)
			}
			rid, _ := awsmiddleware.GetRequestIDMetadata(shardResp.ResultMetadata)
			nativehelper.SetAsyncState(cr, nativehelper.AsyncState{
				Operation: "updating",
				StartedAt: time.Now(),
				RequestID: rid,
			})
			return managed.ExternalUpdate{}, nil
		}
	}

	// Step 2: Auth token rotation.
	// Detect change by comparing spec.AuthTokenUpdateStrategy vs atProvider.AuthTokenUpdateStrategy.
	// After rotation, we set atProvider.AuthTokenUpdateStrategy = spec.AuthTokenUpdateStrategy
	// so subsequent reconciles see them as equal and skip this step.
	atProvider := cr.GetAtProvider()
	if spec.AuthTokenUpdateStrategy != nil {
		specStrategy := *spec.AuthTokenUpdateStrategy
		alreadyApplied := atProvider.AuthTokenUpdateStrategy != nil && *atProvider.AuthTokenUpdateStrategy == specStrategy
		if !alreadyApplied {
			authToken, err := e.readAuthTokenFromRef(ctx, spec.AuthTokenSecretRef)
			if err != nil {
				return managed.ExternalUpdate{}, nativehelper.Wrap(err, errReadToken)
			}
			modResp, err := e.Client.ModifyReplicationGroup(ctx, &awselasticache.ModifyReplicationGroupInput{
				ReplicationGroupId:      aws.String(extName),
				AuthToken:               aws.String(authToken),
				AuthTokenUpdateStrategy: ectypes.AuthTokenUpdateStrategyType(specStrategy),
			})
			if err != nil {
				return managed.ExternalUpdate{}, nativehelper.Wrap(err, errUpdate)
			}
			// Mark strategy as applied in atProvider so isUpToDate returns true for auth token.
			o := cr.GetAtProvider()
			o.AuthTokenUpdateStrategy = spec.AuthTokenUpdateStrategy
			cr.SetAtProvider(o)

			rid, _ := awsmiddleware.GetRequestIDMetadata(modResp.ResultMetadata)
			nativehelper.SetAsyncState(cr, nativehelper.AsyncState{
				Operation: "updating",
				StartedAt: time.Now(),
				RequestID: rid,
			})
			return managed.ExternalUpdate{}, nil
		}
	}

	// Step 3: Other field changes → ModifyReplicationGroup.
	// (SecurityGroupNames intentionally never included — EC2-Classic legacy field)
	if !fieldChangesUpToDate(spec, rg) {
		input := buildModifyInput(spec, extName)
		modResp, err := e.Client.ModifyReplicationGroup(ctx, input)
		if err != nil {
			return managed.ExternalUpdate{}, nativehelper.Wrap(err, errUpdate)
		}
		rid, _ := awsmiddleware.GetRequestIDMetadata(modResp.ResultMetadata)
		nativehelper.SetAsyncState(cr, nativehelper.AsyncState{
			Operation: "updating",
			StartedAt: time.Now(),
			RequestID: rid,
		})
		return managed.ExternalUpdate{}, nil
	}

	// Step 4: Tag-only changes (synchronous — no async needed).
	if rg.ARN != nil {
		if err := e.syncTags(ctx, spec.Tags, *rg.ARN); err != nil {
			return managed.ExternalUpdate{}, err
		}
	}

	return managed.ExternalUpdate{}, nil
}

// Delete deletes the external ReplicationGroupRAW resource.
func (e *ExternalClient) Delete(ctx context.Context, cr ReplicationGroupCR) (managed.ExternalDelete, error) {
	cr.SetConditions(xpv1.Deleting())

	spec := cr.GetForProvider()
	extName := nativehelper.GetExternalName(cr)

	input := &awselasticache.DeleteReplicationGroupInput{
		ReplicationGroupId: aws.String(extName),
	}
	if spec.FinalSnapshotIdentifier != nil {
		input.FinalSnapshotIdentifier = spec.FinalSnapshotIdentifier
	}

	resp, err := e.Client.DeleteReplicationGroup(ctx, input)
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

// readAuthTokenFromRef reads the auth token value from the K8s Secret referenced
// by the SecretKeySelector. Returns ("", nil) when ref is nil or Secret is not found.
func (e *ExternalClient) readAuthTokenFromRef(ctx context.Context, ref *xpv1.SecretKeySelector) (string, error) {
	if ref == nil {
		return "", nil
	}
	s := &corev1.Secret{}
	if err := e.Kube.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, s); err != nil {
		if xpresource.IgnoreNotFound(err) == nil {
			// Secret not found — return empty (no token yet).
			return "", nil
		}
		return "", nativehelper.Wrap(err, errGetSecret)
	}
	return string(s.Data[ref.Key]), nil
}

// writeAuthTokenToSecret writes the auth token to the K8s Secret referenced by
// authTokenSecretRef. Sets an OwnerReference on newly-created Secrets so they are
// garbage-collected when the managed resource is deleted (matching TF PasswordGenerator
// behavior in config/cluster/common/common.go).
func (e *ExternalClient) writeAuthTokenToSecret(ctx context.Context, cr ReplicationGroupCR, token string) error {
	spec := cr.GetForProvider()
	ref := spec.AuthTokenSecretRef
	if ref == nil {
		return fmt.Errorf("authTokenSecretRef is not set, cannot write auth token")
	}

	s := &corev1.Secret{}
	err := e.Kube.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, s)
	if xpresource.IgnoreNotFound(err) != nil {
		return nativehelper.Wrap(err, errGetSecret)
	}

	// Set name/namespace (for the create case where s is empty).
	s.SetName(ref.Name)
	s.SetNamespace(ref.Namespace)

	// Set OwnerReference only when we are creating the Secret (it didn't exist before).
	// If it already exists (WasCreated returns true), someone else created it — don't take ownership.
	if !meta.WasCreated(s) {
		meta.AddOwnerReference(s, meta.AsOwner(meta.TypedReferenceTo(cr, cr.GetObjectKind().GroupVersionKind())))
	}

	if s.Data == nil {
		s.Data = make(map[string][]byte, 1)
	}
	s.Data[ref.Key] = []byte(token)

	return nativehelper.Wrap(xpresource.NewAPIPatchingApplicator(e.Kube).Apply(ctx, s), errWriteToken)
}

// buildConnectionDetails builds the connection details map from the SDK response.
// Also re-reads auth_token from Secret (spec §7: re-publish on every Observe).
//
//nolint:gocyclo
func (e *ExternalClient) buildConnectionDetails(ctx context.Context, cr ReplicationGroupCR, rg ectypes.ReplicationGroup) managed.ConnectionDetails {
	cd := managed.ConnectionDetails{}

	// Port: always published from first node group's primary endpoint.
	if len(rg.NodeGroups) > 0 {
		ng := rg.NodeGroups[0]
		if ng.PrimaryEndpoint != nil && ng.PrimaryEndpoint.Port != nil {
			cd["port"] = []byte(strconv.Itoa(int(*ng.PrimaryEndpoint.Port)))
		}

		if rg.ConfigurationEndpoint != nil {
			// Cluster mode enabled: use configuration endpoint.
			if rg.ConfigurationEndpoint.Address != nil {
				cd["configuration_endpoint_address"] = []byte(*rg.ConfigurationEndpoint.Address)
			}
		} else {
			// Cluster mode disabled: use primary/reader from first node group.
			if ng.PrimaryEndpoint != nil && ng.PrimaryEndpoint.Address != nil {
				cd["primary_endpoint_address"] = []byte(*ng.PrimaryEndpoint.Address)
			}
			if ng.ReaderEndpoint != nil && ng.ReaderEndpoint.Address != nil {
				cd["reader_endpoint_address"] = []byte(*ng.ReaderEndpoint.Address)
			}
		}
	}

	// auth_token: re-read from Secret on every Observe (spec §7).
	spec := cr.GetForProvider()
	if spec.AuthTokenSecretRef != nil {
		authToken, _ := e.readAuthTokenFromRef(ctx, spec.AuthTokenSecretRef)
		if authToken != "" {
			cd["auth_token"] = []byte(authToken)
		}
	}

	return cd
}

// buildCreateInput constructs the CreateReplicationGroupInput from the CR spec.
//
//nolint:gocyclo
func buildCreateInput(spec *clusternativev2.ReplicationGroupRAWParameters, extName, authToken string) *awselasticache.CreateReplicationGroupInput {
	input := &awselasticache.CreateReplicationGroupInput{
		ReplicationGroupId:          aws.String(extName),
		ReplicationGroupDescription: aws.String(aws.ToString(spec.Description)), // required
		AuthToken:                   nillableString(authToken),
		AutomaticFailoverEnabled:    spec.AutomaticFailoverEnabled,
		CacheNodeType:               spec.NodeType,
		CacheParameterGroupName:     spec.ParameterGroupName,
		CacheSubnetGroupName:        spec.SubnetGroupName,
		Engine:                      defaultToRedis(spec.Engine),
		EngineVersion:               spec.EngineVersion,
		GlobalReplicationGroupId:    spec.GlobalReplicationGroupID,
		KmsKeyId:                    spec.KMSKeyID,
		MultiAZEnabled:              spec.MultiAzEnabled,
		NotificationTopicArn:        spec.NotificationTopicArn,
		SnapshotName:                spec.SnapshotName,
		SnapshotRetentionLimit:      int32PtrFromFloat64(spec.SnapshotRetentionLimit),
		SnapshotWindow:              spec.SnapshotWindow,
		TransitEncryptionEnabled:    spec.TransitEncryptionEnabled,
	}

	// AtRestEncryptionEnabled: *string → *bool (spec §13 type mismatch).
	if spec.AtRestEncryptionEnabled != nil {
		if v, err := strconv.ParseBool(*spec.AtRestEncryptionEnabled); err == nil {
			input.AtRestEncryptionEnabled = aws.Bool(v)
		}
	}

	// AutoMinorVersionUpgrade: *string → *bool (spec §13 type mismatch).
	if spec.AutoMinorVersionUpgrade != nil {
		if v, err := strconv.ParseBool(*spec.AutoMinorVersionUpgrade); err == nil {
			input.AutoMinorVersionUpgrade = aws.Bool(v)
		}
	}

	// ClusterMode enum.
	if spec.ClusterMode != nil {
		input.ClusterMode = ectypes.ClusterMode(*spec.ClusterMode)
	}

	// DataTieringEnabled is a *bool in the CreateReplicationGroupInput.
	if spec.DataTieringEnabled != nil {
		input.DataTieringEnabled = spec.DataTieringEnabled
	}

	// IpDiscovery enum.
	if spec.IPDiscovery != nil {
		input.IpDiscovery = ectypes.IpDiscovery(*spec.IPDiscovery)
	}

	// MaintenanceWindow.
	if spec.MaintenanceWindow != nil {
		input.PreferredMaintenanceWindow = spec.MaintenanceWindow
	}

	// NetworkType enum.
	if spec.NetworkType != nil {
		input.NetworkType = ectypes.NetworkType(*spec.NetworkType)
	}

	// NumCacheClusters.
	if spec.NumCacheClusters != nil {
		input.NumCacheClusters = aws.Int32(int32(*spec.NumCacheClusters))
	}

	// NumNodeGroups.
	if spec.NumNodeGroups != nil {
		input.NumNodeGroups = aws.Int32(int32(*spec.NumNodeGroups))
	}

	// Port.
	if spec.Port != nil {
		input.Port = aws.Int32(int32(*spec.Port))
	}

	// PreferredCacheClusterAZs.
	if len(spec.PreferredCacheClusterAzs) > 0 {
		input.PreferredCacheClusterAZs = derefStringSlice(spec.PreferredCacheClusterAzs)
	}

	// ReplicasPerNodeGroup.
	if spec.ReplicasPerNodeGroup != nil {
		input.ReplicasPerNodeGroup = aws.Int32(int32(*spec.ReplicasPerNodeGroup))
	}

	// SecurityGroupIds (never SecurityGroupNames — EC2-Classic legacy).
	if len(spec.SecurityGroupIds) > 0 {
		input.SecurityGroupIds = derefStringSlice(spec.SecurityGroupIds)
	}

	// SnapshotArns.
	if len(spec.SnapshotArns) > 0 {
		input.SnapshotArns = derefStringSlice(spec.SnapshotArns)
	}

	// Tags.
	for k, v := range spec.Tags {
		tagKey := k
		input.Tags = append(input.Tags, ectypes.Tag{Key: aws.String(tagKey), Value: v})
	}

	// TransitEncryptionMode.
	if spec.TransitEncryptionMode != nil {
		input.TransitEncryptionMode = ectypes.TransitEncryptionMode(*spec.TransitEncryptionMode)
	}

	// UserGroupIds.
	if len(spec.UserGroupIds) > 0 {
		input.UserGroupIds = derefStringSlice(spec.UserGroupIds)
	}

	// LogDeliveryConfigurations.
	for _, ldc := range spec.LogDeliveryConfiguration {
		req := logDeliveryConfigToRequest(ldc)
		if req != nil {
			input.LogDeliveryConfigurations = append(input.LogDeliveryConfigurations, *req)
		}
	}

	// NodeGroupConfiguration.
	for _, ngc := range spec.NodeGroupConfiguration {
		sdkNGC := ectypes.NodeGroupConfiguration{
			NodeGroupId:             ngc.NodeGroupID,
			PrimaryAvailabilityZone: ngc.PrimaryAvailabilityZone,
			PrimaryOutpostArn:       ngc.PrimaryOutpostArn,
			Slots:                   ngc.Slots,
		}
		if ngc.ReplicaCount != nil {
			sdkNGC.ReplicaCount = aws.Int32(int32(*ngc.ReplicaCount))
		}
		if len(ngc.ReplicaAvailabilityZones) > 0 {
			sdkNGC.ReplicaAvailabilityZones = derefStringSlice(ngc.ReplicaAvailabilityZones)
		}
		if len(ngc.ReplicaOutpostArns) > 0 {
			sdkNGC.ReplicaOutpostArns = derefStringSlice(ngc.ReplicaOutpostArns)
		}
		input.NodeGroupConfiguration = append(input.NodeGroupConfiguration, sdkNGC)
	}

	return input
}

// buildModifyInput constructs the ModifyReplicationGroupInput from the CR spec.
// SecurityGroupNames is intentionally excluded (EC2-Classic legacy; AWS rejects it in VPC).
//
//nolint:gocyclo
func buildModifyInput(spec *clusternativev2.ReplicationGroupRAWParameters, extName string) *awselasticache.ModifyReplicationGroupInput {
	input := &awselasticache.ModifyReplicationGroupInput{
		ReplicationGroupId:          aws.String(extName),
		ReplicationGroupDescription: spec.Description,
		ApplyImmediately:            spec.ApplyImmediately,
		AutomaticFailoverEnabled:    spec.AutomaticFailoverEnabled,
		CacheNodeType:               spec.NodeType,
		CacheParameterGroupName:     spec.ParameterGroupName,
		Engine:                      spec.Engine,
		EngineVersion:               spec.EngineVersion,
		MultiAZEnabled:              spec.MultiAzEnabled,
		NotificationTopicArn:        spec.NotificationTopicArn,
		SnapshotRetentionLimit:      int32PtrFromFloat64(spec.SnapshotRetentionLimit),
		SnapshotWindow:              spec.SnapshotWindow,
		TransitEncryptionEnabled:    spec.TransitEncryptionEnabled,
	}

	// AutoMinorVersionUpgrade: *string → *bool.
	if spec.AutoMinorVersionUpgrade != nil {
		if v, err := strconv.ParseBool(*spec.AutoMinorVersionUpgrade); err == nil {
			input.AutoMinorVersionUpgrade = aws.Bool(v)
		}
	}

	// MaintenanceWindow.
	if spec.MaintenanceWindow != nil {
		input.PreferredMaintenanceWindow = spec.MaintenanceWindow
	}

	// SecurityGroupIds (never SecurityGroupNames).
	if len(spec.SecurityGroupIds) > 0 {
		input.SecurityGroupIds = derefStringSlice(spec.SecurityGroupIds)
	}

	// TransitEncryptionMode.
	if spec.TransitEncryptionMode != nil {
		input.TransitEncryptionMode = ectypes.TransitEncryptionMode(*spec.TransitEncryptionMode)
	}

	// UserGroup: determine adds and removes based on current UserGroupIds vs spec.
	// We send all desired user group IDs as UserGroupIdsToAdd.
	if len(spec.UserGroupIds) > 0 {
		input.UserGroupIdsToAdd = derefStringSlice(spec.UserGroupIds)
	} else {
		input.RemoveUserGroups = aws.Bool(true)
	}

	// LogDeliveryConfigurations.
	for _, ldc := range spec.LogDeliveryConfiguration {
		req := logDeliveryConfigToRequest(ldc)
		if req != nil {
			input.LogDeliveryConfigurations = append(input.LogDeliveryConfigurations, *req)
		}
	}

	return input
}

// lateInitializeRG copies AWS-defaulted fields from the observed ReplicationGroup
// into spec when the spec fields are nil. Returns true if any field was changed.
//
// This prevents infinite reconciliation loops when optional fields (such as NodeType)
// are omitted from the user manifest and AWS has them set (e.g. an RG created as part
// of a global replication group inherits NodeType). Without late-init, isUpToDate
// would compare "" (nil spec) vs "cache.r7g.medium" (AWS) every cycle and always
// return false, triggering Update on every reconcile iteration (spec §8, §17).
func lateInitializeRG(spec *clusternativev2.ReplicationGroupRAWParameters, rg ectypes.ReplicationGroup) bool {
	changed := false
	changed = nativehelper.LateInitializeStringPtr(&spec.NodeType, rg.CacheNodeType) || changed
	changed = nativehelper.LateInitializeStringPtr(&spec.SnapshotWindow, rg.SnapshotWindow) || changed
	if spec.SnapshotRetentionLimit == nil && rg.SnapshotRetentionLimit != nil {
		v := float64(*rg.SnapshotRetentionLimit)
		spec.SnapshotRetentionLimit = &v
		changed = true
	}
	// ClusterMode: AWS enum string (non-empty) → *string.
	if spec.ClusterMode == nil && rg.ClusterMode != "" {
		cm := string(rg.ClusterMode)
		spec.ClusterMode = &cm
		changed = true
	}
	return changed
}

// fieldChangesUpToDate returns true when the non-shard, non-auth-token, non-tag fields
// match between spec and the observed AWS state. SecurityGroupNames is intentionally
// never compared (EC2-Classic legacy field — spec §6).
//
//nolint:gocyclo
func fieldChangesUpToDate(spec *clusternativev2.ReplicationGroupRAWParameters, rg ectypes.ReplicationGroup) bool {
	// Description.
	if aws.ToString(spec.Description) != aws.ToString(rg.Description) {
		return false
	}

	// NodeType (CacheNodeType in SDK): guard against nil — RGs created from global
	// replication groups may not have NodeType in spec (it is inherited from the GRG).
	// After late-initialization the spec will be populated; this nil check prevents
	// a transient spurious update.
	if spec.NodeType != nil && aws.ToString(spec.NodeType) != aws.ToString(rg.CacheNodeType) {
		return false
	}

	// SnapshotRetentionLimit.
	specRetention := int32(0)
	if spec.SnapshotRetentionLimit != nil {
		specRetention = int32(*spec.SnapshotRetentionLimit)
	}
	obsRetention := int32(0)
	if rg.SnapshotRetentionLimit != nil {
		obsRetention = *rg.SnapshotRetentionLimit
	}
	if specRetention != obsRetention {
		return false
	}

	// SnapshotWindow.
	if spec.SnapshotWindow != nil && aws.ToString(spec.SnapshotWindow) != aws.ToString(rg.SnapshotWindow) {
		return false
	}

	// AutomaticFailoverEnabled.
	if spec.AutomaticFailoverEnabled != nil {
		expected := *spec.AutomaticFailoverEnabled
		actual := rg.AutomaticFailover == ectypes.AutomaticFailoverStatusEnabled ||
			rg.AutomaticFailover == ectypes.AutomaticFailoverStatusEnabling
		if expected != actual {
			return false
		}
	}

	// MultiAzEnabled.
	if spec.MultiAzEnabled != nil {
		expected := *spec.MultiAzEnabled
		actual := rg.MultiAZ == ectypes.MultiAZStatusEnabled
		if expected != actual {
			return false
		}
	}

	// AutoMinorVersionUpgrade: *string → *bool.
	if spec.AutoMinorVersionUpgrade != nil {
		if v, err := strconv.ParseBool(*spec.AutoMinorVersionUpgrade); err == nil {
			if rg.AutoMinorVersionUpgrade != nil && v != *rg.AutoMinorVersionUpgrade {
				return false
			}
		}
	}

	// SecurityGroupIds: not directly comparable from ReplicationGroup response
	// (SGs are reported on member clusters). We rely on always sending them in Modify.
	// SecurityGroupNames is intentionally NOT compared — EC2-Classic legacy.
	_ = spec.SecurityGroupIds // acknowledged — comparison not possible at RG level

	// TransitEncryptionMode.
	if spec.TransitEncryptionMode != nil {
		actual := string(rg.TransitEncryptionMode)
		if *spec.TransitEncryptionMode != actual && actual != "" {
			return false
		}
	}

	// UserGroupIds (order-independent set comparison).
	if spec.UserGroupIds != nil {
		if !sortedStringSliceEqual(derefStringSlice(spec.UserGroupIds), rg.UserGroupIds) {
			return false
		}
	}

	return true
}

// isUpToDate returns true when the CR spec matches the observed AWS state.
// Called only when the resource is in "available" status.
//
//nolint:gocyclo
func isUpToDate(cr ReplicationGroupCR, rg ectypes.ReplicationGroup, observedTags []ectypes.Tag) bool {
	spec := cr.GetForProvider()
	atProvider := cr.GetAtProvider()

	// Check fields that trigger ModifyReplicationGroupShardConfiguration.
	if spec.NumNodeGroups != nil {
		specNG := int32(*spec.NumNodeGroups)
		obsNG := int32(len(rg.NodeGroups))
		if specNG != obsNG {
			return false
		}
	}

	// Check auth token update strategy (tracks pending rotation).
	// spec §8: auth token change detected by spec vs atProvider mismatch.
	if spec.AuthTokenUpdateStrategy != nil {
		specStrategy := *spec.AuthTokenUpdateStrategy
		if atProvider.AuthTokenUpdateStrategy == nil || *atProvider.AuthTokenUpdateStrategy != specStrategy {
			return false
		}
	}

	// Check fields that trigger ModifyReplicationGroup.
	if !fieldChangesUpToDate(spec, rg) {
		return false
	}

	// Check tags.
	if !tagsUpToDate(spec.Tags, observedTags) {
		return false
	}

	return true
}

// syncTags reconciles the desired tags on the resource using its ARN.
func (e *ExternalClient) syncTags(ctx context.Context, specTags map[string]*string, arn string) error {
	tagsResp, err := e.Client.ListTagsForResource(ctx, &awselasticache.ListTagsForResourceInput{
		ResourceName: aws.String(arn),
	})
	if err != nil {
		return nativehelper.Wrap(err, errListTags)
	}
	observed := tagsResp.TagList

	desired := mapToTags(specTags)
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

// observationFromSDK converts the SDK ReplicationGroup struct to the CR observation type.
//
//nolint:gocyclo
func observationFromSDK(rg ectypes.ReplicationGroup) clusternativev2.ReplicationGroupRAWObservation {
	obs := clusternativev2.ReplicationGroupRAWObservation{
		Arn:            rg.ARN,
		ID:             rg.ReplicationGroupId,
		Description:    rg.Description,
		Engine:         rg.Engine,
		KMSKeyID:       rg.KmsKeyId,
		SnapshotWindow: rg.SnapshotWindow,
		NodeType:       rg.CacheNodeType,
	}

	// AtRestEncryptionEnabled: *bool → *string (preserve CRD schema parity with TF).
	if rg.AtRestEncryptionEnabled != nil {
		v := strconv.FormatBool(*rg.AtRestEncryptionEnabled)
		obs.AtRestEncryptionEnabled = &v
	}

	// AutoMinorVersionUpgrade: *bool → *string (preserve CRD schema parity with TF).
	if rg.AutoMinorVersionUpgrade != nil {
		v := strconv.FormatBool(*rg.AutoMinorVersionUpgrade)
		obs.AutoMinorVersionUpgrade = &v
	}

	// AutomaticFailoverEnabled.
	switch rg.AutomaticFailover {
	case ectypes.AutomaticFailoverStatusEnabled, ectypes.AutomaticFailoverStatusEnabling:
		obs.AutomaticFailoverEnabled = aws.Bool(true)
	case ectypes.AutomaticFailoverStatusDisabled, ectypes.AutomaticFailoverStatusDisabling:
		obs.AutomaticFailoverEnabled = aws.Bool(false)
	}

	// ClusterEnabled.
	obs.ClusterEnabled = rg.ClusterEnabled

	// ClusterMode.
	if rg.ClusterMode != "" {
		cm := string(rg.ClusterMode)
		obs.ClusterMode = &cm
	}

	// ConfigurationEndpointAddress (cluster mode enabled).
	if rg.ConfigurationEndpoint != nil {
		obs.ConfigurationEndpointAddress = rg.ConfigurationEndpoint.Address
	}

	// DataTieringEnabled.
	switch rg.DataTiering {
	case ectypes.DataTieringStatusEnabled:
		obs.DataTieringEnabled = aws.Bool(true)
	case ectypes.DataTieringStatusDisabled:
		obs.DataTieringEnabled = aws.Bool(false)
	}

	// GlobalReplicationGroupID.
	if rg.GlobalReplicationGroupInfo != nil {
		obs.GlobalReplicationGroupID = rg.GlobalReplicationGroupInfo.GlobalReplicationGroupId
	}

	// IPDiscovery.
	if rg.IpDiscovery != "" {
		ipd := string(rg.IpDiscovery)
		obs.IPDiscovery = &ipd
	}

	// LogDeliveryConfiguration.
	for _, ldc := range rg.LogDeliveryConfigurations {
		entry := clusternativev2.RGLogDeliveryConfigurationRAWObservation{
			LogType:   aws.String(string(ldc.LogType)),
			LogFormat: aws.String(string(ldc.LogFormat)),
		}
		if ldc.DestinationDetails != nil {
			if ldc.DestinationDetails.CloudWatchLogsDetails != nil {
				entry.Destination = ldc.DestinationDetails.CloudWatchLogsDetails.LogGroup
				entry.DestinationType = aws.String(string(ectypes.DestinationTypeCloudWatchLogs))
			} else if ldc.DestinationDetails.KinesisFirehoseDetails != nil {
				entry.Destination = ldc.DestinationDetails.KinesisFirehoseDetails.DeliveryStream
				entry.DestinationType = aws.String(string(ectypes.DestinationTypeKinesisFirehose))
			}
		}
		obs.LogDeliveryConfiguration = append(obs.LogDeliveryConfiguration, entry)
	}

	// MemberClusters.
	for _, mc := range rg.MemberClusters {
		mcc := mc
		obs.MemberClusters = append(obs.MemberClusters, &mcc)
	}

	// MultiAzEnabled.
	switch rg.MultiAZ {
	case ectypes.MultiAZStatusEnabled:
		obs.MultiAzEnabled = aws.Bool(true)
	case ectypes.MultiAZStatusDisabled:
		obs.MultiAzEnabled = aws.Bool(false)
	}

	// NetworkType.
	if rg.NetworkType != "" {
		nt := string(rg.NetworkType)
		obs.NetworkType = &nt
	}

	// NumNodeGroups.
	if len(rg.NodeGroups) > 0 {
		numNG := float64(len(rg.NodeGroups))
		obs.NumNodeGroups = &numNG

		// Endpoints from first node group.
		ng := rg.NodeGroups[0]
		if ng.PrimaryEndpoint != nil {
			obs.PrimaryEndpointAddress = ng.PrimaryEndpoint.Address
			if ng.PrimaryEndpoint.Port != nil {
				port := float64(*ng.PrimaryEndpoint.Port)
				obs.Port = &port
			}
		}
		if ng.ReaderEndpoint != nil {
			obs.ReaderEndpointAddress = ng.ReaderEndpoint.Address
		}

		// NodeGroupConfiguration (basic mapping from NodeGroup).
		for _, ng := range rg.NodeGroups {
			ngObs := clusternativev2.NodeGroupConfigurationRAWObservation{
				NodeGroupID: ng.NodeGroupId,
				Slots:       ng.Slots,
			}
			// ReplicaCount = total members - 1 (primary).
			if len(ng.NodeGroupMembers) > 1 {
				rc := float64(len(ng.NodeGroupMembers) - 1)
				ngObs.ReplicaCount = &rc
			}
			obs.NodeGroupConfiguration = append(obs.NodeGroupConfiguration, ngObs)
		}
	}

	// SnapshotRetentionLimit.
	if rg.SnapshotRetentionLimit != nil {
		v := float64(*rg.SnapshotRetentionLimit)
		obs.SnapshotRetentionLimit = &v
	}

	// TransitEncryptionEnabled.
	obs.TransitEncryptionEnabled = rg.TransitEncryptionEnabled

	// TransitEncryptionMode.
	if rg.TransitEncryptionMode != "" {
		tem := string(rg.TransitEncryptionMode)
		obs.TransitEncryptionMode = &tem
	}

	// UserGroupIds.
	for _, ugid := range rg.UserGroupIds {
		ug := ugid
		obs.UserGroupIds = append(obs.UserGroupIds, &ug)
	}

	// NOTE: AuthTokenUpdateStrategy is NOT set here — AWS doesn't return it.
	// It is preserved from the existing atProvider in Observe (tracked internally).
	// NOTE: EngineVersionActual requires DescribeCacheClusters; left nil for now.
	// NOTE: SubnetGroupName is not directly in ReplicationGroup response.

	return obs
}

// logDeliveryConfigToRequest converts a spec log delivery config to an SDK request.
func logDeliveryConfigToRequest(ldc clusternativev2.RGLogDeliveryConfigurationRAWParameters) *ectypes.LogDeliveryConfigurationRequest {
	if ldc.DestinationType == nil || ldc.LogType == nil {
		return nil
	}
	req := &ectypes.LogDeliveryConfigurationRequest{
		DestinationType: ectypes.DestinationType(*ldc.DestinationType),
		LogType:         ectypes.LogType(*ldc.LogType),
		Enabled:         aws.Bool(true),
	}
	if ldc.LogFormat != nil {
		req.LogFormat = ectypes.LogFormat(*ldc.LogFormat)
	}
	if ldc.Destination != nil {
		req.DestinationDetails = &ectypes.DestinationDetails{}
		switch req.DestinationType {
		case ectypes.DestinationTypeCloudWatchLogs:
			req.DestinationDetails.CloudWatchLogsDetails = &ectypes.CloudWatchLogsDestinationDetails{
				LogGroup: ldc.Destination,
			}
		case ectypes.DestinationTypeKinesisFirehose:
			req.DestinationDetails.KinesisFirehoseDetails = &ectypes.KinesisFirehoseDestinationDetails{
				DeliveryStream: ldc.Destination,
			}
		}
	}
	return req
}

// ── Tag helpers (same pattern as serverlesscache) ──────────────────────────────

func tagsUpToDate(specTags map[string]*string, observed []ectypes.Tag) bool {
	observedMap := make(map[string]string, len(observed))
	for _, t := range observed {
		if t.Key != nil {
			observedMap[*t.Key] = aws.ToString(t.Value)
		}
	}
	for k, v := range specTags {
		if obsVal, exists := observedMap[k]; !exists || obsVal != aws.ToString(v) {
			return false
		}
	}
	for k := range observedMap {
		if _, exists := specTags[k]; !exists {
			return false
		}
	}
	return true
}

func mapToTags(m map[string]*string) []ectypes.Tag {
	tags := make([]ectypes.Tag, 0, len(m))
	for k, v := range m {
		tagKey := k
		tags = append(tags, ectypes.Tag{Key: aws.String(tagKey), Value: v})
	}
	return tags
}

func tagsToMap(tags []ectypes.Tag) map[string]*string {
	m := make(map[string]*string, len(tags))
	for _, t := range tags {
		if t.Key != nil {
			m[*t.Key] = t.Value
		}
	}
	return m
}

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

// ── General helpers ────────────────────────────────────────────────────────────

// nillableString returns nil for an empty string, or a pointer to the string otherwise.
func nillableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
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

// int32PtrFromFloat64 converts a *float64 to *int32.
func int32PtrFromFloat64(v *float64) *int32 {
	if v == nil {
		return nil
	}
	i := int32(*v)
	return &i
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

// defaultToRedis returns the provided engine string or defaults to "redis" if nil.
// The AWS ElastiCache API requires the Engine field for CreateReplicationGroup.
// The Terraform provider defaults this field to "redis", so native controllers
// must mirror that behaviour for parity.
func defaultToRedis(engine *string) *string {
	if engine != nil {
		return engine
	}
	return aws.String("redis")
}
