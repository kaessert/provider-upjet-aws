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

// ClusterLogDeliveryConfigurationRAWParameters defines the log delivery configuration for a Cluster.
type ClusterLogDeliveryConfigurationRAWParameters struct {
	// Name of either the CloudWatch Logs LogGroup or Kinesis Data Firehose resource.
	// Note: destination auto-ref is intentionally deleted for this field (polymorphic target).
	// +kubebuilder:validation:Optional
	Destination *string `json:"destination"`

	// For CloudWatch Logs use cloudwatch-logs or for Kinesis Data Firehose use kinesis-firehose.
	// +kubebuilder:validation:Optional
	DestinationType *string `json:"destinationType"`

	// Valid values are json or text.
	// +kubebuilder:validation:Optional
	LogFormat *string `json:"logFormat"`

	// Valid values are slow-log or engine-log. Max 1 of each.
	// +kubebuilder:validation:Optional
	LogType *string `json:"logType"`
}

// ClusterLogDeliveryConfigurationRAWInitParameters defines init parameters for log delivery.
type ClusterLogDeliveryConfigurationRAWInitParameters struct {
	// Name of either the CloudWatch Logs LogGroup or Kinesis Data Firehose resource.
	// +kubebuilder:validation:Optional
	Destination *string `json:"destination,omitempty"`

	// For CloudWatch Logs use cloudwatch-logs or for Kinesis Data Firehose use kinesis-firehose.
	// +kubebuilder:validation:Optional
	DestinationType *string `json:"destinationType,omitempty"`

	// Valid values are json or text.
	// +kubebuilder:validation:Optional
	LogFormat *string `json:"logFormat,omitempty"`

	// Valid values are slow-log or engine-log. Max 1 of each.
	// +kubebuilder:validation:Optional
	LogType *string `json:"logType,omitempty"`
}

// ClusterLogDeliveryConfigurationRAWObservation defines the observed log delivery configuration.
type ClusterLogDeliveryConfigurationRAWObservation struct {
	Destination     *string `json:"destination,omitempty"`
	DestinationType *string `json:"destinationType,omitempty"`
	LogFormat       *string `json:"logFormat,omitempty"`
	LogType         *string `json:"logType,omitempty"`
}

// CacheNodeRAWObservation defines the observed state of a cache node.
type CacheNodeRAWObservation struct {
	Address          *string  `json:"address,omitempty"`
	AvailabilityZone *string  `json:"availabilityZone,omitempty"`
	ID               *string  `json:"id,omitempty"`
	OutpostArn       *string  `json:"outpostArn,omitempty"`
	Port             *float64 `json:"port,omitempty"`
}

// ClusterRAWParameters defines the configuration parameters for a native ElastiCache Cluster.
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
	LogDeliveryConfiguration []ClusterLogDeliveryConfigurationRAWParameters `json:"logDeliveryConfiguration,omitempty"`

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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta1/native.ParameterGroupRAW
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta2/native.ReplicationGroupRAW
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/ec2/v1beta1.SecurityGroup
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta1/native.SubnetGroupRAW
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

// ClusterRAWInitParameters defines the init parameters for ClusterRAW.
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
	LogDeliveryConfiguration []ClusterLogDeliveryConfigurationRAWInitParameters `json:"logDeliveryConfiguration,omitempty"`

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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta1/native.ParameterGroupRAW
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/ec2/v1beta1.SecurityGroup
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta1/native.SubnetGroupRAW
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

// ClusterRAWObservation defines the observed state of ClusterRAW.
type ClusterRAWObservation struct {
	// Whether database modifications are applied immediately.
	ApplyImmediately *bool `json:"applyImmediately,omitempty"`

	// The ARN of the created ElastiCache Cluster.
	Arn *string `json:"arn,omitempty"`

	// Specifies whether minor version engine upgrades are applied automatically.
	AutoMinorVersionUpgrade *string `json:"autoMinorVersionUpgrade,omitempty"`

	// Availability Zone for the cache cluster.
	AvailabilityZone *string `json:"availabilityZone,omitempty"`

	// AZ mode.
	AzMode *string `json:"azMode,omitempty"`

	// List of node objects including id, address, port and availability_zone.
	CacheNodes []CacheNodeRAWObservation `json:"cacheNodes,omitempty"`

	// The address of the replication group configuration endpoint (Memcached only).
	ClusterAddress *string `json:"clusterAddress,omitempty"`

	// The configuration endpoint for the cache cluster (Memcached only).
	ConfigurationEndpoint *string `json:"configurationEndpoint,omitempty"`

	// The cache cluster status.
	CacheClusterStatus *string `json:"cacheClusterStatus,omitempty"`

	// Name of the cache engine.
	Engine *string `json:"engine,omitempty"`

	// Version number of the cache engine.
	EngineVersion *string `json:"engineVersion,omitempty"`

	// The running version of the cache engine.
	EngineVersionActual *string `json:"engineVersionActual,omitempty"`

	// ID of the cache cluster.
	ID *string `json:"id,omitempty"`

	// Log delivery configuration.
	LogDeliveryConfiguration []ClusterLogDeliveryConfigurationRAWObservation `json:"logDeliveryConfiguration,omitempty"`

	// The weekly time range for maintenance.
	MaintenanceWindow *string `json:"maintenanceWindow,omitempty"`

	// The instance class used.
	NodeType *string `json:"nodeType,omitempty"`

	// Number of cache nodes.
	NumCacheNodes *float64 `json:"numCacheNodes,omitempty"`

	// The port number.
	Port *float64 `json:"port,omitempty"`

	// Replication group ID.
	ReplicationGroupID *string `json:"replicationGroupId,omitempty"`

	// Security group IDs.
	// +listType=set
	SecurityGroupIds []*string `json:"securityGroupIds,omitempty"`

	// Name of the snapshot.
	SnapshotName *string `json:"snapshotName,omitempty"`

	// Number of days for snapshot retention.
	SnapshotRetentionLimit *float64 `json:"snapshotRetentionLimit,omitempty"`

	// Snapshot window.
	SnapshotWindow *string `json:"snapshotWindow,omitempty"`

	// Name of the subnet group.
	SubnetGroupName *string `json:"subnetGroupName,omitempty"`

	// Tags assigned to the resource.
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// Whether encryption in-transit is enabled.
	TransitEncryptionEnabled *bool `json:"transitEncryptionEnabled,omitempty"`
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

// ClusterRAW is the native (non-Terraform) Schema for AWS ElastiCache Cluster.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
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

// GetInitProvider returns the InitProvider parameters.
func (c *ClusterRAW) GetInitProvider() *ClusterRAWInitParameters { return &c.Spec.InitProvider }

// GetAtProvider returns the current observed state.
func (c *ClusterRAW) GetAtProvider() ClusterRAWObservation { return c.Status.AtProvider }

// SetAtProvider sets the observed state.
func (c *ClusterRAW) SetAtProvider(o ClusterRAWObservation) { c.Status.AtProvider = o }
