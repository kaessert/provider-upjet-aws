// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
)

// BucketPolicyRAWParameters defines the configuration parameters for
// a native (non-Terraform) S3 Bucket Policy resource.
//
// This type is used to validate the reference resolution pipeline for native
// types. It demonstrates the correct extractor annotation pattern for RAW
// types that reference other managed resources.
type BucketPolicyRAWParameters struct {
	// Region is the AWS region where the bucket resides.
	// +optional
	Region *string `json:"region,omitempty"`

	// Bucket is the name of the S3 bucket to which the policy applies.
	//
	// This field demonstrates the correct +crossplane:generate:reference annotation
	// pattern for native RAW types:
	//   - Reference type points to the existing TF-backed Bucket type
	//   - Extractor uses internal/native.ExtractResourceID() which reads
	//     status.atProvider.id via fieldpath WITHOUT casting to upjet's
	//     Terraformed interface (the native-safe alternative to the upjet
	//     TerraformID extractor which would panic for non-Terraformed types).
	//
	// +optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/s3/v1beta1.Bucket
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractResourceID()
	Bucket *string `json:"bucket,omitempty"`

	// BucketRef is a reference to a Bucket to populate bucket.
	// +optional
	BucketRef *xpv1.Reference `json:"bucketRef,omitempty"`

	// BucketSelector selects a reference to a Bucket to populate bucket.
	// +optional
	BucketSelector *xpv1.Selector `json:"bucketSelector,omitempty"`

	// Policy is the JSON-encoded IAM bucket policy document.
	Policy *string `json:"policy"`
}

// BucketPolicyRAWInitParameters contains the fields that can be set during
// initial resource creation. These fields are merged into ForProvider on create.
type BucketPolicyRAWInitParameters struct {
	// Region is the AWS region where the bucket resides.
	// +optional
	Region *string `json:"region,omitempty"`

	// Bucket is the name of the S3 bucket to which the policy applies.
	// +optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/s3/v1beta1.Bucket
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractResourceID()
	Bucket *string `json:"bucket,omitempty"`

	// BucketRef is a reference to a Bucket to populate bucket.
	// +optional
	BucketRef *xpv1.Reference `json:"bucketRef,omitempty"`

	// BucketSelector selects a reference to a Bucket to populate bucket.
	// +optional
	BucketSelector *xpv1.Selector `json:"bucketSelector,omitempty"`

	// Policy is the JSON-encoded IAM bucket policy document.
	// +optional
	Policy *string `json:"policy,omitempty"`
}

// BucketPolicyRAWObservation defines the observed state fields that are
// populated by the controller during reconciliation.
type BucketPolicyRAWObservation struct {
	// ID is the external name / identifier of the S3 bucket policy,
	// which is the bucket name for S3 bucket policies.
	ID *string `json:"id,omitempty"`
}

// BucketPolicyRAWSpec defines the desired state of BucketPolicyRAW.
type BucketPolicyRAWSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider BucketPolicyRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider BucketPolicyRAWInitParameters `json:"initProvider,omitempty"`
}

// BucketPolicyRAWStatus defines the observed state of BucketPolicyRAW.
type BucketPolicyRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider BucketPolicyRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// BucketPolicyRAW is the native (non-Terraform) Schema for S3 bucket policies.
// It is a RAW type used to validate the reference resolution pipeline during
// Phase 0 of the Terraform removal migration.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type BucketPolicyRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketPolicyRAWSpec   `json:"spec"`
	Status BucketPolicyRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BucketPolicyRAWList contains a list of BucketPolicyRAW resources.
type BucketPolicyRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BucketPolicyRAW `json:"items"`
}

// Repository type metadata for BucketPolicyRAW.
var (
	BucketPolicyRAW_Kind             = "BucketPolicyRAW"
	BucketPolicyRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: BucketPolicyRAW_Kind}.String()
	BucketPolicyRAW_KindAPIVersion   = BucketPolicyRAW_Kind + "." + CRDGroupVersion.String()
	BucketPolicyRAW_GroupVersionKind = CRDGroupVersion.WithKind(BucketPolicyRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&BucketPolicyRAW{}, &BucketPolicyRAWList{})
}
