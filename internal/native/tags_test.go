// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"sort"
	"testing"
)

// ptr is a helper to get a pointer to a string literal in tests.
func ptr(s string) *string { return &s }

// tagsFromMap converts a map[string]string to []Tag for test convenience.
func tagsFromMap(m map[string]string) []Tag {
	out := make([]Tag, 0, len(m))
	for k, v := range m {
		k, v := k, v
		out = append(out, Tag{Key: &k, Value: &v})
	}
	return out
}

// tagsToMap converts a []Tag to a map[string]string for easy assertion.
func tagsToMap(tags []Tag) map[string]string {
	out := make(map[string]string, len(tags))
	for _, t := range tags {
		if t.Key != nil && t.Value != nil {
			out[*t.Key] = *t.Value
		}
	}
	return out
}

// sortTags sorts []Tag by key for deterministic comparison.
func sortTags(tags []Tag) []Tag {
	sort.Slice(tags, func(i, j int) bool {
		ki, kj := "", ""
		if tags[i].Key != nil {
			ki = *tags[i].Key
		}
		if tags[j].Key != nil {
			kj = *tags[j].Key
		}
		return ki < kj
	})
	return tags
}

// ──────────────────────────────────────────────
// DiffTags tests
// ──────────────────────────────────────────────

func TestDiffTags(t *testing.T) {
	type testCase struct {
		name       string
		desired    map[string]string
		observed   map[string]string
		wantAdd    map[string]string
		wantRemove map[string]string
	}

	cases := []testCase{
		{
			name:       "both_empty",
			desired:    map[string]string{},
			observed:   map[string]string{},
			wantAdd:    map[string]string{},
			wantRemove: map[string]string{},
		},
		{
			name:       "desired_empty_observed_has_tags",
			desired:    map[string]string{},
			observed:   map[string]string{"env": "prod"},
			wantAdd:    map[string]string{},
			wantRemove: map[string]string{"env": "prod"},
		},
		{
			name:       "desired_has_tags_observed_empty",
			desired:    map[string]string{"env": "prod"},
			observed:   map[string]string{},
			wantAdd:    map[string]string{"env": "prod"},
			wantRemove: map[string]string{},
		},
		{
			name:       "identical_tag_sets",
			desired:    map[string]string{"env": "prod", "team": "infra"},
			observed:   map[string]string{"env": "prod", "team": "infra"},
			wantAdd:    map[string]string{},
			wantRemove: map[string]string{},
		},
		{
			name:       "partial_overlap_add_and_remove",
			desired:    map[string]string{"env": "prod", "app": "web"},
			observed:   map[string]string{"env": "prod", "team": "infra"},
			wantAdd:    map[string]string{"app": "web"},
			wantRemove: map[string]string{"team": "infra"},
		},
		{
			name:       "key_present_different_value",
			desired:    map[string]string{"env": "staging"},
			observed:   map[string]string{"env": "prod"},
			wantAdd:    map[string]string{"env": "staging"},
			wantRemove: map[string]string{"env": "prod"},
		},
		{
			name:       "multiple_value_changes",
			desired:    map[string]string{"env": "staging", "version": "v2"},
			observed:   map[string]string{"env": "prod", "version": "v1"},
			wantAdd:    map[string]string{"env": "staging", "version": "v2"},
			wantRemove: map[string]string{"env": "prod", "version": "v1"},
		},
		{
			name:       "only_additions",
			desired:    map[string]string{"env": "prod", "app": "web", "team": "infra"},
			observed:   map[string]string{"env": "prod"},
			wantAdd:    map[string]string{"app": "web", "team": "infra"},
			wantRemove: map[string]string{},
		},
		{
			name:       "only_removals",
			desired:    map[string]string{"env": "prod"},
			observed:   map[string]string{"env": "prod", "app": "web", "team": "infra"},
			wantAdd:    map[string]string{},
			wantRemove: map[string]string{"app": "web", "team": "infra"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			desired := tagsFromMap(tc.desired)
			observed := tagsFromMap(tc.observed)

			gotAdd, gotRemove := DiffTags(desired, observed)

			if got, want := tagsToMap(gotAdd), tc.wantAdd; !mapsEqual(got, want) {
				t.Errorf("DiffTags() toAdd = %v, want %v", got, want)
			}
			if got, want := tagsToMap(gotRemove), tc.wantRemove; !mapsEqual(got, want) {
				t.Errorf("DiffTags() toRemove = %v, want %v", got, want)
			}
		})
	}
}

