// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package statemachine implements the shared CRUD logic for StateMachineRAW
// resources. It is scope-agnostic: both the cluster-scoped and namespaced
// controllers delegate to ExternalClient here via the StateMachineCR interface.
package statemachine

import (
	"context"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssfn "github.com/aws/aws-sdk-go-v2/service/sfn"
	sfntypes "github.com/aws/aws-sdk-go-v2/service/sfn/types"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1beta2native "github.com/upbound/provider-aws/v2/apis/cluster/sfn/v1beta2/native"
	nativehelper "github.com/upbound/provider-aws/v2/internal/native"
)

const (
	errDescribe      = "cannot describe StateMachine"
	errCreate        = "cannot create StateMachine"
	errUpdate        = "cannot update StateMachine"
	errDelete        = "cannot delete StateMachine"
	errListTags      = "cannot list tags for StateMachine"
	errTagResource   = "cannot tag StateMachine"
	errUntagResource = "cannot untag StateMachine"
)

// SFNClient is the interface for AWS SFN operations required by this controller.
// Defining it as an interface enables mocking in unit tests.
type SFNClient interface {
	DescribeStateMachine(ctx context.Context, params *awssfn.DescribeStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DescribeStateMachineOutput, error)
	CreateStateMachine(ctx context.Context, params *awssfn.CreateStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.CreateStateMachineOutput, error)
	UpdateStateMachine(ctx context.Context, params *awssfn.UpdateStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.UpdateStateMachineOutput, error)
	DeleteStateMachine(ctx context.Context, params *awssfn.DeleteStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DeleteStateMachineOutput, error)
	ListTagsForResource(ctx context.Context, params *awssfn.ListTagsForResourceInput, optFns ...func(*awssfn.Options)) (*awssfn.ListTagsForResourceOutput, error)
	TagResource(ctx context.Context, params *awssfn.TagResourceInput, optFns ...func(*awssfn.Options)) (*awssfn.TagResourceOutput, error)
	UntagResource(ctx context.Context, params *awssfn.UntagResourceInput, optFns ...func(*awssfn.Options)) (*awssfn.UntagResourceOutput, error)
}

// StateMachineCR abstracts over cluster-scoped and namespaced StateMachineRAW
// types. Both scope types implement this interface so that the shared
// ExternalClient can operate on either without scope-specific logic.
type StateMachineCR interface {
	resource.Managed
	GetForProvider() *v1beta2native.StateMachineRAWParameters
	GetInitProvider() *v1beta2native.StateMachineRAWInitParameters
	GetAtProvider() v1beta2native.StateMachineRAWObservation
	SetAtProvider(v1beta2native.StateMachineRAWObservation)
}

// ExternalClient implements the shared CRUD logic for StateMachine resources.
// It is exported so that scope-specific wrapper packages can reference it.
type ExternalClient struct {
	// Client is the AWS Step Functions SDK client (interface for testability).
	Client SFNClient
	// Kube is the Kubernetes client for reading referenced resources.
	Kube client.Client
}

