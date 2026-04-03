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

// GlobalReplicationGroupRAWParameters defines the namespaced configuration parameters for a native ElastiCache Global Replication Group.
// Fields are identical to the cluster-scoped type; defined locally so that angryjet
// generates namespaced resolvers.
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/elasticache/v1beta2/native.ReplicationGroupRAW
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

// GlobalReplicationGroupRAWInitParameters defines the namespaced init parameters for GlobalReplicationGroupRAW.
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/elasticache/v1beta2/native.ReplicationGroupRAW
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

// GlobalReplicationGroupRAWSpec defines the desired state of namespaced GlobalReplicationGroupRAW.
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

// GlobalReplicationGroupRAWStatus defines the observed state of namespaced GlobalReplicationGroupRAW.
type GlobalReplicationGroupRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider clusternative.GlobalReplicationGroupRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// GlobalReplicationGroupRAW is the native (non-Terraform) Schema for AWS ElastiCache Global Replication Group (namespaced scope).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}
type GlobalReplicationGroupRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GlobalReplicationGroupRAWSpec   `json:"spec"`
	Status GlobalReplicationGroupRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GlobalReplicationGroupRAWList contains a list of GlobalReplicationGroupRAW resources (namespaced scope).
type GlobalReplicationGroupRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GlobalReplicationGroupRAW `json:"items"`
}

// Repository type metadata for namespaced GlobalReplicationGroupRAW.
var (
	GlobalReplicationGroupRAW_Kind             = "GlobalReplicationGroupRAW"
	GlobalReplicationGroupRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: GlobalReplicationGroupRAW_Kind}.String()
	GlobalReplicationGroupRAW_KindAPIVersion   = GlobalReplicationGroupRAW_Kind + "." + CRDGroupVersion.String()
	GlobalReplicationGroupRAW_GroupVersionKind = CRDGroupVersion.WithKind(GlobalReplicationGroupRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&GlobalReplicationGroupRAW{}, &GlobalReplicationGroupRAWList{})
}

// GetForProvider returns a cluster-scoped GlobalReplicationGroupRAWParameters populated from
// this namespaced resource's ForProvider fields.
func (g *GlobalReplicationGroupRAW) GetForProvider() *clusternative.GlobalReplicationGroupRAWParameters {
	return &clusternative.GlobalReplicationGroupRAWParameters{
		AutomaticFailoverEnabled:          g.Spec.ForProvider.AutomaticFailoverEnabled,
		CacheNodeType:                     g.Spec.ForProvider.CacheNodeType,
		Engine:                            g.Spec.ForProvider.Engine,
		EngineVersion:                     g.Spec.ForProvider.EngineVersion,
		GlobalReplicationGroupDescription: g.Spec.ForProvider.GlobalReplicationGroupDescription,
		GlobalReplicationGroupIDSuffix:    g.Spec.ForProvider.GlobalReplicationGroupIDSuffix,
		NumNodeGroups:                     g.Spec.ForProvider.NumNodeGroups,
		ParameterGroupName:                g.Spec.ForProvider.ParameterGroupName,
		PrimaryReplicationGroupID:         g.Spec.ForProvider.PrimaryReplicationGroupID,
		Region:                            g.Spec.ForProvider.Region,
	}
}

// GetInitProvider returns a cluster-scoped GlobalReplicationGroupRAWInitParameters populated from
// this namespaced resource's InitProvider fields.
func (g *GlobalReplicationGroupRAW) GetInitProvider() *clusternative.GlobalReplicationGroupRAWInitParameters {
	return &clusternative.GlobalReplicationGroupRAWInitParameters{
		AutomaticFailoverEnabled:          g.Spec.InitProvider.AutomaticFailoverEnabled,
		CacheNodeType:                     g.Spec.InitProvider.CacheNodeType,
		Engine:                            g.Spec.InitProvider.Engine,
		EngineVersion:                     g.Spec.InitProvider.EngineVersion,
		GlobalReplicationGroupDescription: g.Spec.InitProvider.GlobalReplicationGroupDescription,
		GlobalReplicationGroupIDSuffix:    g.Spec.InitProvider.GlobalReplicationGroupIDSuffix,
		NumNodeGroups:                     g.Spec.InitProvider.NumNodeGroups,
		ParameterGroupName:                g.Spec.InitProvider.ParameterGroupName,
		PrimaryReplicationGroupID:         g.Spec.InitProvider.PrimaryReplicationGroupID,
	}
}

// GetAtProvider returns the current observed state.
func (g *GlobalReplicationGroupRAW) GetAtProvider() clusternative.GlobalReplicationGroupRAWObservation {
	return g.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (g *GlobalReplicationGroupRAW) SetAtProvider(o clusternative.GlobalReplicationGroupRAWObservation) {
	g.Status.AtProvider = o
}
