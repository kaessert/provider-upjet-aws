// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package parametergroup implements the shared CRUD logic for ParameterGroupRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the ParameterGroupCR interface.
package parametergroup

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	awselasticache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	ectypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta1/native"
	nativehelper "github.com/upbound/provider-aws/v2/internal/native"
)

const (
	errDescribe       = "cannot describe ElastiCache Parameter Group"
	errDescribeParams = "cannot describe ElastiCache Parameter Group parameters"
	errCreate         = "cannot create ElastiCache Parameter Group"
	errUpdate         = "cannot modify ElastiCache Parameter Group"
	errDelete         = "cannot delete ElastiCache Parameter Group"
	errListTags       = "cannot list tags for ElastiCache Parameter Group"
	errAddTags        = "cannot add tags to ElastiCache Parameter Group"
	errDelTags        = "cannot remove tags from ElastiCache Parameter Group"
)

// ElastiCachePGClient is the interface for AWS ElastiCache operations required
// by the parameter group controller. Defined as an interface to enable mocking
// in unit tests; *awselasticache.Client satisfies it.
type ElastiCachePGClient interface {
	DescribeCacheParameterGroups(ctx context.Context, params *awselasticache.DescribeCacheParameterGroupsInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParameterGroupsOutput, error)
	DescribeCacheParameters(ctx context.Context, params *awselasticache.DescribeCacheParametersInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParametersOutput, error)
	CreateCacheParameterGroup(ctx context.Context, params *awselasticache.CreateCacheParameterGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateCacheParameterGroupOutput, error)
	ModifyCacheParameterGroup(ctx context.Context, params *awselasticache.ModifyCacheParameterGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyCacheParameterGroupOutput, error)
	ResetCacheParameterGroup(ctx context.Context, params *awselasticache.ResetCacheParameterGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ResetCacheParameterGroupOutput, error)
	DeleteCacheParameterGroup(ctx context.Context, params *awselasticache.DeleteCacheParameterGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteCacheParameterGroupOutput, error)
	ListTagsForResource(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error)
	AddTagsToResource(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error)
	RemoveTagsFromResource(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error)
}

// ParameterGroupCR abstracts over cluster-scoped and namespaced ParameterGroupRAW types.
type ParameterGroupCR interface {
	resource.Managed
	GetForProvider() *clusternative.ParameterGroupRAWParameters
	GetInitProvider() *clusternative.ParameterGroupRAWInitParameters
	GetAtProvider() clusternative.ParameterGroupRAWObservation
	SetAtProvider(clusternative.ParameterGroupRAWObservation)
}

// ExternalClient implements the shared CRUD logic for ParameterGroupRAW resources.
type ExternalClient struct {
	// Client is the AWS ElastiCache SDK client (interface for testability).
	Client ElastiCachePGClient
	// Kube is the Kubernetes client (reserved for future use, e.g. secret reads).
	Kube client.Client
}

