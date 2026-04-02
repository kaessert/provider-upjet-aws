// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package topic implements the shared CRUD logic for TopicRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the TopicCR interface.
package topic

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sns/v1beta1/native"
	nativehelper "github.com/upbound/provider-aws/v2/internal/native"
)

const (
	errGetAttrs      = "cannot get SNS topic attributes"
	errCreate        = "cannot create SNS topic"
	errUpdate        = "cannot set SNS topic attribute"
	errDelete        = "cannot delete SNS topic"
	errListTags      = "cannot list SNS topic tags"
	errTagResource   = "cannot tag SNS topic"
	errUntagResource = "cannot untag SNS topic"

	// attrTrue / attrFalse are the string representations of booleans used
	// by the SNS attribute-based API.
	attrTrue  = "true"
	attrFalse = "false"
)

// SNSClient is the interface for AWS SNS operations required by this controller.
// Defining it as an interface enables mocking in unit tests.
type SNSClient interface {
	CreateTopic(ctx context.Context, params *awssns.CreateTopicInput, optFns ...func(*awssns.Options)) (*awssns.CreateTopicOutput, error)
	GetTopicAttributes(ctx context.Context, params *awssns.GetTopicAttributesInput, optFns ...func(*awssns.Options)) (*awssns.GetTopicAttributesOutput, error)
	SetTopicAttributes(ctx context.Context, params *awssns.SetTopicAttributesInput, optFns ...func(*awssns.Options)) (*awssns.SetTopicAttributesOutput, error)
	DeleteTopic(ctx context.Context, params *awssns.DeleteTopicInput, optFns ...func(*awssns.Options)) (*awssns.DeleteTopicOutput, error)
	ListTagsForResource(ctx context.Context, params *awssns.ListTagsForResourceInput, optFns ...func(*awssns.Options)) (*awssns.ListTagsForResourceOutput, error)
	TagResource(ctx context.Context, params *awssns.TagResourceInput, optFns ...func(*awssns.Options)) (*awssns.TagResourceOutput, error)
	UntagResource(ctx context.Context, params *awssns.UntagResourceInput, optFns ...func(*awssns.Options)) (*awssns.UntagResourceOutput, error)
}

// TopicCR abstracts over cluster-scoped and namespaced TopicRAW types.
// Both scope types implement this interface so that the shared ExternalClient
// can operate on either without scope-specific logic.
type TopicCR interface {
	resource.Managed
	GetForProvider() *clusternative.TopicRAWParameters
	GetInitProvider() *clusternative.TopicRAWInitParameters
	GetAtProvider() clusternative.TopicRAWObservation
	SetAtProvider(clusternative.TopicRAWObservation)
	// SetForProvider* setters are needed for late-initialization of AWS-defaulted
	// fields. GetForProvider() returns a field-copied struct for namespaced scope,
	// so mutations through the returned pointer do not propagate back to the spec.
	SetForProviderFifoThroughputScope(*string)
	SetForProviderSignatureVersion(*float64)
	SetForProviderTracingConfig(*string)
}

// ExternalClient implements the shared CRUD logic for Topic resources.
// It is exported so that scope-specific wrapper packages can reference it.
type ExternalClient struct {
	// Client is the AWS SNS SDK client (interface for testability).
	Client SNSClient
	// Kube is the Kubernetes client for reading referenced resources.
	Kube client.Client
}

