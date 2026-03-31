// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"testing"
)

// ──────────────────────────────────────────────────────────────────────────────
// ComputeTagsAll tests
// ──────────────────────────────────────────────────────────────────────────────

func TestComputeTagsAll(t *testing.T) {
	type testCase struct {
		name        string
		specTags    map[string]string // input as map, converted to []Tag in test
		defaultTags map[string]string
		want        map[string]string
	}

	cases := []testCase{
		{
			name:        "both_empty",
			specTags:    map[string]string{},
			defaultTags: map[string]string{},
			want:        map[string]string{},
		},
		{
			name:        "only_default_tags",
			specTags:    map[string]string{},
			defaultTags: map[string]string{"Owner": "platform", "CostCenter": "cc-123"},
			want:        map[string]string{"Owner": "platform", "CostCenter": "cc-123"},
		},
		{
			name:        "only_spec_tags",
			specTags:    map[string]string{"env": "prod", "app": "web"},
			defaultTags: map[string]string{},
			want:        map[string]string{"env": "prod", "app": "web"},
		},
		{
			name:        "no_conflict_merged",
			specTags:    map[string]string{"env": "prod"},
			defaultTags: map[string]string{"Owner": "platform"},
			want:        map[string]string{"env": "prod", "Owner": "platform"},
		},
		{
			name:        "spec_tags_override_defaults_on_conflict",
			specTags:    map[string]string{"Owner": "team-a", "env": "staging"},
			defaultTags: map[string]string{"Owner": "platform", "CostCenter": "cc-123"},
			want:        map[string]string{"Owner": "team-a", "env": "staging", "CostCenter": "cc-123"},
		},
		{
			name:        "nil_default_tags",
			specTags:    map[string]string{"env": "prod"},
			defaultTags: nil,
			want:        map[string]string{"env": "prod"},
		},
		{
			name:        "nil_spec_tags_with_defaults",
			specTags:    nil,
			defaultTags: map[string]string{"Owner": "platform"},
			want:        map[string]string{"Owner": "platform"},
		},
		{
			name:        "nil_both",
			specTags:    nil,
			defaultTags: nil,
			want:        map[string]string{},
		},
		{
			name:        "multiple_defaults_all_in_tagsAll",
			specTags:    map[string]string{"app": "api"},
			defaultTags: map[string]string{"Env": "prod", "Team": "platform", "CostCenter": "cc-999"},
			want: map[string]string{
				"app":        "api",
				"Env":        "prod",
				"Team":       "platform",
				"CostCenter": "cc-999",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			specTags := tagsFromMap(tc.specTags)
			got := ComputeTagsAll(specTags, tc.defaultTags)

			if !mapsEqual(got, tc.want) {
				t.Errorf("ComputeTagsAll() = %v, want %v", got, tc.want)
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// ComputeTagsAll + DiffTagsWithDefaults integration: default tags are not removed
// ──────────────────────────────────────────────────────────────────────────────

// TestDefaultTagsNotRemovedByDiff verifies the acceptance criterion:
// "Tags from default_tags are NOT removed by native controllers."
//
// This simulates the full reconcile flow:
//  1. ComputeTagsAll produces the desired observed state (what should be on AWS).
//  2. DiffTagsWithDefaults computes what to add/remove, protecting default tags.
func TestDefaultTagsNotRemovedByDiff(t *testing.T) {
	// ProviderConfig has two default_tags.
	defaults := map[string]string{
		"Owner":      "platform",
		"CostCenter": "cc-123",
	}

	// spec.forProvider.tags has one resource-specific tag.
	specTags := tagsFromMap(map[string]string{
		"app": "web",
	})

	// tagsAll = ComputeTagsAll(specTags, defaults) — what we expect on AWS.
	tagsAll := ComputeTagsAll(specTags, defaults)

	wantTagsAll := map[string]string{
		"app":        "web",
		"Owner":      "platform",
		"CostCenter": "cc-123",
	}
	if !mapsEqual(tagsAll, wantTagsAll) {
		t.Fatalf("ComputeTagsAll() = %v, want %v", tagsAll, wantTagsAll)
	}

	// Simulate: AWS currently has tagsAll on the resource (steady state).
	observed := tagsFromMap(tagsAll)

	// Next reconcile: spec still only has "app=web"; desired = specTags only.
	// DiffTagsWithDefaults should NOT generate remove operations for default tags.
	toAdd, toRemove := DiffTagsWithDefaults(specTags, observed, defaults)

	if len(toAdd) != 0 {
		t.Errorf("expected no tags to add, got %v", tagsToMap(toAdd))
	}
	if len(toRemove) != 0 {
		t.Errorf("expected no tags to remove (default tags protected), got %v", tagsToMap(toRemove))
	}
}

// TestDefaultTagsOverriddenBySpecTags verifies that when spec.forProvider.tags
// explicitly sets a key that also appears in default_tags, the spec wins and the
// old default value IS removed (replaced with the spec value).
func TestDefaultTagsOverriddenBySpecTags(t *testing.T) {
	defaults := map[string]string{"Owner": "platform"}

	// Spec overrides Owner with a different value.
	specTags := tagsFromMap(map[string]string{"Owner": "team-a"})

	observed := tagsFromMap(map[string]string{"Owner": "platform"})

	toAdd, toRemove := DiffTagsWithDefaults(specTags, observed, defaults)

	addMap := tagsToMap(toAdd)
	removeMap := tagsToMap(toRemove)

	if v, ok := addMap["Owner"]; !ok || v != "team-a" {
		t.Errorf("expected toAdd[Owner]=team-a, got %v", addMap)
	}
	if v, ok := removeMap["Owner"]; !ok || v != "platform" {
		t.Errorf("expected toRemove[Owner]=platform (old value replaced), got %v", removeMap)
	}
}
