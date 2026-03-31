// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"reflect"
	"strings"

	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
)

// LateInitConfig controls which fields are skipped during late-initialization.
type LateInitConfig struct {
	// IgnoredFields lists JSON field names that must never be late-initialized.
	// The names are matched against the lowercase json tag of each struct field.
	IgnoredFields []string

	// ConditionalIgnored maps a JSON field name to a predicate that receives
	// the managed resource. If the predicate returns true the field is skipped
	// for that specific reconcile iteration.
	ConditionalIgnored map[string]func(resource.Managed) bool
}

// IsIgnored reports whether the given JSON field name is listed as ignored in
// the LateInitConfig.
func IsIgnored(field string, cfg LateInitConfig) bool {
	for _, f := range cfg.IgnoredFields {
		if f == field {
			return true
		}
	}
	return false
}

// LateInitializeStringPtr sets *dst = src when dst is nil and src is non-nil.
// Returns true if dst was populated.
func LateInitializeStringPtr(dst **string, src *string) bool {
	if *dst != nil || src == nil {
		return false
	}
	*dst = src
	return true
}

// LateInitializeBoolPtr sets *dst = src when dst is nil and src is non-nil.
// Returns true if dst was populated.
func LateInitializeBoolPtr(dst **bool, src *bool) bool {
	if *dst != nil || src == nil {
		return false
	}
	*dst = src
	return true
}

// LateInitializeInt64Ptr sets *dst = src when dst is nil and src is non-nil.
// Returns true if dst was populated.
func LateInitializeInt64Ptr(dst **int64, src *int64) bool {
	if *dst != nil || src == nil {
		return false
	}
	*dst = src
	return true
}

// LateInitializeInt32Ptr sets *dst = src when dst is nil and src is non-nil.
// Returns true if dst was populated.
func LateInitializeInt32Ptr(dst **int32, src *int32) bool {
	if *dst != nil || src == nil {
		return false
	}
	*dst = src
	return true
}

// LateInitializeFloat64Ptr sets *dst = src when dst is nil and src is non-nil.
// Returns true if dst was populated.
func LateInitializeFloat64Ptr(dst **float64, src *float64) bool {
	if *dst != nil || src == nil {
		return false
	}
	*dst = src
	return true
}

// LateInitializeMapStringPtr sets dst = src when dst is nil and src is non-nil.
// Returns true if dst was populated.
func LateInitializeMapStringPtr(dst *map[string]*string, src map[string]*string) bool {
	if *dst != nil || src == nil {
		return false
	}
	*dst = src
	return true
}

// LateInitialize performs reflection-based late initialization.
// It copies pointer fields from observed into spec when the spec field is nil,
// subject to the ignore rules in cfg.
//
// Both spec and observed must be non-nil pointers to structs of the same type.
// Returns true if any field was populated.
func LateInitialize(cr resource.Managed, spec, observed interface{}, cfg LateInitConfig) bool {
	if spec == nil || observed == nil {
		return false
	}

	sv := reflect.ValueOf(spec)
	ov := reflect.ValueOf(observed)

	// Dereference pointers.
	if sv.Kind() == reflect.Ptr {
		if sv.IsNil() {
			return false
		}
		sv = sv.Elem()
	}
	if ov.Kind() == reflect.Ptr {
		if ov.IsNil() {
			return false
		}
		ov = ov.Elem()
	}

	if sv.Kind() != reflect.Struct || ov.Kind() != reflect.Struct {
		return false
	}

	return lateInitStruct(cr, sv, ov, cfg)
}

// lateInitStruct recursively walks a struct, populating nil pointer fields in
// dst from src when not excluded by cfg.
func lateInitStruct(cr resource.Managed, dst, src reflect.Value, cfg LateInitConfig) bool {
	changed := false
	t := dst.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields.
		if !field.IsExported() {
			continue
		}

		// Derive the JSON field name from the struct tag.
		jsonName := jsonFieldName(field)

		// Apply static ignore list.
		if IsIgnored(jsonName, cfg) {
			continue
		}

		// Apply conditional ignore list.
		if cond, ok := cfg.ConditionalIgnored[jsonName]; ok && cr != nil {
			if cond(cr) {
				continue
			}
		}

		dstField := dst.Field(i)
		srcField := src.Field(i)

		switch dstField.Kind() { //nolint:exhaustive
		case reflect.Ptr:
			if dstField.IsNil() && !srcField.IsNil() {
				// Check whether the pointed-to type is a struct: if so, recurse.
				elem := srcField.Elem()
				if elem.Kind() == reflect.Struct {
					// Allocate a new value and recurse.
					newVal := reflect.New(elem.Type())
					if lateInitStruct(cr, newVal.Elem(), elem, cfg) {
						changed = true
					}
					// Always set when dst was nil and src was non-nil.
					dstField.Set(newVal)
					changed = true
				} else {
					dstField.Set(srcField)
					changed = true
				}
			} else if !dstField.IsNil() && !srcField.IsNil() {
				// Both non-nil: recurse if struct pointer.
				if dstField.Elem().Kind() == reflect.Struct {
					if lateInitStruct(cr, dstField.Elem(), srcField.Elem(), cfg) {
						changed = true
					}
				}
			}
		default:
			// Non-pointer fields are not late-initialized (they already have
			// zero values and cannot distinguish "not set" from "zero").
		}
	}

	return changed
}

// jsonFieldName returns the lowercase JSON key for a struct field, using the
// json struct tag when present. Returns the lowercase field name as fallback.
func jsonFieldName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" {
		return strings.ToLower(f.Name)
	}
	name := strings.Split(tag, ",")[0]
	if name == "" || name == "-" {
		return strings.ToLower(f.Name)
	}
	return name
}
