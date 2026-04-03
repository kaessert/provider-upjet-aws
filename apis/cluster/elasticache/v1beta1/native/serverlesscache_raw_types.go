// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
)

// CacheUsageLimitsRAWParameters defines usage limits for a serverless cache.
type CacheUsageLimitsRAWParameters struct {
	// The maximum data storage limit in the cache, expressed in Gigabytes.
	// +kubebuilder:validation:Optional
	DataStorage []DataStorageRAWParameters `json:"dataStorage,omitempty"`

	// The configuration for the number of ElastiCache Processing Units (ECPU) the cache can consume per second.
	// +kubebuilder:validation:Optional
	EcpuPerSecond []EcpuPerSecondRAWParameters `json:"ecpuPerSecond,omitempty"`
}

// CacheUsageLimitsRAWInitParameters defines init parameters for cache usage limits.
type CacheUsageLimitsRAWInitParameters struct {
	// The maximum data storage limit in the cache, expressed in Gigabytes.
	// +kubebuilder:validation:Optional
	DataStorage []DataStorageRAWInitParameters `json:"dataStorage,omitempty"`

	// The configuration for the number of ECPUs the cache can consume per second.
	// +kubebuilder:validation:Optional
	EcpuPerSecond []EcpuPerSecondRAWInitParameters `json:"ecpuPerSecond,omitempty"`
}

// CacheUsageLimitsRAWObservation defines the observed state of cache usage limits.
type CacheUsageLimitsRAWObservation struct {
	DataStorage   []DataStorageRAWObservation   `json:"dataStorage,omitempty"`
	EcpuPerSecond []EcpuPerSecondRAWObservation `json:"ecpuPerSecond,omitempty"`
}

// DataStorageRAWParameters defines the data storage parameters.
type DataStorageRAWParameters struct {
	// The upper limit for data storage. Must be between 1 and 5000.
	// +kubebuilder:validation:Optional
	Maximum *float64 `json:"maximum,omitempty"`

	// The lower limit for data storage. Must be between 1 and 5000.
	// +kubebuilder:validation:Optional
	Minimum *float64 `json:"minimum,omitempty"`

	// The unit that the storage is measured in, in GB.
	// +kubebuilder:validation:Optional
	Unit *string `json:"unit"`
}

// DataStorageRAWInitParameters defines init parameters for data storage.
type DataStorageRAWInitParameters struct {
	Maximum *float64 `json:"maximum,omitempty"`
	Minimum *float64 `json:"minimum,omitempty"`
	Unit    *string  `json:"unit,omitempty"`
}

// DataStorageRAWObservation defines the observed data storage state.
type DataStorageRAWObservation struct {
	Maximum *float64 `json:"maximum,omitempty"`
	Minimum *float64 `json:"minimum,omitempty"`
	Unit    *string  `json:"unit,omitempty"`
}

// EcpuPerSecondRAWParameters defines the ECPU per second parameters.
type EcpuPerSecondRAWParameters struct {
	// The maximum number of ECPUs the cache can consume per second. Must be between 1000 and 15000000.
	// +kubebuilder:validation:Optional
	Maximum *float64 `json:"maximum,omitempty"`

	// The minimum number of ECPUs the cache can consume per second. Must be between 1000 and 15000000.
	// +kubebuilder:validation:Optional
	Minimum *float64 `json:"minimum,omitempty"`
}

// EcpuPerSecondRAWInitParameters defines init parameters for ECPU per second.
type EcpuPerSecondRAWInitParameters struct {
	Maximum *float64 `json:"maximum,omitempty"`
	Minimum *float64 `json:"minimum,omitempty"`
}

// EcpuPerSecondRAWObservation defines observed ECPU per second state.
type EcpuPerSecondRAWObservation struct {
	Maximum *float64 `json:"maximum,omitempty"`
	Minimum *float64 `json:"minimum,omitempty"`
}

// ServerlessCacheEndpointRAWObservation defines the observed state of a serverless cache endpoint.
type ServerlessCacheEndpointRAWObservation struct {
	// The DNS hostname of the cache node.
	Address *string `json:"address,omitempty"`

	// The port number that the cache engine is listening on.
	Port *float64 `json:"port,omitempty"`
}