// Observe checks whether the external ParameterGroupRAW resource exists and is up-to-date.
func (e *ExternalClient) Observe(ctx context.Context, cr ParameterGroupCR) (managed.ExternalObservation, error) {
	extName := nativehelper.GetExternalName(cr)
	if extName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	resp, err := e.Client.DescribeCacheParameterGroups(ctx, &awselasticache.DescribeCacheParameterGroupsInput{
		CacheParameterGroupName: aws.String(extName),
	})
	if err != nil {
		if nativehelper.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, nativehelper.Wrap(err, errDescribe)
	}

	if len(resp.CacheParameterGroups) == 0 {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	pg := resp.CacheParameterGroups[0]

	// Fetch user-modified parameters via DescribeCacheParameters(Source="user").
	// This returns only parameters the user has explicitly set, not defaults.
	userParams, err := e.fetchUserParameters(ctx, extName)
	if err != nil {
		return managed.ExternalObservation{}, nativehelper.Wrap(err, errDescribeParams)
	}

	// Fetch tags (non-fatal — tag errors should not fail Observe).
	var observedTags []ectypes.Tag
	if pg.ARN != nil {
		tagsResp, tagErr := e.Client.ListTagsForResource(ctx, &awselasticache.ListTagsForResourceInput{
			ResourceName: pg.ARN,
		})
		if tagErr == nil && tagsResp != nil {
			observedTags = tagsResp.TagList
		}
	}

	// Populate atProvider.
	o := clusternative.ParameterGroupRAWObservation{
		Arn:         pg.ARN,
		Description: pg.Description,
		Family:      pg.CacheParameterGroupFamily,
		ID:          pg.CacheParameterGroupName,
		Name:        pg.CacheParameterGroupName,
		Parameter:   userParamsToObservation(userParams),
		Tags:        tagsToMap(observedTags),
	}
	cr.SetAtProvider(o)

	upToDate := isUpToDate(cr.GetForProvider(), userParams, observedTags)
	nativehelper.SetTestConditionIfAnnotated(cr, upToDate)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

// Create creates the external ParameterGroupRAW resource.
// For IdentifierFromProvider, the AWS-assigned name is stored as the external
// name annotation immediately after the successful API call.
func (e *ExternalClient) Create(ctx context.Context, cr ParameterGroupCR) (managed.ExternalCreation, error) {
	spec := cr.GetForProvider()

	// Determine the parameter group name.
	// Prefer spec.forProvider.name if explicitly set; otherwise use the external
	// name annotation (crossplane initializes it from the CR's metadata.name).
	groupName := aws.ToString(spec.Name)
	if groupName == "" {
		groupName = nativehelper.GetExternalName(cr)
	}

	input := &awselasticache.CreateCacheParameterGroupInput{
		CacheParameterGroupName:   aws.String(groupName),
		CacheParameterGroupFamily: spec.Family,
		Description:               spec.Description,
		Tags:                      mapToTags(spec.Tags),
	}

	resp, err := e.Client.CreateCacheParameterGroup(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, nativehelper.Wrap(err, errCreate)
	}

	// Critical (IdentifierFromProvider): store the provider-assigned name as the
	// external name. Without this call, status.atProvider is lost before the next
	// Observe because the reconciler resets it.
	nativehelper.SetExternalName(cr, aws.ToString(resp.CacheParameterGroup.CacheParameterGroupName))

	// Parameters are not settable via CreateCacheParameterGroup — they will be
	// applied by Update on the next reconcile once Observe detects drift.

	return managed.ExternalCreation{}, nil
}

// Update updates the external ParameterGroupRAW resource.
// It calls ModifyCacheParameterGroup with all desired parameters, and
// ResetCacheParameterGroup for any user-set parameters that have been removed
// from the spec. It also syncs tags.
func (e *ExternalClient) Update(ctx context.Context, cr ParameterGroupCR) (managed.ExternalUpdate, error) {
	spec := cr.GetForProvider()
	extName := nativehelper.GetExternalName(cr)

	// Build the desired set of parameter name-values.
	desiredParams := buildParameterNameValues(spec.Parameter)

	if len(desiredParams) > 0 {
		_, err := e.Client.ModifyCacheParameterGroup(ctx, &awselasticache.ModifyCacheParameterGroupInput{
			CacheParameterGroupName: aws.String(extName),
			ParameterNameValues:     desiredParams,
		})
		if err != nil {
			return managed.ExternalUpdate{}, nativehelper.Wrap(err, errUpdate)
		}
	}

	// Reset parameters that exist in atProvider (user-set) but are not in the
	// desired spec — returning them to their default values.
	observedParams := cr.GetAtProvider().Parameter
	toReset := computeParamsToReset(spec.Parameter, observedParams)
	if len(toReset) > 0 {
		resetInput := make([]ectypes.ParameterNameValue, 0, len(toReset))
		for _, name := range toReset {
			name := name
			resetInput = append(resetInput, ectypes.ParameterNameValue{
				ParameterName: aws.String(name),
			})
		}
		_, err := e.Client.ResetCacheParameterGroup(ctx, &awselasticache.ResetCacheParameterGroupInput{
			CacheParameterGroupName: aws.String(extName),
			ParameterNameValues:     resetInput,
		})
		if err != nil {
			return managed.ExternalUpdate{}, nativehelper.Wrap(err, errUpdate)
		}
	}

	// Sync tags using the ARN from atProvider.
	arn := aws.ToString(cr.GetAtProvider().Arn)
	if arn != "" {
		if err := e.syncTags(ctx, arn, spec.Tags); err != nil {
			return managed.ExternalUpdate{}, err
		}
	}

	return managed.ExternalUpdate{}, nil
}

// Delete deletes the external ParameterGroupRAW resource.
// Idempotent: CacheParameterGroupNotFoundFault is treated as success.
func (e *ExternalClient) Delete(ctx context.Context, cr ParameterGroupCR) (managed.ExternalDelete, error) {
	extName := nativehelper.GetExternalName(cr)

	_, err := e.Client.DeleteCacheParameterGroup(ctx, &awselasticache.DeleteCacheParameterGroupInput{
		CacheParameterGroupName: aws.String(extName),
	})
	if err != nil {
		if nativehelper.IsNotFound(err) {
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, nativehelper.Wrap(err, errDelete)
	}
	return managed.ExternalDelete{}, nil
}

// ── Helpers ────────────────────────────────────────────────────────────────────

// fetchUserParameters retrieves all user-modified parameters for the given
// parameter group, paginating through results as needed.
func (e *ExternalClient) fetchUserParameters(ctx context.Context, groupName string) ([]ectypes.Parameter, error) {
	var params []ectypes.Parameter
	var marker *string
	for {
		resp, err := e.Client.DescribeCacheParameters(ctx, &awselasticache.DescribeCacheParametersInput{
			CacheParameterGroupName: aws.String(groupName),
			Source:                  aws.String("user"),
			Marker:                  marker,
		})
		if err != nil {
			return nil, err
		}
		params = append(params, resp.Parameters...)
		if resp.Marker == nil || aws.ToString(resp.Marker) == "" {
			break
		}
		marker = resp.Marker
	}
	return params, nil
}

// isUpToDate returns true when the desired spec matches the observed AWS state.
func isUpToDate(spec *clusternative.ParameterGroupRAWParameters, userParams []ectypes.Parameter, observedTags []ectypes.Tag) bool {
	return paramsUpToDate(spec.Parameter, userParams) && tagsEqual(spec.Tags, observedTags)
}

// paramsUpToDate returns true when the desired parameter list matches the observed
// user-set parameters (both value equality and absence of extra params).
func paramsUpToDate(desired []clusternative.ParameterRAWParameters, userParams []ectypes.Parameter) bool {
	// Build a map of current user-set parameters.
	currentMap := make(map[string]string, len(userParams))
	for _, p := range userParams {
		if p.ParameterName != nil {
			currentMap[aws.ToString(p.ParameterName)] = aws.ToString(p.ParameterValue)
		}
	}

	// Build a map of desired parameters.
	desiredMap := make(map[string]string, len(desired))
	for _, p := range desired {
		if p.Name != nil {
			desiredMap[aws.ToString(p.Name)] = aws.ToString(p.Value)
		}
	}

	// Check all desired params exist in current with the same value.
	for k, v := range desiredMap {
		if cv, ok := currentMap[k]; !ok || cv != v {
			return false
		}
	}

	// Check no extra user-set params exist that aren't in desired.
	for k := range currentMap {
		if _, ok := desiredMap[k]; !ok {
			return false
		}
	}

	return true
}

// tagsEqual returns true when the spec tags match the observed tags.
func tagsEqual(specTags map[string]*string, observed []ectypes.Tag) bool {
	if len(specTags) != len(observed) {
		return false
	}
	observedMap := tagsToMap(observed)
	for k, v := range specTags {
		obsV, ok := observedMap[k]
		if !ok || aws.ToString(v) != aws.ToString(obsV) {
			return false
		}
	}
	return true
}

// syncTags reconciles the desired tags against the current AWS tags for the ARN.
func (e *ExternalClient) syncTags(ctx context.Context, arn string, desired map[string]*string) error {
	// Fetch current tags.
	tagsResp, err := e.Client.ListTagsForResource(ctx, &awselasticache.ListTagsForResourceInput{
		ResourceName: aws.String(arn),
	})
	if err != nil {
		return nativehelper.Wrap(err, errListTags)
	}
	observed := tagsResp.TagList

	toAdd, toRemove := diffTags(desired, observed)

	if len(toRemove) > 0 {
		keys := make([]string, 0, len(toRemove))
		for _, t := range toRemove {
			keys = append(keys, aws.ToString(t.Key))
		}
		if _, err := e.Client.RemoveTagsFromResource(ctx, &awselasticache.RemoveTagsFromResourceInput{
			ResourceName: aws.String(arn),
			TagKeys:      keys,
		}); err != nil {
			return nativehelper.Wrap(err, errDelTags)
		}
	}

	if len(toAdd) > 0 {
		if _, err := e.Client.AddTagsToResource(ctx, &awselasticache.AddTagsToResourceInput{
			ResourceName: aws.String(arn),
			Tags:         toAdd,
		}); err != nil {
			return nativehelper.Wrap(err, errAddTags)
		}
	}

	return nil
}

// buildParameterNameValues converts spec parameters to AWS ParameterNameValue slice.
func buildParameterNameValues(params []clusternative.ParameterRAWParameters) []ectypes.ParameterNameValue {
	pvs := make([]ectypes.ParameterNameValue, 0, len(params))
	for _, p := range params {
		p := p
		pvs = append(pvs, ectypes.ParameterNameValue{
			ParameterName:  p.Name,
			ParameterValue: p.Value,
		})
	}
	return pvs
}

// computeParamsToReset returns the names of parameters that exist in the observed
// user-set state but are not present in the desired spec (they need to be reset
// to their default values).
func computeParamsToReset(desired []clusternative.ParameterRAWParameters, observed []clusternative.ParameterRAWObservation) []string {
	desiredSet := make(map[string]struct{}, len(desired))
	for _, p := range desired {
		if p.Name != nil {
			desiredSet[*p.Name] = struct{}{}
		}
	}

	var toReset []string
	for _, p := range observed {
		if p.Name != nil {
			if _, ok := desiredSet[*p.Name]; !ok {
				toReset = append(toReset, *p.Name)
			}
		}
	}
	sort.Strings(toReset)
	return toReset
}

// userParamsToObservation converts AWS parameter list to observation type.
func userParamsToObservation(params []ectypes.Parameter) []clusternative.ParameterRAWObservation {
	obs := make([]clusternative.ParameterRAWObservation, 0, len(params))
	for _, p := range params {
		p := p
		obs = append(obs, clusternative.ParameterRAWObservation{
			Name:  p.ParameterName,
			Value: p.ParameterValue,
		})
	}
	return obs
}

// mapToTags converts a map[string]*string to a []ectypes.Tag slice.
func mapToTags(m map[string]*string) []ectypes.Tag {
	tags := make([]ectypes.Tag, 0, len(m))
	for k, v := range m {
		tagKey := k
		tags = append(tags, ectypes.Tag{Key: aws.String(tagKey), Value: v})
	}
	return tags
}

// tagsToMap converts a []ectypes.Tag slice to a map[string]*string.
func tagsToMap(tags []ectypes.Tag) map[string]*string {
	m := make(map[string]*string, len(tags))
	for _, t := range tags {
		if t.Key != nil {
			m[*t.Key] = t.Value
		}
	}
	return m
}

// diffTags computes tags to add and tags to remove.
func diffTags(desired map[string]*string, observed []ectypes.Tag) (toAdd []ectypes.Tag, toRemove []ectypes.Tag) {
	observedMap := tagsToMap(observed)

	// Tags to add or update.
	for k, v := range desired {
		tagKey := k
		obsVal, exists := observedMap[tagKey]
		if !exists || aws.ToString(v) != aws.ToString(obsVal) {
			toAdd = append(toAdd, ectypes.Tag{Key: aws.String(tagKey), Value: v})
		}
	}

	// Tags to remove (exist in AWS but not in desired).
	for _, t := range observed {
		if _, exists := desired[aws.ToString(t.Key)]; !exists {
			toRemove = append(toRemove, t)
		}
	}

	return toAdd, toRemove
}
