// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	awspolicy "github.com/hashicorp/awspolicyequivalence"
)

// PoliciesAreEquivalent returns true if two IAM policy JSON strings are
// semantically equivalent (same effect, action, principal, resource,
// condition), even if they differ syntactically (key order, whitespace,
// encoding). Returns false and an error if either string is not a valid JSON
// policy.
func PoliciesAreEquivalent(policy1, policy2 string) (bool, error) {
	return awspolicy.PoliciesAreEquivalent(policy1, policy2)
}

// PolicyNeedsUpdate returns true if the desired policy differs semantically
// from the observed policy. Returns true conservatively when either policy
// cannot be parsed, to trigger an update rather than silently accepting a
// potentially diverged state. Native controllers with a policy field MUST use
// this function instead of reflect.DeepEqual or string comparison to avoid
// infinite reconcile loops caused by AWS normalising IAM policy JSON.
func PolicyNeedsUpdate(desired, observed string) bool {
	equivalent, err := PoliciesAreEquivalent(desired, observed)
	if err != nil {
		// Conservative: treat a parse failure as "needs update" so the
		// controller writes the desired policy and lets AWS validate it.
		return true
	}
	return !equivalent
}