// Observe checks whether the external StateMachine resource exists and is
// up-to-date with the desired state.
func (e *ExternalClient) Observe(ctx context.Context, cr StateMachineCR) (managed.ExternalObservation, error) {
	if nativehelper.GetExternalName(cr) == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	arn := getARN(cr)
	if arn == "" {
		// Cannot describe without an ARN (we need region + account to reconstruct).
		// Return ResourceExists: false to trigger Create; CreateStateMachine is
		// idempotent when called with identical parameters.
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	resp, err := e.Client.DescribeStateMachine(ctx, &awssfn.DescribeStateMachineInput{
		StateMachineArn: aws.String(arn),
	})
	if err != nil {
		if isStateMachineDoesNotExist(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, nativehelper.Wrap(err, errDescribe)
	}

	cr.SetAtProvider(mapDescribeToObservation(resp))

	if resp.Status == sfntypes.StateMachineStatusActive {
		cr.SetConditions(xpv1.Available())
	}

	upToDate, err := e.isUpToDate(ctx, cr, resp)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	// Set the "Test=True" condition when the resource is annotated as a test
	// resource (upjet.upbound.io/test=true) and is fully up-to-date.
	// This allows uptest's --default-conditions="Test" assertion to pass for
	// native controllers the same way it does for upjet-based controllers.
	nativehelper.SetTestConditionIfAnnotated(cr, upToDate)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

// Create creates the external StateMachine resource.
func (e *ExternalClient) Create(ctx context.Context, cr StateMachineCR) (managed.ExternalCreation, error) {
	cr.SetConditions(xpv1.Creating())

	resp, err := e.Client.CreateStateMachine(ctx, buildCreateInput(cr))
	if err != nil {
		return managed.ExternalCreation{}, nativehelper.Wrap(err, errCreate)
	}

	// Store the full ARN as the external name.  Using the ARN (not just the short
	// name) is critical for Observe to work reliably: the reconciler resets
	// status.atProvider between Create and the next Observe (because the
	// annotation update step returns the pre-status object from the API server),
	// so we cannot rely on status.atProvider.arn being present.  The external
	// name annotation IS persisted through annotation updates and is checked by
	// getARN(), which falls back to it when status.atProvider.arn is empty.
	if arn := aws.ToString(resp.StateMachineArn); arn != "" {
		nativehelper.SetExternalName(cr, arn)
	}

	// Store the full ARN in status so Observe/Update/Delete can use it directly.
	obs := cr.GetAtProvider()
	obs.Arn = resp.StateMachineArn
	obs.ID = resp.StateMachineArn
	obs.StateMachineVersionArn = resp.StateMachineVersionArn
	if resp.CreationDate != nil {
		t := resp.CreationDate.Format("2006-01-02T15:04:05Z")
		obs.CreationDate = &t
	}
	cr.SetAtProvider(obs)

	return managed.ExternalCreation{}, nil
}

// Update updates the external StateMachine resource.
func (e *ExternalClient) Update(ctx context.Context, cr StateMachineCR) (managed.ExternalUpdate, error) {
	arn := getARN(cr)
	spec := cr.GetForProvider()

	input := buildUpdateInput(arn, spec)
	resp, err := e.Client.UpdateStateMachine(ctx, input)
	if err != nil {
		return managed.ExternalUpdate{}, nativehelper.Wrap(err, errUpdate)
	}

	obs := cr.GetAtProvider()
	obs.RevisionID = resp.RevisionId
	obs.StateMachineVersionArn = resp.StateMachineVersionArn
	cr.SetAtProvider(obs)

	if err := e.reconcileTags(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, err
	}

	return managed.ExternalUpdate{}, nil
}

// Delete deletes the external StateMachine resource.
func (e *ExternalClient) Delete(ctx context.Context, cr StateMachineCR) (managed.ExternalDelete, error) {
	cr.SetConditions(xpv1.Deleting())

	arn := getARN(cr)
	if arn == "" {
		return managed.ExternalDelete{}, nil
	}

	_, err := e.Client.DeleteStateMachine(ctx, &awssfn.DeleteStateMachineInput{
		StateMachineArn: aws.String(arn),
	})
	if err != nil {
		if isStateMachineDoesNotExist(err) {
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, nativehelper.Wrap(err, errDelete)
	}

	return managed.ExternalDelete{}, nil
}

// ── input builders ─────────────────────────────────────────────────────────────

// buildCreateInput assembles the CreateStateMachineInput from the CR spec.
func buildCreateInput(cr StateMachineCR) *awssfn.CreateStateMachineInput {
	spec := cr.GetForProvider()

	name := stateMachineNameFromCR(cr)

	input := &awssfn.CreateStateMachineInput{
		Name:       aws.String(name),
		Definition: spec.Definition,
		RoleArn:    spec.RoleArn,
	}

	if spec.Type != nil {
		input.Type = sfntypes.StateMachineType(*spec.Type)
	}
	if spec.Publish != nil {
		input.Publish = *spec.Publish
	}
	if spec.LoggingConfiguration != nil {
		input.LoggingConfiguration = mapLoggingConfigToAWS(spec.LoggingConfiguration)
	}
	if spec.TracingConfiguration != nil && spec.TracingConfiguration.Enabled != nil {
		input.TracingConfiguration = &sfntypes.TracingConfiguration{Enabled: *spec.TracingConfiguration.Enabled}
	}
	if spec.EncryptionConfiguration != nil {
		input.EncryptionConfiguration = mapEncryptionConfigToAWS(spec.EncryptionConfiguration)
	}
	if len(spec.Tags) > 0 {
		input.Tags = specTagsToAWS(spec.Tags)
	}

	return input
}

// buildUpdateInput assembles the UpdateStateMachineInput from the CR spec.
func buildUpdateInput(arn string, spec *v1beta2native.StateMachineRAWParameters) *awssfn.UpdateStateMachineInput {
	input := &awssfn.UpdateStateMachineInput{
		StateMachineArn: aws.String(arn),
		Definition:      spec.Definition,
		RoleArn:         spec.RoleArn,
	}

	if spec.Publish != nil {
		input.Publish = *spec.Publish
	}
	if spec.LoggingConfiguration != nil {
		input.LoggingConfiguration = mapLoggingConfigToAWS(spec.LoggingConfiguration)
	}
	if spec.TracingConfiguration != nil && spec.TracingConfiguration.Enabled != nil {
		input.TracingConfiguration = &sfntypes.TracingConfiguration{Enabled: *spec.TracingConfiguration.Enabled}
	}
	if spec.EncryptionConfiguration != nil {
		input.EncryptionConfiguration = mapEncryptionConfigToAWS(spec.EncryptionConfiguration)
	}

	return input
}

// ── helpers ────────────────────────────────────────────────────────────────────

// getARN returns the state machine ARN from status.atProvider.arn.
// Falls back to treating the external name as an ARN when it has the "arn:" prefix
// (import case where status is not yet populated).
func getARN(cr StateMachineCR) string {
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

// stateMachineNameFromCR returns the short AWS state machine name to use in API
// calls.  The external name annotation may hold either the short name or the full
// ARN (set by Create after a successful creation); this helper normalises both
// forms to the short name.
func stateMachineNameFromCR(cr StateMachineCR) string {
	name := nativehelper.GetExternalName(cr)
	if name == "" {
		return cr.GetName()
	}
	// External name was previously stored as the full ARN by Create; extract
	// just the state machine name component.
	if strings.HasPrefix(name, "arn:") {
		if extracted := nameFromARN(name); extracted != "" {
			return extracted
		}
	}
	return name
}

// nameFromARN extracts the state machine name from a full ARN.
// ARN format: arn:aws:states:<region>:<account>:stateMachine:<name>
// Returns an empty string if the ARN format is unexpected.
func nameFromARN(arn string) string {
	if arn == "" {
		return ""
	}
	parts := strings.Split(arn, ":")
	if len(parts) < 7 {
		return ""
	}
	return parts[len(parts)-1]
}

// isStateMachineDoesNotExist returns true if err is a StateMachineDoesNotExist error.
func isStateMachineDoesNotExist(err error) bool {
	var notFound *sfntypes.StateMachineDoesNotExist
	return errors.As(err, &notFound)
}

// mapDescribeToObservation populates an observation struct from a DescribeStateMachine response.
func mapDescribeToObservation(resp *awssfn.DescribeStateMachineOutput) v1beta2native.StateMachineRAWObservation {
	obs := v1beta2native.StateMachineRAWObservation{
		Arn:         resp.StateMachineArn,
		ID:          resp.StateMachineArn,
		RoleArn:     resp.RoleArn,
		Status:      aws.String(string(resp.Status)),
		Type:        aws.String(string(resp.Type)),
		RevisionID:  resp.RevisionId,
		Description: resp.Description,
		Definition:  resp.Definition,
	}
	if resp.CreationDate != nil {
		t := resp.CreationDate.Format("2006-01-02T15:04:05Z")
		obs.CreationDate = &t
	}
	if resp.LoggingConfiguration != nil {
		obs.LoggingConfiguration = mapLoggingConfigFromAWS(resp.LoggingConfiguration)
	}
	if resp.TracingConfiguration != nil {
		enabled := resp.TracingConfiguration.Enabled
		obs.TracingConfiguration = &v1beta2native.TracingConfigurationRAWObservation{Enabled: &enabled}
	}
	if resp.EncryptionConfiguration != nil {
		obs.EncryptionConfiguration = mapEncryptionConfigFromAWS(resp.EncryptionConfiguration)
	}
	return obs
}

// isUpToDate compares the desired spec against the observed AWS state.
// Also checks tags via a ListTagsForResource call.
func (e *ExternalClient) isUpToDate(ctx context.Context, cr StateMachineCR, resp *awssfn.DescribeStateMachineOutput) (bool, error) {
	if !specConfigUpToDate(cr.GetForProvider(), resp) {
		return false, nil
	}

	arn := getARN(cr)
	tagsResp, err := e.Client.ListTagsForResource(ctx, &awssfn.ListTagsForResourceInput{
		ResourceArn: aws.String(arn),
	})
	if err != nil {
		return false, nativehelper.Wrap(err, errListTags)
	}

	toAdd, toRemove := nativehelper.DiffTagsWithDefaults(
		specTagsToNative(cr.GetForProvider().Tags),
		awsTagsToNative(tagsResp.Tags),
		nil,
	)
	if len(toAdd) > 0 || len(toRemove) > 0 {
		return false, nil
	}

	return true, nil
}

// specConfigUpToDate checks non-tag spec fields against the AWS describe response.
func specConfigUpToDate(spec *v1beta2native.StateMachineRAWParameters, resp *awssfn.DescribeStateMachineOutput) bool {
	if aws.ToString(spec.Definition) != aws.ToString(resp.Definition) {
		return false
	}
	if aws.ToString(spec.RoleArn) != aws.ToString(resp.RoleArn) {
		return false
	}
	if spec.Type != nil && *spec.Type != string(resp.Type) {
		return false
	}
	if !loggingConfigUpToDate(spec.LoggingConfiguration, resp.LoggingConfiguration) {
		return false
	}
	if !tracingConfigUpToDate(spec.TracingConfiguration, resp.TracingConfiguration) {
		return false
	}
	if !encryptionConfigUpToDate(spec.EncryptionConfiguration, resp.EncryptionConfiguration) {
		return false
	}
	return true
}

// reconcileTags synchronises the resource's AWS tags with the spec.
func (e *ExternalClient) reconcileTags(ctx context.Context, cr StateMachineCR) error {
	arn := getARN(cr)
	spec := cr.GetForProvider()

	tagsResp, err := e.Client.ListTagsForResource(ctx, &awssfn.ListTagsForResourceInput{
		ResourceArn: aws.String(arn),
	})
	if err != nil {
		return nativehelper.Wrap(err, errListTags)
	}

	toAdd, toRemove := nativehelper.DiffTagsWithDefaults(
		specTagsToNative(spec.Tags),
		awsTagsToNative(tagsResp.Tags),
		nil,
	)

	if len(toAdd) > 0 {
		if _, err := e.Client.TagResource(ctx, &awssfn.TagResourceInput{
			ResourceArn: aws.String(arn),
			Tags:        nativeTagsToAWS(toAdd),
		}); err != nil {
			return nativehelper.Wrap(err, errTagResource)
		}
	}

	if len(toRemove) > 0 {
		keys := make([]string, 0, len(toRemove))
		for _, t := range toRemove {
			if t.Key != nil {
				keys = append(keys, *t.Key)
			}
		}
		if _, err := e.Client.UntagResource(ctx, &awssfn.UntagResourceInput{
			ResourceArn: aws.String(arn),
			TagKeys:     keys,
		}); err != nil {
			return nativehelper.Wrap(err, errUntagResource)
		}
	}

	return nil
}

// ── tag converters ─────────────────────────────────────────────────────────────

// specTagsToNative converts the map[string]*string spec tags to []native.Tag.
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

// awsTagsToNative converts []sfntypes.Tag to []native.Tag.
func awsTagsToNative(tags []sfntypes.Tag) []nativehelper.Tag {
	out := make([]nativehelper.Tag, 0, len(tags))
	for _, t := range tags {
		t := t
		out = append(out, nativehelper.Tag{Key: t.Key, Value: t.Value})
	}
	return out
}

// nativeTagsToAWS converts []native.Tag to []sfntypes.Tag.
func nativeTagsToAWS(tags []nativehelper.Tag) []sfntypes.Tag {
	out := make([]sfntypes.Tag, 0, len(tags))
	for _, t := range tags {
		t := t
		out = append(out, sfntypes.Tag{Key: t.Key, Value: t.Value})
	}
	return out
}

// specTagsToAWS converts map[string]*string spec tags to []sfntypes.Tag for use in Create.
func specTagsToAWS(tags map[string]*string) []sfntypes.Tag {
	out := make([]sfntypes.Tag, 0, len(tags))
	for k, v := range tags {
		k, v := k, v
		if v == nil {
			continue
		}
		out = append(out, sfntypes.Tag{Key: &k, Value: v})
	}
	return out
}

// ── config converters ──────────────────────────────────────────────────────────

// mapLoggingConfigToAWS converts spec logging config to SDK type.
func mapLoggingConfigToAWS(cfg *v1beta2native.LoggingConfigurationRAWParameters) *sfntypes.LoggingConfiguration {
	if cfg == nil {
		return nil
	}
	out := &sfntypes.LoggingConfiguration{}
	if cfg.Level != nil {
		out.Level = sfntypes.LogLevel(*cfg.Level)
	}
	if cfg.IncludeExecutionData != nil {
		out.IncludeExecutionData = *cfg.IncludeExecutionData
	}
	if cfg.LogDestination != nil {
		out.Destinations = []sfntypes.LogDestination{
			{CloudWatchLogsLogGroup: &sfntypes.CloudWatchLogsLogGroup{LogGroupArn: cfg.LogDestination}},
		}
	}
	return out
}

// mapLoggingConfigFromAWS converts SDK logging config to observation type.
func mapLoggingConfigFromAWS(cfg *sfntypes.LoggingConfiguration) *v1beta2native.LoggingConfigurationRAWObservation {
	if cfg == nil {
		return nil
	}
	level := string(cfg.Level)
	out := &v1beta2native.LoggingConfigurationRAWObservation{
		Level:                &level,
		IncludeExecutionData: &cfg.IncludeExecutionData,
	}
	if len(cfg.Destinations) > 0 && cfg.Destinations[0].CloudWatchLogsLogGroup != nil {
		out.LogDestination = cfg.Destinations[0].CloudWatchLogsLogGroup.LogGroupArn
	}
	return out
}

// mapEncryptionConfigToAWS converts spec encryption config to SDK type.
func mapEncryptionConfigToAWS(cfg *v1beta2native.EncryptionConfigurationRAWParameters) *sfntypes.EncryptionConfiguration {
	if cfg == nil {
		return nil
	}
	out := &sfntypes.EncryptionConfiguration{
		KmsKeyId: cfg.KMSKeyID,
	}
	if cfg.Type != nil {
		out.Type = sfntypes.EncryptionType(*cfg.Type)
	}
	if cfg.KMSDataKeyReusePeriodSeconds != nil {
		v := int32(*cfg.KMSDataKeyReusePeriodSeconds)
		out.KmsDataKeyReusePeriodSeconds = &v
	}
	return out
}

// mapEncryptionConfigFromAWS converts SDK encryption config to observation type.
func mapEncryptionConfigFromAWS(cfg *sfntypes.EncryptionConfiguration) *v1beta2native.EncryptionConfigurationRAWObservation {
	if cfg == nil {
		return nil
	}
	encType := string(cfg.Type)
	out := &v1beta2native.EncryptionConfigurationRAWObservation{
		Type:     &encType,
		KMSKeyID: cfg.KmsKeyId,
	}
	if cfg.KmsDataKeyReusePeriodSeconds != nil {
		v := float64(*cfg.KmsDataKeyReusePeriodSeconds)
		out.KMSDataKeyReusePeriodSeconds = &v
	}
	return out
}

// ── up-to-date comparison helpers ─────────────────────────────────────────────

func loggingConfigUpToDate(spec *v1beta2native.LoggingConfigurationRAWParameters, observed *sfntypes.LoggingConfiguration) bool {
	if spec == nil {
		// User did not specify a logging configuration; accept whatever AWS has.
		return true
	}
	if observed == nil {
		return false
	}
	if spec.Level != nil && *spec.Level != string(observed.Level) {
		return false
	}
	if spec.IncludeExecutionData != nil && *spec.IncludeExecutionData != observed.IncludeExecutionData {
		return false
	}
	return logDestinationUpToDate(spec.LogDestination, observed)
}

// logDestinationUpToDate checks whether the single-log-destination setting is unchanged.
func logDestinationUpToDate(specDest *string, observed *sfntypes.LoggingConfiguration) bool {
	specARN := aws.ToString(specDest)
	var obsARN string
	if len(observed.Destinations) > 0 && observed.Destinations[0].CloudWatchLogsLogGroup != nil {
		obsARN = aws.ToString(observed.Destinations[0].CloudWatchLogsLogGroup.LogGroupArn)
	}
	return specARN == obsARN
}

func tracingConfigUpToDate(spec *v1beta2native.TracingConfigurationRAWParameters, observed *sfntypes.TracingConfiguration) bool {
	if spec == nil && (observed == nil || !observed.Enabled) {
		return true
	}
	if spec == nil {
		return !observed.Enabled
	}
	if observed == nil {
		return spec.Enabled == nil || !*spec.Enabled
	}
	if spec.Enabled != nil && *spec.Enabled != observed.Enabled {
		return false
	}
	return true
}

func encryptionConfigUpToDate(spec *v1beta2native.EncryptionConfigurationRAWParameters, observed *sfntypes.EncryptionConfiguration) bool {
	if spec == nil {
		// User did not specify an encryption configuration; accept whatever AWS has.
		return true
	}
	if observed == nil {
		return false
	}
	if spec.Type != nil && *spec.Type != string(observed.Type) {
		return false
	}
	return aws.ToString(spec.KMSKeyID) == aws.ToString(observed.KmsKeyId)
}