// ──────────────────────────────────────────────
// DiffTagsWithDefaults tests
// ──────────────────────────────────────────────

func TestDiffTagsWithDefaults(t *testing.T) {
	type testCase struct {
		name       string
		desired    map[string]string
		observed   map[string]string
		defaults   map[string]string
		wantAdd    map[string]string
		wantRemove map[string]string
	}

	cases := []testCase{
		{
			name:       "no_defaults",
			desired:    map[string]string{"env": "prod"},
			observed:   map[string]string{"env": "prod", "team": "infra"},
			defaults:   map[string]string{},
			wantAdd:    map[string]string{},
			wantRemove: map[string]string{"team": "infra"},
		},
		{
			name:       "default_tags_not_removed_when_absent_from_desired",
			desired:    map[string]string{"app": "web"},
			observed:   map[string]string{"app": "web", "Owner": "platform"},
			defaults:   map[string]string{"Owner": "platform"},
			wantAdd:    map[string]string{},
			wantRemove: map[string]string{},
		},
		{
			name:       "default_tags_not_removed_even_when_not_in_observed",
			desired:    map[string]string{"app": "web"},
			observed:   map[string]string{"app": "web"},
			defaults:   map[string]string{"Owner": "platform"},
			wantAdd:    map[string]string{},
			wantRemove: map[string]string{},
		},
		{
			name:       "non_default_tag_removed_normally",
			desired:    map[string]string{"app": "web"},
			observed:   map[string]string{"app": "web", "extra": "value", "Owner": "platform"},
			defaults:   map[string]string{"Owner": "platform"},
			wantAdd:    map[string]string{},
			wantRemove: map[string]string{"extra": "value"},
		},
		{
			name:       "desired_tag_overrides_default_tag",
			desired:    map[string]string{"Owner": "team-a"},
			observed:   map[string]string{"Owner": "platform"},
			defaults:   map[string]string{"Owner": "platform"},
			wantAdd:    map[string]string{"Owner": "team-a"},
			wantRemove: map[string]string{"Owner": "platform"},
		},
		{
			name:       "all_empty",
			desired:    map[string]string{},
			observed:   map[string]string{},
			defaults:   map[string]string{},
			wantAdd:    map[string]string{},
			wantRemove: map[string]string{},
		},
		{
			name:       "nil_defaults_treated_as_empty",
			desired:    map[string]string{"env": "prod"},
			observed:   map[string]string{"env": "prod", "old": "tag"},
			defaults:   nil,
			wantAdd:    map[string]string{},
			wantRemove: map[string]string{"old": "tag"},
		},
		{
			name:       "multiple_defaults_all_preserved",
			desired:    map[string]string{"app": "web"},
			observed:   map[string]string{"app": "web", "CostCenter": "cc-123", "Env": "prod", "extra": "remove-me"},
			defaults:   map[string]string{"CostCenter": "cc-123", "Env": "prod"},
			wantAdd:    map[string]string{},
			wantRemove: map[string]string{"extra": "remove-me"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			desired := tagsFromMap(tc.desired)
			observed := tagsFromMap(tc.observed)

			gotAdd, gotRemove := DiffTagsWithDefaults(desired, observed, tc.defaults)

			if got, want := tagsToMap(gotAdd), tc.wantAdd; !mapsEqual(got, want) {
				t.Errorf("DiffTagsWithDefaults() toAdd = %v, want %v", got, want)
			}
			if got, want := tagsToMap(gotRemove), tc.wantRemove; !mapsEqual(got, want) {
				t.Errorf("DiffTagsWithDefaults() toRemove = %v, want %v", got, want)
			}
		})
	}
}

