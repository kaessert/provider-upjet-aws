// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta1/native"
)

// ServerlessCacheRAWParameters defines the namespaced configuration parameters for a native ElastiCache Serverless Cache.
// Fields are identical to the cluster-scoped type; defined locally so that angryjet
// generates namespaced resolvers.
type ServerlessCacheRAWParameters struct {
	// Sets the cache usage limits for storage and ECPU for the cache.
	// +kubebuilder:validation:Optional
	CacheUsageLimits []clusternative.CacheUsageLimitsRAWParameters `json:"cacheUsageLimits,omitempty"`

	// The daily time that snapshots will be created.
	// +kubebuilder:validation:Optional
	DailySnapshotTime *string `json:"dailySnapshotTime,omitempty"`

	// User-provided description for the serverless cache.
	// +kubebuilder:validation:Optional
	Description *string `json:"description,omitempty"`

	// Name of the cache engine. Valid values: memcached, redis, valkey.
	// +kubebuilder:validation:Optional
	Engine *string `json:"engine,omitempty"`

	// ARN of the customer managed KMS key for encrypting data at rest.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/kms/v1beta1.Key
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractAtProviderField("arn")
	// +kubebuilder:validation:Optional
	KMSKeyID *string `json:"kmsKeyId,omitempty"`

	// Reference to a Key in kms to populate kmsKeyId.
	// +kubebuilder:validation:Optional
	KMSKeyIDRef *xpv1.NamespacedReference `json:"kmsKeyIdRef,omitempty"`

	// Selector for a Key in kms to populate kmsKeyId.
	// +kubebuilder:validation:Optional
	KMSKeyIDSelector *xpv1.NamespacedSelector `json:"kmsKeyIdSelector,omitempty"`

	// The version of the cache engine.
	// +kubebuilder:validation:Optional
	MajorEngineVersion *string `json:"majorEngineVersion,omitempty"`

	// Region where this resource will be managed.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// References to SecurityGroup in ec2 to populate securityGroupIds.
	// +kubebuilder:validation:Optional
	SecurityGroupIDRefs []xpv1.NamespacedReference `json:"securityGroupIdRefs,omitempty"`

	// Selector for a list of SecurityGroup in ec2 to populate securityGroupIds.
	// +kubebuilder:validation:Optional
	SecurityGroupIDSelector *xpv1.NamespacedSelector `json:"securityGroupIdSelector,omitempty"`

	// A list of VPC security groups to be associated with the serverless cache.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SecurityGroupIDRefs
	// +crossplane:generate:reference:selectorFieldName=SecurityGroupIDSelector
	// +kubebuilder:validation:Optional
	// +listType=set
	SecurityGroupIds []*string `json:"securityGroupIds,omitempty"`

	// The list of ARNs of the snapshots from which the new serverless cache will be created.
	// +kubebuilder:validation:Optional
	SnapshotArnsToRestore []*string `json:"snapshotArnsToRestore,omitempty"`

	// The number of snapshots that will be retained for the serverless cache.
	// +kubebuilder:validation:Optional
	SnapshotRetentionLimit *float64 `json:"snapshotRetentionLimit,omitempty"`

	// References to Subnet in ec2 to populate subnetIds.
	// +kubebuilder:validation:Optional
	SubnetIDRefs []xpv1.NamespacedReference `json:"subnetIdRefs,omitempty"`

	// Selector for a list of Subnet in ec2 to populate subnetIds.
	// +kubebuilder:validation:Optional
	SubnetIDSelector *xpv1.NamespacedSelector `json:"subnetIdSelector,omitempty"`

	// A list of the one or more VPC subnets to be associated with the serverless cache.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/ec2/v1beta1.Subnet
	// +crossplane:generate:reference:refFieldName=SubnetIDRefs
	// +crossplane:generate:reference:selectorFieldName=SubnetIDSelector
	// +kubebuilder:validation:Optional
	// +listType=set
	SubnetIds []*string `json:"subnetIds,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// The identifier of the UserGroup to be associated with the serverless cache.
	// +kubebuilder:validation:Optional
	UserGroupID *string `json:"userGroupId,omitempty"`
}

// ServerlessCacheRAWInitParameters defines the namespaced init parameters for ServerlessCacheRAW.
type ServerlessCacheRAWInitParameters struct {
	// Sets the cache usage limits.
	// +kubebuilder:validation:Optional
	CacheUsageLimits []clusternative.CacheUsageLimitsRAWInitParameters `json:"cacheUsageLimits,omitempty"`

	// The daily time that snapshots will be created.
	// +kubebuilder:validation:Optional
	DailySnapshotTime *string `json:"dailySnapshotTime,omitempty"`

	// User-provided description.
	// +kubebuilder:validation:Optional
	Description *string `json:"description,omitempty"`

	// Name of the cache engine.
	// +kubebuilder:validation:Optional
	Engine *string `json:"engine,omitempty"`

	// ARN of the customer managed KMS key.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/kms/v1beta1.Key
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractAtProviderField("arn")
	// +kubebuilder:validation:Optional
	KMSKeyID *string `json:"kmsKeyId,omitempty"`

	// Reference to a Key in kms.
	// +kubebuilder:validation:Optional
	KMSKeyIDRef *xpv1.NamespacedReference `json:"kmsKeyIdRef,omitempty"`

	// Selector for a Key in kms.
	// +kubebuilder:validation:Optional
	KMSKeyIDSelector *xpv1.NamespacedSelector `json:"kmsKeyIdSelector,omitempty"`

	// The version of the cache engine.
	// +kubebuilder:validation:Optional
	MajorEngineVersion *string `json:"majorEngineVersion,omitempty"`

	// References to SecurityGroup in ec2.
	// +kubebuilder:validation:Optional
	SecurityGroupIDRefs []xpv1.NamespacedReference `json:"securityGroupIdRefs,omitempty"`

	// Selector for a list of SecurityGroup in ec2.
	// +kubebuilder:validation:Optional
	SecurityGroupIDSelector *xpv1.NamespacedSelector `json:"securityGroupIdSelector,omitempty"`

	// A list of VPC security groups.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SecurityGroupIDRefs
	// +crossplane:generate:reference:selectorFieldName=SecurityGroupIDSelector
	// +kubebuilder:validation:Optional
	// +listType=set
	SecurityGroupIds []*string `json:"securityGroupIds,omitempty"`

	// The list of ARNs of the snapshots.
	// +kubebuilder:validation:Optional
	SnapshotArnsToRestore []*string `json:"snapshotArnsToRestore,omitempty"`

	// The number of snapshots that will be retained.
	// +kubebuilder:validation:Optional
	SnapshotRetentionLimit *float64 `json:"snapshotRetentionLimit,omitempty"`

	// References to Subnet in ec2.
	// +kubebuilder:validation:Optional
	SubnetIDRefs []xpv1.NamespacedReference `json:"subnetIdRefs,omitempty"`

	// Selector for a list of Subnet in ec2.
	// +kubebuilder:validation:Optional
	SubnetIDSelector *xpv1.NamespacedSelector `json:"subnetIdSelector,omitempty"`

	// A list of VPC subnets.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/ec2/v1beta1.Subnet
	// +crossplane:generate:reference:refFieldName=SubnetIDRefs
	// +crossplane:generate:reference:selectorFieldName=SubnetIDSelector
	// +kubebuilder:validation:Optional
	// +listType=set
	SubnetIds []*string `json:"subnetIds,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// The identifier of the UserGroup.
	// +kubebuilder:validation:Optional
	UserGroupID *string `json:"userGroupId,omitempty"`
}

