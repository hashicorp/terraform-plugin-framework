package reflect

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Struct builds a new struct using the data in `object`, as long as `object`
// is a `tftypes.Object`. It will take the struct type from `target`, which
// must be a struct type.
//
// The properties on `target` must be tagged with a "tfsdk" label containing
// the field name to map to that property. Every property must be tagged, and
// every property must be present in the type of `object`, and all the
// attributes in the type of `object` must have a corresponding property.
// Properties that don't map to object attributes must have a `tfsdk:"-"` tag,
// explicitly defining them as not part of the object. This is to catch typos
// and other mistakes early.
//
// Struct is meant to be called from Into, not directly.
func Struct(ctx context.Context, typ attr.Type, object tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, error) {
	// this only works with object values, so make sure that constraint is
	// met
	if target.Kind() != reflect.Struct {
		return target, path.NewErrorf("expected a struct type, got %s", target.Type())
	}
	if !object.Type().Is(tftypes.Object{}) {
		return target, path.NewErrorf("can't reflect %s into a struct, must be an object", object.Type().String())
	}
	attrsType, ok := typ.(attr.TypeWithAttributeTypes)
	if !ok {
		return target, path.NewErrorf("can't reflect object using type information provided by %T, %T must be an attr.TypeWithAttributeTypes", typ, typ)
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

	attrTypes := attrsType.AttributeTypes()

	// now that we know they match perfectly, fill the struct with the
	// values in the object
	result := reflect.New(target.Type()).Elem()
	for field, structFieldPos := range targetFields {
		attrType, ok := attrTypes[field]
		if !ok {
			return target, path.WithAttributeName(field).NewErrorf("couldn't find type information for attribute in supplied attr.Type %T", typ)
		}
		structField := result.Field(structFieldPos)
		fieldVal, err := BuildValue(ctx, attrType, objectFields[field], structField, opts, path.WithAttributeName(field))
		if err != nil {
			return target, err
		}
		structField.Set(fieldVal)
	}
	return result, nil
}

// FromStruct builds an attr.Value as produced by `typ` from the data in `val`.
// `val` must be a struct type, and must have all its properties tagged and be
// a 1:1 match with the attributes reported by `typ`. FromStruct will recurse
// into FromValue for each attribute, using the type of the attribute as
// reported by `typ`.
//
// It is meant to be called through OutOf, not directly.
func FromStruct(ctx context.Context, typ attr.TypeWithAttributeTypes, val reflect.Value, path *tftypes.AttributePath) (attr.Value, error) {
	objTypes := map[string]tftypes.Type{}
	objValues := map[string]tftypes.Value{}

	// collect a map of fields that are defined in the tags of the struct
	// passed in
	targetFields, err := getStructTags(ctx, val, path)
	if err != nil {
		return nil, fmt.Errorf("error retrieving field names from struct tags: %w", err)
	}

	attrTypes := typ.AttributeTypes()
	for name, fieldNo := range targetFields {
		path := path.WithAttributeName(name)
		fieldValue := val.Field(fieldNo)

		attrVal, err := FromValue(ctx, attrTypes[name], fieldValue.Interface(), path)
		if err != nil {
			return nil, err
		}

		attrType, ok := attrTypes[name]
		if !ok || attrType == nil {
			return nil, path.NewErrorf("couldn't find type information for attribute in supplied attr.Type %T", typ)
		}

		objTypes[name] = attrType.TerraformType(ctx)

		tfVal, err := attrVal.ToTerraformValue(ctx)
		if err != nil {
			return nil, path.NewError(err)
		}
		err = tftypes.ValidateValue(objTypes[name], tfVal)
		if err != nil {
			return nil, path.NewError(err)
		}
		objValues[name] = tftypes.NewValue(objTypes[name], tfVal)
	}

	tfVal := tftypes.NewValue(tftypes.Object{
		AttributeTypes: objTypes,
	}, objValues)

	retType := typ.WithAttributeTypes(attrTypes)
	ret, err := retType.ValueFromTerraform(ctx, tfVal)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
