// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
)

// QueueRAWParameters defines the configuration parameters for a native SQS Queue.
type QueueRAWParameters struct {
	// Enables content-based deduplication for FIFO queues.
	// +kubebuilder:validation:Optional
	ContentBasedDeduplication *bool `json:"contentBasedDeduplication,omitempty"`

	// Specifies whether message deduplication occurs at the message group or
	// queue level. Valid values are messageGroup and queue (default).
	// +kubebuilder:validation:Optional
	DeduplicationScope *string `json:"deduplicationScope,omitempty"`

	// Time in seconds that the delivery of all messages in the queue will be
	// delayed. An integer from 0 to 900 (15 minutes).
	// +kubebuilder:validation:Optional
	DelaySeconds *float64 `json:"delaySeconds,omitempty"`

	// Boolean designating a FIFO queue. If not set, it defaults to false making
	// it standard.
	// +kubebuilder:validation:Optional
	FifoQueue *bool `json:"fifoQueue,omitempty"`

	// Specifies whether the FIFO queue throughput quota applies to the entire
	// queue or per message group. Valid values are perQueue (default) and
	// perMessageGroupId.
	// +kubebuilder:validation:Optional
	FifoThroughputLimit *string `json:"fifoThroughputLimit,omitempty"`

	// Length of time, in seconds, for which Amazon SQS can reuse a data key to
	// encrypt or decrypt messages before calling AWS KMS again.
	// +kubebuilder:validation:Optional
	KMSDataKeyReusePeriodSeconds *float64 `json:"kmsDataKeyReusePeriodSeconds,omitempty"`

	// ID of an AWS-managed customer master key (CMK) for Amazon SQS or a
	// custom CMK.
	// +kubebuilder:validation:Optional
	KMSMasterKeyID *string `json:"kmsMasterKeyId,omitempty"`

	// Limit of how many bytes a message can contain before Amazon SQS rejects
	// it. An integer from 1024 bytes (1 KiB) up to 1048576 bytes (1024 KiB).
	// +kubebuilder:validation:Optional
	MaxMessageSize *float64 `json:"maxMessageSize,omitempty"`

	// Number of seconds Amazon SQS retains a message.
	// +kubebuilder:validation:Optional
	MessageRetentionSeconds *float64 `json:"messageRetentionSeconds,omitempty"`

	// Name of the queue. Queue names must be made up of only uppercase and
	// lowercase ASCII letters, numbers, underscores, and hyphens, and must be
	// between 1 and 80 characters long.
	// +kubebuilder:validation:Optional
	Name *string `json:"name,omitempty"`

	// JSON policy for the SQS queue.
	// +kubebuilder:validation:Optional
	Policy *string `json:"policy,omitempty"`

	// Time for which a ReceiveMessage call will wait for a message to arrive
	// (long polling) before returning. An integer from 0 to 20 (seconds).
	// +kubebuilder:validation:Optional
	ReceiveWaitTimeSeconds *float64 `json:"receiveWaitTimeSeconds,omitempty"`

	// JSON policy to set up the Dead Letter Queue redrive permission.
	// +kubebuilder:validation:Optional
	RedriveAllowPolicy *string `json:"redriveAllowPolicy,omitempty"`

	// JSON policy to set up the Dead Letter Queue.
	// +kubebuilder:validation:Optional
	RedrivePolicy *string `json:"redrivePolicy,omitempty"`

	// Region where this resource will be managed. Required for credential resolution.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// Boolean to enable server-side encryption (SSE) of message content with
	// SQS-owned encryption keys.
	// +kubebuilder:validation:Optional
	SqsManagedSseEnabled *bool `json:"sqsManagedSseEnabled,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// Visibility timeout for the queue. An integer from 0 to 43200 (12 hours).
	// +kubebuilder:validation:Optional
	VisibilityTimeoutSeconds *float64 `json:"visibilityTimeoutSeconds,omitempty"`
}

// QueueRAWInitParameters defines the init parameters for a native SQS Queue.
type QueueRAWInitParameters struct {
	// Enables content-based deduplication for FIFO queues.
	// +kubebuilder:validation:Optional
	ContentBasedDeduplication *bool `json:"contentBasedDeduplication,omitempty"`

	// Specifies whether message deduplication occurs at the message group or queue level.
	// +kubebuilder:validation:Optional
	DeduplicationScope *string `json:"deduplicationScope,omitempty"`

	// Time in seconds that the delivery of all messages in the queue will be delayed.
	// +kubebuilder:validation:Optional
	DelaySeconds *float64 `json:"delaySeconds,omitempty"`

	// Boolean designating a FIFO queue.
	// +kubebuilder:validation:Optional
	FifoQueue *bool `json:"fifoQueue,omitempty"`

	// Specifies whether the FIFO queue throughput quota applies to the entire queue or per message group.
	// +kubebuilder:validation:Optional
	FifoThroughputLimit *string `json:"fifoThroughputLimit,omitempty"`

	// Length of time, in seconds, for which Amazon SQS can reuse a data key.
	// +kubebuilder:validation:Optional
	KMSDataKeyReusePeriodSeconds *float64 `json:"kmsDataKeyReusePeriodSeconds,omitempty"`

	// ID of an AWS-managed customer master key (CMK) for Amazon SQS.
	// +kubebuilder:validation:Optional
	KMSMasterKeyID *string `json:"kmsMasterKeyId,omitempty"`

	// Limit of how many bytes a message can contain before Amazon SQS rejects it.
	// +kubebuilder:validation:Optional
	MaxMessageSize *float64 `json:"maxMessageSize,omitempty"`

	// Number of seconds Amazon SQS retains a message.
	// +kubebuilder:validation:Optional
	MessageRetentionSeconds *float64 `json:"messageRetentionSeconds,omitempty"`

	// Name of the queue.
	// +kubebuilder:validation:Optional
	Name *string `json:"name,omitempty"`

	// JSON policy for the SQS queue.
	// +kubebuilder:validation:Optional
	Policy *string `json:"policy,omitempty"`

	// Time for which a ReceiveMessage call will wait for a message to arrive.
	// +kubebuilder:validation:Optional
	ReceiveWaitTimeSeconds *float64 `json:"receiveWaitTimeSeconds,omitempty"`

	// JSON policy to set up the Dead Letter Queue redrive permission.
	// +kubebuilder:validation:Optional
	RedriveAllowPolicy *string `json:"redriveAllowPolicy,omitempty"`

	// JSON policy to set up the Dead Letter Queue.
	// +kubebuilder:validation:Optional
	RedrivePolicy *string `json:"redrivePolicy,omitempty"`

	// Boolean to enable server-side encryption (SSE) of message content with SQS-owned encryption keys.
	// +kubebuilder:validation:Optional
	SqsManagedSseEnabled *bool `json:"sqsManagedSseEnabled,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// Visibility timeout for the queue. An integer from 0 to 43200 (12 hours).
	// +kubebuilder:validation:Optional
	VisibilityTimeoutSeconds *float64 `json:"visibilityTimeoutSeconds,omitempty"`
}

