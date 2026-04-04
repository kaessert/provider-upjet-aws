// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sqs/v1beta1/native"
)

// QueueRAWParameters defines the namespaced configuration parameters for a
// native SQS Queue. Fields are identical to the cluster-scoped type; we define
// them locally (rather than embedding the cluster type inline) so that angryjet
// can recognise this package as a v2 managed resource and generate
// zz_generated.resolvers.go for the policy types that live in this package.
// Queue itself has no cross-resource reference annotations so no resolver is
// generated for Queue, but a standalone struct is required for angryjet to
// process the package at all.
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

// QueueRAWInitParameters defines the namespaced init parameters for a native
// SQS Queue. Fields are identical to the cluster-scoped type; defined locally
// for the same reason as QueueRAWParameters above.
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

// QueueRAWSpec defines the desired state of QueueRAW (namespaced scope).
type QueueRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider QueueRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	// +optional
	InitProvider QueueRAWInitParameters `json:"initProvider,omitempty"`
}

// QueueRAWStatus defines the observed state of QueueRAW.
// Note: using xpv1.ResourceStatus (not ConditionedStatus) so that angryjet's
// ManagedV2() matcher recognises this as a v2-style managed resource and
// generates zz_generated.resolvers.go for the policy types in this package.
type QueueRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider clusternative.QueueRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// QueueRAW is the native (non-Terraform) Schema for AWS SQS Queues API (namespaced scope).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}
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

// GetForProvider returns a cluster-scoped QueueRAWParameters populated from
// this namespaced resource's ForProvider fields.
//
// A field-by-field copy is necessary because the namespaced package defines its
// own QueueRAWParameters struct (for angryjet compatibility) while the shared
// CRUD interface expects the cluster-scoped parameter type. Fields are
// value-copied; the returned pointer is a new allocation and mutations to it do
// NOT propagate back to the spec. For late initialization use the explicit
// SetForProvider* setter methods instead.
func (q *QueueRAW) GetForProvider() *clusternative.QueueRAWParameters {
	return &clusternative.QueueRAWParameters{
		ContentBasedDeduplication:    q.Spec.ForProvider.ContentBasedDeduplication,
		DeduplicationScope:           q.Spec.ForProvider.DeduplicationScope,
		DelaySeconds:                 q.Spec.ForProvider.DelaySeconds,
		FifoQueue:                    q.Spec.ForProvider.FifoQueue,
		FifoThroughputLimit:          q.Spec.ForProvider.FifoThroughputLimit,
		KMSDataKeyReusePeriodSeconds: q.Spec.ForProvider.KMSDataKeyReusePeriodSeconds,
		KMSMasterKeyID:               q.Spec.ForProvider.KMSMasterKeyID,
		MaxMessageSize:               q.Spec.ForProvider.MaxMessageSize,
		MessageRetentionSeconds:      q.Spec.ForProvider.MessageRetentionSeconds,
		Name:                         q.Spec.ForProvider.Name,
		Policy:                       q.Spec.ForProvider.Policy,
		ReceiveWaitTimeSeconds:       q.Spec.ForProvider.ReceiveWaitTimeSeconds,
		RedriveAllowPolicy:           q.Spec.ForProvider.RedriveAllowPolicy,
		RedrivePolicy:                q.Spec.ForProvider.RedrivePolicy,
		Region:                       q.Spec.ForProvider.Region,
		SqsManagedSseEnabled:         q.Spec.ForProvider.SqsManagedSseEnabled,
		Tags:                         q.Spec.ForProvider.Tags,
		VisibilityTimeoutSeconds:     q.Spec.ForProvider.VisibilityTimeoutSeconds,
	}
}

// GetInitProvider returns a cluster-scoped QueueRAWInitParameters populated
// from this namespaced resource's InitProvider fields.
// See GetForProvider for rationale on field-by-field copy.
func (q *QueueRAW) GetInitProvider() *clusternative.QueueRAWInitParameters {
	return &clusternative.QueueRAWInitParameters{
		ContentBasedDeduplication:    q.Spec.InitProvider.ContentBasedDeduplication,
		DeduplicationScope:           q.Spec.InitProvider.DeduplicationScope,
		DelaySeconds:                 q.Spec.InitProvider.DelaySeconds,
		FifoQueue:                    q.Spec.InitProvider.FifoQueue,
		FifoThroughputLimit:          q.Spec.InitProvider.FifoThroughputLimit,
		KMSDataKeyReusePeriodSeconds: q.Spec.InitProvider.KMSDataKeyReusePeriodSeconds,
		KMSMasterKeyID:               q.Spec.InitProvider.KMSMasterKeyID,
		MaxMessageSize:               q.Spec.InitProvider.MaxMessageSize,
		MessageRetentionSeconds:      q.Spec.InitProvider.MessageRetentionSeconds,
		Name:                         q.Spec.InitProvider.Name,
		Policy:                       q.Spec.InitProvider.Policy,
		ReceiveWaitTimeSeconds:       q.Spec.InitProvider.ReceiveWaitTimeSeconds,
		RedriveAllowPolicy:           q.Spec.InitProvider.RedriveAllowPolicy,
		RedrivePolicy:                q.Spec.InitProvider.RedrivePolicy,
		SqsManagedSseEnabled:         q.Spec.InitProvider.SqsManagedSseEnabled,
		Tags:                         q.Spec.InitProvider.Tags,
		VisibilityTimeoutSeconds:     q.Spec.InitProvider.VisibilityTimeoutSeconds,
	}
}

