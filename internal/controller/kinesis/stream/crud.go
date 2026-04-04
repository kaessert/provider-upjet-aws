// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package stream implements the shared CRUD logic for StreamRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the StreamCR interface.
package stream

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awskinesis "github.com/aws/aws-sdk-go-v2/service/kinesis"
	ktypes "github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1beta2native "github.com/upbound/provider-aws/v2/apis/cluster/kinesis/v1beta2/native"
	nativehelper "github.com/upbound/provider-aws/v2/internal/native"
)

const (
	errDescribe          = "cannot describe Kinesis Stream"
	errCreate            = "cannot create Kinesis Stream"
	errDelete            = "cannot delete Kinesis Stream"
	errListTags          = "cannot list tags for Kinesis Stream"
	errAddTags           = "cannot add tags to Kinesis Stream"
	errRemoveTags        = "cannot remove tags from Kinesis Stream"
	errUpdateShardCount  = "cannot update shard count for Kinesis Stream"
	errStartEncryption   = "cannot start encryption for Kinesis Stream"
	errStopEncryption    = "cannot stop encryption for Kinesis Stream"
	errIncreaseRetention = "cannot increase retention period for Kinesis Stream"
	errDecreaseRetention = "cannot decrease retention period for Kinesis Stream"
	errUpdateStreamMode  = "cannot update stream mode for Kinesis Stream"
	errEnableMetrics     = "cannot enable enhanced monitoring for Kinesis Stream"
	errDisableMetrics    = "cannot disable enhanced monitoring for Kinesis Stream"
	errUpdateMaxRecSize  = "cannot update max record size for Kinesis Stream"
)

// KinesisStreamClient is the interface for AWS Kinesis operations required by
// the stream controller. Defined as an interface to enable mocking in unit tests.
type KinesisStreamClient interface {
	DescribeStreamSummary(ctx context.Context, params *awskinesis.DescribeStreamSummaryInput, optFns ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamSummaryOutput, error)
	CreateStream(ctx context.Context, params *awskinesis.CreateStreamInput, optFns ...func(*awskinesis.Options)) (*awskinesis.CreateStreamOutput, error)
	DeleteStream(ctx context.Context, params *awskinesis.DeleteStreamInput, optFns ...func(*awskinesis.Options)) (*awskinesis.DeleteStreamOutput, error)
	UpdateShardCount(ctx context.Context, params *awskinesis.UpdateShardCountInput, optFns ...func(*awskinesis.Options)) (*awskinesis.UpdateShardCountOutput, error)
	StartStreamEncryption(ctx context.Context, params *awskinesis.StartStreamEncryptionInput, optFns ...func(*awskinesis.Options)) (*awskinesis.StartStreamEncryptionOutput, error)
	StopStreamEncryption(ctx context.Context, params *awskinesis.StopStreamEncryptionInput, optFns ...func(*awskinesis.Options)) (*awskinesis.StopStreamEncryptionOutput, error)
	IncreaseStreamRetentionPeriod(ctx context.Context, params *awskinesis.IncreaseStreamRetentionPeriodInput, optFns ...func(*awskinesis.Options)) (*awskinesis.IncreaseStreamRetentionPeriodOutput, error)
	DecreaseStreamRetentionPeriod(ctx context.Context, params *awskinesis.DecreaseStreamRetentionPeriodInput, optFns ...func(*awskinesis.Options)) (*awskinesis.DecreaseStreamRetentionPeriodOutput, error)
	UpdateStreamMode(ctx context.Context, params *awskinesis.UpdateStreamModeInput, optFns ...func(*awskinesis.Options)) (*awskinesis.UpdateStreamModeOutput, error)
	EnableEnhancedMonitoring(ctx context.Context, params *awskinesis.EnableEnhancedMonitoringInput, optFns ...func(*awskinesis.Options)) (*awskinesis.EnableEnhancedMonitoringOutput, error)
	DisableEnhancedMonitoring(ctx context.Context, params *awskinesis.DisableEnhancedMonitoringInput, optFns ...func(*awskinesis.Options)) (*awskinesis.DisableEnhancedMonitoringOutput, error)
	ListTagsForStream(ctx context.Context, params *awskinesis.ListTagsForStreamInput, optFns ...func(*awskinesis.Options)) (*awskinesis.ListTagsForStreamOutput, error)
	AddTagsToStream(ctx context.Context, params *awskinesis.AddTagsToStreamInput, optFns ...func(*awskinesis.Options)) (*awskinesis.AddTagsToStreamOutput, error)
	RemoveTagsFromStream(ctx context.Context, params *awskinesis.RemoveTagsFromStreamInput, optFns ...func(*awskinesis.Options)) (*awskinesis.RemoveTagsFromStreamOutput, error)
	UpdateMaxRecordSize(ctx context.Context, params *awskinesis.UpdateMaxRecordSizeInput, optFns ...func(*awskinesis.Options)) (*awskinesis.UpdateMaxRecordSizeOutput, error)
}

