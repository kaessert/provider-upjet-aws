// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package subnetgroup implements the shared CRUD logic for SubnetGroupRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the SubnetGroupCR interface.
package subnetgroup

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
	errDescribe = "cannot describe ElastiCache Subnet Group"
	errCreate   = "cannot create ElastiCache Subnet Group"
	errUpdate   = "cannot modify ElastiCache Subnet Group"
	errDelete   = "cannot delete ElastiCache Subnet Group"
	errListTags = "cannot list tags for ElastiCache Subnet Group"
	errAddTags  = "cannot add tags to ElastiCache Subnet Group"
	errDelTags  = "cannot remove tags from ElastiCache Subnet Group"
)

// ElastiCacheSGClient is the interface for AWS ElastiCache operations
// required by the subnet group controller. Defined as an interface to enable
// mocking in unit tests; *awselasticache.Client satisfies it.
type ElastiCacheSGClient interface {
	DescribeCacheSubnetGroups(ctx context.Context, params *awselasticache.DescribeCacheSubnetGroupsInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheSubnetGroupsOutput, error)
	CreateCacheSubnetGroup(ctx context.Context, params *awselasticache.CreateCacheSubnetGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateCacheSubnetGroupOutput, error)
	ModifyCacheSubnetGroup(ctx context.Context, params *awselasticache.ModifyCacheSubnetGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyCacheSubnetGroupOutput, error)
	DeleteCacheSubnetGroup(ctx context.Context, params *awselasticache.DeleteCacheSubnetGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteCacheSubnetGroupOutput, error)
	ListTagsForResource(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error)
	AddTagsToResource(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error)
	RemoveTagsFromResource(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error)
}

// SubnetGroupCR abstracts over cluster-scoped and namespaced SubnetGroupRAW types.
type SubnetGroupCR interface {
	resource.Managed
	GetForProvider() *clusternative.SubnetGroupRAWParameters
	GetInitProvider() *clusternative.SubnetGroupRAWInitParameters
	GetAtProvider() clusternative.SubnetGroupRAWObservation
	SetAtProvider(clusternative.SubnetGroupRAWObservation)
}

// ExternalClient implements the shared CRUD logic for SubnetGroupRAW resources.
type ExternalClient struct {
	// Client is the AWS ElastiCache SDK client (interface for testability).
	Client ElastiCacheSGClient
	// Kube is the Kubernetes client (reserved for future use, e.g. secret reads).
	Kube client.Client
}

