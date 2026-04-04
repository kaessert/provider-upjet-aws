// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

	clusterv2native "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta2/native"
)

// RGLogDeliveryConfigurationRAWParameters defines the log delivery configuration for namespaced ReplicationGroup.
// Note: destination auto-ref is intentionally deleted (polymorphic target).
type RGLogDeliveryConfigurationRAWParameters struct {
	// Name of either the CloudWatch Logs LogGroup or Kinesis Data Firehose resource.
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

// RGLogDeliveryConfigurationRAWInitParameters defines init parameters for log delivery.
type RGLogDeliveryConfigurationRAWInitParameters struct {
	Destination     *string `json:"destination,omitempty"`
	DestinationType *string `json:"destinationType,omitempty"`
	LogFormat       *string `json:"logFormat,omitempty"`
	LogType         *string `json:"logType,omitempty"`
}

// NodeGroupConfigurationRAWParameters defines node group configuration.
type NodeGroupConfigurationRAWParameters struct {
	// ID for the node group.
	// +kubebuilder:validation:Optional
	NodeGroupID *string `json:"nodeGroupId,omitempty"`

	// Availability zone for the primary node.
	// +kubebuilder:validation:Optional
	PrimaryAvailabilityZone *string `json:"primaryAvailabilityZone,omitempty"`

	// ARN of the Outpost for the primary node.
	// +kubebuilder:validation:Optional
	PrimaryOutpostArn *string `json:"primaryOutpostArn,omitempty"`

	// List of availability zones for the replica nodes.
	// +kubebuilder:validation:Optional
	ReplicaAvailabilityZones []*string `json:"replicaAvailabilityZones,omitempty"`

	// Number of replica nodes in this node group.
	// +kubebuilder:validation:Optional
	ReplicaCount *float64 `json:"replicaCount,omitempty"`

	// List of ARNs of the Outposts for the replica nodes.
	// +kubebuilder:validation:Optional
	ReplicaOutpostArns []*string `json:"replicaOutpostArns,omitempty"`

	// Keyspace for this node group.
	// +kubebuilder:validation:Optional
	Slots *string `json:"slots,omitempty"`
}

// NodeGroupConfigurationRAWInitParameters defines init parameters for node group configuration.
type NodeGroupConfigurationRAWInitParameters struct {
	NodeGroupID              *string   `json:"nodeGroupId,omitempty"`
	PrimaryAvailabilityZone  *string   `json:"primaryAvailabilityZone,omitempty"`
	PrimaryOutpostArn        *string   `json:"primaryOutpostArn,omitempty"`
	ReplicaAvailabilityZones []*string `json:"replicaAvailabilityZones,omitempty"`
	ReplicaCount             *float64  `json:"replicaCount,omitempty"`
	ReplicaOutpostArns       []*string `json:"replicaOutpostArns,omitempty"`
	Slots                    *string   `json:"slots,omitempty"`
}

// ReplicationGroupRAWParameters defines the namespaced configuration parameters for a native ElastiCache Replication Group (v1beta2 hub).
// Fields mirror the cluster-scoped v1beta2 type but reference annotations point to namespaced packages.
type ReplicationGroupRAWParameters struct {
	// Specifies whether any modifications are applied immediately, or during the next maintenance window.
	// +kubebuilder:validation:Optional
	ApplyImmediately *bool `json:"applyImmediately,omitempty"`

	// Whether to enable encryption at rest. *string to match TF CRD schema parity.
	// +kubebuilder:validation:Optional
	AtRestEncryptionEnabled *string `json:"atRestEncryptionEnabled,omitempty"`

	// Password used to access a password protected server.
	// +kubebuilder:validation:Optional
	AuthTokenSecretRef *xpv1.SecretKeySelector `json:"authTokenSecretRef,omitempty"`

	// Strategy used when modifying auth_token. Valid values: SET, ROTATE, DELETE.
	// +kubebuilder:validation:Optional
	AuthTokenUpdateStrategy *string `json:"authTokenUpdateStrategy,omitempty"`

	// Specifies whether minor version engine upgrades will be applied automatically. *string to match TF CRD schema parity.
	// +kubebuilder:validation:Optional
	AutoMinorVersionUpgrade *string `json:"autoMinorVersionUpgrade,omitempty"`

	// If true, generates a random auth token and writes it to the Secret at authTokenSecretRef.
	// +kubebuilder:validation:Optional
	AutoGenerateAuthToken *bool `json:"autoGenerateAuthToken,omitempty"`

	// Specifies whether a read-only replica will be automatically promoted.
	// +kubebuilder:validation:Optional
	AutomaticFailoverEnabled *bool `json:"automaticFailoverEnabled,omitempty"`

	// Specifies whether cluster mode is enabled or disabled. Valid values: enabled, disabled, compatible.
	// +kubebuilder:validation:Optional
	ClusterMode *string `json:"clusterMode,omitempty"`

	// Enables data tiering. Only supported for r6gd node type.
	// +kubebuilder:validation:Optional
	DataTieringEnabled *bool `json:"dataTieringEnabled,omitempty"`

	// User-created description for the replication group.
	// +kubebuilder:validation:Optional
	Description *string `json:"description,omitempty"`

	// Name of the cache engine. Valid values: redis, valkey.
	// +kubebuilder:validation:Optional
	Engine *string `json:"engine,omitempty"`

	// Version number of the cache engine.
	// +kubebuilder:validation:Optional
	EngineVersion *string `json:"engineVersion,omitempty"`

	// The name of your final node group snapshot.
	// +kubebuilder:validation:Optional
	FinalSnapshotIdentifier *string `json:"finalSnapshotIdentifier,omitempty"`

	// The ID of the global replication group.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/elasticache/v1beta1.GlobalReplicationGroup
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractAtProviderField("globalReplicationGroupId")
	// +kubebuilder:validation:Optional
	GlobalReplicationGroupID *string `json:"globalReplicationGroupId,omitempty"`

	// Reference to a GlobalReplicationGroupRAW in elasticache to populate globalReplicationGroupId.
	// +kubebuilder:validation:Optional
	GlobalReplicationGroupIDRef *xpv1.NamespacedReference `json:"globalReplicationGroupIdRef,omitempty"`

	// Selector for a GlobalReplicationGroupRAW in elasticache to populate globalReplicationGroupId.
	// +kubebuilder:validation:Optional
	GlobalReplicationGroupIDSelector *xpv1.NamespacedSelector `json:"globalReplicationGroupIdSelector,omitempty"`

	// The IP version to advertise in the discovery protocol.
	// +kubebuilder:validation:Optional
	IPDiscovery *string `json:"ipDiscovery,omitempty"`

	// The ARN of the KMS key to use if encrypting at rest.
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

	// Log delivery configuration.
	// +kubebuilder:validation:Optional
	LogDeliveryConfiguration []RGLogDeliveryConfigurationRAWParameters `json:"logDeliveryConfiguration,omitempty"`

	// The weekly time range for when maintenance is performed.
	// +kubebuilder:validation:Optional
	MaintenanceWindow *string `json:"maintenanceWindow,omitempty"`

	// Specifies whether to enable Multi-AZ Support.
	// +kubebuilder:validation:Optional
	MultiAzEnabled *bool `json:"multiAzEnabled,omitempty"`

	// The IP versions for cache cluster connections.
	// +kubebuilder:validation:Optional
	NetworkType *string `json:"networkType,omitempty"`

	// Configuration block for node groups (shards).
	// +kubebuilder:validation:Optional
	NodeGroupConfiguration []NodeGroupConfigurationRAWParameters `json:"nodeGroupConfiguration,omitempty"`

	// Instance class to be used.
	// +kubebuilder:validation:Optional
	NodeType *string `json:"nodeType,omitempty"`

	// ARN of an SNS topic to send ElastiCache notifications to.
	// +kubebuilder:validation:Optional
	NotificationTopicArn *string `json:"notificationTopicArn,omitempty"`

	// The number of cache clusters this replication group will have.
	// +kubebuilder:validation:Optional
	NumCacheClusters *float64 `json:"numCacheClusters,omitempty"`

	// Number of node groups (shards).
	// +kubebuilder:validation:Optional
	NumNodeGroups *float64 `json:"numNodeGroups,omitempty"`

	// Name of the parameter group.
	// +kubebuilder:validation:Optional
	ParameterGroupName *string `json:"parameterGroupName,omitempty"`

	// Port number on which each cache node accepts connections.
	// +kubebuilder:validation:Optional
	Port *float64 `json:"port,omitempty"`

	// List of EC2 availability zones for cache clusters.
	// +kubebuilder:validation:Optional
	PreferredCacheClusterAzs []*string `json:"preferredCacheClusterAzs,omitempty"`

	// Number of replica nodes in each node group.
	// +kubebuilder:validation:Optional
	ReplicasPerNodeGroup *float64 `json:"replicasPerNodeGroup,omitempty"`

	// Region where this resource will be managed.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// References to SecurityGroup in ec2 to populate securityGroupIds.
	// +kubebuilder:validation:Optional
	SecurityGroupIDRefs []xpv1.NamespacedReference `json:"securityGroupIdRefs,omitempty"`

	// Selector for a list of SecurityGroup in ec2 to populate securityGroupIds.
	// +kubebuilder:validation:Optional
	SecurityGroupIDSelector *xpv1.NamespacedSelector `json:"securityGroupIdSelector,omitempty"`

	// IDs of one or more Amazon VPC security groups.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/ec2/v1beta1.SecurityGroup
	// +crossplane:generate:reference:refFieldName=SecurityGroupIDRefs
	// +crossplane:generate:reference:selectorFieldName=SecurityGroupIDSelector
	// +kubebuilder:validation:Optional
	// +listType=set
	SecurityGroupIds []*string `json:"securityGroupIds,omitempty"`

	// Names of Amazon VPC security groups (EC2-Classic legacy field, not updated).
	// +kubebuilder:validation:Optional
	// +listType=set
	SecurityGroupNames []*string `json:"securityGroupNames,omitempty"`

	// List of ARNs that identify Redis RDB snapshot files.
	// +kubebuilder:validation:Optional
	// +listType=set
	SnapshotArns []*string `json:"snapshotArns,omitempty"`

	// Name of a snapshot from which to restore data.
	// +kubebuilder:validation:Optional
	SnapshotName *string `json:"snapshotName,omitempty"`

	// Number of days for which ElastiCache will retain automatic snapshots.
	// +kubebuilder:validation:Optional
	SnapshotRetentionLimit *float64 `json:"snapshotRetentionLimit,omitempty"`

	// Daily time range during which ElastiCache will begin taking snapshots.
	// +kubebuilder:validation:Optional
	SnapshotWindow *string `json:"snapshotWindow,omitempty"`

	// Name of the cache subnet group.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/elasticache/v1beta1.SubnetGroup
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

	// Whether to enable encryption in transit.
	// +kubebuilder:validation:Optional
	TransitEncryptionEnabled *bool `json:"transitEncryptionEnabled,omitempty"`

	// A setting that enables clients to migrate to in-transit encryption with no downtime.
	// +kubebuilder:validation:Optional
	TransitEncryptionMode *string `json:"transitEncryptionMode,omitempty"`

	// User Group ID to associate with the replication group.
	// +kubebuilder:validation:Optional
	// +listType=set
	UserGroupIds []*string `json:"userGroupIds,omitempty"`
}

// ReplicationGroupRAWInitParameters defines init parameters for namespaced ReplicationGroupRAW (v1beta2 hub).
type ReplicationGroupRAWInitParameters struct {
	ApplyImmediately                 *bool                                         `json:"applyImmediately,omitempty"`
	AtRestEncryptionEnabled          *string                                       `json:"atRestEncryptionEnabled,omitempty"`
	AuthTokenUpdateStrategy          *string                                       `json:"authTokenUpdateStrategy,omitempty"`
	AutoMinorVersionUpgrade          *string                                       `json:"autoMinorVersionUpgrade,omitempty"`
	AutoGenerateAuthToken            *bool                                         `json:"autoGenerateAuthToken,omitempty"`
	AutomaticFailoverEnabled         *bool                                         `json:"automaticFailoverEnabled,omitempty"`
	ClusterMode                      *string                                       `json:"clusterMode,omitempty"`
	DataTieringEnabled               *bool                                         `json:"dataTieringEnabled,omitempty"`
	Description                      *string                                       `json:"description,omitempty"`
	Engine                           *string                                       `json:"engine,omitempty"`
	EngineVersion                    *string                                       `json:"engineVersion,omitempty"`
	FinalSnapshotIdentifier          *string                                       `json:"finalSnapshotIdentifier,omitempty"`
	GlobalReplicationGroupID         *string                                       `json:"globalReplicationGroupId,omitempty"`
	GlobalReplicationGroupIDRef      *xpv1.NamespacedReference                     `json:"globalReplicationGroupIdRef,omitempty"`
	GlobalReplicationGroupIDSelector *xpv1.NamespacedSelector                      `json:"globalReplicationGroupIdSelector,omitempty"`
	IPDiscovery                      *string                                       `json:"ipDiscovery,omitempty"`
	KMSKeyID                         *string                                       `json:"kmsKeyId,omitempty"`
	KMSKeyIDRef                      *xpv1.NamespacedReference                     `json:"kmsKeyIdRef,omitempty"`
	KMSKeyIDSelector                 *xpv1.NamespacedSelector                      `json:"kmsKeyIdSelector,omitempty"`
	LogDeliveryConfiguration         []RGLogDeliveryConfigurationRAWInitParameters `json:"logDeliveryConfiguration,omitempty"`
	MaintenanceWindow                *string                                       `json:"maintenanceWindow,omitempty"`
	MultiAzEnabled                   *bool                                         `json:"multiAzEnabled,omitempty"`
	NetworkType                      *string                                       `json:"networkType,omitempty"`
	NodeGroupConfiguration           []NodeGroupConfigurationRAWInitParameters     `json:"nodeGroupConfiguration,omitempty"`
	NodeType                         *string                                       `json:"nodeType,omitempty"`
	NotificationTopicArn             *string                                       `json:"notificationTopicArn,omitempty"`
	NumCacheClusters                 *float64                                      `json:"numCacheClusters,omitempty"`
	NumNodeGroups                    *float64                                      `json:"numNodeGroups,omitempty"`
	ParameterGroupName               *string                                       `json:"parameterGroupName,omitempty"`
	Port                             *float64                                      `json:"port,omitempty"`
	PreferredCacheClusterAzs         []*string                                     `json:"preferredCacheClusterAzs,omitempty"`
	ReplicasPerNodeGroup             *float64                                      `json:"replicasPerNodeGroup,omitempty"`
	SecurityGroupIDRefs              []xpv1.NamespacedReference                    `json:"securityGroupIdRefs,omitempty"`
	SecurityGroupIDSelector          *xpv1.NamespacedSelector                      `json:"securityGroupIdSelector,omitempty"`
	SecurityGroupIds                 []*string                                     `json:"securityGroupIds,omitempty"`
	SecurityGroupNames               []*string                                     `json:"securityGroupNames,omitempty"`
	SnapshotArns                     []*string                                     `json:"snapshotArns,omitempty"`
	SnapshotName                     *string                                       `json:"snapshotName,omitempty"`
	SnapshotRetentionLimit           *float64                                      `json:"snapshotRetentionLimit,omitempty"`
	SnapshotWindow                   *string                                       `json:"snapshotWindow,omitempty"`
	SubnetGroupName                  *string                                       `json:"subnetGroupName,omitempty"`
	SubnetGroupNameRef               *xpv1.NamespacedReference                     `json:"subnetGroupNameRef,omitempty"`
	SubnetGroupNameSelector          *xpv1.NamespacedSelector                      `json:"subnetGroupNameSelector,omitempty"`
	Tags                             map[string]*string                            `json:"tags,omitempty"`
	TransitEncryptionEnabled         *bool                                         `json:"transitEncryptionEnabled,omitempty"`
	TransitEncryptionMode            *string                                       `json:"transitEncryptionMode,omitempty"`
	UserGroupIds                     []*string                                     `json:"userGroupIds,omitempty"`
}

// ReplicationGroupRAWSpec defines the desired state of namespaced ReplicationGroupRAW (v1beta2 hub).
type ReplicationGroupRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider ReplicationGroupRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider ReplicationGroupRAWInitParameters `json:"initProvider,omitempty"`
}

