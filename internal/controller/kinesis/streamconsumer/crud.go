// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package streamconsumer implements the shared CRUD logic for StreamConsumerRAW
// resources. It is scope-agnostic: both the cluster-scoped and namespaced
// controllers delegate to ExternalClient here via the StreamConsumerCR interface.
package streamconsumer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awskinesis "github.com/aws/aws-sdk-go-v2/service/kinesis"
	ktypes "github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1beta1native "github.com/upbound/provider-aws/v2/apis/cluster/kinesis/v1beta1/native"
	nativehelper "github.com/upbound/provider-aws/v2/internal/native"
)

const (
	errDescribe  = "cannot describe Kinesis Stream Consumer"
	errCreate    = "cannot create Kinesis Stream Consumer"
	errDelete    = "cannot delete Kinesis Stream Consumer"
	errListTags  = "cannot list tags for Kinesis Stream Consumer"
	errTagAdd    = "cannot add tags to Kinesis Stream Consumer"
	errTagRemove = "cannot remove tags from Kinesis Stream Consumer"
	errUpdate    = "cannot update Kinesis Stream Consumer"
)

// KinesisConsumerClient is the interface for AWS Kinesis operations required by
// the stream consumer controller. Defined as an interface to enable mocking in
// unit tests.
type KinesisConsumerClient interface {
	RegisterStreamConsumer(ctx context.Context, params *awskinesis.RegisterStreamConsumerInput, optFns ...func(*awskinesis.Options)) (*awskinesis.RegisterStreamConsumerOutput, error)
	DescribeStreamConsumer(ctx context.Context, params *awskinesis.DescribeStreamConsumerInput, optFns ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamConsumerOutput, error)
	DeregisterStreamConsumer(ctx context.Context, params *awskinesis.DeregisterStreamConsumerInput, optFns ...func(*awskinesis.Options)) (*awskinesis.DeregisterStreamConsumerOutput, error)
	ListTagsForResource(ctx context.Context, params *awskinesis.ListTagsForResourceInput, optFns ...func(*awskinesis.Options)) (*awskinesis.ListTagsForResourceOutput, error)
	TagResource(ctx context.Context, params *awskinesis.TagResourceInput, optFns ...func(*awskinesis.Options)) (*awskinesis.TagResourceOutput, error)
	UntagResource(ctx context.Context, params *awskinesis.UntagResourceInput, optFns ...func(*awskinesis.Options)) (*awskinesis.UntagResourceOutput, error)
}

// StreamConsumerCR abstracts over cluster-scoped and namespaced StreamConsumerRAW types.
// Both scope types implement this interface so that the shared ExternalClient
// can operate on either without scope-specific logic.
type StreamConsumerCR interface {
	resource.Managed
	GetForProvider() *v1beta1native.StreamConsumerRAWParameters
	GetInitProvider() *v1beta1native.StreamConsumerRAWInitParameters
	GetAtProvider() v1beta1native.StreamConsumerRAWObservation
	SetAtProvider(v1beta1native.StreamConsumerRAWObservation)
}

// ExternalClient implements the shared CRUD logic for StreamConsumerRAW resources.
// It is exported so that scope-specific wrapper packages can reference it.
type ExternalClient struct {
	// Client is the AWS Kinesis SDK client (interface for testability).
	Client KinesisConsumerClient
	// Kube is the Kubernetes client for reading referenced resources.
	Kube client.Client
}

