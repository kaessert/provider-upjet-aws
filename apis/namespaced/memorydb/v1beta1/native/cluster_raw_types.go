// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/memorydb/v1beta1/native"
)

// ClusterRAWParameters defines the namespaced configuration parameters for a native MemoryDB Cluster.
// Fields are identical to the cluster-scoped type; defined locally so that angryjet
// generates namespaced resolvers.
type ClusterRAWParameters struct {
	// The name of the Access Control List to associate with the cluster.
	// +crossplane:generate:reference:type=ACLRAW
	// +kubebuilder:validation:Optional
	ACLName *string `json:"aclName,omitempty"`

	// Reference to a ACLRAW in memorydb to populate aclName.
	// +kubebuilder:validation:Optional
	ACLNameRef *xpv1.NamespacedReference `json:"aclNameRef,omitempty"`

	// Selector for a ACLRAW in memorydb to populate aclName.
	// +kubebuilder:validation:Optional
	ACLNameSelector *xpv1.NamespacedSelector `json:"aclNameSelector,omitempty"`

	// When set to true, the cluster will automatically receive minor engine
	// version upgrades after launch. Defaults to true.
	// +kubebuilder:validation:Optional
	AutoMinorVersionUpgrade *bool `json:"autoMinorVersionUpgrade,omitempty"`

	// Enables data tiering. This option is not supported by all instance types.
	// +kubebuilder:validation:Optional
	DataTiering *bool `json:"dataTiering,omitempty"`

	// Description for the cluster.
	// +kubebuilder:validation:Optional
	Description *string `json:"description,omitempty"`

	// The engine that will run on your nodes. Supported values are redis and valkey.
	// +kubebuilder:validation:Optional
	Engine *string `json:"engine,omitempty"`

	// Version number of the engine to be used for the cluster.
	// +kubebuilder:validation:Optional
	EngineVersion *string `json:"engineVersion,omitempty"`

	// Name of the final cluster snapshot to be created when this resource is deleted.
	// +kubebuilder:validation:Optional
	FinalSnapshotName *string `json:"finalSnapshotName,omitempty"`

	// Mechanism that the cluster uses to discover IP addresses.
	// +kubebuilder:validation:Optional
	IPDiscovery *string `json:"ipDiscovery,omitempty"`

	// ARN of the KMS key used to encrypt the cluster at rest.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/kms/v1beta1.Key
	// +kubebuilder:validation:Optional
	KMSKeyArn *string `json:"kmsKeyArn,omitempty"`

	// Reference to a Key in kms to populate kmsKeyArn.
	// +kubebuilder:validation:Optional
	KMSKeyArnRef *xpv1.NamespacedReference `json:"kmsKeyArnRef,omitempty"`

	// Selector for a Key in kms to populate kmsKeyArn.
	// +kubebuilder:validation:Optional
	KMSKeyArnSelector *xpv1.NamespacedSelector `json:"kmsKeyArnSelector,omitempty"`

	// Specifies the weekly time range during which maintenance on the cluster is performed.
	// +kubebuilder:validation:Optional
	MaintenanceWindow *string `json:"maintenanceWindow,omitempty"`

	// The multi region cluster identifier.
	// +kubebuilder:validation:Optional
	MultiRegionClusterName *string `json:"multiRegionClusterName,omitempty"`

	// IP address type for the cluster.
	// +kubebuilder:validation:Optional
	NetworkType *string `json:"networkType,omitempty"`

	// The compute and memory capacity of the nodes in the cluster.
	// +kubebuilder:validation:Optional
	NodeType *string `json:"nodeType,omitempty"`

	// The number of replicas to apply to each shard.
	// +kubebuilder:validation:Optional
	NumReplicasPerShard *float64 `json:"numReplicasPerShard,omitempty"`

	// The number of shards in the cluster.
	// +kubebuilder:validation:Optional
	NumShards *float64 `json:"numShards,omitempty"`

	// The name of the parameter group associated with the cluster.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/memorydb/v1beta1.ParameterGroup
	// +kubebuilder:validation:Optional
	ParameterGroupName *string `json:"parameterGroupName,omitempty"`

	// Reference to a ParameterGroup in memorydb to populate parameterGroupName.
	// +kubebuilder:validation:Optional
	ParameterGroupNameRef *xpv1.NamespacedReference `json:"parameterGroupNameRef,omitempty"`

	// Selector for a ParameterGroup in memorydb to populate parameterGroupName.
	// +kubebuilder:validation:Optional
	ParameterGroupNameSelector *xpv1.NamespacedSelector `json:"parameterGroupNameSelector,omitempty"`

	// The port number on which each of the nodes accepts connections.
	// +kubebuilder:validation:Optional
	Port *float64 `json:"port,omitempty"`

	// Region where this resource will be managed.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// References to SecurityGroup in ec2 to populate securityGroupIds.
	// +kubebuilder:validation:Optional
	SecurityGroupIDRefs []xpv1.NamespacedReference `json:"securityGroupIdRefs,omitempty"`

	// Selector for a list of SecurityGroup in ec2 to populate securityGroupIds.
	// +kubebuilder:validation:Optional
	SecurityGroupIDSelector *xpv1.NamespacedSelector `json:"securityGroupIdSelector,omitempty"`

	// Set of VPC Security Group ID-s to associate with this cluster.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SecurityGroupIDRefs
	// +crossplane:generate:reference:selectorFieldName=SecurityGroupIDSelector
	// +kubebuilder:validation:Optional
	// +listType=set
	SecurityGroupIds []*string `json:"securityGroupIds,omitempty"`

	// List of ARN-s that uniquely identify RDB snapshot files stored in S3.
	// +kubebuilder:validation:Optional
	SnapshotArns []*string `json:"snapshotArns,omitempty"`

	// The name of a snapshot from which to restore data into the new cluster.
	// +kubebuilder:validation:Optional
	SnapshotName *string `json:"snapshotName,omitempty"`

	// The number of days for which MemoryDB retains automatic snapshots.
	// +kubebuilder:validation:Optional
	SnapshotRetentionLimit *float64 `json:"snapshotRetentionLimit,omitempty"`

	// The daily time range during which MemoryDB begins taking a daily snapshot.
	// +kubebuilder:validation:Optional
	SnapshotWindow *string `json:"snapshotWindow,omitempty"`

	// ARN of the SNS topic to which cluster notifications are sent.
	// +kubebuilder:validation:Optional
	SnsTopicArn *string `json:"snsTopicArn,omitempty"`

	// The name of the subnet group to be used for the cluster.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/memorydb/v1beta1.SubnetGroup
	// +kubebuilder:validation:Optional
	SubnetGroupName *string `json:"subnetGroupName,omitempty"`

	// Reference to a SubnetGroup in memorydb to populate subnetGroupName.
	// +kubebuilder:validation:Optional
	SubnetGroupNameRef *xpv1.NamespacedReference `json:"subnetGroupNameRef,omitempty"`

	// Selector for a SubnetGroup in memorydb to populate subnetGroupName.
	// +kubebuilder:validation:Optional
	SubnetGroupNameSelector *xpv1.NamespacedSelector `json:"subnetGroupNameSelector,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// A flag to enable in-transit encryption on the cluster.
	// +kubebuilder:validation:Optional
	TLSEnabled *bool `json:"tlsEnabled,omitempty"`
}

// ClusterRAWInitParameters defines the namespaced init parameters for ClusterRAW.
type ClusterRAWInitParameters struct {
	// The name of the Access Control List to associate with the cluster.
	// +crossplane:generate:reference:type=ACLRAW
	// +kubebuilder:validation:Optional
	ACLName *string `json:"aclName,omitempty"`

	// Reference to a ACLRAW in memorydb to populate aclName.
	// +kubebuilder:validation:Optional
	ACLNameRef *xpv1.NamespacedReference `json:"aclNameRef,omitempty"`

	// Selector for a ACLRAW in memorydb to populate aclName.
	// +kubebuilder:validation:Optional
	ACLNameSelector *xpv1.NamespacedSelector `json:"aclNameSelector,omitempty"`

	// When set to true, the cluster will automatically receive minor engine
	// version upgrades after launch. Defaults to true.
	// +kubebuilder:validation:Optional
	AutoMinorVersionUpgrade *bool `json:"autoMinorVersionUpgrade,omitempty"`

	// Enables data tiering.
	// +kubebuilder:validation:Optional
	DataTiering *bool `json:"dataTiering,omitempty"`

	// Description for the cluster.
	// +kubebuilder:validation:Optional
	Description *string `json:"description,omitempty"`

	// The engine that will run on your nodes.
	// +kubebuilder:validation:Optional
	Engine *string `json:"engine,omitempty"`

	// Version number of the engine.
	// +kubebuilder:validation:Optional
	EngineVersion *string `json:"engineVersion,omitempty"`

	// Name of the final cluster snapshot.
	// +kubebuilder:validation:Optional
	FinalSnapshotName *string `json:"finalSnapshotName,omitempty"`

	// Mechanism that the cluster uses to discover IP addresses.
	// +kubebuilder:validation:Optional
	IPDiscovery *string `json:"ipDiscovery,omitempty"`

	// ARN of the KMS key used to encrypt the cluster at rest.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/kms/v1beta1.Key
	// +kubebuilder:validation:Optional
	KMSKeyArn *string `json:"kmsKeyArn,omitempty"`

	// Reference to a Key in kms to populate kmsKeyArn.
	// +kubebuilder:validation:Optional
	KMSKeyArnRef *xpv1.NamespacedReference `json:"kmsKeyArnRef,omitempty"`

	// Selector for a Key in kms to populate kmsKeyArn.
	// +kubebuilder:validation:Optional
	KMSKeyArnSelector *xpv1.NamespacedSelector `json:"kmsKeyArnSelector,omitempty"`

	// Specifies the weekly time range for maintenance.
	// +kubebuilder:validation:Optional
	MaintenanceWindow *string `json:"maintenanceWindow,omitempty"`

	// The multi region cluster identifier.
	// +kubebuilder:validation:Optional
	MultiRegionClusterName *string `json:"multiRegionClusterName,omitempty"`

	// IP address type for the cluster.
	// +kubebuilder:validation:Optional
	NetworkType *string `json:"networkType,omitempty"`

	// The compute and memory capacity of the nodes in the cluster.
	// +kubebuilder:validation:Optional
	NodeType *string `json:"nodeType,omitempty"`

	// The number of replicas to apply to each shard.
	// +kubebuilder:validation:Optional
	NumReplicasPerShard *float64 `json:"numReplicasPerShard,omitempty"`

	// The number of shards in the cluster.
	// +kubebuilder:validation:Optional
	NumShards *float64 `json:"numShards,omitempty"`

	// The name of the parameter group associated with the cluster.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/memorydb/v1beta1.ParameterGroup
	// +kubebuilder:validation:Optional
	ParameterGroupName *string `json:"parameterGroupName,omitempty"`

	// Reference to a ParameterGroup in memorydb to populate parameterGroupName.
	// +kubebuilder:validation:Optional
	ParameterGroupNameRef *xpv1.NamespacedReference `json:"parameterGroupNameRef,omitempty"`

	// Selector for a ParameterGroup in memorydb to populate parameterGroupName.
	// +kubebuilder:validation:Optional
	ParameterGroupNameSelector *xpv1.NamespacedSelector `json:"parameterGroupNameSelector,omitempty"`

	// The port number on which each of the nodes accepts connections.
	// +kubebuilder:validation:Optional
	Port *float64 `json:"port,omitempty"`

	// References to SecurityGroup in ec2 to populate securityGroupIds.
	// +kubebuilder:validation:Optional
	SecurityGroupIDRefs []xpv1.NamespacedReference `json:"securityGroupIdRefs,omitempty"`

	// Selector for a list of SecurityGroup in ec2 to populate securityGroupIds.
	// +kubebuilder:validation:Optional
	SecurityGroupIDSelector *xpv1.NamespacedSelector `json:"securityGroupIdSelector,omitempty"`

	// Set of VPC Security Group ID-s to associate with this cluster.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SecurityGroupIDRefs
	// +crossplane:generate:reference:selectorFieldName=SecurityGroupIDSelector
	// +kubebuilder:validation:Optional
	// +listType=set
	SecurityGroupIds []*string `json:"securityGroupIds,omitempty"`

	// List of ARN-s that uniquely identify RDB snapshot files stored in S3.
	// +kubebuilder:validation:Optional
	SnapshotArns []*string `json:"snapshotArns,omitempty"`

	// The name of a snapshot from which to restore data.
	// +kubebuilder:validation:Optional
	SnapshotName *string `json:"snapshotName,omitempty"`

	// The number of days for which MemoryDB retains automatic snapshots.
	// +kubebuilder:validation:Optional
	SnapshotRetentionLimit *float64 `json:"snapshotRetentionLimit,omitempty"`

	// The daily time range for snapshots.
	// +kubebuilder:validation:Optional
	SnapshotWindow *string `json:"snapshotWindow,omitempty"`

	// ARN of the SNS topic to which cluster notifications are sent.
	// +kubebuilder:validation:Optional
	SnsTopicArn *string `json:"snsTopicArn,omitempty"`

	// The name of the subnet group to be used for the cluster.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/memorydb/v1beta1.SubnetGroup
	// +kubebuilder:validation:Optional
	SubnetGroupName *string `json:"subnetGroupName,omitempty"`

	// Reference to a SubnetGroup in memorydb to populate subnetGroupName.
	// +kubebuilder:validation:Optional
	SubnetGroupNameRef *xpv1.NamespacedReference `json:"subnetGroupNameRef,omitempty"`

	// Selector for a SubnetGroup in memorydb to populate subnetGroupName.
	// +kubebuilder:validation:Optional
	SubnetGroupNameSelector *xpv1.NamespacedSelector `json:"subnetGroupNameSelector,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// A flag to enable in-transit encryption on the cluster.
	// +kubebuilder:validation:Optional
	TLSEnabled *bool `json:"tlsEnabled,omitempty"`
}

