// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package native contains cluster-scoped native SQS resource types.
//
// NOTE: QueuePolicyRAW implements LegacyManaged (not ModernManaged) because the
// TF QueuePolicy example includes namespace in writeConnectionSecretToRef, and
// the migration rules require RAW type YAML schemas to mirror TF counterparts.
//
// LegacyManaged requires:
//   - ConnectionSecretWriterTo (SecretReference with namespace) ← writeConnectionSecretToRef
//   - ProviderConfigReferencer (Reference with just name, no kind) ← providerConfigRef
//   - Orphanable (DeletionPolicy) ← deletionPolicy
//
// Because QueuePolicyRAWSpec doesn't embed xpv2.ManagedResourceSpec (which would
// force LocalSecretReference and TypedProviderConfigReferencer for ModernManaged),
// angryjet cannot auto-generate these methods.
// They are maintained manually in this file.
package native

import (
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	resource "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
)

// GetCondition of this QueuePolicyRAW.
func (mg *QueuePolicyRAW) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// SetConditions of this QueuePolicyRAW.
func (mg *QueuePolicyRAW) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// GetManagementPolicies of this QueuePolicyRAW.
func (mg *QueuePolicyRAW) GetManagementPolicies() xpv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// SetManagementPolicies of this QueuePolicyRAW.
func (mg *QueuePolicyRAW) SetManagementPolicies(r xpv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = r
}

// GetDeletionPolicy of this QueuePolicyRAW (implements Orphanable, required by LegacyManaged).
func (mg *QueuePolicyRAW) GetDeletionPolicy() xpv1.DeletionPolicy {
	if mg.Spec.DeletionPolicy == nil {
		return xpv1.DeletionDelete
	}
	return *mg.Spec.DeletionPolicy
}

// SetDeletionPolicy of this QueuePolicyRAW (implements Orphanable, required by LegacyManaged).
func (mg *QueuePolicyRAW) SetDeletionPolicy(p xpv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = &p
}

// GetProviderConfigReference of this QueuePolicyRAW (implements ProviderConfigReferencer,
// required by LegacyManaged — uses *xpv1.Reference, not *xpv1.ProviderConfigReference).
func (mg *QueuePolicyRAW) GetProviderConfigReference() *xpv1.Reference {
	return mg.Spec.ProviderConfigReference
}

// SetProviderConfigReference of this QueuePolicyRAW (implements ProviderConfigReferencer,
// required by LegacyManaged — uses *xpv1.Reference, not *xpv1.ProviderConfigReference).
func (mg *QueuePolicyRAW) SetProviderConfigReference(r *xpv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

// GetWriteConnectionSecretToReference of this QueuePolicyRAW.
// Returns a SecretReference (with namespace) to support cluster-scoped
// connection secret writing to a specific namespace.
// Implements ConnectionSecretWriterTo (LegacyManaged).
func (mg *QueuePolicyRAW) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetWriteConnectionSecretToReference of this QueuePolicyRAW.
func (mg *QueuePolicyRAW) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

// GetItems of this QueuePolicyRAWList.
func (l *QueuePolicyRAWList) GetItems() []resource.Managed {
	items := make([]resource.Managed, len(l.Items))
	for i := range l.Items {
		items[i] = &l.Items[i]
	}
	return items
}
