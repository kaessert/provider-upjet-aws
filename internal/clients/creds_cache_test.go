// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	namespacedv1beta1 "github.com/upbound/provider-aws/v2/apis/namespaced/v1beta1"
)

// writeTokenFile writes content to a temp file and returns its path.
func writeTokenFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "token")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("cannot write token file: %v", err)
	}
	return path
}

// newIRSAProviderConfig returns a minimal ClusterProviderConfig using IRSA auth.
func newIRSAProviderConfig(uid types.UID, gen int64) *namespacedv1beta1.ClusterProviderConfig {
	return &namespacedv1beta1.ClusterProviderConfig{
		ObjectMeta: metav1.ObjectMeta{
			UID:        uid,
			Generation: gen,
		},
		Spec: namespacedv1beta1.ProviderConfigSpec{
			Credentials: namespacedv1beta1.ProviderCredentials{
				Source: v1.CredentialsSource(authKeyIRSA),
			},
		},
	}
}

// newNonIRSAProviderConfig returns a minimal ClusterProviderConfig using Secret auth.
func newNonIRSAProviderConfig(uid types.UID) *namespacedv1beta1.ClusterProviderConfig {
	return &namespacedv1beta1.ClusterProviderConfig{
		ObjectMeta: metav1.ObjectMeta{
			UID: uid,
		},
		Spec: namespacedv1beta1.ProviderConfigSpec{
			Credentials: namespacedv1beta1.ProviderCredentials{
				Source: v1.CredentialsSource("Secret"),
			},
		},
	}
}

// newTestCredsCache creates a new *aws.CredentialsCache wrapping a static provider.
func newTestCredsCache(accessKeyID string) *aws.CredentialsCache {
	return aws.NewCredentialsCache(aws.CredentialsProviderFunc(func(_ context.Context) (aws.Credentials, error) {
		return aws.Credentials{AccessKeyID: accessKeyID}, nil
	}))
}

// TestGetCachedCredentialsProvider_NonIRSA checks that non-IRSA credentials
// are returned unchanged without cache interaction.
func TestGetCachedCredentialsProvider_NonIRSA(t *testing.T) {
	cache := NewAWSCredentialsProviderCache()
	pc := newNonIRSAProviderConfig("uid-1")
	provider := newTestCredsCache("AKIA_NON_IRSA")

	got, err := cache.GetCachedCredentialsProvider(pc, "us-east-1", provider)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// For non-IRSA the original provider must be returned unchanged.
	if got != provider {
		t.Error("expected original provider to be returned for non-IRSA auth")
	}
	if len(cache.cache) != 0 {
		t.Errorf("expected empty cache for non-IRSA, got %d entries", len(cache.cache))
	}
}

// TestGetCachedCredentialsProvider_NotCredentialsCache checks that IRSA auth
// with a non-*aws.CredentialsCache provider is returned unchanged (not cached).
func TestGetCachedCredentialsProvider_NotCredentialsCache(t *testing.T) {
	tokenFile := writeTokenFile(t, "tok1")
	t.Setenv("AWS_WEB_IDENTITY_TOKEN_FILE", tokenFile)
	t.Setenv("AWS_ROLE_ARN", "arn:aws:iam::123456789:role/Role1")

	cache := NewAWSCredentialsProviderCache()
	pc := newIRSAProviderConfig("uid-2", 1)

	// plain CredentialsProviderFunc — not an *aws.CredentialsCache
	provider := aws.CredentialsProviderFunc(func(_ context.Context) (aws.Credentials, error) {
		return aws.Credentials{AccessKeyID: "AKIA"}, nil
	})

	got, err := cache.GetCachedCredentialsProvider(pc, "us-east-1", provider)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Must return original without storing anything.
	if len(cache.cache) != 0 {
		t.Errorf("expected empty cache for non-CredentialsCache provider, got %d entries", len(cache.cache))
	}
	if got == nil {
		t.Error("expected non-nil result")
	}
}

