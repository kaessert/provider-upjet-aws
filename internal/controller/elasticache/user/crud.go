// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package user implements the shared CRUD logic for UserRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the UserCR interface.
package user

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awselasticache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	ectypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusternativev2 "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta2/native"
	nativehelper "github.com/upbound/provider-aws/v2/internal/native"
)

const (
	errDescribe  = "cannot describe ElastiCache User"
	errCreate    = "cannot create ElastiCache User"
	errUpdate    = "cannot modify ElastiCache User"
	errDelete    = "cannot delete ElastiCache User"
	errListTags  = "cannot list tags for ElastiCache User"
	errAddTags   = "cannot add tags to ElastiCache User"
	errDelTags   = "cannot remove tags from ElastiCache User"
	errGetSecret = "cannot get K8s secret for user passwords"
)

// ElastiCacheUserClient is the interface for AWS ElastiCache operations required
// by the user controller. Defined as an interface to enable mocking in unit tests;
// *awselasticache.Client satisfies it.
type ElastiCacheUserClient interface {
	DescribeUsers(ctx context.Context, params *awselasticache.DescribeUsersInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeUsersOutput, error)
	CreateUser(ctx context.Context, params *awselasticache.CreateUserInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateUserOutput, error)
	ModifyUser(ctx context.Context, params *awselasticache.ModifyUserInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyUserOutput, error)
	DeleteUser(ctx context.Context, params *awselasticache.DeleteUserInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteUserOutput, error)
	ListTagsForResource(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error)
	AddTagsToResource(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error)
	RemoveTagsFromResource(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error)
}

// UserCR abstracts over cluster-scoped and namespaced UserRAW types.
type UserCR interface {
	resource.Managed
	GetForProvider() *clusternativev2.UserRAWParameters
	GetInitProvider() *clusternativev2.UserRAWInitParameters
	GetAtProvider() clusternativev2.UserRAWObservation
	SetAtProvider(clusternativev2.UserRAWObservation)
}

// ExternalClient implements the shared CRUD logic for UserRAW resources.
type ExternalClient struct {
	// Client is the AWS ElastiCache SDK client (interface for testability).
	Client ElastiCacheUserClient
	// Kube is the Kubernetes client used to read Secret values for passwords.
	Kube client.Client
}