// ServerlessCacheRAWSpec defines the desired state of namespaced ServerlessCacheRAW.
type ServerlessCacheRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider ServerlessCacheRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider ServerlessCacheRAWInitParameters `json:"initProvider,omitempty"`
}

// ServerlessCacheRAWStatus defines the observed state of namespaced ServerlessCacheRAW.
type ServerlessCacheRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider clusternative.ServerlessCacheRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// ServerlessCacheRAW is the native (non-Terraform) Schema for AWS ElastiCache Serverless Cache (namespaced scope).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}
type ServerlessCacheRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServerlessCacheRAWSpec   `json:"spec"`
	Status ServerlessCacheRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServerlessCacheRAWList contains a list of ServerlessCacheRAW resources (namespaced scope).
type ServerlessCacheRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServerlessCacheRAW `json:"items"`
}

// Repository type metadata for namespaced ServerlessCacheRAW.
var (
	ServerlessCacheRAW_Kind             = "ServerlessCacheRAW"
	ServerlessCacheRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: ServerlessCacheRAW_Kind}.String()
	ServerlessCacheRAW_KindAPIVersion   = ServerlessCacheRAW_Kind + "." + CRDGroupVersion.String()
	ServerlessCacheRAW_GroupVersionKind = CRDGroupVersion.WithKind(ServerlessCacheRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&ServerlessCacheRAW{}, &ServerlessCacheRAWList{})
}

