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

// ClusterRAWParameters defines the namespaced configuration parameters for a native ElastiCache Cluster.
// Fields are identical to the cluster-scoped type; defined locally so that angryjet
// generates namespaced resolvers.
type ClusterRAWParameters struct {
	// Whether any database modifications are applied immediately or during maintenance.
	// +kubebuilder:validation:Optional
	ApplyImmediately *bool `json:"applyImmediately,omitempty"`

	// Specifies whether minor version engine upgrades will be applied automatically.
	// +kubebuilder:validation:Optional
	AutoMinorVersionUpgrade *string `json:"autoMinorVersionUpgrade,omitempty"`

	// Availability Zone for the cache cluster.
	// +kubebuilder:validation:Optional
	AvailabilityZone *string `json:"availabilityZone,omitempty"`

	// Whether nodes are created in a single or multiple AZs. Valid values: single-az, cross-az.
	// +kubebuilder:validation:Optional
	AzMode *string `json:"azMode,omitempty"`

	// Name of the cache engine. Valid values: memcached, redis, valkey.
	// +kubebuilder:validation:Optional
	Engine *string `json:"engine,omitempty"`

	// Version number of the cache engine.
	// +kubebuilder:validation:Optional
	EngineVersion *string `json:"engineVersion,omitempty"`

	// Name of final cluster snapshot. If omitted, no final snapshot will be made.
	// +kubebuilder:validation:Optional
	FinalSnapshotIdentifier *string `json:"finalSnapshotIdentifier,omitempty"`

	// The IP version to advertise in the discovery protocol. Valid values: ipv4, ipv6.
	// +kubebuilder:validation:Optional
	IPDiscovery *string `json:"ipDiscovery,omitempty"`

	// Log delivery configuration for the cluster.
	// +kubebuilder:validation:Optional
	LogDeliveryConfiguration []clusternative.ClusterLogDeliveryConfigurationRAWParameters `json:"logDeliveryConfiguration,omitempty"`

	// The weekly time range for maintenance. Format: ddd:hh24:mi-ddd:hh24:mi.
	// +kubebuilder:validation:Optional
	MaintenanceWindow *string `json:"maintenanceWindow,omitempty"`

	// The IP versions for cache cluster connections. Valid values: ipv4, ipv6, dual_stack.
	// +kubebuilder:validation:Optional
	NetworkType *string `json:"networkType,omitempty"`

	// The instance class used.
	// +kubebuilder:validation:Optional
	NodeType *string `json:"nodeType,omitempty"`

	// ARN of an SNS topic to send ElastiCache notifications to.
	// +kubebuilder:validation:Optional
	NotificationTopicArn *string `json:"notificationTopicArn,omitempty"`

	// The initial number of cache nodes.
	// +kubebuilder:validation:Optional
	NumCacheNodes *float64 `json:"numCacheNodes,omitempty"`

	// Outpost mode. Valid values: single-outpost, cross-outpost.
	// +kubebuilder:validation:Optional
	OutpostMode *string `json:"outpostMode,omitempty"`

	// The name of the parameter group to associate with this cache cluster.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/elasticache/v1beta1/native.ParameterGroupRAW
	// +kubebuilder:validation:Optional
	ParameterGroupName *string `json:"parameterGroupName,omitempty"`

	// Reference to a ParameterGroupRAW in elasticache to populate parameterGroupName.
	// +kubebuilder:validation:Optional
	ParameterGroupNameRef *xpv1.NamespacedReference `json:"parameterGroupNameRef,omitempty"`

	// Selector for a ParameterGroupRAW in elasticache to populate parameterGroupName.
	// +kubebuilder:validation:Optional
	ParameterGroupNameSelector *xpv1.NamespacedSelector `json:"parameterGroupNameSelector,omitempty"`

	// The port number on which cache nodes accept connections.
	// +kubebuilder:validation:Optional
	Port *float64 `json:"port,omitempty"`

	// List of Availability Zones in which cache nodes are created.
	// +kubebuilder:validation:Optional
	PreferredAvailabilityZones []*string `json:"preferredAvailabilityZones,omitempty"`

	// The outpost ARN in which the cache cluster will be created.
	// +kubebuilder:validation:Optional
	PreferredOutpostArn *string `json:"preferredOutpostArn,omitempty"`

	// Region where this resource will be managed.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// ID of the replication group to which this cluster should belong.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/elasticache/v1beta2/native.ReplicationGroupRAW
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractAtProviderField("id")
	// +kubebuilder:validation:Optional
	ReplicationGroupID *string `json:"replicationGroupId,omitempty"`

	// Reference to a ReplicationGroupRAW in elasticache to populate replicationGroupId.
	// +kubebuilder:validation:Optional
	ReplicationGroupIDRef *xpv1.NamespacedReference `json:"replicationGroupIdRef,omitempty"`

	// Selector for a ReplicationGroupRAW in elasticache to populate replicationGroupId.
	// +kubebuilder:validation:Optional
	ReplicationGroupIDSelector *xpv1.NamespacedSelector `json:"replicationGroupIdSelector,omitempty"`

	// References to SecurityGroup in ec2 to populate securityGroupIds.
	// +kubebuilder:validation:Optional
	SecurityGroupIDRefs []xpv1.NamespacedReference `json:"securityGroupIdRefs,omitempty"`

	// Selector for a list of SecurityGroup in ec2 to populate securityGroupIds.
	// +kubebuilder:validation:Optional
	SecurityGroupIDSelector *xpv1.NamespacedSelector `json:"securityGroupIdSelector,omitempty"`

	// One or more VPC security groups associated with the cache cluster.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SecurityGroupIDRefs
	// +crossplane:generate:reference:selectorFieldName=SecurityGroupIDSelector
	// +kubebuilder:validation:Optional
	// +listType=set
	SecurityGroupIds []*string `json:"securityGroupIds,omitempty"`

	// Single-element list with ARN of Redis RDB snapshot file stored in Amazon S3.
	// +kubebuilder:validation:Optional
	SnapshotArns []*string `json:"snapshotArns,omitempty"`

	// Name of a snapshot from which to restore data.
	// +kubebuilder:validation:Optional
	SnapshotName *string `json:"snapshotName,omitempty"`

	// Number of days for which ElastiCache will retain automatic cache cluster snapshots.
	// +kubebuilder:validation:Optional
	SnapshotRetentionLimit *float64 `json:"snapshotRetentionLimit,omitempty"`

	// Daily time range during which ElastiCache will begin taking snapshots.
	// +kubebuilder:validation:Optional
	SnapshotWindow *string `json:"snapshotWindow,omitempty"`

	// Name of the subnet group to be used for the cache cluster.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/elasticache/v1beta1/native.SubnetGroupRAW
	// +kubebuilder:validation:Optional
	SubnetGroupName *string `json:"subnetGroupName,omitempty"`

	// Reference to a SubnetGroupRAW in elasticache to populate subnetGroupName.
	// +kubebuilder:validation:Optional
	SubnetGroupNameRef *xpv1.NamespacedReference `json:"subnetGroupNameRef,omitempty"`

	// Selector for a SubnetGroupRAW in elasticache to populate subnetGroupName.
	// +kubebuilder:validation:Optional
	SubnetGroupNameSelector *xpv1.NamespacedSelector `json:"subnetGroupNameSelector,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// Enable encryption in-transit.
	// +kubebuilder:validation:Optional
	TransitEncryptionEnabled *bool `json:"transitEncryptionEnabled,omitempty"`
}

// ClusterRAWInitParameters defines the namespaced init parameters for ClusterRAW.
type ClusterRAWInitParameters struct {
	// Whether any database modifications are applied immediately or during maintenance.
	// +kubebuilder:validation:Optional
	ApplyImmediately *bool `json:"applyImmediately,omitempty"`

	// Specifies whether minor version engine upgrades will be applied automatically.
	// +kubebuilder:validation:Optional
	AutoMinorVersionUpgrade *string `json:"autoMinorVersionUpgrade,omitempty"`

	// Availability Zone for the cache cluster.
	// +kubebuilder:validation:Optional
	AvailabilityZone *string `json:"availabilityZone,omitempty"`

	// Whether nodes are created in a single or multiple AZs.
	// +kubebuilder:validation:Optional
	AzMode *string `json:"azMode,omitempty"`

	// Name of the cache engine.
	// +kubebuilder:validation:Optional
	Engine *string `json:"engine,omitempty"`

	// Version number of the cache engine.
	// +kubebuilder:validation:Optional
	EngineVersion *string `json:"engineVersion,omitempty"`

	// Name of final cluster snapshot.
	// +kubebuilder:validation:Optional
	FinalSnapshotIdentifier *string `json:"finalSnapshotIdentifier,omitempty"`

	// The IP version to advertise.
	// +kubebuilder:validation:Optional
	IPDiscovery *string `json:"ipDiscovery,omitempty"`

	// Log delivery configuration for the cluster.
	// +kubebuilder:validation:Optional
	LogDeliveryConfiguration []clusternative.ClusterLogDeliveryConfigurationRAWInitParameters `json:"logDeliveryConfiguration,omitempty"`

	// The weekly time range for maintenance.
	// +kubebuilder:validation:Optional
	MaintenanceWindow *string `json:"maintenanceWindow,omitempty"`

	// The IP versions for cache cluster connections.
	// +kubebuilder:validation:Optional
	NetworkType *string `json:"networkType,omitempty"`

	// The instance class used.
	// +kubebuilder:validation:Optional
	NodeType *string `json:"nodeType,omitempty"`

	// ARN of an SNS topic to send ElastiCache notifications to.
	// +kubebuilder:validation:Optional
	NotificationTopicArn *string `json:"notificationTopicArn,omitempty"`

	// The initial number of cache nodes.
	// +kubebuilder:validation:Optional
	NumCacheNodes *float64 `json:"numCacheNodes,omitempty"`

	// Outpost mode.
	// +kubebuilder:validation:Optional
	OutpostMode *string `json:"outpostMode,omitempty"`

	// The name of the parameter group.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/elasticache/v1beta1/native.ParameterGroupRAW
	// +kubebuilder:validation:Optional
	ParameterGroupName *string `json:"parameterGroupName,omitempty"`

	// Reference to a ParameterGroupRAW.
	// +kubebuilder:validation:Optional
	ParameterGroupNameRef *xpv1.NamespacedReference `json:"parameterGroupNameRef,omitempty"`

	// Selector for a ParameterGroupRAW.
	// +kubebuilder:validation:Optional
	ParameterGroupNameSelector *xpv1.NamespacedSelector `json:"parameterGroupNameSelector,omitempty"`

	// The port number on which cache nodes accept connections.
	// +kubebuilder:validation:Optional
	Port *float64 `json:"port,omitempty"`

	// List of Availability Zones in which cache nodes are created.
	// +kubebuilder:validation:Optional
	PreferredAvailabilityZones []*string `json:"preferredAvailabilityZones,omitempty"`

	// The outpost ARN.
	// +kubebuilder:validation:Optional
	PreferredOutpostArn *string `json:"preferredOutpostArn,omitempty"`

	// References to SecurityGroup in ec2.
	// +kubebuilder:validation:Optional
	SecurityGroupIDRefs []xpv1.NamespacedReference `json:"securityGroupIdRefs,omitempty"`

	// Selector for a list of SecurityGroup in ec2.
	// +kubebuilder:validation:Optional
	SecurityGroupIDSelector *xpv1.NamespacedSelector `json:"securityGroupIdSelector,omitempty"`

	// One or more VPC security groups.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SecurityGroupIDRefs
	// +crossplane:generate:reference:selectorFieldName=SecurityGroupIDSelector
	// +kubebuilder:validation:Optional
	// +listType=set
	SecurityGroupIds []*string `json:"securityGroupIds,omitempty"`

	// Single-element list with ARN of Redis RDB snapshot.
	// +kubebuilder:validation:Optional
	SnapshotArns []*string `json:"snapshotArns,omitempty"`

	// Name of a snapshot from which to restore data.
	// +kubebuilder:validation:Optional
	SnapshotName *string `json:"snapshotName,omitempty"`

	// Number of days for snapshot retention.
	// +kubebuilder:validation:Optional
	SnapshotRetentionLimit *float64 `json:"snapshotRetentionLimit,omitempty"`

	// Daily time range for snapshots.
	// +kubebuilder:validation:Optional
	SnapshotWindow *string `json:"snapshotWindow,omitempty"`

	// Name of the subnet group.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/elasticache/v1beta1/native.SubnetGroupRAW
	// +kubebuilder:validation:Optional
	SubnetGroupName *string `json:"subnetGroupName,omitempty"`

	// Reference to a SubnetGroupRAW.
	// +kubebuilder:validation:Optional
	SubnetGroupNameRef *xpv1.NamespacedReference `json:"subnetGroupNameRef,omitempty"`

	// Selector for a SubnetGroupRAW.
	// +kubebuilder:validation:Optional
	SubnetGroupNameSelector *xpv1.NamespacedSelector `json:"subnetGroupNameSelector,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// Enable encryption in-transit.
	// +kubebuilder:validation:Optional
	TransitEncryptionEnabled *bool `json:"transitEncryptionEnabled,omitempty"`
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

// ClusterRAW is the native (non-Terraform) Schema for AWS ElastiCache Cluster (namespaced scope).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}
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
		ApplyImmediately:           c.Spec.ForProvider.ApplyImmediately,
		AutoMinorVersionUpgrade:    c.Spec.ForProvider.AutoMinorVersionUpgrade,
		AvailabilityZone:           c.Spec.ForProvider.AvailabilityZone,
		AzMode:                     c.Spec.ForProvider.AzMode,
		Engine:                     c.Spec.ForProvider.Engine,
		EngineVersion:              c.Spec.ForProvider.EngineVersion,
		FinalSnapshotIdentifier:    c.Spec.ForProvider.FinalSnapshotIdentifier,
		IPDiscovery:                c.Spec.ForProvider.IPDiscovery,
		LogDeliveryConfiguration:   c.Spec.ForProvider.LogDeliveryConfiguration,
		MaintenanceWindow:          c.Spec.ForProvider.MaintenanceWindow,
		NetworkType:                c.Spec.ForProvider.NetworkType,
		NodeType:                   c.Spec.ForProvider.NodeType,
		NotificationTopicArn:       c.Spec.ForProvider.NotificationTopicArn,
		NumCacheNodes:              c.Spec.ForProvider.NumCacheNodes,
		OutpostMode:                c.Spec.ForProvider.OutpostMode,
		ParameterGroupName:         c.Spec.ForProvider.ParameterGroupName,
		Port:                       c.Spec.ForProvider.Port,
		PreferredAvailabilityZones: c.Spec.ForProvider.PreferredAvailabilityZones,
		PreferredOutpostArn:        c.Spec.ForProvider.PreferredOutpostArn,
		Region:                     c.Spec.ForProvider.Region,
		ReplicationGroupID:         c.Spec.ForProvider.ReplicationGroupID,
		SecurityGroupIds:           c.Spec.ForProvider.SecurityGroupIds,
		SnapshotArns:               c.Spec.ForProvider.SnapshotArns,
		SnapshotName:               c.Spec.ForProvider.SnapshotName,
		SnapshotRetentionLimit:     c.Spec.ForProvider.SnapshotRetentionLimit,
		SnapshotWindow:             c.Spec.ForProvider.SnapshotWindow,
		SubnetGroupName:            c.Spec.ForProvider.SubnetGroupName,
		Tags:                       c.Spec.ForProvider.Tags,
		TransitEncryptionEnabled:   c.Spec.ForProvider.TransitEncryptionEnabled,
	}
}

