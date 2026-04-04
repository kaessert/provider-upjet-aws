// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// +kubebuilder:object:generate=true
// +groupName=secretsmanager.aws.upbound.io
// +versionName=v1beta1

package native

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	// CRDGroup is the API Group for all native secretsmanager types in this package.
	CRDGroup = "secretsmanager.aws.upbound.io"

	// CRDVersion is the API Version for all native secretsmanager types in this package.
	CRDVersion = "v1beta1"
)

var (
	// CRDGroupVersion is the API Group Version used to register the native objects.
	CRDGroupVersion = schema.GroupVersion{Group: CRDGroup, Version: CRDVersion}

	// SchemeBuilder is used to add native go types to the GroupVersionKind scheme.
	SchemeBuilder = &scheme.Builder{GroupVersion: CRDGroupVersion}

	// AddToScheme adds the native types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)
