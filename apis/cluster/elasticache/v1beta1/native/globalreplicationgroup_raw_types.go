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

// GlobalNodeGroupRAWObservation defines the observed state of a global node group.
type GlobalNodeGroupRAWObservation struct {
	// The ID of the global node group.
	GlobalNodeGroupID *string `json:"globalNodeGroupId,omitempty"`

	// The keyspace for this node group.
	Slots *string `json:"slots,omitempty"`
}

// GlobalReplicationGroupRAWParameters defines the configuration parameters for a native ElastiCache Global Replication Group.
type GlobalReplicationGroupRAWParameters struct {
	// Specifies whether read-only replicas will be automatically promoted.
	// +kubebuilder:validation:Optional
	AutomaticFailoverEnabled *bool `json:"automaticFailoverEnabled,omitempty"`

	// The instance class used.
	// +kubebuilder:validation:Optional
	CacheNodeType *string `json:"cacheNodeType,omitempty"`

	// The name of the cache engine.
	// +kubebuilder:validation:Optional
	Engine *string `json:"engine,omitempty"`

	// Engine version to use for the Global Replication Group.
	// +kubebuilder:validation:Optional
	EngineVersion *string `json:"engineVersion,omitempty"`

	// A user-created description for the global replication group.
	// +kubebuilder:validation:Optional
	GlobalReplicationGroupDescription *string `json:"globalReplicationGroupDescription,omitempty"`

	// The suffix name of a Global Datastore.
	// +kubebuilder:validation:Optional
	GlobalReplicationGroupIDSuffix *string `json:"globalReplicationGroupIdSuffix,omitempty"`

	// The number of node groups (shards) on the global replication group.
	// +kubebuilder:validation:Optional
	NumNodeGroups *float64 `json:"numNodeGroups,omitempty"`

	// An ElastiCache Parameter Group to use for the Global Replication Group.
	// +kubebuilder:validation:Optional
	ParameterGroupName *string `json:"parameterGroupName,omitempty"`

	// The ID of the primary cluster that accepts writes.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta2/native.ReplicationGroupRAW
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractAtProviderField("id")
	// +kubebuilder:validation:Optional
	PrimaryReplicationGroupID *string `json:"primaryReplicationGroupId,omitempty"`

	// Reference to a ReplicationGroupRAW in elasticache to populate primaryReplicationGroupId.
	// +kubebuilder:validation:Optional
	PrimaryReplicationGroupIDRef *xpv1.NamespacedReference `json:"primaryReplicationGroupIdRef,omitempty"`

	// Selector for a ReplicationGroupRAW in elasticache to populate primaryReplicationGroupId.
	// +kubebuilder:validation:Optional
	PrimaryReplicationGroupIDSelector *xpv1.NamespacedSelector `json:"primaryReplicationGroupIdSelector,omitempty"`

	// Region where this resource will be managed.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`
}

// GlobalReplicationGroupRAWInitParameters defines the init parameters for GlobalReplicationGroupRAW.
type GlobalReplicationGroupRAWInitParameters struct {
	// Specifies whether read-only replicas will be automatically promoted.
	// +kubebuilder:validation:Optional
	AutomaticFailoverEnabled *bool `json:"automaticFailoverEnabled,omitempty"`

	// The instance class used.
	// +kubebuilder:validation:Optional
	CacheNodeType *string `json:"cacheNodeType,omitempty"`

	// The name of the cache engine.
	// +kubebuilder:validation:Optional
	Engine *string `json:"engine,omitempty"`

	// Engine version to use.
	// +kubebuilder:validation:Optional
	EngineVersion *string `json:"engineVersion,omitempty"`

	// A user-created description.
	// +kubebuilder:validation:Optional
	GlobalReplicationGroupDescription *string `json:"globalReplicationGroupDescription,omitempty"`

	// The suffix name of a Global Datastore.
	// +kubebuilder:validation:Optional
	GlobalReplicationGroupIDSuffix *string `json:"globalReplicationGroupIdSuffix,omitempty"`

	// The number of node groups (shards).
	// +kubebuilder:validation:Optional
	NumNodeGroups *float64 `json:"numNodeGroups,omitempty"`

	// An ElastiCache Parameter Group.
	// +kubebuilder:validation:Optional
	ParameterGroupName *string `json:"parameterGroupName,omitempty"`

	// The ID of the primary cluster.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta2/native.ReplicationGroupRAW
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractAtProviderField("id")
	// +kubebuilder:validation:Optional
	PrimaryReplicationGroupID *string `json:"primaryReplicationGroupId,omitempty"`

	// Reference to a ReplicationGroupRAW.
	// +kubebuilder:validation:Optional
	PrimaryReplicationGroupIDRef *xpv1.NamespacedReference `json:"primaryReplicationGroupIdRef,omitempty"`

	// Selector for a ReplicationGroupRAW.
	// +kubebuilder:validation:Optional
	PrimaryReplicationGroupIDSelector *xpv1.NamespacedSelector `json:"primaryReplicationGroupIdSelector,omitempty"`
}