// Observe checks whether the external SubnetGroupRAW resource exists and is up-to-date.
func (e *ExternalClient) Observe(ctx context.Context, cr SubnetGroupCR) (managed.ExternalObservation, error) {
	extName := nativehelper.GetExternalName(cr)
	if extName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	resp, err := e.Client.DescribeCacheSubnetGroups(ctx, &awselasticache.DescribeCacheSubnetGroupsInput{
		CacheSubnetGroupName: aws.String(extName),
	})
	if err != nil {
		if nativehelper.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, nativehelper.Wrap(err, errDescribe)
	}

	if len(resp.CacheSubnetGroups) == 0 {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	sg := resp.CacheSubnetGroups[0]

	// Fetch tags from AWS (non-fatal — tag errors should not fail Observe).
	var observedTags []ectypes.Tag
	if sg.ARN != nil {
		tagsResp, tagErr := e.Client.ListTagsForResource(ctx, &awselasticache.ListTagsForResourceInput{
			ResourceName: sg.ARN,
		})
		if tagErr == nil && tagsResp != nil {
			observedTags = tagsResp.TagList
		}
	}

	// Populate atProvider.
	o := clusternative.SubnetGroupRAWObservation{
		Arn:         sg.ARN,
		Description: sg.CacheSubnetGroupDescription,
		ID:          sg.CacheSubnetGroupName,
		VPCID:       sg.VpcId,
		SubnetIds:   extractSubnetIDs(sg.Subnets),
		Tags:        tagsToMap(observedTags),
	}
	cr.SetAtProvider(o)

	upToDate := isUpToDate(cr.GetForProvider(), sg, observedTags)
	nativehelper.SetTestConditionIfAnnotated(cr, upToDate)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

// Create creates the external SubnetGroupRAW resource.
func (e *ExternalClient) Create(ctx context.Context, cr SubnetGroupCR) (managed.ExternalCreation, error) {
	spec := cr.GetForProvider()
	extName := nativehelper.GetExternalName(cr)

	// CacheSubnetGroupDescription is required by the AWS SDK (non-nil). Default
	// to empty string when not provided, mirroring Terraform's behaviour.
	description := spec.Description
	if description == nil {
		description = aws.String("")
	}

	input := &awselasticache.CreateCacheSubnetGroupInput{
		CacheSubnetGroupName:        aws.String(extName),
		CacheSubnetGroupDescription: description,
		SubnetIds:                   derefStringSlice(spec.SubnetIds),
		Tags:                        mapToTags(spec.Tags),
	}

	_, err := e.Client.CreateCacheSubnetGroup(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, nativehelper.Wrap(err, errCreate)
	}

	// For NameAsIdentifier, the external name is already set to the resource name
	// before Create is called. No additional SetExternalName call is needed.

	return managed.ExternalCreation{}, nil
}

// Update updates the external SubnetGroupRAW resource.
func (e *ExternalClient) Update(ctx context.Context, cr SubnetGroupCR) (managed.ExternalUpdate, error) {
	spec := cr.GetForProvider()
	extName := nativehelper.GetExternalName(cr)

	// Modify description and subnet IDs.
	_, err := e.Client.ModifyCacheSubnetGroup(ctx, &awselasticache.ModifyCacheSubnetGroupInput{
		CacheSubnetGroupName:        aws.String(extName),
		CacheSubnetGroupDescription: spec.Description,
		SubnetIds:                   derefStringSlice(spec.SubnetIds),
	})
	if err != nil {
		return managed.ExternalUpdate{}, nativehelper.Wrap(err, errUpdate)
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

// Delete deletes the external SubnetGroupRAW resource.
// Idempotent: CacheSubnetGroupNotFoundFault is treated as success.
func (e *ExternalClient) Delete(ctx context.Context, cr SubnetGroupCR) (managed.ExternalDelete, error) {
	extName := nativehelper.GetExternalName(cr)

	_, err := e.Client.DeleteCacheSubnetGroup(ctx, &awselasticache.DeleteCacheSubnetGroupInput{
		CacheSubnetGroupName: aws.String(extName),
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

// isUpToDate returns true when the spec is in sync with the observed AWS state.
func isUpToDate(spec *clusternative.SubnetGroupRAWParameters, sg ectypes.CacheSubnetGroup, observedTags []ectypes.Tag) bool {
	// Check description.
	specDesc := aws.ToString(spec.Description)
	awsDesc := aws.ToString(sg.CacheSubnetGroupDescription)
	if specDesc != awsDesc {
		return false
	}

	// Check subnet IDs (order-independent set comparison).
	awsSubnetIDs := derefStringSlice(extractSubnetIDs(sg.Subnets))
	specSubnetIDs := derefStringSlice(spec.SubnetIds)
	if !sortedStringSliceEqual(awsSubnetIDs, specSubnetIDs) {
		return false
	}

	// Check tags.
	if !tagsUpToDate(spec.Tags, observedTags) {
		return false
	}

	return true
}

// syncTags reconciles desired tags on the subnet group using its ARN.
func (e *ExternalClient) syncTags(ctx context.Context, arn string, desired map[string]*string) error {
	tagsResp, err := e.Client.ListTagsForResource(ctx, &awselasticache.ListTagsForResourceInput{
		ResourceName: aws.String(arn),
	})
	if err != nil {
		return nativehelper.Wrap(err, errListTags)
	}

	observed := tagsResp.TagList
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
			keys = append(keys, aws.ToString(t.Key))
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

// tagsUpToDate returns true when the desired spec tags match the observed AWS tags.
func tagsUpToDate(specTags map[string]*string, observed []ectypes.Tag) bool {
	if len(specTags) != len(observed) {
		return false
	}
	observedMap := tagsToMap(observed)
	for k, v := range specTags {
		obsV, ok := observedMap[k]
		if !ok {
			return false
		}
		if aws.ToString(v) != aws.ToString(obsV) {
			return false
		}
	}
	return true
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

	// Tags to remove (exist in AWS but not in spec).
	for _, t := range observed {
		if _, exists := desired[aws.ToString(t.Key)]; !exists {
			toRemove = append(toRemove, t)
		}
	}

	return toAdd, toRemove
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

// extractSubnetIDs extracts subnet identifiers from the AWS Subnet slice.
func extractSubnetIDs(subnets []ectypes.Subnet) []*string {
	ids := make([]*string, 0, len(subnets))
	for i := range subnets {
		if subnets[i].SubnetIdentifier != nil {
			ids = append(ids, subnets[i].SubnetIdentifier)
		}
	}
	return ids
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

// sortedStringSliceEqual compares two []string slices in a set-equal manner
// (order-independent).
func sortedStringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sorted := func(ss []string) []string {
		cp := make([]string, len(ss))
		copy(cp, ss)
		sort.Strings(cp)
		return cp
	}
	a, b = sorted(a), sorted(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