// Observe checks whether the external StreamConsumer resource exists and is up-to-date.
func (e *ExternalClient) Observe(ctx context.Context, cr StreamConsumerCR) (managed.ExternalObservation, error) {
	consumerARN := nativehelper.GetExternalName(cr)
	// IdentifierFromProvider: the external name is only meaningful once it has
	// been set to a consumer ARN by Create. Before that, Crossplane initialises
	// it to the Kubernetes resource name (a non-ARN value).
	// When the external name is not yet an ARN, try to adopt a pre-existing
	// consumer by looking it up by name + stream ARN. This makes the controller
	// idempotent when a consumer was left behind by a previous reconcile cycle
	// (e.g. after a failed test run that did not clean up AWS resources).
	if consumerARN == "" || !strings.HasPrefix(consumerARN, "arn:") {
		adopted, err := e.adoptByName(ctx, cr)
		if err != nil {
			return managed.ExternalObservation{}, err
		}
		if !adopted {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		consumerARN = nativehelper.GetExternalName(cr)
	}

	resp, err := e.Client.DescribeStreamConsumer(ctx, &awskinesis.DescribeStreamConsumerInput{
		ConsumerARN: aws.String(consumerARN),
	})
	if err != nil {
		if isNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, nativehelper.Wrap(err, errDescribe)
	}

	desc := resp.ConsumerDescription
	if desc == nil {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	populateObservation(cr, desc)

	// Handle non-ACTIVE states: do not trigger an Update call.
	// The consumer is transitioning; re-poll on the next PollInterval.
	switch desc.ConsumerStatus {
	case ktypes.ConsumerStatusCreating, ktypes.ConsumerStatusDeleting:
		cr.SetConditions(xpv1.Unavailable())
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, nil
	}

	// Consumer is ACTIVE — check if spec matches observed state.
	upToDate, err := e.isUpToDate(ctx, cr, desc)
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	cr.SetConditions(xpv1.Available())
	nativehelper.SetTestConditionIfAnnotated(cr, upToDate)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

// Create creates the external StreamConsumer resource by calling
// RegisterStreamConsumer. The response consumer ARN is stored as the external
// name so that subsequent Observe calls can look it up.
func (e *ExternalClient) Create(ctx context.Context, cr StreamConsumerCR) (managed.ExternalCreation, error) {
	cr.SetConditions(xpv1.Creating())

	spec := cr.GetForProvider()
	input := &awskinesis.RegisterStreamConsumerInput{
		ConsumerName: spec.Name,
		StreamARN:    spec.StreamArn,
	}

	// Include spec tags in the registration request (AWS supports tags at creation time).
	if len(spec.Tags) > 0 {
		input.Tags = specTagsToAWSMap(spec.Tags)
	}

	out, err := e.Client.RegisterStreamConsumer(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, nativehelper.Wrap(err, errCreate)
	}

	// Store the consumer ARN as the external name — this is the stable
	// identifier used for all subsequent Observe and Delete calls.
	if out.Consumer != nil && out.Consumer.ConsumerARN != nil {
		nativehelper.SetExternalName(cr, *out.Consumer.ConsumerARN)
	}

	return managed.ExternalCreation{}, nil
}

// Update reconciles the StreamConsumer resource.
// Stream consumers are immutable (name and stream_arn cannot be changed).
// Only tag updates are supported. If immutable fields differ from the observed
// AWS state, an error is returned indicating that the resource must be recreated.
func (e *ExternalClient) Update(ctx context.Context, cr StreamConsumerCR) (managed.ExternalUpdate, error) {
	consumerARN := nativehelper.GetExternalName(cr)

	// Fetch current state to detect immutable field changes.
	resp, err := e.Client.DescribeStreamConsumer(ctx, &awskinesis.DescribeStreamConsumerInput{
		ConsumerARN: aws.String(consumerARN),
	})
	if err != nil {
		return managed.ExternalUpdate{}, nativehelper.Wrap(err, errDescribe)
	}
	if resp.ConsumerDescription == nil {
		return managed.ExternalUpdate{}, errors.New("consumer description is nil")
	}
	desc := resp.ConsumerDescription

	spec := cr.GetForProvider()

	// Immutable field check: consumer name.
	if spec.Name != nil && desc.ConsumerName != nil &&
		*spec.Name != *desc.ConsumerName {
		return managed.ExternalUpdate{}, fmt.Errorf("%s: consumer name is immutable (desired=%q, observed=%q); resource must be recreated",
			errUpdate, *spec.Name, *desc.ConsumerName)
	}

	// Immutable field check: stream ARN.
	if spec.StreamArn != nil && desc.StreamARN != nil &&
		*spec.StreamArn != *desc.StreamARN {
		return managed.ExternalUpdate{}, fmt.Errorf("%s: stream_arn is immutable (desired=%q, observed=%q); resource must be recreated",
			errUpdate, *spec.StreamArn, *desc.StreamARN)
	}

	// Reconcile tags (the only mutable aspect of a stream consumer).
	if err := e.reconcileTags(ctx, cr, aws.ToString(desc.ConsumerARN)); err != nil {
		return managed.ExternalUpdate{}, err
	}

	return managed.ExternalUpdate{}, nil
}

// Delete deletes the external StreamConsumer resource.
// The delete is idempotent: if the resource is already gone, nil is returned.
func (e *ExternalClient) Delete(ctx context.Context, cr StreamConsumerCR) (managed.ExternalDelete, error) {
	cr.SetConditions(xpv1.Deleting())

	consumerARN := nativehelper.GetExternalName(cr)
	if consumerARN == "" {
		return managed.ExternalDelete{}, nil
	}

	_, err := e.Client.DeregisterStreamConsumer(ctx, &awskinesis.DeregisterStreamConsumerInput{
		ConsumerARN: aws.String(consumerARN),
	})
	if err != nil {
		if isNotFound(err) {
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, nativehelper.Wrap(err, errDelete)
	}

	return managed.ExternalDelete{}, nil
}

// ── adoption helper ───────────────────────────────────────────────────────────

// adoptByName looks up a pre-existing consumer by name + stream ARN and, if
// found, sets the external name annotation to the consumer's ARN so that the
// normal Observe path can proceed. Returns (true, nil) if adopted, (false, nil)
// if not found, or (false, err) on unexpected errors.
func (e *ExternalClient) adoptByName(ctx context.Context, cr StreamConsumerCR) (bool, error) {
	spec := cr.GetForProvider()
	if spec.Name == nil || spec.StreamArn == nil {
		return false, nil
	}

	resp, err := e.Client.DescribeStreamConsumer(ctx, &awskinesis.DescribeStreamConsumerInput{
		ConsumerName: spec.Name,
		StreamARN:    spec.StreamArn,
	})
	switch {
	case err == nil && resp.ConsumerDescription != nil:
		// Found — adopt by setting external name to the consumer ARN.
		nativehelper.SetExternalName(cr, aws.ToString(resp.ConsumerDescription.ConsumerARN))
		return true, nil
	case isNotFound(err):
		return false, nil
	case err != nil:
		return false, nativehelper.Wrap(err, errDescribe)
	default:
		// Response with nil description — treat as not found.
		return false, nil
	}
}

// ── isUpToDate ────────────────────────────────────────────────────────────────

// isUpToDate returns true when the desired spec matches the observed AWS state.
// For stream consumers, the only mutable state is tags; name and stream_arn are
// compared for detection purposes, but any drift cannot be fixed by Update().
func (e *ExternalClient) isUpToDate(ctx context.Context, cr StreamConsumerCR, desc *ktypes.ConsumerDescription) (bool, error) {
	spec := cr.GetForProvider()

	// Check immutable field parity — if they differ, isUpToDate returns false
	// so that Update() is called and returns the appropriate error.
	if spec.Name != nil && desc.ConsumerName != nil {
		if *spec.Name != *desc.ConsumerName {
			return false, nil
		}
	}
	if spec.StreamArn != nil && desc.StreamARN != nil {
		if *spec.StreamArn != *desc.StreamARN {
			return false, nil
		}
	}

	return e.tagsUpToDate(ctx, cr, aws.ToString(desc.ConsumerARN))
}

// tagsUpToDate calls ListTagsForResource and computes the diff.
// It also populates status.atProvider.tagsAll.
func (e *ExternalClient) tagsUpToDate(ctx context.Context, cr StreamConsumerCR, consumerARN string) (bool, error) {
	tagsResp, err := e.Client.ListTagsForResource(ctx, &awskinesis.ListTagsForResourceInput{
		ResourceARN: aws.String(consumerARN),
	})
	if err != nil {
		return false, nativehelper.Wrap(err, errListTags)
	}

	spec := cr.GetForProvider()
	toAdd, toRemove := nativehelper.DiffTagsWithDefaults(
		specTagsToNative(spec.Tags),
		awsTagsToNative(tagsResp.Tags),
		nil,
	)

	// Populate tagsAll from current AWS tags.
	obs := cr.GetAtProvider()
	obs.TagsAll = awsTagsToSpecMap(tagsResp.Tags)
	cr.SetAtProvider(obs)

	return len(toAdd) == 0 && len(toRemove) == 0, nil
}

// ── reconcileTags ─────────────────────────────────────────────────────────────

// reconcileTags synchronises the stream consumer's AWS tags with the spec.
func (e *ExternalClient) reconcileTags(ctx context.Context, cr StreamConsumerCR, consumerARN string) error {
	tagsResp, err := e.Client.ListTagsForResource(ctx, &awskinesis.ListTagsForResourceInput{
		ResourceARN: aws.String(consumerARN),
	})
	if err != nil {
		return nativehelper.Wrap(err, errListTags)
	}

	spec := cr.GetForProvider()
	toAdd, toRemove := nativehelper.DiffTagsWithDefaults(
		specTagsToNative(spec.Tags),
		awsTagsToNative(tagsResp.Tags),
		nil,
	)

	if err := e.addTags(ctx, consumerARN, toAdd); err != nil {
		return err
	}
	return e.removeTags(ctx, consumerARN, toRemove)
}

func (e *ExternalClient) addTags(ctx context.Context, consumerARN string, tags []nativehelper.Tag) error {
	if len(tags) == 0 {
		return nil
	}
	_, err := e.Client.TagResource(ctx, &awskinesis.TagResourceInput{
		ResourceARN: aws.String(consumerARN),
		Tags:        nativeTagsToAWSMap(tags),
	})
	return nativehelper.Wrap(err, errTagAdd)
}

func (e *ExternalClient) removeTags(ctx context.Context, consumerARN string, tags []nativehelper.Tag) error {
	if len(tags) == 0 {
		return nil
	}
	keys := make([]string, 0, len(tags))
	for _, t := range tags {
		if t.Key != nil {
			keys = append(keys, *t.Key)
		}
	}
	_, err := e.Client.UntagResource(ctx, &awskinesis.UntagResourceInput{
		ResourceARN: aws.String(consumerARN),
		TagKeys:     keys,
	})
	return nativehelper.Wrap(err, errTagRemove)
}

// ── observation helpers ───────────────────────────────────────────────────────

// populateObservation fills atProvider from a ConsumerDescription.
func populateObservation(cr StreamConsumerCR, desc *ktypes.ConsumerDescription) {
	obs := cr.GetAtProvider()
	obs.Arn = desc.ConsumerARN
	obs.ID = desc.ConsumerARN
	if desc.ConsumerCreationTimestamp != nil {
		ts := desc.ConsumerCreationTimestamp.Format("2006-01-02T15:04:05Z")
		obs.CreationTimestamp = &ts
	}
	cr.SetAtProvider(obs)
}

// ── helpers ───────────────────────────────────────────────────────────────────

// isNotFound returns true when err represents a Kinesis ResourceNotFoundException.
func isNotFound(err error) bool {
	var notFound *ktypes.ResourceNotFoundException
	return errors.As(err, &notFound)
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
