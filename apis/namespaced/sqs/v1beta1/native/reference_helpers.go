// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
)

// namespacedRefToRef converts a NamespacedReference (used in namespaced types)
// to a plain Reference (used in cluster-scoped types). The Namespace field is
// intentionally dropped since cluster-scoped resources have no namespace.
func namespacedRefToRef(nr *xpv1.NamespacedReference) *xpv1.Reference {
	if nr == nil {
		return nil
	}
	r := &xpv1.Reference{Name: nr.Name}
	if nr.Policy != nil {
		p := *nr.Policy
		r.Policy = &p
	}
	return r
}

// refToNamespacedRef converts a plain Reference (used in cluster-scoped types)
// to a NamespacedReference (used in namespaced types). The Namespace field is
// left empty because namespaced SQS resources reference within their own
// namespace and the reference resolver fills in the namespace at resolve time.
func refToNamespacedRef(r *xpv1.Reference) *xpv1.NamespacedReference {
	if r == nil {
		return nil
	}
	nr := &xpv1.NamespacedReference{Name: r.Name}
	if r.Policy != nil {
		p := *r.Policy
		nr.Policy = &p
	}
	return nr
}

// namespacedSelectorToSelector converts a NamespacedSelector to a plain
// Selector. The Namespace field is intentionally dropped.
func namespacedSelectorToSelector(ns *xpv1.NamespacedSelector) *xpv1.Selector {
	if ns == nil {
		return nil
	}
	s := &xpv1.Selector{
		MatchControllerRef: ns.MatchControllerRef,
	}
	if ns.MatchLabels != nil {
		s.MatchLabels = make(map[string]string, len(ns.MatchLabels))
		for k, v := range ns.MatchLabels {
			s.MatchLabels[k] = v
		}
	}
	if ns.Policy != nil {
		p := *ns.Policy
		s.Policy = &p
	}
	return s
}

// selectorToNamespacedSelector converts a plain Selector to a
// NamespacedSelector. The Namespace field is left empty.
func selectorToNamespacedSelector(s *xpv1.Selector) *xpv1.NamespacedSelector {
	if s == nil {
		return nil
	}
	ns := &xpv1.NamespacedSelector{
		MatchControllerRef: s.MatchControllerRef,
	}
	if s.MatchLabels != nil {
		ns.MatchLabels = make(map[string]string, len(s.MatchLabels))
		for k, v := range s.MatchLabels {
			ns.MatchLabels[k] = v
		}
	}
	if s.Policy != nil {
		p := *s.Policy
		ns.Policy = &p
	}
	return ns
}