// ReplicationGroupRAWStatus defines the observed state of namespaced ReplicationGroupRAW (v1beta2 hub).
type ReplicationGroupRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider clusterv2native.ReplicationGroupRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// ReplicationGroupRAW is the native (non-Terraform) Schema for AWS ElastiCache Replication Group (namespaced, v1beta2 hub).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}
type ReplicationGroupRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReplicationGroupRAWSpec   `json:"spec"`
	Status ReplicationGroupRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ReplicationGroupRAWList contains a list of ReplicationGroupRAW resources (namespaced, v1beta2).
type ReplicationGroupRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReplicationGroupRAW `json:"items"`
}

// Repository type metadata for namespaced ReplicationGroupRAW v1beta2.
var (
	ReplicationGroupRAW_Kind             = "ReplicationGroupRAW"
	ReplicationGroupRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: ReplicationGroupRAW_Kind}.String()
	ReplicationGroupRAW_KindAPIVersion   = ReplicationGroupRAW_Kind + "." + CRDGroupVersion.String()
	ReplicationGroupRAW_GroupVersionKind = CRDGroupVersion.WithKind(ReplicationGroupRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&ReplicationGroupRAW{}, &ReplicationGroupRAWList{})
}

// Hub marks ReplicationGroupRAW v1beta2 as the hub version for conversion.
func (*ReplicationGroupRAW) Hub() {}