// ClusterRAWSpec defines the desired state of namespaced ClusterRAW.
type ClusterRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider ClusterRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider ClusterRAWInitParameters `json:"initProvider,omitempty"`
}

// ClusterRAWStatus defines the observed state of namespaced ClusterRAW.
type ClusterRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider clusternative.ClusterRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// ClusterRAW is the native (non-Terraform) Schema for AWS MemoryDB Cluster (namespaced scope).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}
// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.nodeType) || (has(self.initProvider) && has(self.initProvider.nodeType))",message="spec.forProvider.nodeType is a required parameter"
type ClusterRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterRAWSpec   `json:"spec"`
	Status ClusterRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterRAWList contains a list of ClusterRAW resources (namespaced scope).
type ClusterRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterRAW `json:"items"`
}

// Repository type metadata for namespaced ClusterRAW.
var (
	ClusterRAW_Kind             = "ClusterRAW"
	ClusterRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: ClusterRAW_Kind}.String()
	ClusterRAW_KindAPIVersion   = ClusterRAW_Kind + "." + CRDGroupVersion.String()
	ClusterRAW_GroupVersionKind = CRDGroupVersion.WithKind(ClusterRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&ClusterRAW{}, &ClusterRAWList{})
}

// GetForProvider returns a cluster-scoped ClusterRAWParameters populated from
// this namespaced resource's ForProvider fields.
func (c *ClusterRAW) GetForProvider() *clusternative.ClusterRAWParameters {
	return &clusternative.ClusterRAWParameters{
		ACLName:                 c.Spec.ForProvider.ACLName,
		AutoMinorVersionUpgrade: c.Spec.ForProvider.AutoMinorVersionUpgrade,
		DataTiering:             c.Spec.ForProvider.DataTiering,
		Description:             c.Spec.ForProvider.Description,
		Engine:                  c.Spec.ForProvider.Engine,
		EngineVersion:           c.Spec.ForProvider.EngineVersion,
		FinalSnapshotName:       c.Spec.ForProvider.FinalSnapshotName,
		IPDiscovery:             c.Spec.ForProvider.IPDiscovery,
		KMSKeyArn:               c.Spec.ForProvider.KMSKeyArn,
		MaintenanceWindow:       c.Spec.ForProvider.MaintenanceWindow,
		MultiRegionClusterName:  c.Spec.ForProvider.MultiRegionClusterName,
		NetworkType:             c.Spec.ForProvider.NetworkType,
		NodeType:                c.Spec.ForProvider.NodeType,
		NumReplicasPerShard:     c.Spec.ForProvider.NumReplicasPerShard,
		NumShards:               c.Spec.ForProvider.NumShards,
		ParameterGroupName:      c.Spec.ForProvider.ParameterGroupName,
		Port:                    c.Spec.ForProvider.Port,
		Region:                  c.Spec.ForProvider.Region,
		SecurityGroupIds:        c.Spec.ForProvider.SecurityGroupIds,
		SnapshotArns:            c.Spec.ForProvider.SnapshotArns,
		SnapshotName:            c.Spec.ForProvider.SnapshotName,
		SnapshotRetentionLimit:  c.Spec.ForProvider.SnapshotRetentionLimit,
		SnapshotWindow:          c.Spec.ForProvider.SnapshotWindow,
		SnsTopicArn:             c.Spec.ForProvider.SnsTopicArn,
		SubnetGroupName:         c.Spec.ForProvider.SubnetGroupName,
		Tags:                    c.Spec.ForProvider.Tags,
		TLSEnabled:              c.Spec.ForProvider.TLSEnabled,
	}
}

