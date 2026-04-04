// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

	v1beta2native "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta2/native"
)

// RGLogDeliveryConfigurationRAWParameters defines log delivery configuration for ReplicationGroup (v1beta1).
// Note: destination auto-ref is intentionally deleted (polymorphic target).
type RGLogDeliveryConfigurationRAWParameters struct {
	Destination     *string `json:"destination"`
	DestinationType *string `json:"destinationType"`
	LogFormat       *string `json:"logFormat"`
	LogType         *string `json:"logType"`
}

// RGLogDeliveryConfigurationRAWInitParameters defines init parameters for RG log delivery (v1beta1).
type RGLogDeliveryConfigurationRAWInitParameters struct {
	Destination     *string `json:"destination,omitempty"`
	DestinationType *string `json:"destinationType,omitempty"`
	LogFormat       *string `json:"logFormat,omitempty"`
	LogType         *string `json:"logType,omitempty"`
}

// RGLogDeliveryConfigurationRAWObservation defines observed RG log delivery state (v1beta1).
type RGLogDeliveryConfigurationRAWObservation struct {
	Destination     *string `json:"destination,omitempty"`
	DestinationType *string `json:"destinationType,omitempty"`
	LogFormat       *string `json:"logFormat,omitempty"`
	LogType         *string `json:"logType,omitempty"`
}

// NodeGroupConfigurationRAWParameters defines node group configuration (v1beta1).
type NodeGroupConfigurationRAWParameters struct {
	NodeGroupID              *string   `json:"nodeGroupId,omitempty"`
	PrimaryAvailabilityZone  *string   `json:"primaryAvailabilityZone,omitempty"`
	PrimaryOutpostArn        *string   `json:"primaryOutpostArn,omitempty"`
	ReplicaAvailabilityZones []*string `json:"replicaAvailabilityZones,omitempty"`
	ReplicaCount             *float64  `json:"replicaCount,omitempty"`
	ReplicaOutpostArns       []*string `json:"replicaOutpostArns,omitempty"`
	Slots                    *string   `json:"slots,omitempty"`
}

// NodeGroupConfigurationRAWInitParameters defines init parameters for node group configuration (v1beta1).
type NodeGroupConfigurationRAWInitParameters struct {
	NodeGroupID              *string   `json:"nodeGroupId,omitempty"`
	PrimaryAvailabilityZone  *string   `json:"primaryAvailabilityZone,omitempty"`
	PrimaryOutpostArn        *string   `json:"primaryOutpostArn,omitempty"`
	ReplicaAvailabilityZones []*string `json:"replicaAvailabilityZones,omitempty"`
	ReplicaCount             *float64  `json:"replicaCount,omitempty"`
	ReplicaOutpostArns       []*string `json:"replicaOutpostArns,omitempty"`
	Slots                    *string   `json:"slots,omitempty"`
}

// NodeGroupConfigurationRAWObservation defines observed node group configuration (v1beta1).
type NodeGroupConfigurationRAWObservation struct {
	NodeGroupID              *string   `json:"nodeGroupId,omitempty"`
	PrimaryAvailabilityZone  *string   `json:"primaryAvailabilityZone,omitempty"`
	PrimaryOutpostArn        *string   `json:"primaryOutpostArn,omitempty"`
	ReplicaAvailabilityZones []*string `json:"replicaAvailabilityZones,omitempty"`
	ReplicaCount             *float64  `json:"replicaCount,omitempty"`
	ReplicaOutpostArns       []*string `json:"replicaOutpostArns,omitempty"`
	Slots                    *string   `json:"slots,omitempty"`
}

// ClusterModeConfigurationRAWParameters defines v1beta1-only cluster mode configuration.
// In v1beta2, numNodeGroups and replicasPerNodeGroup are top-level fields.
type ClusterModeConfigurationRAWParameters struct {
	NumNodeGroups        *float64 `json:"numNodeGroups,omitempty"`
	ReplicasPerNodeGroup *float64 `json:"replicasPerNodeGroup,omitempty"`
}