// StreamCR abstracts over cluster-scoped and namespaced StreamRAW types.
// Both scope types implement this interface so that the shared ExternalClient
// can operate on either without scope-specific logic.
type StreamCR interface {
	resource.Managed
	GetForProvider() *v1beta2native.StreamRAWParameters
	// SetForProvider writes back the full ForProvider parameters.
	// ⚠️ MANDATORY: Namespaced types return a *copy* from GetForProvider() (to
	// convert reference types). Without SetForProvider(), late-initialized fields
	// are silently discarded every reconcile, causing an infinite
	// ResourceLateInitialized loop where Ready never becomes True.
	SetForProvider(v1beta2native.StreamRAWParameters)
	GetInitProvider() *v1beta2native.StreamRAWInitParameters
	GetAtProvider() v1beta2native.StreamRAWObservation
	SetAtProvider(v1beta2native.StreamRAWObservation)
}

// ExternalClient implements the shared CRUD logic for StreamRAW resources.
// It is exported so that scope-specific wrapper packages can reference it.
type ExternalClient struct {
	// Client is the AWS Kinesis SDK client (interface for testability).
	Client KinesisStreamClient
	// Kube is the Kubernetes client for reading referenced resources.
	Kube client.Client
}

// Observe checks whether the external Stream resource exists and is up-to-date.
func (e *ExternalClient) Observe(ctx context.Context, cr StreamCR) (managed.ExternalObservation, error) {
	if nativehelper.GetExternalName(cr) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	resp, err := e.Client.DescribeStreamSummary(ctx, buildDescribeInput(cr))
	if err != nil {
		if isNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, nativehelper.Wrap(err, errDescribe)
	}

	summary := resp.StreamDescriptionSummary
	if summary == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	populateObservation(cr, summary)
	spec := cr.GetForProvider()
	lateInited := lateInitialize(spec, summary)
	if lateInited {
		cr.SetForProvider(*spec)
	}
	setStreamCondition(cr, summary.StreamStatus)

	// While transitioning, skip the upToDate check to prevent spurious Update
	// calls that would fail with ResourceInUseException.
	if summary.StreamStatus != ktypes.StreamStatusActive {
		return managed.ExternalObservation{
			ResourceExists:          true,
			ResourceUpToDate:        true,
			ResourceLateInitialized: lateInited,
		}, nil
	}

	upToDate, err := e.isUpToDate(ctx, cr, summary)
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	nativehelper.SetTestConditionIfAnnotated(cr, upToDate)

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        upToDate,
		ResourceLateInitialized: lateInited,
	}, nil
}

// Create creates the external Stream resource.
func (e *ExternalClient) Create(ctx context.Context, cr StreamCR) (managed.ExternalCreation, error) {
	cr.SetConditions(xpv1.Creating())

	input := buildCreateInput(cr)
	if _, err := e.Client.CreateStream(ctx, input); err != nil {
		return managed.ExternalCreation{}, nativehelper.Wrap(err, errCreate)
	}

	// CreateStream has no return payload. Call DescribeStreamSummary to get the
	// ARN and store it as the external name so subsequent Observe calls use the
	// ARN-based lookup (robust to external name annotation resets).
	descResp, err := e.Client.DescribeStreamSummary(ctx, &awskinesis.DescribeStreamSummaryInput{
		StreamName: input.StreamName,
	})
	if err == nil && descResp.StreamDescriptionSummary != nil {
		populateObservation(cr, descResp.StreamDescriptionSummary)
	}

	return managed.ExternalCreation{}, nil
}

