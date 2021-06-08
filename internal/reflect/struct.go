package reflect

import (
	"context"
	"errors"
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

func reflectObjectOutOfStruct(ctx context.Context, val reflect.Value, opts OutOfOptions, path *tftypes.AttributePath) (attr.Value, attr.TypeWithAttributeTypes, error) {
	typ := trueReflectValue(val).Type()

	objTypes := map[string]tftypes.Type{}
	attrTypes := map[string]attr.Type{}
	objValues := map[string]tftypes.Value{}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" {
			// skip unexported fields
			continue
		}
		tag := field.Tag.Get(`tfsdk`)
		if tag == "-" {
			// skip explicitly excluded fields
			continue
		}
		if tag == "" {
			return nil, nil, path.NewErrorf(`need a struct tag for "tfsdk" on %s`, field.Name)
		}
		path := path.WithAttributeName(tag)
		if !isValidFieldName(tag) {
			return nil, nil, path.NewError(errors.New("invalid field name, must only use lowercase letters, underscores, and numbers, and must start with a letter"))
		}

		fieldValue := trueReflectValue(val).Field(i)

		var attrVal attr.Value
		var attrType attr.Type
		var err error
		switch field.Type.Kind() {
		case reflect.String:
			attrVal, attrType, err = reflectOutOfString(ctx, fieldValue.Interface().(string), opts, path)
			if err != nil {
				return nil, nil, path.NewErrorf("error when reflecting field %s: %s", tag, err)
			}
			attrTypes[tag] = attrType
			objTypes[tag] = attrType.TerraformType(ctx)

			tfVal, err := attrVal.ToTerraformValue(ctx)
			if err != nil {
				return nil, nil, path.NewErrorf("error when reflecting field %s: %s", tag, err)
			}
			objValues[tag] = tftypes.NewValue(objTypes[tag], tfVal)
		case reflect.Struct:
			attrVal, attrType, err = reflectObjectOutOfStruct(ctx, fieldValue, opts, path)
			if err != nil {
				return nil, nil, path.NewErrorf("error when recursing into field %s: %s", tag, err)
			}

			attrTypes[tag] = attrType
			objTypes[tag] = attrType.TerraformType(ctx)
		default:
			return nil, nil, path.NewErrorf("don't know how to reflect %s", field.Type.Kind())
		}

		tfVal, err := attrVal.ToTerraformValue(ctx)
		if err != nil {
			return nil, nil, path.NewErrorf("error when reflecting field %s: %s", tag, err)
		}
		objValues[tag] = tftypes.NewValue(objTypes[tag], tfVal)
	}

	tfVal := tftypes.NewValue(tftypes.Object{
		AttributeTypes: objTypes,
	}, objValues)

	retType := opts.Structs.WithAttributeTypes(attrTypes)
	ret, err := retType.ValueFromTerraform(ctx, tfVal)
	if err != nil {
		return nil, nil, err
	}

	return ret, retType, nil
}