// ClusterModeConfigurationRAWInitParameters defines init params for v1beta1 cluster mode.
type ClusterModeConfigurationRAWInitParameters struct {
	NumNodeGroups        *float64 `json:"numNodeGroups,omitempty"`
	ReplicasPerNodeGroup *float64 `json:"replicasPerNodeGroup,omitempty"`
}

// ClusterModeConfigurationRAWObservation defines observed v1beta1 cluster mode state.
type ClusterModeConfigurationRAWObservation struct {
	NumNodeGroups        *float64 `json:"numNodeGroups,omitempty"`
	ReplicasPerNodeGroup *float64 `json:"replicasPerNodeGroup,omitempty"`
}

// ReplicationGroupRAWParameters defines the v1beta1 spoke parameters for ReplicationGroupRAW.
type ReplicationGroupRAWParameters struct {
	ApplyImmediately                 *bool                                     `json:"applyImmediately,omitempty"`
	AtRestEncryptionEnabled          *string                                   `json:"atRestEncryptionEnabled,omitempty"`
	AuthTokenSecretRef               *xpv1.SecretKeySelector                   `json:"authTokenSecretRef,omitempty"`
	AuthTokenUpdateStrategy          *string                                   `json:"authTokenUpdateStrategy,omitempty"`
	AutoMinorVersionUpgrade          *string                                   `json:"autoMinorVersionUpgrade,omitempty"`
	AutoGenerateAuthToken            *bool                                     `json:"autoGenerateAuthToken,omitempty"`
	AutomaticFailoverEnabled         *bool                                     `json:"automaticFailoverEnabled,omitempty"`
	ClusterMode                      *string                                   `json:"clusterMode,omitempty"`
	ClusterModeConfiguration         []ClusterModeConfigurationRAWParameters   `json:"clusterModeConfiguration,omitempty"`
	DataTieringEnabled               *bool                                     `json:"dataTieringEnabled,omitempty"`
	Description                      *string                                   `json:"description,omitempty"`
	Engine                           *string                                   `json:"engine,omitempty"`
	EngineVersion                    *string                                   `json:"engineVersion,omitempty"`
	FinalSnapshotIdentifier          *string                                   `json:"finalSnapshotIdentifier,omitempty"`
	GlobalReplicationGroupID         *string                                   `json:"globalReplicationGroupId,omitempty"`
	GlobalReplicationGroupIDRef      *xpv1.NamespacedReference                 `json:"globalReplicationGroupIdRef,omitempty"`
	GlobalReplicationGroupIDSelector *xpv1.NamespacedSelector                  `json:"globalReplicationGroupIdSelector,omitempty"`
	IPDiscovery                      *string                                   `json:"ipDiscovery,omitempty"`
	KMSKeyID                         *string                                   `json:"kmsKeyId,omitempty"`
	KMSKeyIDRef                      *xpv1.NamespacedReference                 `json:"kmsKeyIdRef,omitempty"`
	KMSKeyIDSelector                 *xpv1.NamespacedSelector                  `json:"kmsKeyIdSelector,omitempty"`
	LogDeliveryConfiguration         []RGLogDeliveryConfigurationRAWParameters `json:"logDeliveryConfiguration,omitempty"`
	MaintenanceWindow                *string                                   `json:"maintenanceWindow,omitempty"`
	MultiAzEnabled                   *bool                                     `json:"multiAzEnabled,omitempty"`
	NetworkType                      *string                                   `json:"networkType,omitempty"`
	NodeGroupConfiguration           []NodeGroupConfigurationRAWParameters     `json:"nodeGroupConfiguration,omitempty"`
	NodeType                         *string                                   `json:"nodeType,omitempty"`
	NotificationTopicArn             *string                                   `json:"notificationTopicArn,omitempty"`
	NumCacheClusters                 *float64                                  `json:"numCacheClusters,omitempty"`
	NumNodeGroups                    *float64                                  `json:"numNodeGroups,omitempty"`
	ParameterGroupName               *string                                   `json:"parameterGroupName,omitempty"`
	Port                             *float64                                  `json:"port,omitempty"`
	PreferredCacheClusterAzs         []*string                                 `json:"preferredCacheClusterAzs,omitempty"`
	ReplicasPerNodeGroup             *float64                                  `json:"replicasPerNodeGroup,omitempty"`
	// Region where this resource will be managed.
	// +kubebuilder:validation:Required
	Region                   *string                    `json:"region"`
	SecurityGroupIDRefs      []xpv1.NamespacedReference `json:"securityGroupIdRefs,omitempty"`
	SecurityGroupIDSelector  *xpv1.NamespacedSelector   `json:"securityGroupIdSelector,omitempty"`
	SecurityGroupIds         []*string                  `json:"securityGroupIds,omitempty"`
	SecurityGroupNames       []*string                  `json:"securityGroupNames,omitempty"`
	SnapshotArns             []*string                  `json:"snapshotArns,omitempty"`
	SnapshotName             *string                    `json:"snapshotName,omitempty"`
	SnapshotRetentionLimit   *float64                   `json:"snapshotRetentionLimit,omitempty"`
	SnapshotWindow           *string                    `json:"snapshotWindow,omitempty"`
	SubnetGroupName          *string                    `json:"subnetGroupName,omitempty"`
	SubnetGroupNameRef       *xpv1.NamespacedReference  `json:"subnetGroupNameRef,omitempty"`
	SubnetGroupNameSelector  *xpv1.NamespacedSelector   `json:"subnetGroupNameSelector,omitempty"`
	Tags                     map[string]*string         `json:"tags,omitempty"`
	TransitEncryptionEnabled *bool                      `json:"transitEncryptionEnabled,omitempty"`
	TransitEncryptionMode    *string                    `json:"transitEncryptionMode,omitempty"`
	UserGroupIds             []*string                  `json:"userGroupIds,omitempty"`
}