// Update updates the external Stream resource.
// Kinesis streams do not have a single Update API; instead, multiple targeted
// operations are used depending on which field has changed.
func (e *ExternalClient) Update(ctx context.Context, cr StreamCR) (managed.ExternalUpdate, error) {
	descResp, err := e.Client.DescribeStreamSummary(ctx, buildDescribeInput(cr))
	if err != nil {
		return managed.ExternalUpdate{}, nativehelper.Wrap(err, errDescribe)
	}
	if descResp.StreamDescriptionSummary == nil {
		return managed.ExternalUpdate{}, errors.New("stream summary is nil")
	}
	summary := descResp.StreamDescriptionSummary
	arn := aws.ToString(summary.StreamARN)

	if err := e.updateStreamMode(ctx, cr, summary, arn); err != nil {
		return managed.ExternalUpdate{}, err
	}
	if err := e.updateShardCount(ctx, cr, summary, arn); err != nil {
		return managed.ExternalUpdate{}, err
	}
	if err := e.updateEncryption(ctx, cr, summary, arn); err != nil {
		return managed.ExternalUpdate{}, err
	}
	if err := e.updateRetentionPeriod(ctx, cr, summary, arn); err != nil {
		return managed.ExternalUpdate{}, err
	}
	if err := e.reconcileEnhancedMonitoring(ctx, cr, summary, arn); err != nil {
		return managed.ExternalUpdate{}, err
	}
	if err := e.updateMaxRecordSize(ctx, cr, summary, arn); err != nil {
		return managed.ExternalUpdate{}, err
	}
	if err := e.reconcileTags(ctx, cr, arn); err != nil {
		return managed.ExternalUpdate{}, err
	}
	return managed.ExternalUpdate{}, nil
}

