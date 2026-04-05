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

// ClusterRAWClusterEndpointObservation defines the observed cluster endpoint.
type ClusterRAWClusterEndpointObservation struct {
	// DNS hostname of the cluster configuration endpoint.
	Address *string `json:"address,omitempty"`

	// The port number on which each of the nodes accepts connections.
	Port *float64 `json:"port,omitempty"`
}

// ClusterRAWEndpointObservation defines the observed endpoint for a shard node.
type ClusterRAWEndpointObservation struct {
	// DNS hostname.
	Address *string `json:"address,omitempty"`

	// Port number.
	Port *float64 `json:"port,omitempty"`
}

// ClusterRAWNodesObservation defines the observed state of a shard node.
type ClusterRAWNodesObservation struct {
	// The Availability Zone in which the node resides.
	AvailabilityZone *string `json:"availabilityZone,omitempty"`

	// The date and time when the node was created.
	CreateTime *string `json:"createTime,omitempty"`

	// The endpoint of the node.
	Endpoint []ClusterRAWEndpointObservation `json:"endpoint,omitempty"`

	// The name of the node.
	Name *string `json:"name,omitempty"`
}

// ClusterRAWShardsObservation defines the observed state of a shard.
type ClusterRAWShardsObservation struct {
	// The name of the shard.
	Name *string `json:"name,omitempty"`

	// Set of nodes in this shard.
	Nodes []ClusterRAWNodesObservation `json:"nodes,omitempty"`

	// Number of individual nodes in this shard.
	NumNodes *float64 `json:"numNodes,omitempty"`

	// Keyspace for this shard. Example: 0-16383.
	Slots *string `json:"slots,omitempty"`
}

// ClusterRAWParameters defines the configuration parameters for a native MemoryDB Cluster.
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

	// Version number of the engine to be used for the cluster. Downgrades are not supported.
	// +kubebuilder:validation:Optional
	EngineVersion *string `json:"engineVersion,omitempty"`

	// Name of the final cluster snapshot to be created when this resource is deleted.
	// If omitted, no final snapshot will be made.
	// +kubebuilder:validation:Optional
	FinalSnapshotName *string `json:"finalSnapshotName,omitempty"`

	// Mechanism that the cluster uses to discover IP addresses.
	// Valid values are ipv4 and ipv6. Defaults to ipv4.
	// +kubebuilder:validation:Optional
	IPDiscovery *string `json:"ipDiscovery,omitempty"`

	// ARN of the KMS key used to encrypt the cluster at rest.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/kms/v1beta1.Key
	// +kubebuilder:validation:Optional
	KMSKeyArn *string `json:"kmsKeyArn,omitempty"`

	// Reference to a Key in kms to populate kmsKeyArn.
	// +kubebuilder:validation:Optional
	KMSKeyArnRef *xpv1.NamespacedReference `json:"kmsKeyArnRef,omitempty"`

	// Selector for a Key in kms to populate kmsKeyArn.
	// +kubebuilder:validation:Optional
	KMSKeyArnSelector *xpv1.NamespacedSelector `json:"kmsKeyArnSelector,omitempty"`

	// Specifies the weekly time range during which maintenance on the cluster is performed.
	// Format: ddd:hh24:mi-ddd:hh24:mi (24H Clock UTC).
	// +kubebuilder:validation:Optional
	MaintenanceWindow *string `json:"maintenanceWindow,omitempty"`

	// The multi region cluster identifier specified on aws_memorydb_multi_region_cluster.
	// +kubebuilder:validation:Optional
	MultiRegionClusterName *string `json:"multiRegionClusterName,omitempty"`

	// IP address type for the cluster. Valid values are ipv4, ipv6 and dual_stack.
	// Defaults to ipv4.
	// +kubebuilder:validation:Optional
	NetworkType *string `json:"networkType,omitempty"`

	// The compute and memory capacity of the nodes in the cluster.
	// +kubebuilder:validation:Optional
	NodeType *string `json:"nodeType,omitempty"`

	// The number of replicas to apply to each shard, up to a maximum of 5.
	// Defaults to 1 (i.e. 2 nodes per shard).
	// +kubebuilder:validation:Optional
	NumReplicasPerShard *float64 `json:"numReplicasPerShard,omitempty"`

	// The number of shards in the cluster. Defaults to 1.
	// +kubebuilder:validation:Optional
	NumShards *float64 `json:"numShards,omitempty"`

	// The name of the parameter group associated with the cluster.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/memorydb/v1beta1.ParameterGroup
	// +kubebuilder:validation:Optional
	ParameterGroupName *string `json:"parameterGroupName,omitempty"`

	// Reference to a ParameterGroup in memorydb to populate parameterGroupName.
	// +kubebuilder:validation:Optional
	ParameterGroupNameRef *xpv1.NamespacedReference `json:"parameterGroupNameRef,omitempty"`

	// Selector for a ParameterGroup in memorydb to populate parameterGroupName.
	// +kubebuilder:validation:Optional
	ParameterGroupNameSelector *xpv1.NamespacedSelector `json:"parameterGroupNameSelector,omitempty"`

	// The port number on which each of the nodes accepts connections. Defaults to 6379.
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/ec2/v1beta1.SecurityGroup
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
	// When set to 0, automatic backups are disabled. Defaults to 0.
	// +kubebuilder:validation:Optional
	SnapshotRetentionLimit *float64 `json:"snapshotRetentionLimit,omitempty"`

	// The daily time range (in UTC) during which MemoryDB begins taking a daily snapshot.
	// Example: 05:00-09:00.
	// +kubebuilder:validation:Optional
	SnapshotWindow *string `json:"snapshotWindow,omitempty"`

	// ARN of the SNS topic to which cluster notifications are sent.
	// +kubebuilder:validation:Optional
	SnsTopicArn *string `json:"snsTopicArn,omitempty"`

	// The name of the subnet group to be used for the cluster.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/memorydb/v1beta1.SubnetGroup
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
	// When set to false, the acl_name must be open-access. Defaults to true.
	// +kubebuilder:validation:Optional
	TLSEnabled *bool `json:"tlsEnabled,omitempty"`
}

