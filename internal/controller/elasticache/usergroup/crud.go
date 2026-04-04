// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package usergroup implements the shared CRUD logic for UserGroupRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the UserGroupCR interface.
package usergroup

import (
	"context"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	errDescribe = "cannot describe ElastiCache User Group"
	errCreate   = "cannot create ElastiCache User Group"
	errUpdate   = "cannot modify ElastiCache User Group"
	errDelete   = "cannot delete ElastiCache User Group"
	errListTags = "cannot list tags for ElastiCache User Group"
	errAddTags  = "cannot add tags to ElastiCache User Group"
	errDelTags  = "cannot remove tags from ElastiCache User Group"
)

// ElastiCacheUGClient is the interface for AWS ElastiCache operations required
// by the user group controller. Defined as an interface to enable mocking in
// unit tests; *awselasticache.Client satisfies it.
type ElastiCacheUGClient interface {
	DescribeUserGroups(ctx context.Context, params *awselasticache.DescribeUserGroupsInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeUserGroupsOutput, error)
	CreateUserGroup(ctx context.Context, params *awselasticache.CreateUserGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateUserGroupOutput, error)
	ModifyUserGroup(ctx context.Context, params *awselasticache.ModifyUserGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyUserGroupOutput, error)
	DeleteUserGroup(ctx context.Context, params *awselasticache.DeleteUserGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteUserGroupOutput, error)
	ListTagsForResource(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error)
	AddTagsToResource(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error)
	RemoveTagsFromResource(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error)
}

// UserGroupCR abstracts over cluster-scoped and namespaced UserGroupRAW types.
type UserGroupCR interface {
	resource.Managed
	GetForProvider() *clusternative.UserGroupRAWParameters
	GetInitProvider() *clusternative.UserGroupRAWInitParameters
	GetAtProvider() clusternative.UserGroupRAWObservation
	SetAtProvider(clusternative.UserGroupRAWObservation)
}

// ExternalClient implements the shared CRUD logic for UserGroupRAW resources.
type ExternalClient struct {
	// Client is the AWS ElastiCache SDK client (interface for testability).
	Client ElastiCacheUGClient
	// Kube is the Kubernetes client (reserved for future use, e.g. secret reads).
	Kube client.Client
}