// ReplicationGroupRAWInitParameters defines the v1beta1 spoke init parameters.
type ReplicationGroupRAWInitParameters struct {
	ApplyImmediately                 *bool                                         `json:"applyImmediately,omitempty"`
	AtRestEncryptionEnabled          *string                                       `json:"atRestEncryptionEnabled,omitempty"`
	AuthTokenUpdateStrategy          *string                                       `json:"authTokenUpdateStrategy,omitempty"`
	AutoMinorVersionUpgrade          *string                                       `json:"autoMinorVersionUpgrade,omitempty"`
	AutoGenerateAuthToken            *bool                                         `json:"autoGenerateAuthToken,omitempty"`
	AutomaticFailoverEnabled         *bool                                         `json:"automaticFailoverEnabled,omitempty"`
	ClusterMode                      *string                                       `json:"clusterMode,omitempty"`
	ClusterModeConfiguration         []ClusterModeConfigurationRAWInitParameters   `json:"clusterModeConfiguration,omitempty"`
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

// ReplicationGroupRAWObservation defines the observed state of ReplicationGroupRAW v1beta1 spoke.
type ReplicationGroupRAWObservation struct {
	ApplyImmediately             *bool                                      `json:"applyImmediately,omitempty"`
	Arn                          *string                                    `json:"arn,omitempty"`
	AtRestEncryptionEnabled      *string                                    `json:"atRestEncryptionEnabled,omitempty"`
	AuthTokenUpdateStrategy      *string                                    `json:"authTokenUpdateStrategy,omitempty"`
	AutoMinorVersionUpgrade      *string                                    `json:"autoMinorVersionUpgrade,omitempty"`
	AutomaticFailoverEnabled     *bool                                      `json:"automaticFailoverEnabled,omitempty"`
	ClusterEnabled               *bool                                      `json:"clusterEnabled,omitempty"`
	ClusterMode                  *string                                    `json:"clusterMode,omitempty"`
	ClusterModeConfiguration     []ClusterModeConfigurationRAWObservation   `json:"clusterModeConfiguration,omitempty"`
	ConfigurationEndpointAddress *string                                    `json:"configurationEndpointAddress,omitempty"`
	DataTieringEnabled           *bool                                      `json:"dataTieringEnabled,omitempty"`
	Description                  *string                                    `json:"description,omitempty"`
	Engine                       *string                                    `json:"engine,omitempty"`
	EngineVersion                *string                                    `json:"engineVersion,omitempty"`
	EngineVersionActual          *string                                    `json:"engineVersionActual,omitempty"`
	FinalSnapshotIdentifier      *string                                    `json:"finalSnapshotIdentifier,omitempty"`
	GlobalReplicationGroupID     *string                                    `json:"globalReplicationGroupId,omitempty"`
	ID                           *string                                    `json:"id,omitempty"`
	IPDiscovery                  *string                                    `json:"ipDiscovery,omitempty"`
	KMSKeyID                     *string                                    `json:"kmsKeyId,omitempty"`
	LogDeliveryConfiguration     []RGLogDeliveryConfigurationRAWObservation `json:"logDeliveryConfiguration,omitempty"`
	MaintenanceWindow            *string                                    `json:"maintenanceWindow,omitempty"`
	MemberClusters               []*string                                  `json:"memberClusters,omitempty"`
	MultiAzEnabled               *bool                                      `json:"multiAzEnabled,omitempty"`
	NetworkType                  *string                                    `json:"networkType,omitempty"`
	NodeGroupConfiguration       []NodeGroupConfigurationRAWObservation     `json:"nodeGroupConfiguration,omitempty"`
	NodeType                     *string                                    `json:"nodeType,omitempty"`
	NotificationTopicArn         *string                                    `json:"notificationTopicArn,omitempty"`
	NumCacheClusters             *float64                                   `json:"numCacheClusters,omitempty"`
	NumNodeGroups                *float64                                   `json:"numNodeGroups,omitempty"`
	ParameterGroupName           *string                                    `json:"parameterGroupName,omitempty"`
	Port                         *float64                                   `json:"port,omitempty"`
	PreferredCacheClusterAzs     []*string                                  `json:"preferredCacheClusterAzs,omitempty"`
	PrimaryEndpointAddress       *string                                    `json:"primaryEndpointAddress,omitempty"`
	ReaderEndpointAddress        *string                                    `json:"readerEndpointAddress,omitempty"`
	ReplicasPerNodeGroup         *float64                                   `json:"replicasPerNodeGroup,omitempty"`
	SecurityGroupIds             []*string                                  `json:"securityGroupIds,omitempty"`
	SecurityGroupNames           []*string                                  `json:"securityGroupNames,omitempty"`
	SnapshotRetentionLimit       *float64                                   `json:"snapshotRetentionLimit,omitempty"`
	SnapshotWindow               *string                                    `json:"snapshotWindow,omitempty"`
	SubnetGroupName              *string                                    `json:"subnetGroupName,omitempty"`
	Tags                         map[string]*string                         `json:"tags,omitempty"`
	TransitEncryptionEnabled     *bool                                      `json:"transitEncryptionEnabled,omitempty"`
	TransitEncryptionMode        *string                                    `json:"transitEncryptionMode,omitempty"`
	UserGroupIds                 []*string                                  `json:"userGroupIds,omitempty"`
}

// ReplicationGroupRAWSpec defines the desired state of ReplicationGroupRAW v1beta1 spoke.
type ReplicationGroupRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`
	ForProvider              ReplicationGroupRAWParameters     `json:"forProvider"`
	InitProvider             ReplicationGroupRAWInitParameters `json:"initProvider,omitempty"`
}

// ReplicationGroupRAWStatus defines the observed state of ReplicationGroupRAW v1beta1 spoke.
type ReplicationGroupRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ReplicationGroupRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ReplicationGroupRAW is the v1beta1 spoke for the native ElastiCache Replication Group resource.
// The hub version is ReplicationGroupRAW in v1beta2. Use v1beta2 for new resources.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type ReplicationGroupRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ReplicationGroupRAWSpec   `json:"spec"`
	Status            ReplicationGroupRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ReplicationGroupRAWList contains a list of ReplicationGroupRAW resources (v1beta1 spoke).
type ReplicationGroupRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReplicationGroupRAW `json:"items"`
}

var (
	ReplicationGroupRAW_Kind             = "ReplicationGroupRAW"
	ReplicationGroupRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: ReplicationGroupRAW_Kind}.String()
	ReplicationGroupRAW_KindAPIVersion   = ReplicationGroupRAW_Kind + "." + CRDGroupVersion.String()
	ReplicationGroupRAW_GroupVersionKind = CRDGroupVersion.WithKind(ReplicationGroupRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&ReplicationGroupRAW{}, &ReplicationGroupRAWList{})
}

// ConvertTo converts ReplicationGroupRAW v1beta1 to the hub v1beta2.
func (src *ReplicationGroupRAW) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*v1beta2native.ReplicationGroupRAW)
	if !ok {
		return fmt.Errorf("expected *v1beta2native.ReplicationGroupRAW, got %T", dstRaw)
	}
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return err
	}
	if len(src.Spec.ForProvider.ClusterModeConfiguration) > 0 {
		cm := src.Spec.ForProvider.ClusterModeConfiguration[0]
		if cm.NumNodeGroups != nil && dst.Spec.ForProvider.NumNodeGroups == nil {
			dst.Spec.ForProvider.NumNodeGroups = cm.NumNodeGroups
		}
		if cm.ReplicasPerNodeGroup != nil && dst.Spec.ForProvider.ReplicasPerNodeGroup == nil {
			dst.Spec.ForProvider.ReplicasPerNodeGroup = cm.ReplicasPerNodeGroup
		}
	}
	dst.TypeMeta = metav1.TypeMeta{
		APIVersion: v1beta2native.CRDGroup + "/" + v1beta2native.CRDVersion,
		Kind:       "ReplicationGroupRAW",
	}
	return nil
}

// ConvertFrom converts from the hub ReplicationGroupRAW v1beta2 to v1beta1.
func (dst *ReplicationGroupRAW) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*v1beta2native.ReplicationGroupRAW)
	if !ok {
		return fmt.Errorf("expected *v1beta2native.ReplicationGroupRAW, got %T", srcRaw)
	}
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return err
	}
	if src.Spec.ForProvider.NumNodeGroups != nil || src.Spec.ForProvider.ReplicasPerNodeGroup != nil {
		dst.Spec.ForProvider.ClusterModeConfiguration = []ClusterModeConfigurationRAWParameters{{
			NumNodeGroups:        src.Spec.ForProvider.NumNodeGroups,
			ReplicasPerNodeGroup: src.Spec.ForProvider.ReplicasPerNodeGroup,
		}}
	}
	dst.TypeMeta = metav1.TypeMeta{
		APIVersion: CRDGroup + "/" + CRDVersion,
		Kind:       "ReplicationGroupRAW",
	}
	return nil
}

// GetForProvider returns the ForProvider parameters.
func (r *ReplicationGroupRAW) GetForProvider() *ReplicationGroupRAWParameters {
	return &r.Spec.ForProvider
}

// GetInitProvider returns the InitProvider parameters.
func (r *ReplicationGroupRAW) GetInitProvider() *ReplicationGroupRAWInitParameters {
	return &r.Spec.InitProvider
}

// GetAtProvider returns the current observed state.
func (r *ReplicationGroupRAW) GetAtProvider() ReplicationGroupRAWObservation {
	return r.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (r *ReplicationGroupRAW) SetAtProvider(o ReplicationGroupRAWObservation) {
	r.Status.AtProvider = o
}
