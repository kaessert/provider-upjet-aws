// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpfake "github.com/crossplane/crossplane-runtime/v2/pkg/resource/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetTestConditionIfAnnotated(t *testing.T) {
	tests := []struct {
		name      string
		annotated bool
		upToDate  bool
		wantCond  bool
	}{
		{
			name:      "annotated and up-to-date: sets Test=True",
			annotated: true,
			upToDate:  true,
			wantCond:  true,
		},
		{
			name:      "annotated but not up-to-date: no condition",
			annotated: true,
			upToDate:  false,
			wantCond:  false,
		},
		{
			name:      "not annotated and up-to-date: no condition",
			annotated: false,
			upToDate:  true,
			wantCond:  false,
		},
		{
			name:      "not annotated and not up-to-date: no condition",
			annotated: false,
			upToDate:  false,
			wantCond:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mg := &xpfake.Managed{
				ObjectMeta: metav1.ObjectMeta{},
			}
			if tt.annotated {
				mg.SetAnnotations(map[string]string{AnnotationKeyTestResource: "true"})
			}

			SetTestConditionIfAnnotated(mg, tt.upToDate)

			cond := mg.GetCondition(conditionTypeTest)
			hasTestCond := cond != (xpv1.Condition{}) && cond.Status == corev1.ConditionTrue

			if hasTestCond != tt.wantCond {
				t.Errorf("SetTestConditionIfAnnotated(%v, %v): hasTestCond=%v, want %v",
					tt.annotated, tt.upToDate, hasTestCond, tt.wantCond)
			}
		})
	}
}
