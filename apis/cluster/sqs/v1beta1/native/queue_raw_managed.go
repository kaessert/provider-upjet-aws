// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package native contains cluster-scoped native SQS resource types.
//
// NOTE: QueueRAW implements LegacyManaged (not ModernManaged) because the TF
// Queue example includes namespace in writeConnectionSecretToRef, and the
// migration rules require RAW type YAML schemas to mirror TF counterparts.
//
// LegacyManaged requires:
//   - ConnectionSecretWriterTo (SecretReference with namespace) ← writeConnectionSecretToRef
//   - ProviderConfigReferencer (Reference with just name, no kind) ← providerConfigRef
//   - Orphanable (DeletionPolicy) ← deletionPolicy
//
// Because QueueRAWSpec doesn't embed xpv2.ManagedResourceSpec (which would
// force LocalSecretReference and TypedProviderConfigReferencer for ModernManaged),
// angryjet cannot auto-generate these methods.
// They are maintained manually in this file.
package native

import (
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	resource "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
)

// GetCondition of this QueueRAW.
func (mg *QueueRAW) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// SetConditions of this QueueRAW.
func (mg *QueueRAW) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// GetManagementPolicies of this QueueRAW.
func (mg *QueueRAW) GetManagementPolicies() xpv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// SetManagementPolicies of this QueueRAW.
func (mg *QueueRAW) SetManagementPolicies(r xpv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = r
}

// GetDeletionPolicy of this QueueRAW (implements Orphanable, required by LegacyManaged).
func (mg *QueueRAW) GetDeletionPolicy() xpv1.DeletionPolicy {
	if mg.Spec.DeletionPolicy == nil {
		return xpv1.DeletionDelete
	}
	return *mg.Spec.DeletionPolicy
}

// SetDeletionPolicy of this QueueRAW (implements Orphanable, required by LegacyManaged).
func (mg *QueueRAW) SetDeletionPolicy(p xpv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = &p
}

// GetProviderConfigReference of this QueueRAW (implements ProviderConfigReferencer,
// required by LegacyManaged — uses *xpv1.Reference, not *xpv1.ProviderConfigReference).
func (mg *QueueRAW) GetProviderConfigReference() *xpv1.Reference {
	return mg.Spec.ProviderConfigReference
}

// SetProviderConfigReference of this QueueRAW (implements ProviderConfigReferencer,
// required by LegacyManaged — uses *xpv1.Reference, not *xpv1.ProviderConfigReference).
func (mg *QueueRAW) SetProviderConfigReference(r *xpv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

// GetWriteConnectionSecretToReference of this QueueRAW.
// Returns a SecretReference (with namespace) to support cluster-scoped
// connection secret writing to a specific namespace.
// Implements ConnectionSecretWriterTo (LegacyManaged).
func (mg *QueueRAW) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetWriteConnectionSecretToReference of this QueueRAW.
func (mg *QueueRAW) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

// GetItems of this QueueRAWList.
func (l *QueueRAWList) GetItems() []resource.Managed {
	items := make([]resource.Managed, len(l.Items))
	for i := range l.Items {
		items[i] = &l.Items[i]
	}
	return items
}