// SetForProvider writes back a cluster-scoped ClusterRAWParameters to this
// namespaced resource's ForProvider fields. This is the symmetric inverse of
// GetForProvider, required so that late-initialized fields are persisted.
func (c *ClusterRAW) SetForProvider(p clusternative.ClusterRAWParameters) {
	c.Spec.ForProvider.ACLName = p.ACLName
	c.Spec.ForProvider.AutoMinorVersionUpgrade = p.AutoMinorVersionUpgrade
	c.Spec.ForProvider.DataTiering = p.DataTiering
	c.Spec.ForProvider.Description = p.Description
	c.Spec.ForProvider.Engine = p.Engine
	c.Spec.ForProvider.EngineVersion = p.EngineVersion
	c.Spec.ForProvider.FinalSnapshotName = p.FinalSnapshotName
	c.Spec.ForProvider.IPDiscovery = p.IPDiscovery
	c.Spec.ForProvider.KMSKeyArn = p.KMSKeyArn
	c.Spec.ForProvider.MaintenanceWindow = p.MaintenanceWindow
	c.Spec.ForProvider.MultiRegionClusterName = p.MultiRegionClusterName
	c.Spec.ForProvider.NetworkType = p.NetworkType
	c.Spec.ForProvider.NodeType = p.NodeType
	c.Spec.ForProvider.NumReplicasPerShard = p.NumReplicasPerShard
	c.Spec.ForProvider.NumShards = p.NumShards
	c.Spec.ForProvider.ParameterGroupName = p.ParameterGroupName
	c.Spec.ForProvider.Port = p.Port
	c.Spec.ForProvider.Region = p.Region
	c.Spec.ForProvider.SecurityGroupIds = p.SecurityGroupIds
	c.Spec.ForProvider.SnapshotArns = p.SnapshotArns
	c.Spec.ForProvider.SnapshotName = p.SnapshotName
	c.Spec.ForProvider.SnapshotRetentionLimit = p.SnapshotRetentionLimit
	c.Spec.ForProvider.SnapshotWindow = p.SnapshotWindow
	c.Spec.ForProvider.SnsTopicArn = p.SnsTopicArn
	c.Spec.ForProvider.SubnetGroupName = p.SubnetGroupName
	c.Spec.ForProvider.Tags = p.Tags
	c.Spec.ForProvider.TLSEnabled = p.TLSEnabled
}