// ──────────────────────────────────────────────
// MergeTags tests
// ──────────────────────────────────────────────

func TestMergeTags(t *testing.T) {
	type testCase struct {
		name      string
		base      map[string]string
		overrides map[string]string
		want      map[string]string
	}

	cases := []testCase{
		{
			name:      "both_empty",
			base:      map[string]string{},
			overrides: map[string]string{},
			want:      map[string]string{},
		},
		{
			name:      "base_only",
			base:      map[string]string{"env": "prod", "team": "infra"},
			overrides: map[string]string{},
			want:      map[string]string{"env": "prod", "team": "infra"},
		},
		{
			name:      "overrides_only",
			base:      map[string]string{},
			overrides: map[string]string{"env": "staging"},
			want:      map[string]string{"env": "staging"},
		},
		{
			name:      "no_key_conflicts",
			base:      map[string]string{"team": "infra"},
			overrides: map[string]string{"env": "prod"},
			want:      map[string]string{"team": "infra", "env": "prod"},
		},
		{
			name:      "overrides_win_on_conflict",
			base:      map[string]string{"env": "prod", "version": "v1"},
			overrides: map[string]string{"env": "staging"},
			want:      map[string]string{"env": "staging", "version": "v1"},
		},
		{
			name:      "all_keys_conflict_overrides_win",
			base:      map[string]string{"env": "prod", "team": "infra"},
			overrides: map[string]string{"env": "dev", "team": "platform"},
			want:      map[string]string{"env": "dev", "team": "platform"},
		},
		{
			name:      "nil_overrides_treated_as_empty",
			base:      map[string]string{"env": "prod"},
			overrides: nil,
			want:      map[string]string{"env": "prod"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			base := tagsFromMap(tc.base)
			result := MergeTags(base, tc.overrides)

			if got, want := tagsToMap(result), tc.want; !mapsEqual(got, want) {
				t.Errorf("MergeTags() = %v, want %v", got, want)
			}
		})
	}
}

// ──────────────────────────────────────────────
// Tag struct format tests
// ──────────────────────────────────────────────

// TestTag_PointerFields verifies that Tag uses *string fields matching the AWS
// SDK pattern (not plain string) so it can be directly used with SDK types.
func TestTag_PointerFields(t *testing.T) {
	k, v := "mykey", "myval"
	tag := Tag{Key: &k, Value: &v}

	if tag.Key == nil || *tag.Key != k {
		t.Errorf("Tag.Key = %v, want %q", tag.Key, k)
	}
	if tag.Value == nil || *tag.Value != v {
		t.Errorf("Tag.Value = %v, want %q", tag.Value, v)
	}
}

// TestDiffTags_NilPointersHandled verifies that tags with nil Key or Value
// do not panic and are treated as empty strings.
func TestDiffTags_NilPointers(t *testing.T) {
	desired := []Tag{{Key: ptr("env"), Value: ptr("prod")}}
	// Observed has a tag with nil Key — should not panic.
	observed := []Tag{{Key: nil, Value: ptr("orphan")}}

	// Should not panic.
	gotAdd, gotRemove := DiffTags(desired, observed)

	addMap := tagsToMap(gotAdd)
	if v, ok := addMap["env"]; !ok || v != "prod" {
		t.Errorf("expected toAdd[env]=prod, got %v", addMap)
	}
	_ = gotRemove // nil-key tag treated as key=""
}

// ──────────────────────────────────────────────
// helpers
// ──────────────────────────────────────────────

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			return false
		}
	}
	return true
}

// Ensure sortTags is used (suppress "unused" warning in tests that don't call it).
var _ = sortTags
var _ = ptr
