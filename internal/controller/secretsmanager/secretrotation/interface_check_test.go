package secretrotation_test

import (
	"testing"

	clusternative1 "github.com/upbound/provider-aws/v2/apis/cluster/secretsmanager/v1beta1/native"
	clusternative2 "github.com/upbound/provider-aws/v2/apis/cluster/secretsmanager/v1beta2/native"
	namespacednative "github.com/upbound/provider-aws/v2/apis/namespaced/secretsmanager/v1beta1/native"
	secretrotation "github.com/upbound/provider-aws/v2/internal/controller/secretsmanager/secretrotation"
)

// Compile-time interface checks.
var _ secretrotation.SecretRotationCR = (*clusternative1.SecretRotationRAW)(nil)
var _ secretrotation.SecretRotationCR = (*clusternative2.SecretRotationRAW)(nil)
var _ secretrotation.SecretRotationCR = (*namespacednative.SecretRotationRAW)(nil)

func TestInterfaceCheck(t *testing.T) {
	// This test is intentionally empty — the compile-time checks above are sufficient.
}