// ClusterRAWInitParameters defines the init parameters for ClusterRAW.
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

	// Version number of the engine to be used for the cluster.
	// +kubebuilder:validation:Optional
	EngineVersion *string `json:"engineVersion,omitempty"`

	// Name of the final cluster snapshot.
	// +kubebuilder:validation:Optional
	FinalSnapshotName *string `json:"finalSnapshotName,omitempty"`

	// Mechanism that the cluster uses to discover IP addresses.
	// +kubebuilder:validation:Optional
	IPDiscovery *string `json:"ipDiscovery,omitempty"`

	// ARN of the KMS key used to encrypt the cluster at rest.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/kms/v1beta1.Key
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/memorydb/v1beta1.ParameterGroup
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/ec2/v1beta1.SecurityGroup
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

	// The daily time range during which MemoryDB begins taking a daily snapshot.
	// +kubebuilder:validation:Optional
	SnapshotWindow *string `json:"snapshotWindow,omitempty"`

	// ARN of the SNS topic to which cluster notifications are sent.
	// +kubebuilder:validation:Optional
	SnsTopicArn *string `json:"snsTopicArn,omitempty"`

	// The name of the subnet group to be used for the cluster.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/memorydb/v1beta1.SubnetGroup
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

// ClusterRAWObservation defines the observed state of ClusterRAW.
type ClusterRAWObservation struct {
	// The ARN of the cluster.
	Arn *string `json:"arn,omitempty"`

	// The configuration endpoint for the cluster.
	ClusterEndpoint []ClusterRAWClusterEndpointObservation `json:"clusterEndpoint,omitempty"`

	// Patch version number of the engine used by the cluster.
	EnginePatchVersion *string `json:"enginePatchVersion,omitempty"`

	// Same as name.
	ID *string `json:"id,omitempty"`

	// Set of shards in this cluster.
	Shards []ClusterRAWShardsObservation `json:"shards,omitempty"`

	// A map of tags assigned to the resource, including those inherited from
	// the provider default_tags configuration block.
	// +mapType=granular
	TagsAll map[string]*string `json:"tagsAll,omitempty"`
}

// ClusterRAWSpec defines the desired state of ClusterRAW.
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

// ClusterRAWStatus defines the observed state of ClusterRAW.
type ClusterRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider ClusterRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// ClusterRAW is the native (non-Terraform) Schema for AWS MemoryDB Cluster.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.nodeType) || (has(self.initProvider) && has(self.initProvider.nodeType))",message="spec.forProvider.nodeType is a required parameter"
type ClusterRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterRAWSpec   `json:"spec"`
	Status ClusterRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterRAWList contains a list of ClusterRAW resources.
type ClusterRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterRAW `json:"items"`
}

// Repository type metadata for ClusterRAW.
var (
	ClusterRAW_Kind             = "ClusterRAW"
	ClusterRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: ClusterRAW_Kind}.String()
	ClusterRAW_KindAPIVersion   = ClusterRAW_Kind + "." + CRDGroupVersion.String()
	ClusterRAW_GroupVersionKind = CRDGroupVersion.WithKind(ClusterRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&ClusterRAW{}, &ClusterRAWList{})
}

// GetForProvider returns the ForProvider parameters.
func (c *ClusterRAW) GetForProvider() *ClusterRAWParameters { return &c.Spec.ForProvider }

// SetForProvider writes back the ForProvider parameters.
func (c *ClusterRAW) SetForProvider(p ClusterRAWParameters) { c.Spec.ForProvider = p }

// GetInitProvider returns the InitProvider parameters.
func (c *ClusterRAW) GetInitProvider() *ClusterRAWInitParameters { return &c.Spec.InitProvider }

// GetAtProvider returns the current observed state.
func (c *ClusterRAW) GetAtProvider() ClusterRAWObservation { return c.Status.AtProvider }

// SetAtProvider sets the observed state.
func (c *ClusterRAW) SetAtProvider(o ClusterRAWObservation) { c.Status.AtProvider = o }
