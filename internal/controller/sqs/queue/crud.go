// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package queue implements the shared CRUD logic for QueueRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the QueueCR interface.
package queue

import (
	"context"
	"errors"
	"strconv"
	"strings"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sqs/v1beta1/native"
	nativehelper "github.com/upbound/provider-aws/v2/internal/native"
)

const (
	errGetAttrs   = "cannot get SQS queue attributes"
	errCreate     = "cannot create SQS queue"
	errUpdate     = "cannot update SQS queue attributes"
	errDelete     = "cannot delete SQS queue"
	errListTags   = "cannot list SQS queue tags"
	errTagQueue   = "cannot tag SQS queue"
	errUntagQueue = "cannot untag SQS queue"
)

// SQSClient is the interface for AWS SQS operations required by this controller.
// Defining it as an interface enables mocking in unit tests.
type SQSClient interface {
	CreateQueue(ctx context.Context, params *awssqs.CreateQueueInput, optFns ...func(*awssqs.Options)) (*awssqs.CreateQueueOutput, error)
	DeleteQueue(ctx context.Context, params *awssqs.DeleteQueueInput, optFns ...func(*awssqs.Options)) (*awssqs.DeleteQueueOutput, error)
	GetQueueAttributes(ctx context.Context, params *awssqs.GetQueueAttributesInput, optFns ...func(*awssqs.Options)) (*awssqs.GetQueueAttributesOutput, error)
	GetQueueUrl(ctx context.Context, params *awssqs.GetQueueUrlInput, optFns ...func(*awssqs.Options)) (*awssqs.GetQueueUrlOutput, error)
	SetQueueAttributes(ctx context.Context, params *awssqs.SetQueueAttributesInput, optFns ...func(*awssqs.Options)) (*awssqs.SetQueueAttributesOutput, error)
	ListQueueTags(ctx context.Context, params *awssqs.ListQueueTagsInput, optFns ...func(*awssqs.Options)) (*awssqs.ListQueueTagsOutput, error)
	TagQueue(ctx context.Context, params *awssqs.TagQueueInput, optFns ...func(*awssqs.Options)) (*awssqs.TagQueueOutput, error)
	UntagQueue(ctx context.Context, params *awssqs.UntagQueueInput, optFns ...func(*awssqs.Options)) (*awssqs.UntagQueueOutput, error)
}

// QueueCR abstracts over cluster-scoped and namespaced QueueRAW types.
// Both scope types implement this interface so that the shared ExternalClient
// can operate on either without scope-specific logic.
type QueueCR interface {
	resource.Managed
	GetForProvider() *clusternative.QueueRAWParameters
	GetInitProvider() *clusternative.QueueRAWInitParameters
	GetAtProvider() clusternative.QueueRAWObservation
	SetAtProvider(clusternative.QueueRAWObservation)
	// SetForProviderDeduplicationScope sets spec.forProvider.deduplicationScope.
	// GetForProvider() for namespaced resources returns a field-copied struct;
	// mutations to it do not propagate back to the spec. These explicit setters
	// are used by late-initialization so that the persisted spec is updated
	// correctly for both cluster and namespaced scope types.
	SetForProviderDeduplicationScope(*string)
	SetForProviderFifoThroughputLimit(*string)
	SetForProviderKMSDataKeyReusePeriodSeconds(*float64)
	SetForProviderSqsManagedSseEnabled(*bool)
}

// ExternalClient implements the shared CRUD logic for Queue resources.
// It is exported so that scope-specific wrapper packages can reference it.
type ExternalClient struct {
	// Client is the AWS SQS SDK client (interface for testability).
	Client SQSClient
	// Kube is the Kubernetes client for reading referenced resources.
	Kube client.Client
}