// TestGetCachedCredentialsProvider_IRSACacheMissThenHit verifies that:
//   - First call (miss) stores the provider.
//   - Second call (hit) returns the SAME *aws.CredentialsCache pointer.
func TestGetCachedCredentialsProvider_IRSACacheMissThenHit(t *testing.T) {
	tokenFile := writeTokenFile(t, "my-irsa-token")
	t.Setenv("AWS_WEB_IDENTITY_TOKEN_FILE", tokenFile)
	t.Setenv("AWS_ROLE_ARN", "arn:aws:iam::123456789:role/Role2")

	cache := NewAWSCredentialsProviderCache()
	pc := newIRSAProviderConfig("uid-3", 1)
	provider1 := newTestCredsCache("AKIA1")
	provider2 := newTestCredsCache("AKIA2") // different instance, same key

	// First call — cache miss: stores provider1.
	got1, err := cache.GetCachedCredentialsProvider(pc, "us-east-1", provider1)
	if err != nil {
		t.Fatalf("first call: unexpected error: %v", err)
	}
	if got1 != provider1 {
		t.Errorf("first call: expected provider1 to be stored and returned, got different pointer")
	}
	if len(cache.cache) != 1 {
		t.Errorf("expected 1 cache entry after miss, got %d", len(cache.cache))
	}

	// Second call — cache hit: should return provider1 (the cached one), not provider2.
	got2, err := cache.GetCachedCredentialsProvider(pc, "us-east-1", provider2)
	if err != nil {
		t.Fatalf("second call: unexpected error: %v", err)
	}
	if got2 != provider1 {
		t.Errorf("second call: expected cached provider1, got different pointer")
	}
	if len(cache.cache) != 1 {
		t.Errorf("expected still 1 cache entry after hit, got %d", len(cache.cache))
	}
}

// TestGetCachedCredentialsProvider_DifferentRegions verifies that different
// regions produce independent cache entries.
func TestGetCachedCredentialsProvider_DifferentRegions(t *testing.T) {
	tokenFile := writeTokenFile(t, "regiontoken")
	t.Setenv("AWS_WEB_IDENTITY_TOKEN_FILE", tokenFile)
	t.Setenv("AWS_ROLE_ARN", "arn:aws:iam::123456789:role/Role3")

	cache := NewAWSCredentialsProviderCache()
	pc := newIRSAProviderConfig("uid-4", 1)
	provider1 := newTestCredsCache("AKIA_east")
	provider2 := newTestCredsCache("AKIA_west")

	got1, err := cache.GetCachedCredentialsProvider(pc, "us-east-1", provider1)
	if err != nil {
		t.Fatalf("us-east-1 call: %v", err)
	}
	got2, err := cache.GetCachedCredentialsProvider(pc, "us-west-2", provider2)
	if err != nil {
		t.Fatalf("us-west-2 call: %v", err)
	}
	if got1 == got2 {
		t.Error("expected different cache entries for different regions")
	}
	if len(cache.cache) != 2 {
		t.Errorf("expected 2 cache entries for 2 regions, got %d", len(cache.cache))
	}
}

// TestGetCachedCredentialsProvider_DifferentPCs verifies that different
// ProviderConfigs produce independent cache entries.
func TestGetCachedCredentialsProvider_DifferentPCs(t *testing.T) {
	tokenFile := writeTokenFile(t, "pctoken")
	t.Setenv("AWS_WEB_IDENTITY_TOKEN_FILE", tokenFile)
	t.Setenv("AWS_ROLE_ARN", "arn:aws:iam::123456789:role/Role4")

	cache := NewAWSCredentialsProviderCache()
	pc1 := newIRSAProviderConfig("uid-pc1", 1)
	pc2 := newIRSAProviderConfig("uid-pc2", 1)
	provider1 := newTestCredsCache("AKIA_pc1")
	provider2 := newTestCredsCache("AKIA_pc2")

	got1, err := cache.GetCachedCredentialsProvider(pc1, "us-east-1", provider1)
	if err != nil {
		t.Fatalf("pc1 call: %v", err)
	}
	got2, err := cache.GetCachedCredentialsProvider(pc2, "us-east-1", provider2)
	if err != nil {
		t.Fatalf("pc2 call: %v", err)
	}
	if got1 == got2 {
		t.Error("expected different cache entries for different PCs")
	}
}