// GetForProvider returns a cluster-scoped ReplicationGroupRAWParameters populated from
// this namespaced resource's ForProvider fields.
func (r *ReplicationGroupRAW) GetForProvider() *clusterv2native.ReplicationGroupRAWParameters {
	// Convert local log delivery configs to cluster types.
	logDelivery := make([]clusterv2native.RGLogDeliveryConfigurationRAWParameters, 0, len(r.Spec.ForProvider.LogDeliveryConfiguration))
	for _, ld := range r.Spec.ForProvider.LogDeliveryConfiguration {
		logDelivery = append(logDelivery, clusterv2native.RGLogDeliveryConfigurationRAWParameters{
			Destination:     ld.Destination,
			DestinationType: ld.DestinationType,
			LogFormat:       ld.LogFormat,
			LogType:         ld.LogType,
		})
	}
	// Convert local node group configs to cluster types.
	nodeGroupCfg := make([]clusterv2native.NodeGroupConfigurationRAWParameters, 0, len(r.Spec.ForProvider.NodeGroupConfiguration))
	for _, ng := range r.Spec.ForProvider.NodeGroupConfiguration {
		nodeGroupCfg = append(nodeGroupCfg, clusterv2native.NodeGroupConfigurationRAWParameters{
			NodeGroupID:              ng.NodeGroupID,
			PrimaryAvailabilityZone:  ng.PrimaryAvailabilityZone,
			PrimaryOutpostArn:        ng.PrimaryOutpostArn,
			ReplicaAvailabilityZones: ng.ReplicaAvailabilityZones,
			ReplicaCount:             ng.ReplicaCount,
			ReplicaOutpostArns:       ng.ReplicaOutpostArns,
			Slots:                    ng.Slots,
		})
	}
	return &clusterv2native.ReplicationGroupRAWParameters{
		ApplyImmediately:         r.Spec.ForProvider.ApplyImmediately,
		AtRestEncryptionEnabled:  r.Spec.ForProvider.AtRestEncryptionEnabled,
		AuthTokenSecretRef:       r.Spec.ForProvider.AuthTokenSecretRef,
		AuthTokenUpdateStrategy:  r.Spec.ForProvider.AuthTokenUpdateStrategy,
		AutoMinorVersionUpgrade:  r.Spec.ForProvider.AutoMinorVersionUpgrade,
		AutoGenerateAuthToken:    r.Spec.ForProvider.AutoGenerateAuthToken,
		AutomaticFailoverEnabled: r.Spec.ForProvider.AutomaticFailoverEnabled,
		ClusterMode:              r.Spec.ForProvider.ClusterMode,
		DataTieringEnabled:       r.Spec.ForProvider.DataTieringEnabled,
		Description:              r.Spec.ForProvider.Description,
		Engine:                   r.Spec.ForProvider.Engine,
		EngineVersion:            r.Spec.ForProvider.EngineVersion,
		FinalSnapshotIdentifier:  r.Spec.ForProvider.FinalSnapshotIdentifier,
		GlobalReplicationGroupID: r.Spec.ForProvider.GlobalReplicationGroupID,
		IPDiscovery:              r.Spec.ForProvider.IPDiscovery,
		KMSKeyID:                 r.Spec.ForProvider.KMSKeyID,
		LogDeliveryConfiguration: logDelivery,
		MaintenanceWindow:        r.Spec.ForProvider.MaintenanceWindow,
		MultiAzEnabled:           r.Spec.ForProvider.MultiAzEnabled,
		NetworkType:              r.Spec.ForProvider.NetworkType,
		NodeGroupConfiguration:   nodeGroupCfg,
		NodeType:                 r.Spec.ForProvider.NodeType,
		NotificationTopicArn:     r.Spec.ForProvider.NotificationTopicArn,
		NumCacheClusters:         r.Spec.ForProvider.NumCacheClusters,
		NumNodeGroups:            r.Spec.ForProvider.NumNodeGroups,
		ParameterGroupName:       r.Spec.ForProvider.ParameterGroupName,
		Port:                     r.Spec.ForProvider.Port,
		PreferredCacheClusterAzs: r.Spec.ForProvider.PreferredCacheClusterAzs,
		ReplicasPerNodeGroup:     r.Spec.ForProvider.ReplicasPerNodeGroup,
		Region:                   r.Spec.ForProvider.Region,
		SecurityGroupIds:         r.Spec.ForProvider.SecurityGroupIds,
		SecurityGroupNames:       r.Spec.ForProvider.SecurityGroupNames,
		SnapshotArns:             r.Spec.ForProvider.SnapshotArns,
		SnapshotName:             r.Spec.ForProvider.SnapshotName,
		SnapshotRetentionLimit:   r.Spec.ForProvider.SnapshotRetentionLimit,
		SnapshotWindow:           r.Spec.ForProvider.SnapshotWindow,
		SubnetGroupName:          r.Spec.ForProvider.SubnetGroupName,
		Tags:                     r.Spec.ForProvider.Tags,
		TransitEncryptionEnabled: r.Spec.ForProvider.TransitEncryptionEnabled,
		TransitEncryptionMode:    r.Spec.ForProvider.TransitEncryptionMode,
		UserGroupIds:             r.Spec.ForProvider.UserGroupIds,
	}
}

