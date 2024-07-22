// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reflect

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
)

// trueReflectValue returns the reflect.Value for `in` after derefencing all
// the pointers and unwrapping all the interfaces. It's the concrete value
// beneath it all.
func trueReflectValue(val reflect.Value) reflect.Value {
	kind := val.Type().Kind()
	for kind == reflect.Interface || kind == reflect.Ptr {
		innerVal := val.Elem()
		if !innerVal.IsValid() {
			break
		}
		val = innerVal
		kind = val.Type().Kind()
	}
	return val
}

// commaSeparatedString returns an English joining of the strings in `in`,
// using "and" and commas as appropriate.
func commaSeparatedString(in []string) string {
	switch len(in) {
	case 0:
		return ""
	case 1:
		return in[0]
	case 2:
		return strings.Join(in, " and ")
	default:
		in[len(in)-1] = "and " + in[len(in)-1]
		return strings.Join(in, ", ")
	}
}

// getStructTags returns a map of Terraform field names to their position in
// the fields of the struct `in`. `in` must be a struct.
//
// The position of the field in a struct is represented as an index sequence to support type embedding
// in structs. This index sequence can be used to retrieve the field with the Go "reflect" package FieldByIndex methods:
//   - https://pkg.go.dev/reflect#Type.FieldByIndex
//   - https://pkg.go.dev/reflect#Value.FieldByIndex
//
// The following are not supported and will return an error if detected in a struct (including embedded structs):
//   - Duplicate "tfsdk" tags
//   - Exported fields without a "tfsdk" tag
//   - Exported fields with an invalid "tfsdk" tag (must be a valid Terraform identifier)
func getStructTags(ctx context.Context, typ reflect.Type, path path.Path) (map[string][]int, error) { //nolint:unparam // False positive, ctx is used below.
	tags := make(map[string][]int, 0)

	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%s: can't get struct tags of %s, is not a struct", path, typ)
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			// skip all unexported fields
			continue
		}

		// This index sequence is the location of the field within the struct.
		// For embedded structs, the length of this sequence will be > 1
		fieldIndexSequence := []int{i}
		tag := field.Tag.Get(`tfsdk`)

		switch tag {
		// "tfsdk" tags can only be omitted on embedded structs
		case "":
			if field.Anonymous {
				embeddedTags, err := getStructTags(ctx, field.Type, path)
				if err != nil {
					return nil, fmt.Errorf(`error retrieving embedded struct %q field tags: %w`, field.Name, err)
				}
				for k, v := range embeddedTags {
					if other, ok := tags[k]; ok {
						otherField := typ.FieldByIndex(other)
						return nil, fmt.Errorf("embedded struct %q contains a duplicate field name %q", field.Name, otherField.Name)
					}

					tags[k] = append(fieldIndexSequence, v...)
				}
				continue
			}

			return nil, fmt.Errorf(`%s: need a struct tag for "tfsdk" on %s`, path, field.Name)

		// "tfsdk" tags with "-" are being explicitly excluded
		case "-":
			continue

		// validate the "tfsdk" tag and ensure there are no duplicates before storing
		default:
			path := path.AtName(tag)
			if !isValidFieldName(tag) {
				return nil, fmt.Errorf("%s: invalid field name, must only use lowercase letters, underscores, and numbers, and must start with a letter", path)
			}
			if other, ok := tags[tag]; ok {
				otherField := typ.FieldByIndex(other)
				return nil, fmt.Errorf("%s: can't use field name for both %s and %s", path, otherField.Name, field.Name)
			}

			tags[tag] = fieldIndexSequence
		}
	}
	return tags, nil
}

// isValidFieldName returns true if `name` can be used as a field name in a
// Terraform resource or data source.
func isValidFieldName(name string) bool {
	re := regexp.MustCompile("^[a-z][a-z0-9_]*$")
	return re.MatchString(name)
}

// canBeNil returns true if `target`'s type can hold a nil value
func canBeNil(target reflect.Value) bool {
	switch target.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Interface:
		// these types can all hold nils
		return true
	default:
		// nothing else can be set to nil
		return false
	}
}