// Observe checks whether the external UserRAW resource exists and is up-to-date.
//
// Transitional state handling (spec §3):
//
//	"active"    → Available; proceed to isUpToDate
//	"modifying" → Unavailable; return UpToDate=true (prevent spurious Update)
//	"deleting"  → Deleting;   return UpToDate=true
//
// Without this, Observe calls isUpToDate on a "modifying" User where AWS
// changes haven't settled yet → may return false → triggers Update → AWS
// returns InvalidUserState.
//
//nolint:gocyclo
func (e *ExternalClient) Observe(ctx context.Context, cr UserCR) (managed.ExternalObservation, error) {
	extName := nativehelper.GetExternalName(cr)
	if extName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	resp, err := e.Client.DescribeUsers(ctx, &awselasticache.DescribeUsersInput{
		UserId: aws.String(extName),
	})
	if err != nil {
		if nativehelper.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, nativehelper.Wrap(err, errDescribe)
	}

	if len(resp.Users) == 0 {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	u := resp.Users[0]
	status := aws.ToString(u.Status)

	// Handle transitional states: skip isUpToDate to prevent spurious Updates.
	switch status {
	case "modifying":
		cr.SetConditions(xpv1.Unavailable())
		setAtProviderFromUser(cr, u, nil)
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	case "deleting":
		cr.SetConditions(xpv1.Deleting())
		setAtProviderFromUser(cr, u, nil)
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	}

	// Status is "active" (or unknown) — proceed to full observation.
	cr.SetConditions(xpv1.Available())

	// Fetch tags (non-fatal — tag errors should not fail Observe).
	var observedTags []ectypes.Tag
	if u.ARN != nil {
		tagsResp, tagErr := e.Client.ListTagsForResource(ctx, &awselasticache.ListTagsForResourceInput{
			ResourceName: u.ARN,
		})
		if tagErr == nil && tagsResp != nil {
			observedTags = tagsResp.TagList
		}
	}

	setAtProviderFromUser(cr, u, observedTags)

	// Late-initialize AWS-defaulted fields (spec §8).
	// Engine is set by AWS when a user is created — spec may omit it.
	// Without late-init, isUpToDate would compare "" (nil) vs "redis" every cycle.
	if nativehelper.LateInitializeStringPtr(&cr.GetForProvider().Engine, u.Engine) {
		return managed.ExternalObservation{
			ResourceExists:          true,
			ResourceUpToDate:        false,
			ResourceLateInitialized: true,
		}, nil
	}

	upToDate := isUpToDate(cr.GetForProvider(), u, observedTags)
	nativehelper.SetTestConditionIfAnnotated(cr, upToDate)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

// Create creates the external UserRAW resource.
// It reads passwords from whichever path is populated (top-level or nested),
// passes them to CreateUser, and publishes password hashes as connection details.
func (e *ExternalClient) Create(ctx context.Context, cr UserCR) (managed.ExternalCreation, error) {
	spec := cr.GetForProvider()
	extName := nativehelper.GetExternalName(cr)

	passwords, err := e.readPasswords(ctx, spec)
	if err != nil {
		return managed.ExternalCreation{}, nativehelper.Wrap(err, errCreate)
	}

	input := &awselasticache.CreateUserInput{
		UserId:             aws.String(extName),
		UserName:           spec.UserName,
		Engine:             spec.Engine,
		AccessString:       spec.AccessString,
		NoPasswordRequired: spec.NoPasswordRequired,
		Passwords:          passwords,
		Tags:               mapToTags(spec.Tags),
	}

	// If AuthenticationMode.Type is set (without passwords), pass authentication mode.
	if spec.AuthenticationMode != nil && spec.AuthenticationMode.Type != nil {
		authType := ectypes.InputAuthenticationType(aws.ToString(spec.AuthenticationMode.Type))
		input.AuthenticationMode = &ectypes.AuthenticationMode{
			Type: authType,
		}
		// Passwords already included directly in input.Passwords above.
	}

	_, err = e.Client.CreateUser(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, nativehelper.Wrap(err, errCreate)
	}

	// For ParameterAsIdentifier, the external name is already set from the CR's
	// external name annotation before Create is called. No SetExternalName needed.

	// Publish password hashes as connection details (sensitive field mechanism).
	connDetails := hashPasswordsToConnDetails(passwords)
	return managed.ExternalCreation{ConnectionDetails: connDetails}, nil
}

// Update updates the external UserRAW resource.
// It reads passwords from whichever path is populated and updates the user.
// Tags are synced separately.
func (e *ExternalClient) Update(ctx context.Context, cr UserCR) (managed.ExternalUpdate, error) {
	spec := cr.GetForProvider()
	extName := nativehelper.GetExternalName(cr)

	passwords, err := e.readPasswords(ctx, spec)
	if err != nil {
		return managed.ExternalUpdate{}, nativehelper.Wrap(err, errUpdate)
	}

	input := &awselasticache.ModifyUserInput{
		UserId:             aws.String(extName),
		AccessString:       spec.AccessString,
		NoPasswordRequired: spec.NoPasswordRequired,
		Passwords:          passwords,
	}

	// If AuthenticationMode.Type is set, pass authentication mode.
	if spec.AuthenticationMode != nil && spec.AuthenticationMode.Type != nil {
		authType := ectypes.InputAuthenticationType(aws.ToString(spec.AuthenticationMode.Type))
		input.AuthenticationMode = &ectypes.AuthenticationMode{
			Type: authType,
		}
	}

	_, err = e.Client.ModifyUser(ctx, input)
	if err != nil {
		return managed.ExternalUpdate{}, nativehelper.Wrap(err, errUpdate)
	}

	// Sync tags using the ARN from atProvider.
	arn := aws.ToString(cr.GetAtProvider().Arn)
	if arn != "" {
		if err := e.syncTags(ctx, arn, spec.Tags); err != nil {
			return managed.ExternalUpdate{}, err
		}
	}

	return managed.ExternalUpdate{}, nil
}

// Delete deletes the external UserRAW resource.
// Idempotent: UserNotFoundFault is treated as success.
func (e *ExternalClient) Delete(ctx context.Context, cr UserCR) (managed.ExternalDelete, error) {
	extName := nativehelper.GetExternalName(cr)

	_, err := e.Client.DeleteUser(ctx, &awselasticache.DeleteUserInput{
		UserId: aws.String(extName),
	})
	if err != nil {
		if nativehelper.IsNotFound(err) {
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, nativehelper.Wrap(err, errDelete)
	}
	return managed.ExternalDelete{}, nil
}

// ── Helpers ────────────────────────────────────────────────────────────────────

// readPasswords reads password values from whichever password path is populated.
// The spec supports two mutually-exclusive paths (spec §12):
//  1. Top-level: spec.forProvider.passwordsSecretRef
//  2. Nested:    spec.forProvider.authenticationMode.passwordsSecretRef
//
// Returns an empty slice if no password path is set (valid for IAM / no-password users).
func (e *ExternalClient) readPasswords(ctx context.Context, spec *clusternativev2.UserRAWParameters) ([]string, error) {
	// Check nested path first (authenticationMode.passwordsSecretRef).
	if spec.AuthenticationMode != nil && spec.AuthenticationMode.PasswordsSecretRef != nil {
		return e.readSecretSelectors(ctx, *spec.AuthenticationMode.PasswordsSecretRef)
	}
	// Fall back to top-level path.
	if spec.PasswordsSecretRef != nil {
		return e.readSecretSelectors(ctx, *spec.PasswordsSecretRef)
	}
	return nil, nil
}

// readSecretSelectors reads secret values for each SecretKeySelector in the slice.
func (e *ExternalClient) readSecretSelectors(ctx context.Context, refs []xpv1.SecretKeySelector) ([]string, error) {
	if e.Kube == nil {
		return nil, nil
	}
	passwords := make([]string, 0, len(refs))
	for i := range refs {
		ref := refs[i]
		s := &corev1.Secret{}
		if err := e.Kube.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, s); err != nil {
			return nil, nativehelper.Wrap(err, errGetSecret)
		}
		passwords = append(passwords, string(s.Data[ref.Key]))
	}
	return passwords, nil
}

// hashPasswordsToConnDetails hashes each password with SHA-256 and returns
// them as connection details with keys "password_hash_0", "password_hash_1", etc.
// This matches the TF sensitive field mechanism and ensures password values
// never appear in status.atProvider.
func hashPasswordsToConnDetails(passwords []string) managed.ConnectionDetails {
	if len(passwords) == 0 {
		return managed.ConnectionDetails{}
	}
	details := managed.ConnectionDetails{}
	for i, pw := range passwords {
		sum := sha256.Sum256([]byte(pw))
		details[fmt.Sprintf("password_hash_%d", i)] = []byte(fmt.Sprintf("%x", sum))
	}
	return details
}

// setAtProviderFromUser populates the atProvider observation from the AWS User state.
// NOTE: password values are NEVER included (AWS API does not return them).
func setAtProviderFromUser(cr UserCR, u ectypes.User, tags []ectypes.Tag) {
	obs := clusternativev2.UserRAWObservation{
		Arn:          u.ARN,
		AccessString: u.AccessString,
		Engine:       u.Engine,
		ID:           u.UserId,
		Status:       u.Status,
		UserName:     u.UserName,
		Tags:         tagsToMap(tags),
	}
	// Populate authenticationMode observation from Authentication.
	if u.Authentication != nil {
		passwordCount := float64(aws.ToInt32(u.Authentication.PasswordCount))
		obs.AuthenticationMode = &clusternativev2.AuthenticationModeRAWObservation{
			PasswordCount: &passwordCount,
			Type:          aws.String(string(u.Authentication.Type)),
		}
	}
	cr.SetAtProvider(obs)
}

// isUpToDate returns true when the spec is in sync with the observed AWS state.
// Passwords are write-only at the AWS API level — they are never compared.
func isUpToDate(spec *clusternativev2.UserRAWParameters, u ectypes.User, observedTags []ectypes.Tag) bool {
	// Check AccessString.
	if aws.ToString(spec.AccessString) != aws.ToString(u.AccessString) {
		return false
	}

	// Check Engine: guard against nil — AWS may default the engine value.
	// After late-initialization the spec will be populated; this nil check prevents
	// a transient spurious update before late-init has run.
	if spec.Engine != nil && aws.ToString(spec.Engine) != aws.ToString(u.Engine) {
		return false
	}

	// Check tags.
	if !tagsUpToDate(spec.Tags, observedTags) {
		return false
	}

	return true
}

// syncTags reconciles desired tags on the user using its ARN.
func (e *ExternalClient) syncTags(ctx context.Context, arn string, desired map[string]*string) error {
	tagsResp, err := e.Client.ListTagsForResource(ctx, &awselasticache.ListTagsForResourceInput{
		ResourceName: aws.String(arn),
	})
	if err != nil {
		return nativehelper.Wrap(err, errListTags)
	}

	observed := tagsResp.TagList
	toAdd, toRemove := diffTags(desired, observed)

	if len(toAdd) > 0 {
		if _, err := e.Client.AddTagsToResource(ctx, &awselasticache.AddTagsToResourceInput{
			ResourceName: aws.String(arn),
			Tags:         toAdd,
		}); err != nil {
			return nativehelper.Wrap(err, errAddTags)
		}
	}

	if len(toRemove) > 0 {
		keys := make([]string, 0, len(toRemove))
		for _, t := range toRemove {
			keys = append(keys, aws.ToString(t.Key))
		}
		if _, err := e.Client.RemoveTagsFromResource(ctx, &awselasticache.RemoveTagsFromResourceInput{
			ResourceName: aws.String(arn),
			TagKeys:      keys,
		}); err != nil {
			return nativehelper.Wrap(err, errDelTags)
		}
	}

	return nil
}

// tagsUpToDate returns true when the desired spec tags match the observed AWS tags.
func tagsUpToDate(specTags map[string]*string, observed []ectypes.Tag) bool {
	if len(specTags) != len(observed) {
		return false
	}
	observedMap := tagsToMap(observed)
	for k, v := range specTags {
		obsV, ok := observedMap[k]
		if !ok {
			return false
		}
		if aws.ToString(v) != aws.ToString(obsV) {
			return false
		}
	}
	return true
}

// diffTags computes tags to add and tags to remove.
func diffTags(desired map[string]*string, observed []ectypes.Tag) (toAdd []ectypes.Tag, toRemove []ectypes.Tag) {
	observedMap := tagsToMap(observed)

	// Tags to add or update.
	for k, v := range desired {
		tagKey := k
		obsVal, exists := observedMap[tagKey]
		if !exists || aws.ToString(v) != aws.ToString(obsVal) {
			toAdd = append(toAdd, ectypes.Tag{Key: aws.String(tagKey), Value: v})
		}
	}

	// Tags to remove (exist in AWS but not in desired).
	for _, t := range observed {
		if _, exists := desired[aws.ToString(t.Key)]; !exists {
			toRemove = append(toRemove, t)
		}
	}

	return toAdd, toRemove
}

// mapToTags converts a map[string]*string to a []ectypes.Tag slice.
func mapToTags(m map[string]*string) []ectypes.Tag {
	tags := make([]ectypes.Tag, 0, len(m))
	for k, v := range m {
		tagKey := k
		tags = append(tags, ectypes.Tag{Key: aws.String(tagKey), Value: v})
	}
	return tags
}

// tagsToMap converts a []ectypes.Tag slice to a map[string]*string.
func tagsToMap(tags []ectypes.Tag) map[string]*string {
	m := make(map[string]*string, len(tags))
	for _, t := range tags {
		if t.Key != nil {
			m[*t.Key] = t.Value
		}
	}
	return m
}