// GlobalReplicationGroupRAWObservation defines the observed state of GlobalReplicationGroupRAW.
type GlobalReplicationGroupRAWObservation struct {
	// The ARN of the ElastiCache Global Replication Group.
	Arn *string `json:"arn,omitempty"`

	// A flag that indicates whether encryption at rest is enabled.
	AtRestEncryptionEnabled *bool `json:"atRestEncryptionEnabled,omitempty"`

	// A flag that indicates whether AuthToken (password) is enabled.
	AuthTokenEnabled *bool `json:"authTokenEnabled,omitempty"`

	// Specifies whether read-only replicas will be automatically promoted.
	AutomaticFailoverEnabled *bool `json:"automaticFailoverEnabled,omitempty"`

	// The instance class used.
	CacheNodeType *string `json:"cacheNodeType,omitempty"`

	// Indicates whether the Global Datastore is cluster enabled.
	ClusterEnabled *bool `json:"clusterEnabled,omitempty"`

	// The name of the cache engine.
	Engine *string `json:"engine,omitempty"`

	// Engine version used.
	EngineVersion *string `json:"engineVersion,omitempty"`

	// The full version number.
	EngineVersionActual *string `json:"engineVersionActual,omitempty"`

	// Set of node groups on the global replication group.
	GlobalNodeGroups []GlobalNodeGroupRAWObservation `json:"globalNodeGroups,omitempty"`

	// A user-created description.
	GlobalReplicationGroupDescription *string `json:"globalReplicationGroupDescription,omitempty"`

	// The full ID of the global replication group.
	GlobalReplicationGroupID *string `json:"globalReplicationGroupId,omitempty"`

	// The suffix name.
	GlobalReplicationGroupIDSuffix *string `json:"globalReplicationGroupIdSuffix,omitempty"`

	// The ID.
	ID *string `json:"id,omitempty"`

	// The number of node groups.
	NumNodeGroups *float64 `json:"numNodeGroups,omitempty"`

	// An ElastiCache Parameter Group.
	ParameterGroupName *string `json:"parameterGroupName,omitempty"`

	// The ID of the primary cluster.
	PrimaryReplicationGroupID *string `json:"primaryReplicationGroupId,omitempty"`

	// A flag that indicates whether encryption in transit is enabled.
	TransitEncryptionEnabled *bool `json:"transitEncryptionEnabled,omitempty"`
}

// GlobalReplicationGroupRAWSpec defines the desired state of GlobalReplicationGroupRAW.
type GlobalReplicationGroupRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider GlobalReplicationGroupRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider GlobalReplicationGroupRAWInitParameters `json:"initProvider,omitempty"`
}

// GlobalReplicationGroupRAWStatus defines the observed state of GlobalReplicationGroupRAW.
type GlobalReplicationGroupRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider GlobalReplicationGroupRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// GlobalReplicationGroupRAW is the native (non-Terraform) Schema for AWS ElastiCache Global Replication Group.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type GlobalReplicationGroupRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GlobalReplicationGroupRAWSpec   `json:"spec"`
	Status GlobalReplicationGroupRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GlobalReplicationGroupRAWList contains a list of GlobalReplicationGroupRAW resources.
type GlobalReplicationGroupRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalReplicationGroupRAW `json:"items"`
}

// Repository type metadata for GlobalReplicationGroupRAW.
var (
	GlobalReplicationGroupRAW_Kind             = "GlobalReplicationGroupRAW"
	GlobalReplicationGroupRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: GlobalReplicationGroupRAW_Kind}.String()
	GlobalReplicationGroupRAW_KindAPIVersion   = GlobalReplicationGroupRAW_Kind + "." + CRDGroupVersion.String()
	GlobalReplicationGroupRAW_GroupVersionKind = CRDGroupVersion.WithKind(GlobalReplicationGroupRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&GlobalReplicationGroupRAW{}, &GlobalReplicationGroupRAWList{})
}

// GetForProvider returns the ForProvider parameters.
func (g *GlobalReplicationGroupRAW) GetForProvider() *GlobalReplicationGroupRAWParameters {
	return &g.Spec.ForProvider
}

// SetForProvider writes back the ForProvider parameters.
func (g *GlobalReplicationGroupRAW) SetForProvider(p GlobalReplicationGroupRAWParameters) {
	g.Spec.ForProvider = p
}

// GetInitProvider returns the InitProvider parameters.
func (g *GlobalReplicationGroupRAW) GetInitProvider() *GlobalReplicationGroupRAWInitParameters {
	return &g.Spec.InitProvider
}

// GetAtProvider returns the current observed state.
func (g *GlobalReplicationGroupRAW) GetAtProvider() GlobalReplicationGroupRAWObservation {
	return g.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (g *GlobalReplicationGroupRAW) SetAtProvider(o GlobalReplicationGroupRAWObservation) {
	g.Status.AtProvider = o
}