// Observe checks whether the external UserGroupRAW resource exists and is up-to-date.
//
// Transitional state handling (spec §3):
//
//	"active"    → Available; proceed to isUpToDate
//	"creating"  → Unavailable; return UpToDate=true (prevent spurious Update)
//	"modifying" → Unavailable; return UpToDate=true
//	"deleting"  → Deleting; return UpToDate=true
//
// Without this, Observe calls isUpToDate on a "creating" UserGroup where AWS
// defaults aren't applied yet → returns false → triggers Update → AWS returns
// InvalidUserGroupState.
func (e *ExternalClient) Observe(ctx context.Context, cr UserGroupCR) (managed.ExternalObservation, error) {
	extName := nativehelper.GetExternalName(cr)
	if extName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	resp, err := e.Client.DescribeUserGroups(ctx, &awselasticache.DescribeUserGroupsInput{
		UserGroupId: aws.String(extName),
	})
	if err != nil {
		if nativehelper.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, nativehelper.Wrap(err, errDescribe)
	}

	if len(resp.UserGroups) == 0 {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	ug := resp.UserGroups[0]
	status := aws.ToString(ug.Status)

	// Handle transitional states: skip isUpToDate to prevent spurious Updates.
	switch status {
	case "creating", "modifying":
		cr.SetConditions(xpv1.Unavailable())
		setAtProviderFromUserGroup(cr, ug, nil)
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	case "deleting":
		cr.SetConditions(xpv1.Deleting())
		setAtProviderFromUserGroup(cr, ug, nil)
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	}

	// Status is "active" (or unknown) — proceed to full observation.
	cr.SetConditions(xpv1.Available())

	// Fetch tags (non-fatal — tag errors should not fail Observe).
	var observedTags []ectypes.Tag
	if ug.ARN != nil {
		tagsResp, tagErr := e.Client.ListTagsForResource(ctx, &awselasticache.ListTagsForResourceInput{
			ResourceName: ug.ARN,
		})
		if tagErr == nil && tagsResp != nil {
			observedTags = tagsResp.TagList
		}
	}

	setAtProviderFromUserGroup(cr, ug, observedTags)

	upToDate := isUpToDate(cr.GetForProvider(), ug, observedTags)
	nativehelper.SetTestConditionIfAnnotated(cr, upToDate)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

// Create creates the external UserGroupRAW resource.
func (e *ExternalClient) Create(ctx context.Context, cr UserGroupCR) (managed.ExternalCreation, error) {
	spec := cr.GetForProvider()
	extName := nativehelper.GetExternalName(cr)

	input := &awselasticache.CreateUserGroupInput{
		UserGroupId: aws.String(extName),
		Engine:      spec.Engine,
		UserIds:     derefStringSlice(spec.UserIds),
		Tags:        mapToTags(spec.Tags),
	}

	_, err := e.Client.CreateUserGroup(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, nativehelper.Wrap(err, errCreate)
	}

	// For ParameterAsIdentifier, the external name is already set from the CR's
	// external name annotation before Create is called. No SetExternalName needed.

	// UserGroup publishes no connection details.
	return managed.ExternalCreation{}, nil
}

// Update updates the external UserGroupRAW resource.
// It computes the diff of desired UserIds vs observed UserIds, passing adds to
// UserIdsToAdd and removals to UserIdsToRemove. Tags are synced separately.
func (e *ExternalClient) Update(ctx context.Context, cr UserGroupCR) (managed.ExternalUpdate, error) {
	spec := cr.GetForProvider()
	extName := nativehelper.GetExternalName(cr)

	// Compute UserIds diff from atProvider.
	observed := cr.GetAtProvider()
	toAdd, toRemove := diffUserIds(spec.UserIds, observed.UserIds)

	if len(toAdd) > 0 || len(toRemove) > 0 {
		_, err := e.Client.ModifyUserGroup(ctx, &awselasticache.ModifyUserGroupInput{
			UserGroupId:     aws.String(extName),
			UserIdsToAdd:    toAdd,
			UserIdsToRemove: toRemove,
		})
		if err != nil {
			return managed.ExternalUpdate{}, nativehelper.Wrap(err, errUpdate)
		}
	}

	// Sync tags using the ARN from atProvider.
	arn := aws.ToString(observed.Arn)
	if arn != "" {
		if err := e.syncTags(ctx, arn, spec.Tags); err != nil {
			return managed.ExternalUpdate{}, err
		}
	}

	return managed.ExternalUpdate{}, nil
}

// Delete deletes the external UserGroupRAW resource.
// Idempotent: UserGroupNotFoundFault and "already deleting" state are treated as success.
func (e *ExternalClient) Delete(ctx context.Context, cr UserGroupCR) (managed.ExternalDelete, error) {
	extName := nativehelper.GetExternalName(cr)

	_, err := e.Client.DeleteUserGroup(ctx, &awselasticache.DeleteUserGroupInput{
		UserGroupId: aws.String(extName),
	})
	if err != nil {
		if nativehelper.IsNotFound(err) {
			return managed.ExternalDelete{}, nil
		}
		// AWS rejects delete requests when the user group is already in "deleting"
		// state (InvalidUserGroupStateFault). Treat this as a no-op: the deletion
		// is already in progress and the resource will be cleaned up shortly.
		if nativehelper.IsErrorCode(err, "InvalidUserGroupStateFault") {
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, nativehelper.Wrap(err, errDelete)
	}
	return managed.ExternalDelete{}, nil
}

// ── Helpers ────────────────────────────────────────────────────────────────────

// setAtProviderFromUserGroup populates the atProvider observation from the AWS
// UserGroup state and the provided tag list.
func setAtProviderFromUserGroup(cr UserGroupCR, ug ectypes.UserGroup, tags []ectypes.Tag) {
	userIds := make([]*string, 0, len(ug.UserIds))
	for i := range ug.UserIds {
		id := ug.UserIds[i]
		userIds = append(userIds, &id)
	}
	cr.SetAtProvider(clusternative.UserGroupRAWObservation{
		Arn:     ug.ARN,
		Engine:  ug.Engine,
		ID:      ug.UserGroupId,
		Status:  ug.Status,
		UserIds: userIds,
		Tags:    tagsToMap(tags),
	})
}

// isUpToDate returns true when the spec is in sync with the observed AWS state.
func isUpToDate(spec *clusternative.UserGroupRAWParameters, ug ectypes.UserGroup, observedTags []ectypes.Tag) bool {
	// Check UserIds (order-independent set comparison).
	specIDs := derefStringSlice(spec.UserIds)
	awsIDs := ug.UserIds
	if !sortedStringSliceEqual(specIDs, awsIDs) {
		return false
	}

	// Check tags.
	if !tagsUpToDate(spec.Tags, observedTags) {
		return false
	}

	return true
}

// diffUserIds computes which user IDs to add and which to remove.
// desired is the spec list; observed is the atProvider list.
func diffUserIds(desired []*string, observed []*string) (toAdd []string, toRemove []string) {
	desiredSet := make(map[string]struct{}, len(desired))
	for _, id := range desired {
		if id != nil {
			desiredSet[*id] = struct{}{}
		}
	}

	observedSet := make(map[string]struct{}, len(observed))
	for _, id := range observed {
		if id != nil {
			observedSet[*id] = struct{}{}
		}
	}

	// IDs in desired but not in observed → add.
	for id := range desiredSet {
		if _, exists := observedSet[id]; !exists {
			toAdd = append(toAdd, id)
		}
	}

	// IDs in observed but not in desired → remove.
	for id := range observedSet {
		if _, exists := desiredSet[id]; !exists {
			toRemove = append(toRemove, id)
		}
	}

	// Sort for determinism in tests.
	sort.Strings(toAdd)
	sort.Strings(toRemove)

	return toAdd, toRemove
}

// syncTags reconciles desired tags on the user group using its ARN.
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

	// Tags to remove (exist in AWS but not in desired).
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

// derefStringSlice dereferences a []*string slice to []string, skipping nil
// pointers.
func derefStringSlice(s []*string) []string {
	result := make([]string, 0, len(s))
	for _, v := range s {
		if v != nil {
			result = append(result, *v)
		}
	}
	return result
}

// sortedStringSliceEqual compares two string slices in a set-equal manner
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
