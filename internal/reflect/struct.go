package reflect

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// build a struct with a type matching that of `target` and populate it with
// the values in `object`. `target` must be a struct type. The properties on
// `target` must be tagged with a "tfsdk" label, and every property must be
// present in the type of `object`, and all the attributes in the type of
// `object` must have a corresponding property. Properties that don't map to
// object attributes must have a `tfsdk:"-"` tag, explicitly defining them as
// not part of the object.
func reflectStructFromObject(ctx context.Context, object tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
	// this only works with object values, so make sure that constraint is
	// met
	if target.Kind() != reflect.Struct {
		return target, path.NewErrorf("expected a struct type, got %s", target.Type())
	}
	if !object.Type().Is(tftypes.Object{}) {
		return target, path.NewErrorf("can't reflect %s into a struct, must be an object", object.Type().String())
	}

	// collect a map of fields that are in the object passed in
	var objectFields map[string]tftypes.Value
	err := object.As(&objectFields)
	if err != nil {
		return target, path.NewErrorf("unexpected error converting object: %w", err)
	}

	// collect a map of fields that are defined in the tags of the struct
	// passed in
	targetFields, err := getStructTags(ctx, target, path)
	if err != nil {
		return target, fmt.Errorf("error retrieving field names from struct tags: %w", err)
	}

	// we require an exact, 1:1 match of these fields to avoid typos
	// leading to surprises, so let's ensure they have the exact same
	// fields defined
	var objectMissing, targetMissing []string
	for field := range targetFields {
		if _, ok := objectFields[field]; !ok {
			objectMissing = append(objectMissing, field)
		}
	}
	for field := range objectFields {
		if _, ok := targetFields[field]; !ok {
			targetMissing = append(targetMissing, field)
		}
	}
	if len(objectMissing) > 0 || len(targetMissing) > 0 {
		var missing []string
		if len(objectMissing) > 0 {
			missing = append(missing, fmt.Sprintf("Struct defines fields not found in object: %s.", commaSeparatedString(objectMissing)))
		}
		if len(targetMissing) > 0 {
			missing = append(missing, fmt.Sprintf("Object defines fields not found in struct: %s.", commaSeparatedString(targetMissing)))
		}
		return target, path.NewErrorf("mismatch between struct and object: %s", strings.Join(missing, " "))
	}

	// now that we know they match perfectly, fill the struct with the
	// values in the object
	result := reflect.New(target.Type()).Elem()
	for field, structFieldPos := range targetFields {
		structField := result.Field(structFieldPos)
		fieldVal, err := buildReflectValue(ctx, objectFields[field], structField, opts, path.WithAttributeName(field))
		if err != nil {
			return target, err
		}
		structField.Set(fieldVal)
	}
	return result, nil
}
