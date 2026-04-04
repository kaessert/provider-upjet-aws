// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// NOTE: QueueRedriveAllowPolicyRAW implements LegacyManaged (not ModernManaged)
// because the TF QueueRedriveAllowPolicy example includes namespace in
// writeConnectionSecretToRef, and the migration rules require RAW type YAML
// schemas to mirror TF counterparts.
//
// LegacyManaged requires:
//   - ConnectionSecretWriterTo (SecretReference with namespace) ← writeConnectionSecretToRef
//   - ProviderConfigReferencer (Reference with just name, no kind) ← providerConfigRef
//   - Orphanable (DeletionPolicy) ← deletionPolicy
//
// Because QueueRedriveAllowPolicyRAWSpec doesn't embed xpv2.ManagedResourceSpec
// (which would force LocalSecretReference and TypedProviderConfigReferencer for
// ModernManaged), angryjet cannot auto-generate these methods.
// They are maintained manually in this file.
package native

import (
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	resource "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
)

// GetCondition of this QueueRedriveAllowPolicyRAW.
func (mg *QueueRedriveAllowPolicyRAW) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	return mg.Status.GetCondition(ct)
}

// SetConditions of this QueueRedriveAllowPolicyRAW.
func (mg *QueueRedriveAllowPolicyRAW) SetConditions(c ...xpv1.Condition) {
	mg.Status.SetConditions(c...)
}

// GetManagementPolicies of this QueueRedriveAllowPolicyRAW.
func (mg *QueueRedriveAllowPolicyRAW) GetManagementPolicies() xpv1.ManagementPolicies {
	return mg.Spec.ManagementPolicies
}

// SetManagementPolicies of this QueueRedriveAllowPolicyRAW.
func (mg *QueueRedriveAllowPolicyRAW) SetManagementPolicies(r xpv1.ManagementPolicies) {
	mg.Spec.ManagementPolicies = r
}

// GetDeletionPolicy of this QueueRedriveAllowPolicyRAW (implements Orphanable,
// required by LegacyManaged).
func (mg *QueueRedriveAllowPolicyRAW) GetDeletionPolicy() xpv1.DeletionPolicy {
	if mg.Spec.DeletionPolicy == nil {
		return xpv1.DeletionDelete
	}
	return *mg.Spec.DeletionPolicy
}

// SetDeletionPolicy of this QueueRedriveAllowPolicyRAW (implements Orphanable,
// required by LegacyManaged).
func (mg *QueueRedriveAllowPolicyRAW) SetDeletionPolicy(p xpv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = &p
}

// GetProviderConfigReference of this QueueRedriveAllowPolicyRAW (implements
// ProviderConfigReferencer, required by LegacyManaged — uses *xpv1.Reference,
// not *xpv1.ProviderConfigReference).
func (mg *QueueRedriveAllowPolicyRAW) GetProviderConfigReference() *xpv1.Reference {
	return mg.Spec.ProviderConfigReference
}

// SetProviderConfigReference of this QueueRedriveAllowPolicyRAW (implements
// ProviderConfigReferencer, required by LegacyManaged — uses *xpv1.Reference,
// not *xpv1.ProviderConfigReference).
func (mg *QueueRedriveAllowPolicyRAW) SetProviderConfigReference(r *xpv1.Reference) {
	mg.Spec.ProviderConfigReference = r
}

// GetWriteConnectionSecretToReference of this QueueRedriveAllowPolicyRAW.
// Returns a SecretReference (with namespace) to support cluster-scoped
// connection secret writing to a specific namespace.
// Implements ConnectionSecretWriterTo (LegacyManaged).
func (mg *QueueRedriveAllowPolicyRAW) GetWriteConnectionSecretToReference() *xpv1.SecretReference {
	return mg.Spec.WriteConnectionSecretToReference
}

// SetWriteConnectionSecretToReference of this QueueRedriveAllowPolicyRAW.
func (mg *QueueRedriveAllowPolicyRAW) SetWriteConnectionSecretToReference(r *xpv1.SecretReference) {
	mg.Spec.WriteConnectionSecretToReference = r
}

// GetItems of this QueueRedriveAllowPolicyRAWList.
func (l *QueueRedriveAllowPolicyRAWList) GetItems() []resource.Managed {
	items := make([]resource.Managed, len(l.Items))
	for i := range l.Items {
		items[i] = &l.Items[i]
	}
	return items
}
