// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// AnnotationKeyTestResource is the upjet compatibility annotation that marks a
	// managed resource as a test resource.  When set to "true" and the resource is
	// observed to be up-to-date, the controller should set a "Test" condition so that
	// uptest's assertion `status.conditions[?type == 'Test'].status == "True"` passes.
	AnnotationKeyTestResource = "upjet.upbound.io/test"

	// conditionTypeTest is the condition type checked by uptest.
	conditionTypeTest xpv1.ConditionType = "Test"

	// conditionReasonUpToDate is the reason set when the resource is up-to-date.
	conditionReasonUpToDate xpv1.ConditionReason = "UpToDate"
)

// SetTestConditionIfAnnotated sets a "Test=True" condition on mg when ALL of the
// following are true:
//  1. mg carries the annotation `upjet.upbound.io/test=true`
//  2. upToDate is true
//
// This mirrors the behaviour of upjet's SetUpToDateCondition so that uptest's
// default assertion (`--default-conditions="Test"`) passes for native controllers.
func SetTestConditionIfAnnotated(mg resource.Managed, upToDate bool) {
	if !upToDate {
		return
	}
	if mg.GetAnnotations()[AnnotationKeyTestResource] != "true" {
		return
	}
	mg.SetConditions(xpv1.Condition{
		Type:               conditionTypeTest,
		Status:             corev1.ConditionTrue,
		LastTransitionTime: metav1.Now(),
		Reason:             conditionReasonUpToDate,
	})
}
