// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package globalreplicationgroup implements the shared CRUD logic for GlobalReplicationGroupRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the GlobalReplicationGroupCR interface.
package globalreplicationgroup

import (
	"context"

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
	errDescribe = "cannot describe ElastiCache Global Replication Group"
	errCreate   = "cannot create ElastiCache Global Replication Group"
	errUpdate   = "cannot modify ElastiCache Global Replication Group"
	errDelete   = "cannot delete ElastiCache Global Replication Group"
)

// ElastiCacheGRGClient is the interface for AWS ElastiCache operations required
// by the global replication group controller. Defined as an interface to enable
// mocking in unit tests; *awselasticache.Client satisfies it.
type ElastiCacheGRGClient interface {
	DescribeGlobalReplicationGroups(ctx context.Context, params *awselasticache.DescribeGlobalReplicationGroupsInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeGlobalReplicationGroupsOutput, error)
	CreateGlobalReplicationGroup(ctx context.Context, params *awselasticache.CreateGlobalReplicationGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateGlobalReplicationGroupOutput, error)
	ModifyGlobalReplicationGroup(ctx context.Context, params *awselasticache.ModifyGlobalReplicationGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyGlobalReplicationGroupOutput, error)
	DeleteGlobalReplicationGroup(ctx context.Context, params *awselasticache.DeleteGlobalReplicationGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteGlobalReplicationGroupOutput, error)
}

// GlobalReplicationGroupCR abstracts over cluster-scoped and namespaced GlobalReplicationGroupRAW types.
type GlobalReplicationGroupCR interface {
	resource.Managed
	GetForProvider() *clusternative.GlobalReplicationGroupRAWParameters
	GetInitProvider() *clusternative.GlobalReplicationGroupRAWInitParameters
	GetAtProvider() clusternative.GlobalReplicationGroupRAWObservation
	SetAtProvider(clusternative.GlobalReplicationGroupRAWObservation)
}

// ExternalClient implements the shared CRUD logic for GlobalReplicationGroupRAW resources.
type ExternalClient struct {
	// Client is the AWS ElastiCache SDK client (interface for testability).
	Client ElastiCacheGRGClient
	// Kube is the Kubernetes client (reserved for future use, e.g. secret reads).
	Kube client.Client
}