// Observe checks whether the external Topic resource exists and is up-to-date.
func (e *ExternalClient) Observe(ctx context.Context, cr TopicCR) (managed.ExternalObservation, error) {
	arn := topicARN(cr)
	if arn == "" {
		// No ARN stored — resource hasn't been created yet.
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	resp, err := e.Client.GetTopicAttributes(ctx, &awssns.GetTopicAttributesInput{
		TopicArn: &arn,
	})
	if err != nil {
		if isTopicNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, nativehelper.Wrap(err, errGetAttrs)
	}

	attrs := resp.Attributes

	// Populate status.atProvider from the attributes.
	cr.SetAtProvider(mapAttrsToObservation(attrs, cr.GetAtProvider()))

	// Late-initialize AWS-defaulted fields.
	lateInited := e.lateInitialize(cr, attrs)

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

// Create creates the external Topic resource.
func (e *ExternalClient) Create(ctx context.Context, cr TopicCR) (managed.ExternalCreation, error) {
	cr.SetConditions(xpv1.Creating())

	topicName := topicNameFromCR(cr)
	spec := cr.GetForProvider()

	input := &awssns.CreateTopicInput{
		Name:       &topicName,
		Attributes: buildCreateAttributes(spec),
	}
	// Pass tags at creation time.
	if len(spec.Tags) > 0 {
		input.Tags = specTagsToSNSTags(spec.Tags)
	}

	resp, err := e.Client.CreateTopic(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, nativehelper.Wrap(err, errCreate)
	}

	topicArn := ""
	if resp.TopicArn != nil {
		topicArn = *resp.TopicArn
	}

	// Store the full ARN as the external-name annotation.  The external-name
	// annotation is persisted atomically by the managed reconciler immediately
	// after Create returns, unlike status.atProvider fields.  By storing the
	// ARN here we guarantee that topicARN() can reconstruct it on the very
	// next Observe call even when status has not been flushed yet.
	if topicArn != "" {
		meta.SetExternalName(cr, topicArn)
	} else {
		// Fallback (unexpected): store the topic name so we at least have
		// something in the annotation.
		name := nameFromARN(topicArn)
		if name == "" {
			name = topicName
		}
		meta.SetExternalName(cr, name)
	}

	// Store the full ARN in atProvider as well (best-effort; the reconciler
	// may not flush status before the next Observe, but we store it anyway).
	obs := cr.GetAtProvider()
	obs.Arn = &topicArn
	obs.ID = &topicArn
	cr.SetAtProvider(obs)

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{
			"arn": []byte(topicArn),
		},
	}, nil
}

// Update updates the external Topic resource.
func (e *ExternalClient) Update(ctx context.Context, cr TopicCR) (managed.ExternalUpdate, error) {
	arn := topicARN(cr)
	if arn == "" {
		return managed.ExternalUpdate{}, nil
	}

	// Fetch current attributes to compare against spec.
	resp, err := e.Client.GetTopicAttributes(ctx, &awssns.GetTopicAttributesInput{
		TopicArn: &arn,
	})
	if err != nil {
		return managed.ExternalUpdate{}, nativehelper.Wrap(err, errGetAttrs)
	}

	attrs := resp.Attributes
	spec := cr.GetForProvider()

	// Build a list of attributes that need to be changed.
	updates := buildUpdateAttributes(spec, attrs)

	// Call SetTopicAttributes once per attribute (SNS API sets one at a time).
	for attrName, attrValue := range updates {
		name := attrName
		val := attrValue
		if _, err := e.Client.SetTopicAttributes(ctx, &awssns.SetTopicAttributesInput{
			TopicArn:       &arn,
			AttributeName:  &name,
			AttributeValue: &val,
		}); err != nil {
			return managed.ExternalUpdate{}, nativehelper.Wrap(err,
				fmt.Sprintf("%s (attribute: %s)", errUpdate, attrName))
		}
	}

	// Reconcile tags.
	if err := e.reconcileTags(ctx, arn, spec.Tags); err != nil {
		return managed.ExternalUpdate{}, err
	}

	return managed.ExternalUpdate{}, nil
}

// Delete deletes the external Topic resource.
func (e *ExternalClient) Delete(ctx context.Context, cr TopicCR) (managed.ExternalDelete, error) {
	cr.SetConditions(xpv1.Deleting())

	arn := topicARN(cr)
	if arn == "" {
		// Nothing to delete — resource was never created.
		return managed.ExternalDelete{}, nil
	}

	_, err := e.Client.DeleteTopic(ctx, &awssns.DeleteTopicInput{
		TopicArn: &arn,
	})
	if err != nil {
		if isTopicNotFound(err) {
			// Already gone — idempotent success.
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, nativehelper.Wrap(err, errDelete)
	}

	return managed.ExternalDelete{}, nil
}

// ── helpers ────────────────────────────────────────────────────────────────────

// topicARN returns the SNS topic ARN to use in API calls.
// It prefers the ARN stored in atProvider.Arn; if not set, it checks whether
// the external name annotation looks like a full ARN.
func topicARN(cr TopicCR) string {
	obs := cr.GetAtProvider()
	if obs.Arn != nil && *obs.Arn != "" {
		return *obs.Arn
	}
	extName := meta.GetExternalName(cr)
	if strings.HasPrefix(extName, "arn:aws:sns:") {
		return extName
	}
	return ""
}

// topicNameFromCR returns the topic name to use in CreateTopic.
// Uses the external name annotation if it's distinct from the K8s name,
// otherwise falls back to the K8s metadata name.  If the external name is
// stored as a full ARN (e.g., after a successful Create), the topic name
// is extracted from the last segment of the ARN.
func topicNameFromCR(cr TopicCR) string {
	extName := meta.GetExternalName(cr)
	if extName != "" && extName != cr.GetName() {
		// If the external name is a full SNS ARN, extract just the topic name.
		if strings.HasPrefix(extName, "arn:aws:sns:") {
			return nameFromARN(extName)
		}
		return extName
	}
	return cr.GetName()
}

// nameFromARN extracts the topic name (last segment after ":") from a full ARN.
func nameFromARN(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

// isTopicNotFound returns true if the error indicates the SNS topic does not exist.
func isTopicNotFound(err error) bool {
	var notFound *snstypes.NotFoundException
	if errors.As(err, &notFound) {
		return true
	}
	return nativehelper.IsNotFound(err)
}

// mapAttrsToObservation converts the GetTopicAttributes response map to an observation struct.
// It uses helper functions to avoid high cyclomatic complexity.
func mapAttrsToObservation(attrs map[string]string, existing clusternative.TopicRAWObservation) clusternative.TopicRAWObservation {
	obs := existing
	mapCoreObsAttrs(attrs, &obs)
	mapFifoObsAttrs(attrs, &obs)
	mapFeedbackObsAttrs(attrs, &obs)
	return obs
}

// mapCoreObsAttrs populates core/basic observation fields from the attribute map.
func mapCoreObsAttrs(attrs map[string]string, obs *clusternative.TopicRAWObservation) {
	setStrObs(&obs.Arn, attrs, "TopicArn")
	setStrObs(&obs.ID, attrs, "TopicArn")
	setStrObs(&obs.Owner, attrs, "Owner")
	setStrObs(&obs.DisplayName, attrs, "DisplayName")
	setStrObs(&obs.DeliveryPolicy, attrs, "DeliveryPolicy")
	setStrObs(&obs.Policy, attrs, "Policy")
	setStrObs(&obs.KMSMasterKeyID, attrs, "KmsMasterKeyId")
	setStrObs(&obs.TracingConfig, attrs, "TracingConfig")
	setStrObs(&obs.ArchivePolicy, attrs, "ArchivePolicy")
	setStrObs(&obs.BeginningArchiveTime, attrs, "BeginningArchiveTime")
}

// mapFifoObsAttrs populates FIFO-specific observation fields from the attribute map.
func mapFifoObsAttrs(attrs map[string]string, obs *clusternative.TopicRAWObservation) {
	setStrObs(&obs.FifoThroughputScope, attrs, "FifoThroughputScope")
	setBoolObs(&obs.FifoTopic, attrs, "FifoTopic")
	setBoolObs(&obs.ContentBasedDeduplication, attrs, "ContentBasedDeduplication")
	setFloatObs(&obs.SignatureVersion, attrs, "SignatureVersion")
}

// mapFeedbackObsAttrs populates feedback role/rate observation fields.
func mapFeedbackObsAttrs(attrs map[string]string, obs *clusternative.TopicRAWObservation) {
	setStrObs(&obs.HTTPSuccessFeedbackRoleArn, attrs, "HTTPSuccessFeedbackRoleArn")
	setFloatObs(&obs.HTTPSuccessFeedbackSampleRate, attrs, "HTTPSuccessFeedbackSampleRate")
	setStrObs(&obs.HTTPFailureFeedbackRoleArn, attrs, "HTTPFailureFeedbackRoleArn")
	setStrObs(&obs.ApplicationSuccessFeedbackRoleArn, attrs, "ApplicationSuccessFeedbackRoleArn")
	setFloatObs(&obs.ApplicationSuccessFeedbackSampleRate, attrs, "ApplicationSuccessFeedbackSampleRate")
	setStrObs(&obs.ApplicationFailureFeedbackRoleArn, attrs, "ApplicationFailureFeedbackRoleArn")
	setStrObs(&obs.FirehoseSuccessFeedbackRoleArn, attrs, "FirehoseSuccessFeedbackRoleArn")
	setFloatObs(&obs.FirehoseSuccessFeedbackSampleRate, attrs, "FirehoseSuccessFeedbackSampleRate")
	setStrObs(&obs.FirehoseFailureFeedbackRoleArn, attrs, "FirehoseFailureFeedbackRoleArn")
	setStrObs(&obs.LambdaSuccessFeedbackRoleArn, attrs, "LambdaSuccessFeedbackRoleArn")
	setFloatObs(&obs.LambdaSuccessFeedbackSampleRate, attrs, "LambdaSuccessFeedbackSampleRate")
	setStrObs(&obs.LambdaFailureFeedbackRoleArn, attrs, "LambdaFailureFeedbackRoleArn")
	// Note: AWS uses "SQS" prefix in attribute names.
	setStrObs(&obs.SqsSuccessFeedbackRoleArn, attrs, "SQSSuccessFeedbackRoleArn")
	setFloatObs(&obs.SqsSuccessFeedbackSampleRate, attrs, "SQSSuccessFeedbackSampleRate")
	setStrObs(&obs.SqsFailureFeedbackRoleArn, attrs, "SQSFailureFeedbackRoleArn")
}

// setStrObs sets *dst from attrs[key] when the key exists in the map.
func setStrObs(dst **string, attrs map[string]string, key string) {
	if v, ok := attrs[key]; ok {
		v := v
		*dst = &v
	}
}

// setBoolObs sets *dst from attrs[key] parsed as a bool ("true"/"false").
func setBoolObs(dst **bool, attrs map[string]string, key string) {
	if v, ok := attrs[key]; ok {
		b := v == attrTrue
		*dst = &b
	}
}

// setFloatObs sets *dst from attrs[key] parsed as a float64.
func setFloatObs(dst **float64, attrs map[string]string, key string) {
	if v, ok := attrs[key]; ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			*dst = &f
		}
	}
}

// lateInitialize copies AWS-defaulted field values into the spec when those
// fields were not explicitly set. Returns true if any field was populated.
//
// Uses SetForProvider* setters rather than mutating through GetForProvider()
// because the namespaced GetForProvider() returns a field-copied struct.
func (e *ExternalClient) lateInitialize(cr TopicCR, attrs map[string]string) bool {
	spec := cr.GetForProvider()
	changed := false

	if v, ok := lateInitStringVal(spec.TracingConfig, attrs, "TracingConfig"); ok {
		cr.SetForProviderTracingConfig(v)
		changed = true
	}
	if v, ok := lateInitStringVal(spec.FifoThroughputScope, attrs, "FifoThroughputScope"); ok {
		cr.SetForProviderFifoThroughputScope(v)
		changed = true
	}
	if v, ok := lateInitSignatureVersion(spec.SignatureVersion, attrs); ok {
		cr.SetForProviderSignatureVersion(v)
		changed = true
	}

	return changed
}

// lateInitStringVal returns the string value to late-initialise and true when
// the field is unset and the attribute exists and is non-empty.
func lateInitStringVal(current *string, attrs map[string]string, key string) (*string, bool) {
	v, ok := attrs[key]
	if !ok || v == "" || current != nil {
		return nil, false
	}
	copy := v
	return &copy, true
}

// lateInitSignatureVersion converts the "SignatureVersion" attribute string
// ("1" or "2") to *float64 for late-init.
func lateInitSignatureVersion(current *float64, attrs map[string]string) (*float64, bool) {
	v, ok := attrs["SignatureVersion"]
	if !ok || v == "" || current != nil {
		return nil, false
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return nil, false
	}
	return &f, true
}

// isUpToDate compares the desired spec against the observed AWS state.
// Returns (true, nil) when the resource matches the spec.
func (e *ExternalClient) isUpToDate(ctx context.Context, cr TopicCR, attrs map[string]string) (bool, error) {
	spec := cr.GetForProvider()

	if !basicAttrsUpToDate(spec, attrs) {
		return false, nil
	}

	policyOK, err := policyUpToDate(spec, attrs)
	if err != nil || !policyOK {
		return false, err
	}

	if !feedbackRoleAttrsUpToDate(spec, attrs) {
		return false, nil
	}

	if !sampleRateAttrsUpToDate(spec, attrs) {
		return false, nil
	}

	// Tags comparison requires a separate API call.
	arn := topicARN(cr)
	tagsResp, err := e.Client.ListTagsForResource(ctx, &awssns.ListTagsForResourceInput{
		ResourceArn: &arn,
	})
	if err != nil {
		return false, nativehelper.Wrap(err, errListTags)
	}

	toAdd, toRemove := nativehelper.DiffTagsWithDefaults(
		specTagsToNative(spec.Tags),
		snsTagsToNative(tagsResp.Tags),
		nil,
	)
	return len(toAdd) == 0 && len(toRemove) == 0, nil
}

// basicAttrsUpToDate checks string, bool and numeric attributes that can be
// compared without special semantics (excludes policy and tags).
func basicAttrsUpToDate(spec *clusternative.TopicRAWParameters, attrs map[string]string) bool {
	return basicStringAttrsUpToDate(spec, attrs) &&
		basicBoolAttrsUpToDate(spec, attrs) &&
		basicNumericAttrsUpToDate(spec, attrs)
}

// basicStringAttrsUpToDate checks basic string attributes.
func basicStringAttrsUpToDate(spec *clusternative.TopicRAWParameters, attrs map[string]string) bool {
	return strAttrUpToDate(spec.DisplayName, attrs, "DisplayName") &&
		strAttrUpToDate(spec.DeliveryPolicy, attrs, "DeliveryPolicy") &&
		strAttrUpToDate(spec.TracingConfig, attrs, "TracingConfig") &&
		strAttrUpToDate(spec.KMSMasterKeyID, attrs, "KmsMasterKeyId") &&
		strAttrUpToDate(spec.ArchivePolicy, attrs, "ArchivePolicy") &&
		strAttrUpToDate(spec.FifoThroughputScope, attrs, "FifoThroughputScope")
}

// basicBoolAttrsUpToDate checks boolean attributes.
func basicBoolAttrsUpToDate(spec *clusternative.TopicRAWParameters, attrs map[string]string) bool {
	if spec.ContentBasedDeduplication == nil {
		return true
	}
	obsVal := attrs["ContentBasedDeduplication"] == attrTrue
	return *spec.ContentBasedDeduplication == obsVal
}

// basicNumericAttrsUpToDate checks numeric attributes stored as integer strings.
func basicNumericAttrsUpToDate(spec *clusternative.TopicRAWParameters, attrs map[string]string) bool {
	if spec.SignatureVersion == nil {
		return true
	}
	want := strconv.FormatInt(int64(*spec.SignatureVersion), 10)
	return attrs["SignatureVersion"] == want
}

// strAttrUpToDate returns true when the spec *string is nil (unmanaged) or
// equals the observed AWS attribute value.
func strAttrUpToDate(spec *string, attrs map[string]string, key string) bool {
	if spec == nil {
		return true
	}
	return attrs[key] == *spec
}

// policyUpToDate checks the IAM policy field using semantic equivalence.
// Policy is a special case: AWS normalises the JSON, so we use
// awspolicyequivalence to avoid spurious updates.
// Per config, policy is excluded from late-init (TopicPolicy manages it separately).
func policyUpToDate(spec *clusternative.TopicRAWParameters, attrs map[string]string) (bool, error) {
	if spec.Policy == nil {
		// spec.Policy not set — do not manage it (TopicPolicy owns it).
		return true, nil
	}
	if *spec.Policy == "" {
		// Empty string means the caller wants to clear the policy.
		if attrs["Policy"] != "" {
			return false, nil
		}
		return true, nil
	}
	if nativehelper.PolicyNeedsUpdate(*spec.Policy, attrs["Policy"]) {
		return false, nil
	}
	return true, nil
}

// feedbackRoleAttrsUpToDate checks the 10 feedback role ARN fields.
// Split across two helpers to keep per-function complexity below the lint threshold.
func feedbackRoleAttrsUpToDate(spec *clusternative.TopicRAWParameters, attrs map[string]string) bool {
	return httpAndAppFeedbackRoleAttrsUpToDate(spec, attrs) &&
		firehoseLambdaSQSFeedbackRoleAttrsUpToDate(spec, attrs)
}

// httpAndAppFeedbackRoleAttrsUpToDate checks HTTP and Application feedback role ARNs.
func httpAndAppFeedbackRoleAttrsUpToDate(spec *clusternative.TopicRAWParameters, attrs map[string]string) bool {
	return strAttrUpToDate(spec.HTTPSuccessFeedbackRoleArn, attrs, "HTTPSuccessFeedbackRoleArn") &&
		strAttrUpToDate(spec.HTTPFailureFeedbackRoleArn, attrs, "HTTPFailureFeedbackRoleArn") &&
		strAttrUpToDate(spec.ApplicationSuccessFeedbackRoleArn, attrs, "ApplicationSuccessFeedbackRoleArn") &&
		strAttrUpToDate(spec.ApplicationFailureFeedbackRoleArn, attrs, "ApplicationFailureFeedbackRoleArn")
}

// firehoseLambdaSQSFeedbackRoleAttrsUpToDate checks Firehose, Lambda, and SQS feedback role ARNs.
func firehoseLambdaSQSFeedbackRoleAttrsUpToDate(spec *clusternative.TopicRAWParameters, attrs map[string]string) bool {
	return strAttrUpToDate(spec.FirehoseSuccessFeedbackRoleArn, attrs, "FirehoseSuccessFeedbackRoleArn") &&
		strAttrUpToDate(spec.FirehoseFailureFeedbackRoleArn, attrs, "FirehoseFailureFeedbackRoleArn") &&
		strAttrUpToDate(spec.LambdaSuccessFeedbackRoleArn, attrs, "LambdaSuccessFeedbackRoleArn") &&
		strAttrUpToDate(spec.LambdaFailureFeedbackRoleArn, attrs, "LambdaFailureFeedbackRoleArn") &&
		// AWS uses "SQS" prefix (uppercase) for SQS feedback attributes.
		strAttrUpToDate(spec.SqsSuccessFeedbackRoleArn, attrs, "SQSSuccessFeedbackRoleArn") &&
		strAttrUpToDate(spec.SqsFailureFeedbackRoleArn, attrs, "SQSFailureFeedbackRoleArn")
}

// sampleRateAttrsUpToDate checks the 5 feedback sample rate fields.
func sampleRateAttrsUpToDate(spec *clusternative.TopicRAWParameters, attrs map[string]string) bool {
	return floatAttrMatches(spec.HTTPSuccessFeedbackSampleRate, attrs, "HTTPSuccessFeedbackSampleRate") &&
		floatAttrMatches(spec.ApplicationSuccessFeedbackSampleRate, attrs, "ApplicationSuccessFeedbackSampleRate") &&
		floatAttrMatches(spec.FirehoseSuccessFeedbackSampleRate, attrs, "FirehoseSuccessFeedbackSampleRate") &&
		floatAttrMatches(spec.LambdaSuccessFeedbackSampleRate, attrs, "LambdaSuccessFeedbackSampleRate") &&
		floatAttrMatches(spec.SqsSuccessFeedbackSampleRate, attrs, "SQSSuccessFeedbackSampleRate")
}

// floatAttrMatches returns false (indicating drift) when the spec value is set
// and does not match the integer-string AWS attribute value. Returns true when
// the spec value is nil (not managed).
func floatAttrMatches(specVal *float64, attrs map[string]string, key string) bool {
	if specVal == nil {
		return true
	}
	want := strconv.FormatInt(int64(*specVal), 10)
	return attrs[key] == want
}

// buildCreateAttributes assembles the SNS CreateTopic attribute map from spec.
// FifoTopic is included because it is immutable and must be set at creation.
func buildCreateAttributes(spec *clusternative.TopicRAWParameters) map[string]string {
	attrs := make(map[string]string)
	setStrAttr(attrs, "DisplayName", spec.DisplayName)
	setStrAttr(attrs, "DeliveryPolicy", spec.DeliveryPolicy)
	setStrAttr(attrs, "Policy", spec.Policy)
	setStrAttr(attrs, "TracingConfig", spec.TracingConfig)
	setStrAttr(attrs, "KmsMasterKeyId", spec.KMSMasterKeyID)
	setStrAttr(attrs, "ArchivePolicy", spec.ArchivePolicy)
	setStrAttr(attrs, "FifoThroughputScope", spec.FifoThroughputScope)
	setBoolAttr(attrs, "FifoTopic", spec.FifoTopic) // immutable; set at creation only
	setBoolAttr(attrs, "ContentBasedDeduplication", spec.ContentBasedDeduplication)
	setFloatAsIntAttr(attrs, "SignatureVersion", spec.SignatureVersion)
	setFeedbackAttrs(attrs, spec)
	return attrs
}

// setFeedbackAttrs adds all feedback role and sample-rate attributes to attrs.
func setFeedbackAttrs(attrs map[string]string, spec *clusternative.TopicRAWParameters) {
	setStrAttr(attrs, "HTTPSuccessFeedbackRoleArn", spec.HTTPSuccessFeedbackRoleArn)
	setStrAttr(attrs, "HTTPFailureFeedbackRoleArn", spec.HTTPFailureFeedbackRoleArn)
	setFloatAsIntAttr(attrs, "HTTPSuccessFeedbackSampleRate", spec.HTTPSuccessFeedbackSampleRate)
	setStrAttr(attrs, "ApplicationSuccessFeedbackRoleArn", spec.ApplicationSuccessFeedbackRoleArn)
	setStrAttr(attrs, "ApplicationFailureFeedbackRoleArn", spec.ApplicationFailureFeedbackRoleArn)
	setFloatAsIntAttr(attrs, "ApplicationSuccessFeedbackSampleRate", spec.ApplicationSuccessFeedbackSampleRate)
	setStrAttr(attrs, "FirehoseSuccessFeedbackRoleArn", spec.FirehoseSuccessFeedbackRoleArn)
	setStrAttr(attrs, "FirehoseFailureFeedbackRoleArn", spec.FirehoseFailureFeedbackRoleArn)
	setFloatAsIntAttr(attrs, "FirehoseSuccessFeedbackSampleRate", spec.FirehoseSuccessFeedbackSampleRate)
	setStrAttr(attrs, "LambdaSuccessFeedbackRoleArn", spec.LambdaSuccessFeedbackRoleArn)
	setStrAttr(attrs, "LambdaFailureFeedbackRoleArn", spec.LambdaFailureFeedbackRoleArn)
	setFloatAsIntAttr(attrs, "LambdaSuccessFeedbackSampleRate", spec.LambdaSuccessFeedbackSampleRate)
	// AWS attribute names use "SQS" prefix.
	setStrAttr(attrs, "SQSSuccessFeedbackRoleArn", spec.SqsSuccessFeedbackRoleArn)
	setStrAttr(attrs, "SQSFailureFeedbackRoleArn", spec.SqsFailureFeedbackRoleArn)
	setFloatAsIntAttr(attrs, "SQSSuccessFeedbackSampleRate", spec.SqsSuccessFeedbackSampleRate)
}

// buildUpdateAttributes returns a map of attribute name → desired value for
// attributes that differ from the observed AWS state. FifoTopic is excluded
// because it is immutable after creation.
func buildUpdateAttributes(spec *clusternative.TopicRAWParameters, attrs map[string]string) map[string]string {
	updates := make(map[string]string)
	buildBasicUpdates(spec, attrs, updates)
	buildFeedbackUpdates(spec, attrs, updates)
	return updates
}

// buildBasicUpdates adds basic (non-feedback) changed attributes to updates.
func buildBasicUpdates(spec *clusternative.TopicRAWParameters, attrs map[string]string, updates map[string]string) {
	maybeSetStr(updates, "DisplayName", spec.DisplayName, attrs["DisplayName"])
	maybeSetStr(updates, "DeliveryPolicy", spec.DeliveryPolicy, attrs["DeliveryPolicy"])
	maybeSetStr(updates, "TracingConfig", spec.TracingConfig, attrs["TracingConfig"])
	maybeSetStr(updates, "KmsMasterKeyId", spec.KMSMasterKeyID, attrs["KmsMasterKeyId"])
	maybeSetStr(updates, "ArchivePolicy", spec.ArchivePolicy, attrs["ArchivePolicy"])
	maybeSetStr(updates, "FifoThroughputScope", spec.FifoThroughputScope, attrs["FifoThroughputScope"])
	maybeSetBool(updates, "ContentBasedDeduplication", spec.ContentBasedDeduplication, attrs["ContentBasedDeduplication"])
	maybeSetFloatAsInt(updates, "SignatureVersion", spec.SignatureVersion, attrs["SignatureVersion"])
	// Policy uses policy-equivalence comparison so must be handled separately.
	if spec.Policy != nil && *spec.Policy != "" && nativehelper.PolicyNeedsUpdate(*spec.Policy, attrs["Policy"]) {
		updates["Policy"] = *spec.Policy
	}
}

// buildFeedbackUpdates adds changed feedback role and sample-rate attributes to updates.
func buildFeedbackUpdates(spec *clusternative.TopicRAWParameters, attrs map[string]string, updates map[string]string) {
	maybeSetStr(updates, "HTTPSuccessFeedbackRoleArn", spec.HTTPSuccessFeedbackRoleArn, attrs["HTTPSuccessFeedbackRoleArn"])
	maybeSetStr(updates, "HTTPFailureFeedbackRoleArn", spec.HTTPFailureFeedbackRoleArn, attrs["HTTPFailureFeedbackRoleArn"])
	maybeSetFloatAsInt(updates, "HTTPSuccessFeedbackSampleRate", spec.HTTPSuccessFeedbackSampleRate, attrs["HTTPSuccessFeedbackSampleRate"])
	maybeSetStr(updates, "ApplicationSuccessFeedbackRoleArn", spec.ApplicationSuccessFeedbackRoleArn, attrs["ApplicationSuccessFeedbackRoleArn"])
	maybeSetStr(updates, "ApplicationFailureFeedbackRoleArn", spec.ApplicationFailureFeedbackRoleArn, attrs["ApplicationFailureFeedbackRoleArn"])
	maybeSetFloatAsInt(updates, "ApplicationSuccessFeedbackSampleRate", spec.ApplicationSuccessFeedbackSampleRate, attrs["ApplicationSuccessFeedbackSampleRate"])
	maybeSetStr(updates, "FirehoseSuccessFeedbackRoleArn", spec.FirehoseSuccessFeedbackRoleArn, attrs["FirehoseSuccessFeedbackRoleArn"])
	maybeSetStr(updates, "FirehoseFailureFeedbackRoleArn", spec.FirehoseFailureFeedbackRoleArn, attrs["FirehoseFailureFeedbackRoleArn"])
	maybeSetFloatAsInt(updates, "FirehoseSuccessFeedbackSampleRate", spec.FirehoseSuccessFeedbackSampleRate, attrs["FirehoseSuccessFeedbackSampleRate"])
	maybeSetStr(updates, "LambdaSuccessFeedbackRoleArn", spec.LambdaSuccessFeedbackRoleArn, attrs["LambdaSuccessFeedbackRoleArn"])
	maybeSetStr(updates, "LambdaFailureFeedbackRoleArn", spec.LambdaFailureFeedbackRoleArn, attrs["LambdaFailureFeedbackRoleArn"])
	maybeSetFloatAsInt(updates, "LambdaSuccessFeedbackSampleRate", spec.LambdaSuccessFeedbackSampleRate, attrs["LambdaSuccessFeedbackSampleRate"])
	maybeSetStr(updates, "SQSSuccessFeedbackRoleArn", spec.SqsSuccessFeedbackRoleArn, attrs["SQSSuccessFeedbackRoleArn"])
	maybeSetStr(updates, "SQSFailureFeedbackRoleArn", spec.SqsFailureFeedbackRoleArn, attrs["SQSFailureFeedbackRoleArn"])
	maybeSetFloatAsInt(updates, "SQSSuccessFeedbackSampleRate", spec.SqsSuccessFeedbackSampleRate, attrs["SQSSuccessFeedbackSampleRate"])
}

// reconcileTags synchronises the topic's AWS tags with the spec.
func (e *ExternalClient) reconcileTags(ctx context.Context, arn string, specTags map[string]*string) error {
	tagsResp, err := e.Client.ListTagsForResource(ctx, &awssns.ListTagsForResourceInput{
		ResourceArn: &arn,
	})
	if err != nil {
		return nativehelper.Wrap(err, errListTags)
	}

	toAdd, toRemove := nativehelper.DiffTagsWithDefaults(
		specTagsToNative(specTags),
		snsTagsToNative(tagsResp.Tags),
		nil,
	)

	if err := e.applyTagsAdd(ctx, arn, toAdd); err != nil {
		return err
	}
	return e.applyTagsRemove(ctx, arn, toRemove)
}

// applyTagsAdd tags the SNS topic with any tags that need to be added or updated.
func (e *ExternalClient) applyTagsAdd(ctx context.Context, arn string, toAdd []nativehelper.Tag) error {
	if len(toAdd) == 0 {
		return nil
	}
	snsTags := make([]snstypes.Tag, 0, len(toAdd))
	for _, t := range toAdd {
		t := t
		if t.Key != nil && t.Value != nil {
			snsTags = append(snsTags, snstypes.Tag{Key: t.Key, Value: t.Value})
		}
	}
	_, err := e.Client.TagResource(ctx, &awssns.TagResourceInput{
		ResourceArn: &arn,
		Tags:        snsTags,
	})
	return nativehelper.Wrap(err, errTagResource)
}

// applyTagsRemove removes tags from the SNS topic.
func (e *ExternalClient) applyTagsRemove(ctx context.Context, arn string, toRemove []nativehelper.Tag) error {
	if len(toRemove) == 0 {
		return nil
	}
	keys := make([]string, 0, len(toRemove))
	for _, t := range toRemove {
		if t.Key != nil {
			keys = append(keys, *t.Key)
		}
	}
	_, err := e.Client.UntagResource(ctx, &awssns.UntagResourceInput{
		ResourceArn: &arn,
		TagKeys:     keys,
	})
	return nativehelper.Wrap(err, errUntagResource)
}

// ── attribute set helpers ──────────────────────────────────────────────────────

// setStrAttr sets an attribute key to a string value when the pointer is non-nil.
func setStrAttr(attrs map[string]string, key string, val *string) {
	if val == nil {
		return
	}
	attrs[key] = *val
}

// setBoolAttr sets an attribute key to attrTrue or attrFalse when the pointer is non-nil.
func setBoolAttr(attrs map[string]string, key string, val *bool) {
	if val == nil {
		return
	}
	if *val {
		attrs[key] = attrTrue
	} else {
		attrs[key] = attrFalse
	}
}

// setFloatAsIntAttr sets an attribute key to an integer string when the pointer is non-nil.
func setFloatAsIntAttr(attrs map[string]string, key string, val *float64) {
	if val == nil {
		return
	}
	attrs[key] = strconv.FormatInt(int64(*val), 10)
}

// maybeSetStr adds an attribute to updates when the spec value differs from observed.
func maybeSetStr(updates map[string]string, key string, spec *string, observed string) {
	if spec == nil {
		return
	}
	if *spec != observed {
		updates[key] = *spec
	}
}

// maybeSetBool adds an attribute to updates when the spec bool differs from observed.
func maybeSetBool(updates map[string]string, key string, spec *bool, observed string) {
	if spec == nil {
		return
	}
	want := attrFalse
	if *spec {
		want = attrTrue
	}
	if want != observed {
		updates[key] = want
	}
}

// maybeSetFloatAsInt adds an attribute to updates when the spec float (as integer
// string) differs from observed.
func maybeSetFloatAsInt(updates map[string]string, key string, spec *float64, observed string) {
	if spec == nil {
		return
	}
	want := strconv.FormatInt(int64(*spec), 10)
	if want != observed {
		updates[key] = want
	}
}

// ── tag helpers ────────────────────────────────────────────────────────────────

// specTagsToNative converts map[string]*string spec tags to []nativehelper.Tag.
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

// snsTagsToNative converts []snstypes.Tag to []nativehelper.Tag.
func snsTagsToNative(tags []snstypes.Tag) []nativehelper.Tag {
	out := make([]nativehelper.Tag, 0, len(tags))
	for _, t := range tags {
		t := t
		out = append(out, nativehelper.Tag{Key: t.Key, Value: t.Value})
	}
	return out
}

// specTagsToSNSTags converts map[string]*string spec tags to []snstypes.Tag
// for use with the CreateTopic Tags parameter.
func specTagsToSNSTags(tags map[string]*string) []snstypes.Tag {
	out := make([]snstypes.Tag, 0, len(tags))
	for k, v := range tags {
		k, v := k, v
		if v == nil {
			continue
		}
		out = append(out, snstypes.Tag{Key: &k, Value: v})
	}
	return out
}