// Delete deletes the external Stream resource.
func (e *ExternalClient) Delete(ctx context.Context, cr StreamCR) (managed.ExternalDelete, error) {
	cr.SetConditions(xpv1.Deleting())

	extName := nativehelper.GetExternalName(cr)
	if extName == "" {
		return managed.ExternalDelete{}, nil
	}

	input := buildDeleteInput(extName, cr.GetForProvider())
	if _, err := e.Client.DeleteStream(ctx, input); err != nil {
		if isNotFound(err) {
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, nativehelper.Wrap(err, errDelete)
	}
	return managed.ExternalDelete{}, nil
}

// ── targeted update helpers ──────────────────────────────────────────────────

// updateStreamMode applies StreamMode changes (PROVISIONED ↔ ON_DEMAND).
// This must run before updateShardCount since ON_DEMAND ignores shard count.
func (e *ExternalClient) updateStreamMode(ctx context.Context, cr StreamCR, summary *ktypes.StreamDescriptionSummary, arn string) error {
	spec := cr.GetForProvider()
	if spec.StreamModeDetails == nil {
		return nil
	}
	desired := ktypes.StreamMode(aws.ToString(spec.StreamModeDetails.StreamMode))
	current := streamModeFromSummary(summary)
	if desired == current {
		return nil
	}
	_, err := e.Client.UpdateStreamMode(ctx, &awskinesis.UpdateStreamModeInput{
		StreamARN:         aws.String(arn),
		StreamModeDetails: &ktypes.StreamModeDetails{StreamMode: desired},
	})
	return nativehelper.Wrap(err, errUpdateStreamMode)
}

// updateShardCount applies shard count changes (only for PROVISIONED mode).
func (e *ExternalClient) updateShardCount(ctx context.Context, cr StreamCR, summary *ktypes.StreamDescriptionSummary, arn string) error {
	spec := cr.GetForProvider()
	if spec.ShardCount == nil || summary.OpenShardCount == nil {
		return nil
	}
	desired := int32(*spec.ShardCount)
	if desired == *summary.OpenShardCount {
		return nil
	}
	// Determine effective mode after any pending mode change.
	mode := effectiveStreamMode(spec, summary)
	if mode != ktypes.StreamModeProvisioned {
		return nil
	}
	_, err := e.Client.UpdateShardCount(ctx, &awskinesis.UpdateShardCountInput{
		StreamARN:        aws.String(arn),
		TargetShardCount: aws.Int32(desired),
		ScalingType:      ktypes.ScalingTypeUniformScaling,
	})
	return nativehelper.Wrap(err, errUpdateShardCount)
}

// updateEncryption applies encryption changes (start or stop KMS encryption).
func (e *ExternalClient) updateEncryption(ctx context.Context, cr StreamCR, summary *ktypes.StreamDescriptionSummary, arn string) error {
	spec := cr.GetForProvider()
	if spec.EncryptionType == nil {
		return nil
	}
	desired := ktypes.EncryptionType(*spec.EncryptionType)
	desiredKeyID := aws.ToString(spec.KMSKeyID)
	currentKeyID := aws.ToString(summary.KeyId)
	if desired == summary.EncryptionType && desiredKeyID == currentKeyID {
		return nil
	}
	return e.applyEncryptionChange(ctx, desired, spec.KMSKeyID, summary, arn)
}

// applyEncryptionChange performs the actual StartStreamEncryption or StopStreamEncryption call.
func (e *ExternalClient) applyEncryptionChange(ctx context.Context, desired ktypes.EncryptionType, keyID *string, summary *ktypes.StreamDescriptionSummary, arn string) error {
	switch desired {
	case ktypes.EncryptionTypeKms:
		_, err := e.Client.StartStreamEncryption(ctx, &awskinesis.StartStreamEncryptionInput{
			StreamARN:      aws.String(arn),
			EncryptionType: ktypes.EncryptionTypeKms,
			KeyId:          keyID,
		})
		return nativehelper.Wrap(err, errStartEncryption)
	case ktypes.EncryptionTypeNone:
		_, err := e.Client.StopStreamEncryption(ctx, &awskinesis.StopStreamEncryptionInput{
			StreamARN:      aws.String(arn),
			EncryptionType: summary.EncryptionType,
			KeyId:          summary.KeyId,
		})
		return nativehelper.Wrap(err, errStopEncryption)
	default:
		return nil
	}
}

// updateRetentionPeriod applies retention period changes.
func (e *ExternalClient) updateRetentionPeriod(ctx context.Context, cr StreamCR, summary *ktypes.StreamDescriptionSummary, arn string) error {
	spec := cr.GetForProvider()
	if spec.RetentionPeriod == nil || summary.RetentionPeriodHours == nil {
		return nil
	}
	desired := int32(*spec.RetentionPeriod)
	current := *summary.RetentionPeriodHours
	switch {
	case desired > current:
		_, err := e.Client.IncreaseStreamRetentionPeriod(ctx, &awskinesis.IncreaseStreamRetentionPeriodInput{
			StreamARN:            aws.String(arn),
			RetentionPeriodHours: aws.Int32(desired),
		})
		return nativehelper.Wrap(err, errIncreaseRetention)
	case desired < current:
		_, err := e.Client.DecreaseStreamRetentionPeriod(ctx, &awskinesis.DecreaseStreamRetentionPeriodInput{
			StreamARN:            aws.String(arn),
			RetentionPeriodHours: aws.Int32(desired),
		})
		return nativehelper.Wrap(err, errDecreaseRetention)
	}
	return nil
}

// updateMaxRecordSize applies max record size changes.
func (e *ExternalClient) updateMaxRecordSize(ctx context.Context, cr StreamCR, summary *ktypes.StreamDescriptionSummary, arn string) error {
	spec := cr.GetForProvider()
	if spec.MaxRecordSizeInKib == nil || summary.MaxRecordSizeInKiB == nil {
		return nil
	}
	desired := int32(*spec.MaxRecordSizeInKib)
	if desired == *summary.MaxRecordSizeInKiB {
		return nil
	}
	_, err := e.Client.UpdateMaxRecordSize(ctx, &awskinesis.UpdateMaxRecordSizeInput{
		StreamARN:          aws.String(arn),
		MaxRecordSizeInKiB: aws.Int32(desired),
	})
	return nativehelper.Wrap(err, errUpdateMaxRecSize)
}

// ── input builders ────────────────────────────────────────────────────────────

// buildDescribeInput builds a DescribeStreamSummaryInput using the ARN when
// available (from status or external name), falling back to the stream name.
func buildDescribeInput(cr StreamCR) *awskinesis.DescribeStreamSummaryInput {
	if arn := streamARN(cr); arn != "" {
		return &awskinesis.DescribeStreamSummaryInput{StreamARN: aws.String(arn)}
	}
	return &awskinesis.DescribeStreamSummaryInput{StreamName: aws.String(streamName(cr))}
}

// buildCreateInput assembles the CreateStreamInput from the CR spec.
func buildCreateInput(cr StreamCR) *awskinesis.CreateStreamInput {
	spec := cr.GetForProvider()
	input := &awskinesis.CreateStreamInput{StreamName: aws.String(streamName(cr))}
	if spec.ShardCount != nil {
		input.ShardCount = aws.Int32(int32(*spec.ShardCount))
	}
	if spec.StreamModeDetails != nil && spec.StreamModeDetails.StreamMode != nil {
		input.StreamModeDetails = &ktypes.StreamModeDetails{
			StreamMode: ktypes.StreamMode(*spec.StreamModeDetails.StreamMode),
		}
	}
	if spec.MaxRecordSizeInKib != nil {
		input.MaxRecordSizeInKiB = aws.Int32(int32(*spec.MaxRecordSizeInKib))
	}
	if len(spec.Tags) > 0 {
		input.Tags = specTagsToAWSMap(spec.Tags)
	}
	return input
}

// buildDeleteInput assembles the DeleteStreamInput.
func buildDeleteInput(extName string, spec *v1beta2native.StreamRAWParameters) *awskinesis.DeleteStreamInput {
	input := &awskinesis.DeleteStreamInput{}
	if strings.HasPrefix(extName, "arn:") {
		input.StreamARN = aws.String(extName)
	} else {
		input.StreamName = aws.String(extName)
	}
	if spec.EnforceConsumerDeletion != nil && *spec.EnforceConsumerDeletion {
		input.EnforceConsumerDeletion = aws.Bool(true)
	}
	return input
}

// ── observation helpers ───────────────────────────────────────────────────────

// populateObservation updates status.atProvider and the external name annotation
// from a DescribeStreamSummary response.
func populateObservation(cr StreamCR, summary *ktypes.StreamDescriptionSummary) {
	if summary == nil {
		return
	}
	obs := cr.GetAtProvider()
	obs.Arn = summary.StreamARN
	obs.ID = summary.StreamARN
	cr.SetAtProvider(obs)
	// Set external name to the ARN (idempotent — already ARN after first create).
	if summary.StreamARN != nil && *summary.StreamARN != "" {
		nativehelper.SetExternalName(cr, *summary.StreamARN)
	}
}

// setStreamCondition sets the condition based on the current stream status.
func setStreamCondition(cr StreamCR, status ktypes.StreamStatus) {
	switch status {
	case ktypes.StreamStatusActive:
		cr.SetConditions(xpv1.Available())
	case ktypes.StreamStatusDeleting:
		cr.SetConditions(xpv1.Deleting())
	default:
		// CREATING, UPDATING
		cr.SetConditions(xpv1.Unavailable())
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// streamARN returns the ARN for the stream: first from status.atProvider.arn,
// then from the external name annotation if it has an "arn:" prefix.
func streamARN(cr StreamCR) string {
	obs := cr.GetAtProvider()
	if obs.Arn != nil && *obs.Arn != "" {
		return *obs.Arn
	}
	extName := nativehelper.GetExternalName(cr)
	if strings.HasPrefix(extName, "arn:") {
		return extName
	}
	return ""
}

// streamName returns the plain stream name for use in API calls.
// If the external name looks like an ARN, it extracts the stream name from the
// trailing path component (arn:aws:kinesis:<region>:<account>:stream/<name>).
func streamName(cr StreamCR) string {
	extName := nativehelper.GetExternalName(cr)
	if strings.HasPrefix(extName, "arn:") {
		if idx := strings.LastIndex(extName, "/"); idx >= 0 {
			return extName[idx+1:]
		}
	}
	if extName != "" {
		return extName
	}
	return cr.GetName()
}

// isNotFound returns true when err represents a Kinesis ResourceNotFoundException.
func isNotFound(err error) bool {
	var notFound *ktypes.ResourceNotFoundException
	return errors.As(err, &notFound)
}

// streamModeFromSummary returns the current stream mode from a describe summary.
func streamModeFromSummary(summary *ktypes.StreamDescriptionSummary) ktypes.StreamMode {
	if summary.StreamModeDetails != nil {
		return summary.StreamModeDetails.StreamMode
	}
	return ktypes.StreamMode("")
}

// effectiveStreamMode returns the stream mode that will be in effect after any
// pending mode change (spec takes precedence over current AWS state).
func effectiveStreamMode(spec *v1beta2native.StreamRAWParameters, summary *ktypes.StreamDescriptionSummary) ktypes.StreamMode {
	if spec.StreamModeDetails != nil && spec.StreamModeDetails.StreamMode != nil {
		return ktypes.StreamMode(*spec.StreamModeDetails.StreamMode)
	}
	return streamModeFromSummary(summary)
}

// ── isUpToDate ────────────────────────────────────────────────────────────────

// isUpToDate returns true when the desired spec matches the observed AWS state.
func (e *ExternalClient) isUpToDate(ctx context.Context, cr StreamCR, summary *ktypes.StreamDescriptionSummary) (bool, error) {
	spec := cr.GetForProvider()

	if !specFieldsUpToDate(spec, summary) {
		return false, nil
	}

	return e.tagsUpToDate(ctx, cr, summary)
}

// specFieldsUpToDate checks non-tag spec fields against the describe response.
func specFieldsUpToDate(spec *v1beta2native.StreamRAWParameters, summary *ktypes.StreamDescriptionSummary) bool {
	if !shardCountUpToDate(spec, summary) {
		return false
	}
	if !retentionPeriodUpToDate(spec, summary) {
		return false
	}
	if !encryptionUpToDate(spec, summary) {
		return false
	}
	if !streamModeUpToDate(spec, summary) {
		return false
	}
	if !maxRecordSizeUpToDate(spec, summary) {
		return false
	}
	if spec.ShardLevelMetrics != nil {
		if !shardLevelMetricsUpToDate(spec.ShardLevelMetrics, summary.EnhancedMonitoring) {
			return false
		}
	}
	return true
}

func shardCountUpToDate(spec *v1beta2native.StreamRAWParameters, summary *ktypes.StreamDescriptionSummary) bool {
	if spec.ShardCount == nil || summary.OpenShardCount == nil {
		return true
	}
	mode := streamModeFromSummary(summary)
	if mode != ktypes.StreamModeProvisioned {
		return true // ON_DEMAND ignores shard count
	}
	return int32(*spec.ShardCount) == *summary.OpenShardCount
}

func retentionPeriodUpToDate(spec *v1beta2native.StreamRAWParameters, summary *ktypes.StreamDescriptionSummary) bool {
	if spec.RetentionPeriod == nil || summary.RetentionPeriodHours == nil {
		return true
	}
	return int32(*spec.RetentionPeriod) == *summary.RetentionPeriodHours
}

func encryptionUpToDate(spec *v1beta2native.StreamRAWParameters, summary *ktypes.StreamDescriptionSummary) bool {
	if spec.EncryptionType == nil {
		return true
	}
	if *spec.EncryptionType != string(summary.EncryptionType) {
		return false
	}
	if spec.KMSKeyID != nil {
		return aws.ToString(spec.KMSKeyID) == aws.ToString(summary.KeyId)
	}
	return true
}

func streamModeUpToDate(spec *v1beta2native.StreamRAWParameters, summary *ktypes.StreamDescriptionSummary) bool {
	if spec.StreamModeDetails == nil {
		return true // accept AWS defaults
	}
	desired := ktypes.StreamMode(aws.ToString(spec.StreamModeDetails.StreamMode))
	return desired == streamModeFromSummary(summary)
}

func maxRecordSizeUpToDate(spec *v1beta2native.StreamRAWParameters, summary *ktypes.StreamDescriptionSummary) bool {
	if spec.MaxRecordSizeInKib == nil || summary.MaxRecordSizeInKiB == nil {
		return true
	}
	return int32(*spec.MaxRecordSizeInKib) == *summary.MaxRecordSizeInKiB
}

// tagsUpToDate calls ListTagsForStream and computes the diff.
// It also populates status.atProvider.tagsAll from the AWS response.
func (e *ExternalClient) tagsUpToDate(ctx context.Context, cr StreamCR, summary *ktypes.StreamDescriptionSummary) (bool, error) {
	arn := aws.ToString(summary.StreamARN)
	tagsResp, err := e.Client.ListTagsForStream(ctx, &awskinesis.ListTagsForStreamInput{
		StreamARN: aws.String(arn),
	})
	if err != nil {
		return false, nativehelper.Wrap(err, errListTags)
	}

	toAdd, toRemove := nativehelper.DiffTagsWithDefaults(
		specTagsToNative(cr.GetForProvider().Tags),
		awsTagsToNative(tagsResp.Tags),
		nil,
	)

	// Populate tagsAll from current AWS tags.
	obs := cr.GetAtProvider()
	obs.TagsAll = awsTagsToSpecMap(tagsResp.Tags)
	cr.SetAtProvider(obs)

	return len(toAdd) == 0 && len(toRemove) == 0, nil
}

// shardLevelMetricsUpToDate returns true when the spec shard-level metrics
// match the currently-enabled metrics (order-independent set comparison).
func shardLevelMetricsUpToDate(specMetrics []*string, enhanced []ktypes.EnhancedMetrics) bool {
	current := metricsSet(enhanced)
	desired := make(map[string]bool, len(specMetrics))
	for _, m := range specMetrics {
		if m != nil {
			desired[*m] = true
		}
	}
	if len(desired) != len(current) {
		return false
	}
	for k := range desired {
		if !current[k] {
			return false
		}
	}
	return true
}

// metricsSet flattens EnhancedMetrics into a set of metric name strings.
func metricsSet(enhanced []ktypes.EnhancedMetrics) map[string]bool {
	m := make(map[string]bool)
	for _, em := range enhanced {
		for _, n := range em.ShardLevelMetrics {
			m[string(n)] = true
		}
	}
	return m
}

// ── enhanced monitoring ───────────────────────────────────────────────────────

// reconcileEnhancedMonitoring enables/disables shard-level metrics to match spec.
func (e *ExternalClient) reconcileEnhancedMonitoring(
	ctx context.Context,
	cr StreamCR,
	summary *ktypes.StreamDescriptionSummary,
	arn string,
) error {
	spec := cr.GetForProvider()
	if spec.ShardLevelMetrics == nil {
		return nil
	}

	toEnable, toDisable := metricsDiff(spec.ShardLevelMetrics, summary.EnhancedMonitoring)
	if err := e.enableMetrics(ctx, arn, toEnable); err != nil {
		return err
	}
	return e.disableMetrics(ctx, arn, toDisable)
}

// metricsDiff computes the sets of metrics to enable and disable.
func metricsDiff(specMetrics []*string, enhanced []ktypes.EnhancedMetrics) (toEnable, toDisable []ktypes.MetricsName) {
	desired := make(map[string]bool, len(specMetrics))
	for _, m := range specMetrics {
		if m != nil {
			desired[*m] = true
		}
	}
	current := metricsSet(enhanced)

	for k := range desired {
		if !current[k] {
			toEnable = append(toEnable, ktypes.MetricsName(k))
		}
	}
	for k := range current {
		if !desired[k] {
			toDisable = append(toDisable, ktypes.MetricsName(k))
		}
	}
	// Sort for deterministic API calls.
	sort.Slice(toEnable, func(i, j int) bool { return toEnable[i] < toEnable[j] })
	sort.Slice(toDisable, func(i, j int) bool { return toDisable[i] < toDisable[j] })
	return toEnable, toDisable
}

func (e *ExternalClient) enableMetrics(ctx context.Context, arn string, metrics []ktypes.MetricsName) error {
	if len(metrics) == 0 {
		return nil
	}
	_, err := e.Client.EnableEnhancedMonitoring(ctx, &awskinesis.EnableEnhancedMonitoringInput{
		StreamARN:         aws.String(arn),
		ShardLevelMetrics: metrics,
	})
	return nativehelper.Wrap(err, errEnableMetrics)
}

func (e *ExternalClient) disableMetrics(ctx context.Context, arn string, metrics []ktypes.MetricsName) error {
	if len(metrics) == 0 {
		return nil
	}
	_, err := e.Client.DisableEnhancedMonitoring(ctx, &awskinesis.DisableEnhancedMonitoringInput{
		StreamARN:         aws.String(arn),
		ShardLevelMetrics: metrics,
	})
	return nativehelper.Wrap(err, errDisableMetrics)
}

// ── tag helpers ───────────────────────────────────────────────────────────────

// reconcileTags synchronises the stream's AWS tags with the spec.
func (e *ExternalClient) reconcileTags(ctx context.Context, cr StreamCR, arn string) error {
	tagsResp, err := e.Client.ListTagsForStream(ctx, &awskinesis.ListTagsForStreamInput{
		StreamARN: aws.String(arn),
	})
	if err != nil {
		return nativehelper.Wrap(err, errListTags)
	}

	toAdd, toRemove := nativehelper.DiffTagsWithDefaults(
		specTagsToNative(cr.GetForProvider().Tags),
		awsTagsToNative(tagsResp.Tags),
		nil,
	)

	if err := e.addTags(ctx, arn, toAdd); err != nil {
		return err
	}
	return e.removeTags(ctx, arn, toRemove)
}

func (e *ExternalClient) addTags(ctx context.Context, arn string, tags []nativehelper.Tag) error {
	if len(tags) == 0 {
		return nil
	}
	_, err := e.Client.AddTagsToStream(ctx, &awskinesis.AddTagsToStreamInput{
		StreamARN: aws.String(arn),
		Tags:      nativeTagsToAWSMap(tags),
	})
	return nativehelper.Wrap(err, errAddTags)
}

func (e *ExternalClient) removeTags(ctx context.Context, arn string, tags []nativehelper.Tag) error {
	if len(tags) == 0 {
		return nil
	}
	keys := make([]string, 0, len(tags))
	for _, t := range tags {
		if t.Key != nil {
			keys = append(keys, *t.Key)
		}
	}
	_, err := e.Client.RemoveTagsFromStream(ctx, &awskinesis.RemoveTagsFromStreamInput{
		StreamARN: aws.String(arn),
		TagKeys:   keys,
	})
	return nativehelper.Wrap(err, errRemoveTags)
}

// ── tag converters ────────────────────────────────────────────────────────────

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

func awsTagsToNative(tags []ktypes.Tag) []nativehelper.Tag {
	out := make([]nativehelper.Tag, 0, len(tags))
	for _, t := range tags {
		t := t
		out = append(out, nativehelper.Tag{Key: t.Key, Value: t.Value})
	}
	return out
}

func nativeTagsToAWSMap(tags []nativehelper.Tag) map[string]string {
	out := make(map[string]string, len(tags))
	for _, t := range tags {
		if t.Key != nil && t.Value != nil {
			out[*t.Key] = *t.Value
		}
	}
	return out
}

func specTagsToAWSMap(tags map[string]*string) map[string]string {
	out := make(map[string]string, len(tags))
	for k, v := range tags {
		if v != nil {
			out[k] = *v
		}
	}
	return out
}

func awsTagsToSpecMap(tags []ktypes.Tag) map[string]*string {
	if len(tags) == 0 {
		return nil
	}
	out := make(map[string]*string, len(tags))
	for _, t := range tags {
		t := t
		if t.Key != nil {
			out[*t.Key] = t.Value
		}
	}
	return out
}

// ── late initialization ───────────────────────────────────────────────────────

// lateInitialize copies AWS-defaulted field values into the spec when those
// fields were not explicitly set. Returns true if any field was populated.
//
// Fields late-initialized:
//   - spec.forProvider.encryptionType   → AWS defaults to "NONE".
//   - spec.forProvider.retentionPeriod  → AWS defaults to 24 hours.
//   - spec.forProvider.shardCount       → from OpenShardCount.
//   - spec.forProvider.maxRecordSizeInKib → from MaxRecordSizeInKiB.
//   - spec.forProvider.streamModeDetails  → AWS always returns PROVISIONED.
//
// Note: enforce_consumer_deletion is intentionally excluded from late-init.
//
// Mutates the spec pointer directly; callers must call cr.SetForProvider(*spec)
// afterwards to persist the changes in both cluster and namespaced scopes.
func lateInitialize(spec *v1beta2native.StreamRAWParameters, summary *ktypes.StreamDescriptionSummary) bool {
	if summary == nil {
		return false
	}
	changed := false
	changed = lateInitEncryptionType(spec, summary) || changed
	changed = lateInitRetentionPeriod(spec, summary) || changed
	changed = lateInitShardCount(spec, summary) || changed
	changed = lateInitMaxRecordSize(spec, summary) || changed
	changed = lateInitStreamModeDetails(spec, summary) || changed
	return changed
}

func lateInitEncryptionType(spec *v1beta2native.StreamRAWParameters, summary *ktypes.StreamDescriptionSummary) bool {
	if spec.EncryptionType != nil || summary.EncryptionType == "" {
		return false
	}
	encType := string(summary.EncryptionType)
	spec.EncryptionType = &encType
	return true
}

func lateInitRetentionPeriod(spec *v1beta2native.StreamRAWParameters, summary *ktypes.StreamDescriptionSummary) bool {
	if spec.RetentionPeriod != nil || summary.RetentionPeriodHours == nil {
		return false
	}
	v := float64(*summary.RetentionPeriodHours)
	spec.RetentionPeriod = &v
	return true
}

func lateInitShardCount(spec *v1beta2native.StreamRAWParameters, summary *ktypes.StreamDescriptionSummary) bool {
	if spec.ShardCount != nil || summary.OpenShardCount == nil {
		return false
	}
	v := float64(*summary.OpenShardCount)
	spec.ShardCount = &v
	return true
}

func lateInitMaxRecordSize(spec *v1beta2native.StreamRAWParameters, summary *ktypes.StreamDescriptionSummary) bool {
	if spec.MaxRecordSizeInKib != nil || summary.MaxRecordSizeInKiB == nil {
		return false
	}
	v := float64(*summary.MaxRecordSizeInKiB)
	spec.MaxRecordSizeInKib = &v
	return true
}

func lateInitStreamModeDetails(spec *v1beta2native.StreamRAWParameters, summary *ktypes.StreamDescriptionSummary) bool {
	if spec.StreamModeDetails != nil || summary.StreamModeDetails == nil {
		return false
	}
	mode := string(summary.StreamModeDetails.StreamMode)
	spec.StreamModeDetails = &v1beta2native.StreamModeDetailsRAWParameters{
		StreamMode: &mode,
	}
	return true
}