// Observe checks whether the external Queue resource exists and is up-to-date.
func (e *ExternalClient) Observe(ctx context.Context, cr QueueCR) (managed.ExternalObservation, error) {
	queueURL := nativehelper.GetExternalName(cr)

	// External-name guard: before first creation the annotation holds the K8s
	// resource name, not a URL.  Only proceed when we have a real queue URL.
	if !strings.HasPrefix(queueURL, "https://sqs.") {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	resp, err := e.Client.GetQueueAttributes(ctx, &awssqs.GetQueueAttributesInput{
		QueueUrl:       &queueURL,
		AttributeNames: []sqstypes.QueueAttributeName{sqstypes.QueueAttributeNameAll},
	})
	if err != nil {
		if isQueueDoesNotExist(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, nativehelper.Wrap(err, errGetAttrs)
	}

	attrs := resp.Attributes
	cr.SetAtProvider(mapAttrsToObservation(attrs, queueURL))

	// Late-initialize AWS-defaulted fields.
	lateInited := lateInitialize(cr, attrs)

	cr.SetConditions(xpv1.Available())

	upToDate, err := e.isUpToDate(ctx, cr, attrs)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	// Set the "Test=True" condition when the resource is annotated as a test
	// resource and is fully up-to-date. This allows uptest's
	// --default-conditions="Test" assertion to pass for native controllers.
	nativehelper.SetTestConditionIfAnnotated(cr, upToDate)

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        upToDate,
		ResourceLateInitialized: lateInited,
	}, nil
}

// Create creates the external Queue resource.
func (e *ExternalClient) Create(ctx context.Context, cr QueueCR) (managed.ExternalCreation, error) {
	cr.SetConditions(xpv1.Creating())

	spec := cr.GetForProvider()
	name := queueNameFromCR(cr)

	input := &awssqs.CreateQueueInput{
		QueueName:  &name,
		Attributes: buildCreateAttributes(spec),
	}
	if len(spec.Tags) > 0 {
		input.Tags = specTagsToAWSMap(spec.Tags)
	}

	resp, err := e.Client.CreateQueue(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, nativehelper.Wrap(err, errCreate)
	}

	queueURL := ""
	if resp.QueueUrl != nil {
		queueURL = *resp.QueueUrl
	}

	// Set the queue URL as the external name so subsequent Observe calls
	// can locate the resource via GetQueueAttributes(QueueUrl=...).
	nativehelper.SetExternalName(cr, queueURL)

	obs := cr.GetAtProvider()
	obs.URL = resp.QueueUrl
	obs.ID = resp.QueueUrl
	cr.SetAtProvider(obs)

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{
			"url": []byte(queueURL),
		},
	}, nil
}

// Update updates the external Queue resource.
func (e *ExternalClient) Update(ctx context.Context, cr QueueCR) (managed.ExternalUpdate, error) {
	queueURL := nativehelper.GetExternalName(cr)

	attrs := buildUpdateAttributes(cr.GetForProvider())
	if len(attrs) > 0 {
		if _, err := e.Client.SetQueueAttributes(ctx, &awssqs.SetQueueAttributesInput{
			QueueUrl:   &queueURL,
			Attributes: attrs,
		}); err != nil {
			return managed.ExternalUpdate{}, nativehelper.Wrap(err, errUpdate)
		}
	}

	if err := e.reconcileTags(ctx, queueURL, cr.GetForProvider().Tags); err != nil {
		return managed.ExternalUpdate{}, err
	}

	return managed.ExternalUpdate{}, nil
}