// Observe checks whether the external GlobalReplicationGroupRAW resource exists and is up-to-date.
//
// Transitional state handling (spec §3):
//
//	"available" → Available; proceed to isUpToDate
//	"creating"  → Unavailable; return UpToDate=true (prevent spurious Update)
//	"modifying" → Unavailable; return UpToDate=true
//	"deleting"  → Deleting; return UpToDate=true
//
// Without this, Observe calls isUpToDate on a "creating" GlobalReplicationGroup
// where AWS defaults aren't applied yet → returns false → triggers Update →
// AWS returns InvalidGlobalReplicationGroupState.
func (e *ExternalClient) Observe(ctx context.Context, cr GlobalReplicationGroupCR) (managed.ExternalObservation, error) {
	extName := nativehelper.GetExternalName(cr)
	if extName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	resp, err := e.Client.DescribeGlobalReplicationGroups(ctx, &awselasticache.DescribeGlobalReplicationGroupsInput{
		GlobalReplicationGroupId: aws.String(extName),
		ShowMemberInfo:           aws.Bool(true),
	})
	if err != nil {
		if nativehelper.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, nativehelper.Wrap(err, errDescribe)
	}

	if len(resp.GlobalReplicationGroups) == 0 {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	grg := resp.GlobalReplicationGroups[0]
	status := aws.ToString(grg.Status)

	// Handle transitional states: skip isUpToDate to prevent spurious Updates.
	switch status {
	case "creating", "modifying":
		cr.SetConditions(xpv1.Unavailable())
		setAtProviderFromGRG(cr, grg)
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	case "deleting":
		cr.SetConditions(xpv1.Deleting())
		setAtProviderFromGRG(cr, grg)
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	}

	// Status is "available" (or unknown) — proceed to full observation.
	cr.SetConditions(xpv1.Available())
	setAtProviderFromGRG(cr, grg)

	upToDate := isUpToDate(cr.GetForProvider(), grg)
	nativehelper.SetTestConditionIfAnnotated(cr, upToDate)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

// Create creates the external GlobalReplicationGroupRAW resource.
// For IdentifierFromProvider (spec §11b), the AWS-assigned GlobalReplicationGroupId
// is stored as the external name annotation immediately after the successful API call.
// Without this call, status.atProvider is lost before the next Observe because the
// reconciler resets it.
func (e *ExternalClient) Create(ctx context.Context, cr GlobalReplicationGroupCR) (managed.ExternalCreation, error) {
	spec := cr.GetForProvider()

	input := &awselasticache.CreateGlobalReplicationGroupInput{
		GlobalReplicationGroupIdSuffix:    spec.GlobalReplicationGroupIDSuffix,
		PrimaryReplicationGroupId:         spec.PrimaryReplicationGroupID,
		GlobalReplicationGroupDescription: spec.GlobalReplicationGroupDescription,
	}

	resp, err := e.Client.CreateGlobalReplicationGroup(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, nativehelper.Wrap(err, errCreate)
	}

	// Critical (IdentifierFromProvider §11b): store the provider-assigned ID as
	// the external name. Without this call, status.atProvider is lost before the
	// next Observe because the reconciler resets it.
	nativehelper.SetExternalName(cr, aws.ToString(resp.GlobalReplicationGroup.GlobalReplicationGroupId))

	// GlobalReplicationGroup publishes no connection details.
	return managed.ExternalCreation{}, nil
}

// Update updates the external GlobalReplicationGroupRAW resource.
// All mutable fields are passed to ModifyGlobalReplicationGroup with ApplyImmediately=true.
func (e *ExternalClient) Update(ctx context.Context, cr GlobalReplicationGroupCR) (managed.ExternalUpdate, error) {
	spec := cr.GetForProvider()
	extName := nativehelper.GetExternalName(cr)

	_, err := e.Client.ModifyGlobalReplicationGroup(ctx, &awselasticache.ModifyGlobalReplicationGroupInput{
		GlobalReplicationGroupId:          aws.String(extName),
		ApplyImmediately:                  aws.Bool(true),
		AutomaticFailoverEnabled:          spec.AutomaticFailoverEnabled,
		CacheNodeType:                     spec.CacheNodeType,
		CacheParameterGroupName:           spec.ParameterGroupName,
		Engine:                            spec.Engine,
		EngineVersion:                     spec.EngineVersion,
		GlobalReplicationGroupDescription: spec.GlobalReplicationGroupDescription,
	})
	if err != nil {
		return managed.ExternalUpdate{}, nativehelper.Wrap(err, errUpdate)
	}

	return managed.ExternalUpdate{}, nil
}

// Delete deletes the external GlobalReplicationGroupRAW resource.
// Idempotent: GlobalReplicationGroupNotFoundFault is treated as success.
// RetainPrimaryReplicationGroup=true preserves the primary replication group
// after the global replication group is deleted (matching TF provider behavior).
func (e *ExternalClient) Delete(ctx context.Context, cr GlobalReplicationGroupCR) (managed.ExternalDelete, error) {
	extName := nativehelper.GetExternalName(cr)

	_, err := e.Client.DeleteGlobalReplicationGroup(ctx, &awselasticache.DeleteGlobalReplicationGroupInput{
		GlobalReplicationGroupId:      aws.String(extName),
		RetainPrimaryReplicationGroup: aws.Bool(true),
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

// setAtProviderFromGRG populates the atProvider observation from the AWS
// GlobalReplicationGroup state.
func setAtProviderFromGRG(cr GlobalReplicationGroupCR, grg ectypes.GlobalReplicationGroup) {
	nodeGroups := make([]clusternative.GlobalNodeGroupRAWObservation, 0, len(grg.GlobalNodeGroups))
	for _, ng := range grg.GlobalNodeGroups {
		ng := ng
		nodeGroups = append(nodeGroups, clusternative.GlobalNodeGroupRAWObservation{
			GlobalNodeGroupID: ng.GlobalNodeGroupId,
			Slots:             ng.Slots,
		})
	}

	numNodeGroups := float64(len(grg.GlobalNodeGroups))

	cr.SetAtProvider(clusternative.GlobalReplicationGroupRAWObservation{
		Arn:                               grg.ARN,
		AtRestEncryptionEnabled:           grg.AtRestEncryptionEnabled,
		AuthTokenEnabled:                  grg.AuthTokenEnabled,
		CacheNodeType:                     grg.CacheNodeType,
		ClusterEnabled:                    grg.ClusterEnabled,
		Engine:                            grg.Engine,
		EngineVersion:                     grg.EngineVersion,
		GlobalNodeGroups:                  nodeGroups,
		GlobalReplicationGroupDescription: grg.GlobalReplicationGroupDescription,
		GlobalReplicationGroupID:          grg.GlobalReplicationGroupId,
		ID:                                grg.GlobalReplicationGroupId,
		NumNodeGroups:                     &numNodeGroups,
		TransitEncryptionEnabled:          grg.TransitEncryptionEnabled,
	})
}

// isUpToDate returns true when the desired spec is in sync with the observed
// AWS GlobalReplicationGroup state.
//
// Fields NOT checked (not reliably returned by DescribeGlobalReplicationGroups):
//   - AutomaticFailoverEnabled (set per-member, not at GRG level)
//   - ParameterGroupName (not returned by Describe)
//   - NumNodeGroups (requires IncreaseNodeGroups / DecreaseNodeGroups)
//   - GlobalReplicationGroupIDSuffix (immutable after create)
//   - PrimaryReplicationGroupID (immutable after create)
func isUpToDate(spec *clusternative.GlobalReplicationGroupRAWParameters, grg ectypes.GlobalReplicationGroup) bool {
	// Description
	if aws.ToString(spec.GlobalReplicationGroupDescription) != aws.ToString(grg.GlobalReplicationGroupDescription) {
		return false
	}

	// CacheNodeType — only compare when spec explicitly sets it.
	if spec.CacheNodeType != nil && aws.ToString(spec.CacheNodeType) != aws.ToString(grg.CacheNodeType) {
		return false
	}

	// Engine — only compare when spec explicitly sets it.
	if spec.Engine != nil && aws.ToString(spec.Engine) != aws.ToString(grg.Engine) {
		return false
	}

	// EngineVersion — only compare when spec explicitly sets it.
	// AWS may return a more specific version (e.g., "7.2.4" for requested "7.2").
	// We compare using a prefix match to avoid spurious drift from minor-version
	// suffixes added by AWS.
	if spec.EngineVersion != nil {
		if !versionMatchesSpec(aws.ToString(spec.EngineVersion), aws.ToString(grg.EngineVersion)) {
			return false
		}
	}

	return true
}

// versionMatchesSpec returns true when the AWS-reported version (awsVer) is
// compatible with the spec-requested version (specVer). A spec of "7.2" matches
// an AWS version of "7.2" or "7.2.4". An exact match always returns true.
func versionMatchesSpec(specVer, awsVer string) bool {
	if specVer == awsVer {
		return true
	}
	// AWS may append a patch version suffix. Consider spec satisfied if awsVer
	// starts with specVer followed by "." (e.g., "7.2" matches "7.2.4").
	return len(awsVer) > len(specVer) &&
		awsVer[:len(specVer)] == specVer &&
		awsVer[len(specVer)] == '.'
}