// SetForProvider writes back a cluster-scoped ClusterRAWParameters to this
// namespaced resource's ForProvider fields. Symmetric inverse of GetForProvider,
// required so that late-initialized fields are persisted.
func (c *ClusterRAW) SetForProvider(p clusternative.ClusterRAWParameters) {
	c.Spec.ForProvider.ApplyImmediately = p.ApplyImmediately
	c.Spec.ForProvider.AutoMinorVersionUpgrade = p.AutoMinorVersionUpgrade
	c.Spec.ForProvider.AvailabilityZone = p.AvailabilityZone
	c.Spec.ForProvider.AzMode = p.AzMode
	c.Spec.ForProvider.Engine = p.Engine
	c.Spec.ForProvider.EngineVersion = p.EngineVersion
	c.Spec.ForProvider.FinalSnapshotIdentifier = p.FinalSnapshotIdentifier
	c.Spec.ForProvider.IPDiscovery = p.IPDiscovery
	c.Spec.ForProvider.LogDeliveryConfiguration = p.LogDeliveryConfiguration
	c.Spec.ForProvider.MaintenanceWindow = p.MaintenanceWindow
	c.Spec.ForProvider.NetworkType = p.NetworkType
	c.Spec.ForProvider.NodeType = p.NodeType
	c.Spec.ForProvider.NotificationTopicArn = p.NotificationTopicArn
	c.Spec.ForProvider.NumCacheNodes = p.NumCacheNodes
	c.Spec.ForProvider.OutpostMode = p.OutpostMode
	c.Spec.ForProvider.ParameterGroupName = p.ParameterGroupName
	c.Spec.ForProvider.Port = p.Port
	c.Spec.ForProvider.PreferredAvailabilityZones = p.PreferredAvailabilityZones
	c.Spec.ForProvider.PreferredOutpostArn = p.PreferredOutpostArn
	c.Spec.ForProvider.Region = p.Region
	c.Spec.ForProvider.ReplicationGroupID = p.ReplicationGroupID
	c.Spec.ForProvider.SecurityGroupIds = p.SecurityGroupIds
	c.Spec.ForProvider.SnapshotArns = p.SnapshotArns
	c.Spec.ForProvider.SnapshotName = p.SnapshotName
	c.Spec.ForProvider.SnapshotRetentionLimit = p.SnapshotRetentionLimit
	c.Spec.ForProvider.SnapshotWindow = p.SnapshotWindow
	c.Spec.ForProvider.SubnetGroupName = p.SubnetGroupName
	c.Spec.ForProvider.Tags = p.Tags
	c.Spec.ForProvider.TransitEncryptionEnabled = p.TransitEncryptionEnabled
}