// TestGetCachedCredentialsProvider_LRUEviction verifies that when the cache
// exceeds maxSize, the least-recently-accessed entry is evicted.
func TestGetCachedCredentialsProvider_LRUEviction(t *testing.T) {
	tokenFile := writeTokenFile(t, "evicttoken")
	t.Setenv("AWS_WEB_IDENTITY_TOKEN_FILE", tokenFile)
	t.Setenv("AWS_ROLE_ARN", "arn:aws:iam::123456789:role/Role5")

	// Pre-populate cache with 2 entries where entry "old" has oldest access.
	oldEntry := &awsCredentialsProviderCacheEntry{awsCredCache: newTestCredsCache("old")}
	oldEntry.accessedAt.Store(time.Now().Add(-time.Hour))
	oldEntry.accountID.Store("")
	newEntry := &awsCredentialsProviderCacheEntry{awsCredCache: newTestCredsCache("new")}
	newEntry.accessedAt.Store(time.Now())
	newEntry.accountID.Store("")

	preloaded := map[string]*awsCredentialsProviderCacheEntry{
		"key-old": oldEntry,
		"key-new": newEntry,
	}

	cache := NewAWSCredentialsProviderCache(
		WithCacheStore(preloaded),
		WithCacheMaxSize(2),
	)

	// Add a 3rd entry — should evict "key-old".
	pc := newIRSAProviderConfig("uid-evict", 1)
	provider := newTestCredsCache("AKIA_evict")
	_, err := cache.GetCachedCredentialsProvider(pc, "us-east-1", provider)
	if err != nil {
		t.Fatalf("evict call: %v", err)
	}

	if len(cache.cache) > 2 {
		t.Errorf("expected cache size <= 2 after eviction, got %d", len(cache.cache))
	}
	if _, exists := cache.cache["key-old"]; exists {
		t.Error("expected oldest entry to be evicted, but it still exists")
	}
}

// TestGetCachedCredentialsProvider_Concurrency verifies that concurrent calls
// for the same key converge on a single cache entry (no data races).
func TestGetCachedCredentialsProvider_Concurrency(t *testing.T) {
	tokenFile := writeTokenFile(t, "concurtoken")
	t.Setenv("AWS_WEB_IDENTITY_TOKEN_FILE", tokenFile)
	t.Setenv("AWS_ROLE_ARN", "arn:aws:iam::123456789:role/Role6")

	cache := NewAWSCredentialsProviderCache()
	pc := newIRSAProviderConfig("uid-concur", 1)

	const goroutines = 20
	var wg sync.WaitGroup
	results := make([]aws.CredentialsProvider, goroutines)
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			p := newTestCredsCache("AKIA_concurrent")
			got, err := cache.GetCachedCredentialsProvider(pc, "us-east-1", p)
			if err != nil {
				t.Errorf("goroutine %d: unexpected error: %v", idx, err)
				return
			}
			results[idx] = got
		}(i)
	}
	wg.Wait()

	// All results should point to the same cached instance.
	first := results[0]
	for i, r := range results {
		if r != first {
			t.Errorf("goroutine %d: expected same cached provider, got different pointer", i)
		}
	}
	if len(cache.cache) != 1 {
		t.Errorf("expected exactly 1 cache entry after concurrent calls, got %d", len(cache.cache))
	}
}

// TestGlobalAWSCredentialsProviderCache verifies the global cache is exported
// and is a valid *AWSCredentialsProviderCache.
func TestGlobalAWSCredentialsProviderCache(t *testing.T) {
	if GlobalAWSCredentialsProviderCache == nil {
		t.Fatal("GlobalAWSCredentialsProviderCache must not be nil")
	}
	if GlobalAWSCredentialsProviderCache.cache == nil {
		t.Error("GlobalAWSCredentialsProviderCache.cache must not be nil")
	}
}

// TestAWSCredentialsProviderCache_SetLogger verifies SetLogger does not panic.
func TestAWSCredentialsProviderCache_SetLogger(t *testing.T) {
	cache := NewAWSCredentialsProviderCache()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetLogger panicked: %v", r)
		}
	}()
	// SetLogger should not panic even with a nil-ish logger.
	// In production, callers use logging.NewNopLogger().
	cache.SetLogger(nil)
}

// TestGetAWSConfigWithTrackingAndCache_IsExported verifies the function
// is exported from the clients package. Compilation is the test.
func TestGetAWSConfigWithTrackingAndCache_IsExported(t *testing.T) {
	// If GetAWSConfigWithTrackingAndCache did not exist, this file would not compile.
	// Assign to a typed variable to confirm it matches the expected signature.
	// A function value can never be nil, so we just use it to prevent the compiler
	// from optimising the assignment away.
	fn := GetAWSConfigWithTrackingAndCache
	_ = fn
	t.Log("GetAWSConfigWithTrackingAndCache is exported and callable")
}