// Delete deletes the external Queue resource.
func (e *ExternalClient) Delete(ctx context.Context, cr QueueCR) (managed.ExternalDelete, error) {
	cr.SetConditions(xpv1.Deleting())

	queueURL := nativehelper.GetExternalName(cr)
	if !strings.HasPrefix(queueURL, "https://sqs.") {
		return managed.ExternalDelete{}, nil
	}

	_, err := e.Client.DeleteQueue(ctx, &awssqs.DeleteQueueInput{
		QueueUrl: &queueURL,
	})
	if err != nil {
		if isQueueDoesNotExist(err) {
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, nativehelper.Wrap(err, errDelete)
	}

	return managed.ExternalDelete{}, nil
}

// ── helpers ────────────────────────────────────────────────────────────────────

// queueNameFromCR returns the SQS queue name to use in CreateQueue.
// Uses spec.forProvider.name when set, otherwise falls back to the K8s name.
func queueNameFromCR(cr QueueCR) string {
	spec := cr.GetForProvider()
	if spec.Name != nil && *spec.Name != "" {
		return *spec.Name
	}
	return cr.GetName()
}

// isQueueDoesNotExist returns true if err is a QueueDoesNotExist error.
func isQueueDoesNotExist(err error) bool {
	var notFound *sqstypes.QueueDoesNotExist
	return errors.As(err, &notFound)
}

// mapAttrsToObservation converts the SQS attribute map to an observation struct.
func mapAttrsToObservation(attrs map[string]string, queueURL string) clusternative.QueueRAWObservation {
	obs := clusternative.QueueRAWObservation{
		URL: &queueURL,
		ID:  &queueURL,
	}
	if arn, ok := attrs["QueueArn"]; ok {
		arn := arn
		obs.Arn = &arn
	}
	return obs
}

// lateInitialize copies AWS-defaulted field values into the spec when those
// fields were not explicitly set.  Returns true if any field was populated.
//
// Note: we use explicit SetForProvider* setters rather than mutating through
// the pointer returned by GetForProvider(). For namespaced resources,
// GetForProvider() returns a field-copied struct (not a direct pointer to
// spec.forProvider), so mutations through that pointer would not persist.
// The setter methods work correctly for both cluster and namespaced scopes.
func lateInitialize(cr QueueCR, attrs map[string]string) bool {
	spec := cr.GetForProvider()

	var changed bool

	if v, ok := lateInitStringVal(spec.DeduplicationScope, attrs, "DeduplicationScope"); ok {
		cr.SetForProviderDeduplicationScope(v)
		changed = true
	}
	if v, ok := lateInitStringVal(spec.FifoThroughputLimit, attrs, "FifoThroughputLimit"); ok {
		cr.SetForProviderFifoThroughputLimit(v)
		changed = true
	}
	if v, ok := lateInitFloat64Val(spec.KMSDataKeyReusePeriodSeconds, attrs, "KmsDataKeyReusePeriodSeconds"); ok {
		cr.SetForProviderKMSDataKeyReusePeriodSeconds(v)
		changed = true
	}
	if v, ok := lateInitBoolVal(spec.SqsManagedSseEnabled, attrs, "SqsManagedSseEnabled"); ok {
		cr.SetForProviderSqsManagedSseEnabled(v)
		changed = true
	}

	return changed
}

// lateInitStringVal returns the string value to late-initialise and true when
// the field is unset and the attribute exists and is non-empty. The returned
// *string is always a freshly allocated copy safe for use as a setter argument.
func lateInitStringVal(current *string, attrs map[string]string, key string) (*string, bool) {
	v, ok := attrs[key]
	if !ok || v == "" || current != nil {
		return nil, false
	}
	copy := v
	return &copy, true
}

// lateInitFloat64Val returns the float64 value to late-initialise and true when
// the field is unset and the attribute exists and is non-empty.
func lateInitFloat64Val(current *float64, attrs map[string]string, key string) (*float64, bool) {
	v, ok := attrs[key]
	if !ok || v == "" || current != nil {
		return nil, false
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil, false
	}
	return &f, true
}

// lateInitBoolVal returns the bool value to late-initialise and true when the
// field is unset and the attribute exists and is non-empty.
func lateInitBoolVal(current *bool, attrs map[string]string, key string) (*bool, bool) {
	v, ok := attrs[key]
	if !ok || v == "" || current != nil {
		return nil, false
	}
	b := v == "true"
	return &b, true
}

// isUpToDate compares the desired spec against the observed AWS state.
func (e *ExternalClient) isUpToDate(ctx context.Context, cr QueueCR, attrs map[string]string) (bool, error) {
	spec := cr.GetForProvider()

	if !attrsBasicUpToDate(spec, attrs) {
		return false, nil
	}
	if !attrsEncryptionUpToDate(spec, attrs) {
		return false, nil
	}
	if !attrsFifoUpToDate(spec, attrs) {
		return false, nil
	}

	policyOK, err := attrsPoliciesUpToDate(spec, attrs)
	if err != nil || !policyOK {
		return false, err
	}

	queueURL := nativehelper.GetExternalName(cr)
	tagsResp, err := e.Client.ListQueueTags(ctx, &awssqs.ListQueueTagsInput{
		QueueUrl: &queueURL,
	})
	if err != nil {
		return false, nativehelper.Wrap(err, errListTags)
	}

	toAdd, toRemove := nativehelper.DiffTagsWithDefaults(
		specTagsToNative(spec.Tags),
		awsTagsToNative(tagsResp.Tags),
		nil,
	)
	return len(toAdd) == 0 && len(toRemove) == 0, nil
}

// attrsBasicUpToDate checks basic numeric spec fields against the AWS attribute map.
func attrsBasicUpToDate(spec *clusternative.QueueRAWParameters, attrs map[string]string) bool {
	return attrsMsgRetentionUpToDate(spec, attrs) && attrsTimeoutsUpToDate(spec, attrs)
}

// attrsMsgRetentionUpToDate checks delay, message size, and retention attributes.
func attrsMsgRetentionUpToDate(spec *clusternative.QueueRAWParameters, attrs map[string]string) bool {
	if spec.DelaySeconds != nil && !floatAttrMatches(attrs, "DelaySeconds", *spec.DelaySeconds) {
		return false
	}
	if spec.MaxMessageSize != nil && !floatAttrMatches(attrs, "MaximumMessageSize", *spec.MaxMessageSize) {
		return false
	}
	if spec.MessageRetentionSeconds != nil && !floatAttrMatches(attrs, "MessageRetentionPeriod", *spec.MessageRetentionSeconds) {
		return false
	}
	return true
}

// attrsTimeoutsUpToDate checks receive-wait and visibility timeout attributes.
func attrsTimeoutsUpToDate(spec *clusternative.QueueRAWParameters, attrs map[string]string) bool {
	if spec.ReceiveWaitTimeSeconds != nil && !floatAttrMatches(attrs, "ReceiveMessageWaitTimeSeconds", *spec.ReceiveWaitTimeSeconds) {
		return false
	}
	if spec.VisibilityTimeoutSeconds != nil && !floatAttrMatches(attrs, "VisibilityTimeout", *spec.VisibilityTimeoutSeconds) {
		return false
	}
	return true
}

// attrsEncryptionUpToDate checks encryption-related spec fields.
func attrsEncryptionUpToDate(spec *clusternative.QueueRAWParameters, attrs map[string]string) bool {
	if spec.KMSMasterKeyID != nil && attrs["KmsMasterKeyId"] != *spec.KMSMasterKeyID {
		return false
	}
	if spec.KMSDataKeyReusePeriodSeconds != nil && !floatAttrMatches(attrs, "KmsDataKeyReusePeriodSeconds", *spec.KMSDataKeyReusePeriodSeconds) {
		return false
	}
	if spec.SqsManagedSseEnabled != nil && attrs["SqsManagedSseEnabled"] != boolToStr(*spec.SqsManagedSseEnabled) {
		return false
	}
	return true
}

// attrsFifoUpToDate checks FIFO-related spec fields (excluding FifoQueue which is immutable).
func attrsFifoUpToDate(spec *clusternative.QueueRAWParameters, attrs map[string]string) bool {
	if spec.ContentBasedDeduplication != nil && attrs["ContentBasedDeduplication"] != boolToStr(*spec.ContentBasedDeduplication) {
		return false
	}
	if spec.DeduplicationScope != nil && attrs["DeduplicationScope"] != *spec.DeduplicationScope {
		return false
	}
	if spec.FifoThroughputLimit != nil && attrs["FifoThroughputLimit"] != *spec.FifoThroughputLimit {
		return false
	}
	return true
}

// attrsPoliciesUpToDate checks JSON policy fields.
// The IAM queue Policy field uses semantic equivalence comparison (whitespace,
// key-order agnostic). RedrivePolicy and RedriveAllowPolicy are SQS-specific
// JSON (not IAM policy format) and use simple string comparison.
func attrsPoliciesUpToDate(spec *clusternative.QueueRAWParameters, attrs map[string]string) (bool, error) {
	if spec.Policy != nil && *spec.Policy != "" {
		// Use semantic IAM policy comparison to avoid drift from AWS normalisation.
		if nativehelper.PolicyNeedsUpdate(*spec.Policy, attrs["Policy"]) {
			return false, nil
		}
	} else if spec.Policy != nil && attrs["Policy"] != "" {
		// spec.Policy is set to empty string — treat as clearing the policy.
		return false, nil
	}

	// RedrivePolicy and RedriveAllowPolicy use SQS-specific JSON structures,
	// not IAM policy format. Use direct string comparison.
	if spec.RedrivePolicy != nil && attrs["RedrivePolicy"] != *spec.RedrivePolicy {
		return false, nil
	}

	if spec.RedriveAllowPolicy != nil && attrs["RedriveAllowPolicy"] != *spec.RedriveAllowPolicy {
		return false, nil
	}

	return true, nil
}

// buildCreateAttributes assembles the SQS attribute map for CreateQueue.
// Note: FifoQueue can only be set at creation time (immutable).
func buildCreateAttributes(spec *clusternative.QueueRAWParameters) map[string]string {
	attrs := make(map[string]string)
	setFloatAttr(attrs, "DelaySeconds", spec.DelaySeconds)
	setFloatAttr(attrs, "MaximumMessageSize", spec.MaxMessageSize)
	setFloatAttr(attrs, "MessageRetentionPeriod", spec.MessageRetentionSeconds)
	setFloatAttr(attrs, "ReceiveMessageWaitTimeSeconds", spec.ReceiveWaitTimeSeconds)
	setFloatAttr(attrs, "VisibilityTimeout", spec.VisibilityTimeoutSeconds)
	setFloatAttr(attrs, "KmsDataKeyReusePeriodSeconds", spec.KMSDataKeyReusePeriodSeconds)
	setStrAttr(attrs, "KmsMasterKeyId", spec.KMSMasterKeyID)
	setBoolAttr(attrs, "SqsManagedSseEnabled", spec.SqsManagedSseEnabled)
	setBoolAttr(attrs, "FifoQueue", spec.FifoQueue)
	setBoolAttr(attrs, "ContentBasedDeduplication", spec.ContentBasedDeduplication)
	setStrAttr(attrs, "DeduplicationScope", spec.DeduplicationScope)
	setStrAttr(attrs, "FifoThroughputLimit", spec.FifoThroughputLimit)
	setStrAttr(attrs, "Policy", spec.Policy)
	setStrAttr(attrs, "RedrivePolicy", spec.RedrivePolicy)
	setStrAttr(attrs, "RedriveAllowPolicy", spec.RedriveAllowPolicy)
	return attrs
}

// buildUpdateAttributes assembles the SQS attribute map for SetQueueAttributes.
// FifoQueue is intentionally excluded — it is immutable after queue creation.
func buildUpdateAttributes(spec *clusternative.QueueRAWParameters) map[string]string {
	attrs := make(map[string]string)
	setFloatAttr(attrs, "DelaySeconds", spec.DelaySeconds)
	setFloatAttr(attrs, "MaximumMessageSize", spec.MaxMessageSize)
	setFloatAttr(attrs, "MessageRetentionPeriod", spec.MessageRetentionSeconds)
	setFloatAttr(attrs, "ReceiveMessageWaitTimeSeconds", spec.ReceiveWaitTimeSeconds)
	setFloatAttr(attrs, "VisibilityTimeout", spec.VisibilityTimeoutSeconds)
	setFloatAttr(attrs, "KmsDataKeyReusePeriodSeconds", spec.KMSDataKeyReusePeriodSeconds)
	setStrAttr(attrs, "KmsMasterKeyId", spec.KMSMasterKeyID)
	setBoolAttr(attrs, "SqsManagedSseEnabled", spec.SqsManagedSseEnabled)
	// FifoQueue is OMITTED — immutable after creation.
	setBoolAttr(attrs, "ContentBasedDeduplication", spec.ContentBasedDeduplication)
	setStrAttr(attrs, "DeduplicationScope", spec.DeduplicationScope)
	setStrAttr(attrs, "FifoThroughputLimit", spec.FifoThroughputLimit)
	setStrAttr(attrs, "Policy", spec.Policy)
	setStrAttr(attrs, "RedrivePolicy", spec.RedrivePolicy)
	setStrAttr(attrs, "RedriveAllowPolicy", spec.RedriveAllowPolicy)
	return attrs
}

// reconcileTags synchronises the queue's AWS tags with the spec.
func (e *ExternalClient) reconcileTags(ctx context.Context, queueURL string, specTags map[string]*string) error {
	tagsResp, err := e.Client.ListQueueTags(ctx, &awssqs.ListQueueTagsInput{
		QueueUrl: &queueURL,
	})
	if err != nil {
		return nativehelper.Wrap(err, errListTags)
	}

	toAdd, toRemove := nativehelper.DiffTagsWithDefaults(
		specTagsToNative(specTags),
		awsTagsToNative(tagsResp.Tags),
		nil,
	)
	if err := e.applyTagsAdd(ctx, queueURL, toAdd); err != nil {
		return err
	}
	return e.applyTagsRemove(ctx, queueURL, toRemove)
}

// applyTagsAdd tags the queue with any tags that need to be added or updated.
func (e *ExternalClient) applyTagsAdd(ctx context.Context, queueURL string, toAdd []nativehelper.Tag) error {
	if len(toAdd) == 0 {
		return nil
	}
	addMap := make(map[string]string, len(toAdd))
	for _, t := range toAdd {
		if t.Key != nil && t.Value != nil {
			addMap[*t.Key] = *t.Value
		}
	}
	_, err := e.Client.TagQueue(ctx, &awssqs.TagQueueInput{
		QueueUrl: &queueURL,
		Tags:     addMap,
	})
	return nativehelper.Wrap(err, errTagQueue)
}

// applyTagsRemove removes tags from the queue.
func (e *ExternalClient) applyTagsRemove(ctx context.Context, queueURL string, toRemove []nativehelper.Tag) error {
	if len(toRemove) == 0 {
		return nil
	}
	keys := make([]string, 0, len(toRemove))
	for _, t := range toRemove {
		if t.Key != nil {
			keys = append(keys, *t.Key)
		}
	}
	_, err := e.Client.UntagQueue(ctx, &awssqs.UntagQueueInput{
		QueueUrl: &queueURL,
		TagKeys:  keys,
	})
	return nativehelper.Wrap(err, errUntagQueue)
}

// ── attribute converters ───────────────────────────────────────────────────────

// setFloatAttr sets an SQS attribute from a *float64. SQS expects integer strings.
func setFloatAttr(attrs map[string]string, key string, val *float64) {
	if val == nil {
		return
	}
	attrs[key] = strconv.FormatInt(int64(*val), 10)
}

// setStrAttr sets an SQS attribute from a *string.
func setStrAttr(attrs map[string]string, key string, val *string) {
	if val == nil {
		return
	}
	attrs[key] = *val
}

// setBoolAttr sets an SQS attribute from a *bool as "true" or "false".
func setBoolAttr(attrs map[string]string, key string, val *bool) {
	if val == nil {
		return
	}
	attrs[key] = boolToStr(*val)
}

// boolToStr converts a bool to the string "true" or "false".
func boolToStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// floatAttrMatches returns true when the attribute value in attrs matches the
// given float64 spec value (compared as integer strings).
func floatAttrMatches(attrs map[string]string, key string, val float64) bool {
	obs, ok := attrs[key]
	if !ok {
		return false
	}
	want := strconv.FormatInt(int64(val), 10)
	return obs == want
}

// ── tag converters ─────────────────────────────────────────────────────────────

// specTagsToNative converts map[string]*string spec tags to []native.Tag.
func specTagsToNative(tags map[string]*string) []nativehelper.Tag {
	out := make([]nativehelper.Tag, 0, len(tags))
	for k, v := range tags {
		k, v := k, v
		if v == nil {
			continue
		}
		out = append(out, nativehelper.Tag{Key: &k, Value: v})
	}
	return out
}

// awsTagsToNative converts map[string]string AWS tags to []native.Tag.
func awsTagsToNative(tags map[string]string) []nativehelper.Tag {
	out := make([]nativehelper.Tag, 0, len(tags))
	for k, v := range tags {
		k, v := k, v
		out = append(out, nativehelper.Tag{Key: &k, Value: &v})
	}
	return out
}

// specTagsToAWSMap converts map[string]*string spec tags to map[string]string
// for use with the CreateQueue Tags parameter.
func specTagsToAWSMap(tags map[string]*string) map[string]string {
	out := make(map[string]string, len(tags))
	for k, v := range tags {
		if v == nil {
			continue
		}
		out[k] = *v
	}
	return out
}