// GetForProvider returns a cluster-scoped ServerlessCacheRAWParameters populated from
// this namespaced resource's ForProvider fields.
func (s *ServerlessCacheRAW) GetForProvider() *clusternative.ServerlessCacheRAWParameters {
	return &clusternative.ServerlessCacheRAWParameters{
		CacheUsageLimits:       s.Spec.ForProvider.CacheUsageLimits,
		DailySnapshotTime:      s.Spec.ForProvider.DailySnapshotTime,
		Description:            s.Spec.ForProvider.Description,
		Engine:                 s.Spec.ForProvider.Engine,
		KMSKeyID:               s.Spec.ForProvider.KMSKeyID,
		MajorEngineVersion:     s.Spec.ForProvider.MajorEngineVersion,
		Region:                 s.Spec.ForProvider.Region,
		SecurityGroupIds:       s.Spec.ForProvider.SecurityGroupIds,
		SnapshotArnsToRestore:  s.Spec.ForProvider.SnapshotArnsToRestore,
		SnapshotRetentionLimit: s.Spec.ForProvider.SnapshotRetentionLimit,
		SubnetIds:              s.Spec.ForProvider.SubnetIds,
		Tags:                   s.Spec.ForProvider.Tags,
		UserGroupID:            s.Spec.ForProvider.UserGroupID,
	}
}

// SetForProvider writes back a cluster-scoped ServerlessCacheRAWParameters to this
// namespaced resource's ForProvider fields. Symmetric inverse of GetForProvider,
// required so that late-initialized fields (MajorEngineVersion, DailySnapshotTime)
// are persisted.
func (s *ServerlessCacheRAW) SetForProvider(p clusternative.ServerlessCacheRAWParameters) {
	s.Spec.ForProvider.CacheUsageLimits = p.CacheUsageLimits
	s.Spec.ForProvider.DailySnapshotTime = p.DailySnapshotTime
	s.Spec.ForProvider.Description = p.Description
	s.Spec.ForProvider.Engine = p.Engine
	s.Spec.ForProvider.KMSKeyID = p.KMSKeyID
	s.Spec.ForProvider.MajorEngineVersion = p.MajorEngineVersion
	s.Spec.ForProvider.Region = p.Region
	s.Spec.ForProvider.SecurityGroupIds = p.SecurityGroupIds
	s.Spec.ForProvider.SnapshotArnsToRestore = p.SnapshotArnsToRestore
	s.Spec.ForProvider.SnapshotRetentionLimit = p.SnapshotRetentionLimit
	s.Spec.ForProvider.SubnetIds = p.SubnetIds
	s.Spec.ForProvider.Tags = p.Tags
	s.Spec.ForProvider.UserGroupID = p.UserGroupID
}

// GetInitProvider returns a cluster-scoped ServerlessCacheRAWInitParameters populated from
// this namespaced resource's InitProvider fields.
func (s *ServerlessCacheRAW) GetInitProvider() *clusternative.ServerlessCacheRAWInitParameters {
	return &clusternative.ServerlessCacheRAWInitParameters{
		CacheUsageLimits:       s.Spec.InitProvider.CacheUsageLimits,
		DailySnapshotTime:      s.Spec.InitProvider.DailySnapshotTime,
		Description:            s.Spec.InitProvider.Description,
		Engine:                 s.Spec.InitProvider.Engine,
		KMSKeyID:               s.Spec.InitProvider.KMSKeyID,
		MajorEngineVersion:     s.Spec.InitProvider.MajorEngineVersion,
		SecurityGroupIds:       s.Spec.InitProvider.SecurityGroupIds,
		SnapshotArnsToRestore:  s.Spec.InitProvider.SnapshotArnsToRestore,
		SnapshotRetentionLimit: s.Spec.InitProvider.SnapshotRetentionLimit,
		SubnetIds:              s.Spec.InitProvider.SubnetIds,
		Tags:                   s.Spec.InitProvider.Tags,
		UserGroupID:            s.Spec.InitProvider.UserGroupID,
	}
}

// GetAtProvider returns the current observed state.
func (s *ServerlessCacheRAW) GetAtProvider() clusternative.ServerlessCacheRAWObservation {
	return s.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (s *ServerlessCacheRAW) SetAtProvider(o clusternative.ServerlessCacheRAWObservation) {
	s.Status.AtProvider = o
}