// GetAtProvider returns the current observed state.
func (q *QueueRAW) GetAtProvider() clusternative.QueueRAWObservation { return q.Status.AtProvider }

// SetAtProvider sets the observed state.
func (q *QueueRAW) SetAtProvider(o clusternative.QueueRAWObservation) { q.Status.AtProvider = o }

// SetForProvider copies cluster-scoped QueueRAWParameters back to the
// namespaced spec. Required by the QueueCR interface for late-init write-back.
// Namespaced GetForProvider() returns a freshly-allocated copy, so any
// mutations made by the shared CRUD late-init logic must be written back via
// this method to actually persist in the spec.
func (q *QueueRAW) SetForProvider(p clusternative.QueueRAWParameters) {
	q.Spec.ForProvider.ContentBasedDeduplication = p.ContentBasedDeduplication
	q.Spec.ForProvider.DeduplicationScope = p.DeduplicationScope
	q.Spec.ForProvider.DelaySeconds = p.DelaySeconds
	q.Spec.ForProvider.FifoQueue = p.FifoQueue
	q.Spec.ForProvider.FifoThroughputLimit = p.FifoThroughputLimit
	q.Spec.ForProvider.KMSDataKeyReusePeriodSeconds = p.KMSDataKeyReusePeriodSeconds
	q.Spec.ForProvider.KMSMasterKeyID = p.KMSMasterKeyID
	q.Spec.ForProvider.MaxMessageSize = p.MaxMessageSize
	q.Spec.ForProvider.MessageRetentionSeconds = p.MessageRetentionSeconds
	q.Spec.ForProvider.Name = p.Name
	q.Spec.ForProvider.Policy = p.Policy
	q.Spec.ForProvider.ReceiveWaitTimeSeconds = p.ReceiveWaitTimeSeconds
	q.Spec.ForProvider.RedriveAllowPolicy = p.RedriveAllowPolicy
	q.Spec.ForProvider.RedrivePolicy = p.RedrivePolicy
	q.Spec.ForProvider.Region = p.Region
	q.Spec.ForProvider.SqsManagedSseEnabled = p.SqsManagedSseEnabled
	q.Spec.ForProvider.Tags = p.Tags
	q.Spec.ForProvider.VisibilityTimeoutSeconds = p.VisibilityTimeoutSeconds
}

// SetForProviderDeduplicationScope sets spec.forProvider.deduplicationScope.
// Called by the shared CRUD late-initialization logic, which cannot mutate the
// spec through GetForProvider() (since that returns a field-copied struct, not
// a direct pointer to spec.forProvider). This setter persists late-initialized
// values in the namespaced resource's spec.
func (q *QueueRAW) SetForProviderDeduplicationScope(v *string) {
	q.Spec.ForProvider.DeduplicationScope = v
}

// SetForProviderFifoThroughputLimit sets spec.forProvider.fifoThroughputLimit.
// See SetForProviderDeduplicationScope for rationale.
func (q *QueueRAW) SetForProviderFifoThroughputLimit(v *string) {
	q.Spec.ForProvider.FifoThroughputLimit = v
}

// SetForProviderKMSDataKeyReusePeriodSeconds sets
// spec.forProvider.kmsDataKeyReusePeriodSeconds.
// See SetForProviderDeduplicationScope for rationale.
func (q *QueueRAW) SetForProviderKMSDataKeyReusePeriodSeconds(v *float64) {
	q.Spec.ForProvider.KMSDataKeyReusePeriodSeconds = v
}

// SetForProviderSqsManagedSseEnabled sets spec.forProvider.sqsManagedSseEnabled.
// See SetForProviderDeduplicationScope for rationale.
func (q *QueueRAW) SetForProviderSqsManagedSseEnabled(v *bool) {
	q.Spec.ForProvider.SqsManagedSseEnabled = v
}