// GetInitProvider returns a cluster-scoped ClusterRAWInitParameters populated from
// this namespaced resource's InitProvider fields.
func (c *ClusterRAW) GetInitProvider() *clusternative.ClusterRAWInitParameters {
	return &clusternative.ClusterRAWInitParameters{
		ACLName:                 c.Spec.InitProvider.ACLName,
		AutoMinorVersionUpgrade: c.Spec.InitProvider.AutoMinorVersionUpgrade,
		DataTiering:             c.Spec.InitProvider.DataTiering,
		Description:             c.Spec.InitProvider.Description,
		Engine:                  c.Spec.InitProvider.Engine,
		EngineVersion:           c.Spec.InitProvider.EngineVersion,
		FinalSnapshotName:       c.Spec.InitProvider.FinalSnapshotName,
		IPDiscovery:             c.Spec.InitProvider.IPDiscovery,
		KMSKeyArn:               c.Spec.InitProvider.KMSKeyArn,
		MaintenanceWindow:       c.Spec.InitProvider.MaintenanceWindow,
		MultiRegionClusterName:  c.Spec.InitProvider.MultiRegionClusterName,
		NetworkType:             c.Spec.InitProvider.NetworkType,
		NodeType:                c.Spec.InitProvider.NodeType,
		NumReplicasPerShard:     c.Spec.InitProvider.NumReplicasPerShard,
		NumShards:               c.Spec.InitProvider.NumShards,
		ParameterGroupName:      c.Spec.InitProvider.ParameterGroupName,
		Port:                    c.Spec.InitProvider.Port,
		SecurityGroupIds:        c.Spec.InitProvider.SecurityGroupIds,
		SnapshotArns:            c.Spec.InitProvider.SnapshotArns,
		SnapshotName:            c.Spec.InitProvider.SnapshotName,
		SnapshotRetentionLimit:  c.Spec.InitProvider.SnapshotRetentionLimit,
		SnapshotWindow:          c.Spec.InitProvider.SnapshotWindow,
		SnsTopicArn:             c.Spec.InitProvider.SnsTopicArn,
		SubnetGroupName:         c.Spec.InitProvider.SubnetGroupName,
		Tags:                    c.Spec.InitProvider.Tags,
		TLSEnabled:              c.Spec.InitProvider.TLSEnabled,
	}
}

// GetAtProvider returns the current observed state.
func (c *ClusterRAW) GetAtProvider() clusternative.ClusterRAWObservation {
	return c.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (c *ClusterRAW) SetAtProvider(o clusternative.ClusterRAWObservation) {
	c.Status.AtProvider = o
}
