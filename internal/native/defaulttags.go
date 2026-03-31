// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

// ComputeTagsAll returns the merged tag map that should be written to
// status.atProvider.tagsAll. ProviderConfig default_tags (from
// ProviderConfigSpec.DefaultTags) provide the base; resource-level tags from
// spec.forProvider.tags override them on key conflicts.
//
// Usage in a native controller's Observe/Create/Update:
//
//	tagsAll := native.ComputeTagsAll(specTags, pc.Spec.DefaultTags)
//	cr.Status.AtProvider.TagsAll = tagsAll
//
// Combined with DiffTagsWithDefaults, this ensures that:
//  1. Default tags appear in status.atProvider.tagsAll.
//  2. Default tags are never removed by the controller (they are not in the
//     desired set but are protected from deletion by DiffTagsWithDefaults).
//  3. Resource-level tags can override default tag values for the same key.
func ComputeTagsAll(specTags []Tag, defaultTags map[string]string) map[string]string {
	result := make(map[string]string, len(defaultTags)+len(specTags))

	// Seed with ProviderConfig default_tags.
	for k, v := range defaultTags {
		result[k] = v
	}

	// Resource-level tags override defaults on key conflict.
	for _, t := range specTags {
		if t.Key != nil && t.Value != nil {
			result[*t.Key] = *t.Value
		}
	}

	return result
}
