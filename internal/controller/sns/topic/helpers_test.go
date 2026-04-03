// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package topic — internal unit tests for unexported ARN helper functions.
package topic

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sns/v1beta1/native"
)

// makeMinimalCR creates a TopicRAW with just the external name set.
func makeMinimalCR(extName string) *clusternative.TopicRAW {
	cr := &clusternative.TopicRAW{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-topic",
		},
	}
	meta.SetExternalName(cr, extName)
	return cr
}

// makeMinimalCRWithARN creates a TopicRAW with an ARN stored in atProvider.
func makeMinimalCRWithARN(arnVal string) *clusternative.TopicRAW {
	cr := makeMinimalCR("test-topic")
	cr.Status.AtProvider = clusternative.TopicRAWObservation{
		Arn: &arnVal,
	}
	return cr
}

// ── topicARN tests ─────────────────────────────────────────────────────────────

func TestTopicARN_StandardAWS(t *testing.T) {
	const arn = "arn:aws:sns:us-east-1:123456789012:my-topic"
	cr := makeMinimalCR(arn)
	got := topicARN(cr)
	if got != arn {
		t.Errorf("topicARN() with standard AWS ARN: got %q, want %q", got, arn)
	}
}

func TestTopicARN_GovCloud(t *testing.T) {
	const arn = "arn:aws-us-gov:sns:us-gov-west-1:123456789012:my-topic"
	cr := makeMinimalCR(arn)
	got := topicARN(cr)
	if got != arn {
		t.Errorf("topicARN() with GovCloud ARN: got %q, want %q", got, arn)
	}
}

func TestTopicARN_China(t *testing.T) {
	const arn = "arn:aws-cn:sns:cn-north-1:123456789012:my-topic"
	cr := makeMinimalCR(arn)
	got := topicARN(cr)
	if got != arn {
		t.Errorf("topicARN() with China ARN: got %q, want %q", got, arn)
	}
}

func TestTopicARN_PlainName_ReturnsEmpty(t *testing.T) {
	cr := makeMinimalCR("my-topic")
	got := topicARN(cr)
	if got != "" {
		t.Errorf("topicARN() with plain name: got %q, want empty string", got)
	}
}

func TestTopicARN_PrefersAtProviderARN(t *testing.T) {
	const storedARN = "arn:aws:sns:us-east-1:123456789012:my-topic"
	cr := makeMinimalCRWithARN(storedARN)
	// External name is a plain name but atProvider.Arn is set
	got := topicARN(cr)
	if got != storedARN {
		t.Errorf("topicARN() should prefer atProvider.Arn: got %q, want %q", got, storedARN)
	}
}

// ── topicNameFromCR tests ──────────────────────────────────────────────────────

func TestTopicNameFromCR_StandardAWS(t *testing.T) {
	const arn = "arn:aws:sns:us-east-1:123456789012:my-topic"
	cr := makeMinimalCR(arn)
	got := topicNameFromCR(cr)
	if got != "my-topic" {
		t.Errorf("topicNameFromCR() with standard AWS ARN: got %q, want %q", got, "my-topic")
	}
}

func TestTopicNameFromCR_GovCloud(t *testing.T) {
	const arn = "arn:aws-us-gov:sns:us-gov-west-1:123456789012:my-topic"
	cr := makeMinimalCR(arn)
	got := topicNameFromCR(cr)
	if got != "my-topic" {
		t.Errorf("topicNameFromCR() with GovCloud ARN: got %q, want %q", got, "my-topic")
	}
}

func TestTopicNameFromCR_China(t *testing.T) {
	const arn = "arn:aws-cn:sns:cn-north-1:123456789012:my-topic"
	cr := makeMinimalCR(arn)
	got := topicNameFromCR(cr)
	if got != "my-topic" {
		t.Errorf("topicNameFromCR() with China ARN: got %q, want %q", got, "my-topic")
	}
}

func TestTopicNameFromCR_PlainExtName(t *testing.T) {
	cr := makeMinimalCR("my-custom-name")
	got := topicNameFromCR(cr)
	if got != "my-custom-name" {
		t.Errorf("topicNameFromCR() with plain ext name: got %q, want %q", got, "my-custom-name")
	}
}

func TestTopicNameFromCR_FallsBackToK8sName(t *testing.T) {
	// When external name equals the K8s object name, return the K8s name
	cr := makeMinimalCR("test-topic") // same as ObjectMeta.Name
	got := topicNameFromCR(cr)
	if got != "test-topic" {
		t.Errorf("topicNameFromCR() fallback to k8s name: got %q, want %q", got, "test-topic")
	}
}