// GetInitProvider returns a cluster-scoped ReplicationGroupRAWInitParameters populated from
// this namespaced resource's InitProvider fields.
func (r *ReplicationGroupRAW) GetInitProvider() *clusterv2native.ReplicationGroupRAWInitParameters {
	logDelivery := make([]clusterv2native.RGLogDeliveryConfigurationRAWInitParameters, 0, len(r.Spec.InitProvider.LogDeliveryConfiguration))
	for _, ld := range r.Spec.InitProvider.LogDeliveryConfiguration {
		logDelivery = append(logDelivery, clusterv2native.RGLogDeliveryConfigurationRAWInitParameters{
			Destination:     ld.Destination,
			DestinationType: ld.DestinationType,
			LogFormat:       ld.LogFormat,
			LogType:         ld.LogType,
		})
	}
	nodeGroupCfg := make([]clusterv2native.NodeGroupConfigurationRAWInitParameters, 0, len(r.Spec.InitProvider.NodeGroupConfiguration))
	for _, ng := range r.Spec.InitProvider.NodeGroupConfiguration {
		nodeGroupCfg = append(nodeGroupCfg, clusterv2native.NodeGroupConfigurationRAWInitParameters{
			NodeGroupID:              ng.NodeGroupID,
			PrimaryAvailabilityZone:  ng.PrimaryAvailabilityZone,
			PrimaryOutpostArn:        ng.PrimaryOutpostArn,
			ReplicaAvailabilityZones: ng.ReplicaAvailabilityZones,
			ReplicaCount:             ng.ReplicaCount,
			ReplicaOutpostArns:       ng.ReplicaOutpostArns,
			Slots:                    ng.Slots,
		})
	}
	return &clusterv2native.ReplicationGroupRAWInitParameters{
		ApplyImmediately:         r.Spec.InitProvider.ApplyImmediately,
		AtRestEncryptionEnabled:  r.Spec.InitProvider.AtRestEncryptionEnabled,
		AuthTokenUpdateStrategy:  r.Spec.InitProvider.AuthTokenUpdateStrategy,
		AutoMinorVersionUpgrade:  r.Spec.InitProvider.AutoMinorVersionUpgrade,
		AutoGenerateAuthToken:    r.Spec.InitProvider.AutoGenerateAuthToken,
		AutomaticFailoverEnabled: r.Spec.InitProvider.AutomaticFailoverEnabled,
		ClusterMode:              r.Spec.InitProvider.ClusterMode,
		DataTieringEnabled:       r.Spec.InitProvider.DataTieringEnabled,
		Description:              r.Spec.InitProvider.Description,
		Engine:                   r.Spec.InitProvider.Engine,
		EngineVersion:            r.Spec.InitProvider.EngineVersion,
		FinalSnapshotIdentifier:  r.Spec.InitProvider.FinalSnapshotIdentifier,
		GlobalReplicationGroupID: r.Spec.InitProvider.GlobalReplicationGroupID,
		IPDiscovery:              r.Spec.InitProvider.IPDiscovery,
		KMSKeyID:                 r.Spec.InitProvider.KMSKeyID,
		LogDeliveryConfiguration: logDelivery,
		MaintenanceWindow:        r.Spec.InitProvider.MaintenanceWindow,
		MultiAzEnabled:           r.Spec.InitProvider.MultiAzEnabled,
		NetworkType:              r.Spec.InitProvider.NetworkType,
		NodeGroupConfiguration:   nodeGroupCfg,
		NodeType:                 r.Spec.InitProvider.NodeType,
		NotificationTopicArn:     r.Spec.InitProvider.NotificationTopicArn,
		NumCacheClusters:         r.Spec.InitProvider.NumCacheClusters,
		NumNodeGroups:            r.Spec.InitProvider.NumNodeGroups,
		ParameterGroupName:       r.Spec.InitProvider.ParameterGroupName,
		Port:                     r.Spec.InitProvider.Port,
		PreferredCacheClusterAzs: r.Spec.InitProvider.PreferredCacheClusterAzs,
		ReplicasPerNodeGroup:     r.Spec.InitProvider.ReplicasPerNodeGroup,
		SecurityGroupIds:         r.Spec.InitProvider.SecurityGroupIds,
		SecurityGroupNames:       r.Spec.InitProvider.SecurityGroupNames,
		SnapshotArns:             r.Spec.InitProvider.SnapshotArns,
		SnapshotName:             r.Spec.InitProvider.SnapshotName,
		SnapshotRetentionLimit:   r.Spec.InitProvider.SnapshotRetentionLimit,
		SnapshotWindow:           r.Spec.InitProvider.SnapshotWindow,
		SubnetGroupName:          r.Spec.InitProvider.SubnetGroupName,
		Tags:                     r.Spec.InitProvider.Tags,
		TransitEncryptionEnabled: r.Spec.InitProvider.TransitEncryptionEnabled,
		TransitEncryptionMode:    r.Spec.InitProvider.TransitEncryptionMode,
		UserGroupIds:             r.Spec.InitProvider.UserGroupIds,
	}
}

// GetAtProvider returns the current observed state.
func (r *ReplicationGroupRAW) GetAtProvider() clusterv2native.ReplicationGroupRAWObservation {
	return r.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (r *ReplicationGroupRAW) SetAtProvider(o clusterv2native.ReplicationGroupRAWObservation) {
	r.Status.AtProvider = o
}