// ServerlessCacheRAWParameters defines the configuration parameters for a native ElastiCache Serverless Cache.
type ServerlessCacheRAWParameters struct {
	// Sets the cache usage limits for storage and ECPU for the cache.
	// +kubebuilder:validation:Optional
	CacheUsageLimits []CacheUsageLimitsRAWParameters `json:"cacheUsageLimits,omitempty"`

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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/kms/v1beta1.Key
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/ec2/v1beta1.SecurityGroup
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/ec2/v1beta1.Subnet
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

// ServerlessCacheRAWInitParameters defines init parameters for ServerlessCacheRAW.
type ServerlessCacheRAWInitParameters struct {
	// Sets the cache usage limits.
	// +kubebuilder:validation:Optional
	CacheUsageLimits []CacheUsageLimitsRAWInitParameters `json:"cacheUsageLimits,omitempty"`

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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/kms/v1beta1.Key
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/ec2/v1beta1.SecurityGroup
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/ec2/v1beta1.Subnet
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

// ServerlessCacheRAWObservation defines the observed state of ServerlessCacheRAW.
type ServerlessCacheRAWObservation struct {
	// The ARN of the serverless cache.
	Arn *string `json:"arn,omitempty"`

	// Cache usage limits.
	CacheUsageLimits []CacheUsageLimitsRAWObservation `json:"cacheUsageLimits,omitempty"`

	// When the serverless cache was created.
	CreateTime *string `json:"createTime,omitempty"`

	// The daily time that snapshots will be created.
	DailySnapshotTime *string `json:"dailySnapshotTime,omitempty"`

	// User-provided description.
	Description *string `json:"description,omitempty"`

	// Represents the information required for client programs to connect to the cache.
	Endpoint []ServerlessCacheEndpointRAWObservation `json:"endpoint,omitempty"`

	// Name of the cache engine.
	Engine *string `json:"engine,omitempty"`

	// The full engine version that is utilized to run the serverless cache.
	FullEngineVersion *string `json:"fullEngineVersion,omitempty"`

	// The ID.
	ID *string `json:"id,omitempty"`

	// ARN of the customer managed KMS key.
	KMSKeyID *string `json:"kmsKeyId,omitempty"`

	// The version of the cache engine.
	MajorEngineVersion *string `json:"majorEngineVersion,omitempty"`

	// Represents the information required for client programs to connect to a cache node.
	ReaderEndpoint []ServerlessCacheEndpointRAWObservation `json:"readerEndpoint,omitempty"`

	// A list of VPC security groups.
	// +listType=set
	SecurityGroupIds []*string `json:"securityGroupIds,omitempty"`

	// The number of snapshots that will be retained.
	SnapshotRetentionLimit *float64 `json:"snapshotRetentionLimit,omitempty"`

	// The current status of the serverless cache.
	Status *string `json:"status,omitempty"`

	// A list of VPC subnets.
	// +listType=set
	SubnetIds []*string `json:"subnetIds,omitempty"`

	// Tags assigned to the resource.
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// The identifier of the UserGroup.
	UserGroupID *string `json:"userGroupId,omitempty"`
}

// ServerlessCacheRAWSpec defines the desired state of ServerlessCacheRAW.
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

// ServerlessCacheRAWStatus defines the observed state of ServerlessCacheRAW.
type ServerlessCacheRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider ServerlessCacheRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// ServerlessCacheRAW is the native (non-Terraform) Schema for AWS ElastiCache Serverless Cache.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type ServerlessCacheRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServerlessCacheRAWSpec   `json:"spec"`
	Status ServerlessCacheRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServerlessCacheRAWList contains a list of ServerlessCacheRAW resources.
type ServerlessCacheRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServerlessCacheRAW `json:"items"`
}

// Repository type metadata for ServerlessCacheRAW.
var (
	ServerlessCacheRAW_Kind             = "ServerlessCacheRAW"
	ServerlessCacheRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: ServerlessCacheRAW_Kind}.String()
	ServerlessCacheRAW_KindAPIVersion   = ServerlessCacheRAW_Kind + "." + CRDGroupVersion.String()
	ServerlessCacheRAW_GroupVersionKind = CRDGroupVersion.WithKind(ServerlessCacheRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&ServerlessCacheRAW{}, &ServerlessCacheRAWList{})
}

// GetForProvider returns the ForProvider parameters.
func (s *ServerlessCacheRAW) GetForProvider() *ServerlessCacheRAWParameters { return &s.Spec.ForProvider }

// GetInitProvider returns the InitProvider parameters.
func (s *ServerlessCacheRAW) GetInitProvider() *ServerlessCacheRAWInitParameters {
	return &s.Spec.InitProvider
}

// GetAtProvider returns the current observed state.
func (s *ServerlessCacheRAW) GetAtProvider() ServerlessCacheRAWObservation { return s.Status.AtProvider }

// SetAtProvider sets the observed state.
func (s *ServerlessCacheRAW) SetAtProvider(o ServerlessCacheRAWObservation) {
	s.Status.AtProvider = o
}