// GetInitProvider returns a cluster-scoped ClusterRAWInitParameters populated from
// this namespaced resource's InitProvider fields.
func (c *ClusterRAW) GetInitProvider() *clusternative.ClusterRAWInitParameters {
	return &clusternative.ClusterRAWInitParameters{
		ApplyImmediately:           c.Spec.InitProvider.ApplyImmediately,
		AutoMinorVersionUpgrade:    c.Spec.InitProvider.AutoMinorVersionUpgrade,
		AvailabilityZone:           c.Spec.InitProvider.AvailabilityZone,
		AzMode:                     c.Spec.InitProvider.AzMode,
		Engine:                     c.Spec.InitProvider.Engine,
		EngineVersion:              c.Spec.InitProvider.EngineVersion,
		FinalSnapshotIdentifier:    c.Spec.InitProvider.FinalSnapshotIdentifier,
		IPDiscovery:                c.Spec.InitProvider.IPDiscovery,
		LogDeliveryConfiguration:   c.Spec.InitProvider.LogDeliveryConfiguration,
		MaintenanceWindow:          c.Spec.InitProvider.MaintenanceWindow,
		NetworkType:                c.Spec.InitProvider.NetworkType,
		NodeType:                   c.Spec.InitProvider.NodeType,
		NotificationTopicArn:       c.Spec.InitProvider.NotificationTopicArn,
		NumCacheNodes:              c.Spec.InitProvider.NumCacheNodes,
		OutpostMode:                c.Spec.InitProvider.OutpostMode,
		ParameterGroupName:         c.Spec.InitProvider.ParameterGroupName,
		Port:                       c.Spec.InitProvider.Port,
		PreferredAvailabilityZones: c.Spec.InitProvider.PreferredAvailabilityZones,
		PreferredOutpostArn:        c.Spec.InitProvider.PreferredOutpostArn,
		SecurityGroupIds:           c.Spec.InitProvider.SecurityGroupIds,
		SnapshotArns:               c.Spec.InitProvider.SnapshotArns,
		SnapshotName:               c.Spec.InitProvider.SnapshotName,
		SnapshotRetentionLimit:     c.Spec.InitProvider.SnapshotRetentionLimit,
		SnapshotWindow:             c.Spec.InitProvider.SnapshotWindow,
		SubnetGroupName:            c.Spec.InitProvider.SubnetGroupName,
		Tags:                       c.Spec.InitProvider.Tags,
		TransitEncryptionEnabled:   c.Spec.InitProvider.TransitEncryptionEnabled,
	}
}

// GetAtProvider returns the current observed state.
func (c *ClusterRAW) GetAtProvider() clusternative.ClusterRAWObservation { return c.Status.AtProvider }

// SetAtProvider sets the observed state.
func (c *ClusterRAW) SetAtProvider(o clusternative.ClusterRAWObservation) { c.Status.AtProvider = o }