// QueueRAWObservation defines the observed state of a native SQS Queue.
type QueueRAWObservation struct {
	// ARN of the SQS queue.
	Arn *string `json:"arn,omitempty"`

	// URL for the created Amazon SQS queue (same as ID).
	ID *string `json:"id,omitempty"`

	// URL for the created Amazon SQS queue.
	URL *string `json:"url,omitempty"`

	// Map of tags assigned to the resource, including those inherited from the provider.
	// +mapType=granular
	TagsAll map[string]*string `json:"tagsAll,omitempty"`
}

// QueueRAWSpec defines the desired state of QueueRAW.
//
// NOTE: cluster-scoped QueueRAW implements LegacyManaged (not ModernManaged)
// because the TF Queue example specifies namespace in writeConnectionSecretToRef.
// LegacyManaged uses ConnectionSecretWriterTo (SecretReference with namespace)
// and ProviderConfigReferencer (Reference with just name, no kind).
// The providerConfigRef default is {"name": "default"} matching the TF Queue CRD schema.
type QueueRAWSpec struct {
	// DeletionPolicy specifies what will happen to the underlying external
	// when this managed resource is deleted - either "Delete" or "Orphan" the
	// external resource. This field is planned to be deprecated in favour of
	// the ManagementPolicies field in a future release. Currently, both could be
	// set independently and non-default values would be honored if the feature
	// flag is enabled. See the design doc for more information:
	// https://github.com/crossplane/crossplane/blob/499895a25d1a1a0ba1604944ef98ac7a1a71f197/design/design-doc-observe-only-resources.md?plain=1#L223
	// +optional
	// +kubebuilder:default=Delete
	DeletionPolicy *xpv1.DeletionPolicy `json:"deletionPolicy,omitempty"`

	// WriteConnectionSecretToReference specifies the namespace and name of a
	// Secret to which any connection details for this managed resource should
	// be written. Connection details frequently include the endpoint, username,
	// and password required to connect to the managed resource.
	// +optional
	WriteConnectionSecretToReference *xpv1.SecretReference `json:"writeConnectionSecretToRef,omitempty"`

	// ProviderConfigReference specifies how the provider that will be used to
	// create, observe, update, and delete this managed resource should be
	// configured.
	// +kubebuilder:default={"name": "default"}
	ProviderConfigReference *xpv1.Reference `json:"providerConfigRef,omitempty"`

	// ManagementPolicies specify the array of actions Crossplane is allowed to
	// take on the managed and external resources.
	// +optional
	// +kubebuilder:default={"*"}
	ManagementPolicies xpv1.ManagementPolicies `json:"managementPolicies,omitempty"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider QueueRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider QueueRAWInitParameters `json:"initProvider,omitempty"`
}

// QueueRAWStatus defines the observed state of QueueRAW.
type QueueRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider QueueRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// QueueRAW is the native (non-Terraform) Schema for AWS SQS Queues API.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type QueueRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QueueRAWSpec   `json:"spec"`
	Status QueueRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// QueueRAWList contains a list of QueueRAW resources.
type QueueRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QueueRAW `json:"items"`
}

// Repository type metadata for QueueRAW.
var (
	QueueRAW_Kind             = "QueueRAW"
	QueueRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: QueueRAW_Kind}.String()
	QueueRAW_KindAPIVersion   = QueueRAW_Kind + "." + CRDGroupVersion.String()
	QueueRAW_GroupVersionKind = CRDGroupVersion.WithKind(QueueRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&QueueRAW{}, &QueueRAWList{})
}

// GetForProvider returns the ForProvider parameters.
func (q *QueueRAW) GetForProvider() *QueueRAWParameters { return &q.Spec.ForProvider }

// GetInitProvider returns the InitProvider parameters.
func (q *QueueRAW) GetInitProvider() *QueueRAWInitParameters { return &q.Spec.InitProvider }

// GetAtProvider returns the current observed state.
func (q *QueueRAW) GetAtProvider() QueueRAWObservation { return q.Status.AtProvider }

// SetAtProvider sets the observed state.
func (q *QueueRAW) SetAtProvider(o QueueRAWObservation) { q.Status.AtProvider = o }

// SetForProviderDeduplicationScope sets spec.forProvider.deduplicationScope.
// Required by the QueueCR interface to support late-initialization in both
// cluster and namespaced scopes.
func (q *QueueRAW) SetForProviderDeduplicationScope(v *string) {
	q.Spec.ForProvider.DeduplicationScope = v
}

// SetForProviderFifoThroughputLimit sets spec.forProvider.fifoThroughputLimit.
func (q *QueueRAW) SetForProviderFifoThroughputLimit(v *string) {
	q.Spec.ForProvider.FifoThroughputLimit = v
}

// SetForProviderKMSDataKeyReusePeriodSeconds sets
// spec.forProvider.kmsDataKeyReusePeriodSeconds.
func (q *QueueRAW) SetForProviderKMSDataKeyReusePeriodSeconds(v *float64) {
	q.Spec.ForProvider.KMSDataKeyReusePeriodSeconds = v
}

// SetForProviderSqsManagedSseEnabled sets spec.forProvider.sqsManagedSseEnabled.
func (q *QueueRAW) SetForProviderSqsManagedSseEnabled(v *bool) {
	q.Spec.ForProvider.SqsManagedSseEnabled = v
}
