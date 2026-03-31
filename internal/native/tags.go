// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

// Tag represents a key-value tag pair using pointer fields, matching the AWS
// SDK style so Tag slices can be used directly with SDK request types.
type Tag struct {
	// Key is the tag key. Pointer to align with AWS SDK tag types.
	Key *string

	// Value is the tag value. Pointer to align with AWS SDK tag types.
	Value *string
}

// tagKey returns the string value of a tag's Key, or "" if nil.
func tagKey(t Tag) string {
	if t.Key == nil {
		return ""
	}
	return *t.Key
}

// tagValue returns the string value of a tag's Value, or "" if nil.
func tagValue(t Tag) string {
	if t.Value == nil {
		return ""
	}
	return *t.Value
}

// buildTagMap converts a []Tag to a map[string]string keyed by tag key.
func buildTagMap(tags []Tag) map[string]string {
	m := make(map[string]string, len(tags))
	for _, t := range tags {
		m[tagKey(t)] = tagValue(t)
	}
	return m
}

// ensureSlice returns an initialised (non-nil) empty slice when s is nil,
// so callers never receive a nil slice from Diff functions.
func ensureSlice(s []Tag) []Tag {
	if s == nil {
		return []Tag{}
	}
	return s
}

// DiffTags computes the set of tags to add and the set to remove when
// reconciling from observed to desired.
//
//   - toAdd contains every tag in desired that is absent from observed or has a
//     different value.
//   - toRemove contains every tag in observed that is absent from desired OR
//     that has a different value in desired (the old observed value is removed
//     and the new desired value is added).
func DiffTags(desired, observed []Tag) (toAdd, toRemove []Tag) {
	observedMap := buildTagMap(observed)
	desiredMap := buildTagMap(desired)

	// Determine tags to add (new or changed).
	for _, t := range desired {
		if v, exists := observedMap[tagKey(t)]; !exists || v != tagValue(t) {
			toAdd = append(toAdd, t)
		}
	}

	// Determine tags to remove: present in observed but absent from desired,
	// OR present in both but with a different value (stale value must be
	// removed before/alongside the new value being added).
	for _, t := range observed {
		if desiredVal, exists := desiredMap[tagKey(t)]; !exists || desiredVal != tagValue(t) {
			toRemove = append(toRemove, t)
		}
	}

	return ensureSlice(toAdd), ensureSlice(toRemove)
}

// DiffTagsWithDefaults is like DiffTags but protects default tags provided by
// the ProviderConfig from being removed. A default tag is NOT included in
// toRemove when it is absent from the desired set (i.e., not explicitly
// overridden by the controller). If a desired tag explicitly overrides a
// default tag key with a different value, the old observed value IS included
// in toRemove so that it can be replaced.
//
// This preserves ProviderConfig default_tags that are injected by AWS
// infrastructure and should never be cleaned up by the controller.
func DiffTagsWithDefaults(desired, observed []Tag, defaults map[string]string) (toAdd, toRemove []Tag) {
	toAdd, toRemove = DiffTags(desired, observed)

	if len(defaults) == 0 {
		return toAdd, toRemove
	}

	// Build desired key set for O(1) lookup.
	desiredKeys := make(map[string]struct{}, len(desired))
	for _, t := range desired {
		desiredKeys[tagKey(t)] = struct{}{}
	}

	// Filter out default-tag keys from toRemove, but only when the desired
	// set does not explicitly provide a different value for the same key.
	// An explicit override (key present in desired) means the user wants to
	// change the value, so the old value must be removed.
	filtered := make([]Tag, 0, len(toRemove))
	for _, t := range toRemove {
		k := tagKey(t)
		_, isDefault := defaults[k]
		_, inDesired := desiredKeys[k]
		if isDefault && !inDesired {
			// Protected default tag not being overridden — skip removal.
			continue
		}
		filtered = append(filtered, t)
	}
	toRemove = filtered

	// Ensure nil-safety.
	if toRemove == nil {
		toRemove = []Tag{}
	}
	return toAdd, toRemove
}

// MergeTags merges base and overrides tag slices into a single slice. When the
// same key appears in both, the value from overrides wins.
func MergeTags(base []Tag, overrides map[string]string) []Tag {
	// Build a map from base tags.
	merged := make(map[string]string, len(base))
	for _, t := range base {
		merged[tagKey(t)] = tagValue(t)
	}

	// Apply overrides.
	for k, v := range overrides {
		k, v := k, v
		merged[k] = v
	}

	// Reconstruct as a Tag slice.
	out := make([]Tag, 0, len(merged))
	for k, v := range merged {
		k, v := k, v
		out = append(out, Tag{Key: &k, Value: &v})
	}
	return out
}
