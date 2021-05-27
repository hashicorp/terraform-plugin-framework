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

func reflectObjectIntoStruct(ctx context.Context, object tftypes.Value, target interface{}, opts Options, path *tftypes.AttributePath) error {
	// this only works with object values, so make sure that constraint is
	// met
	if !object.Type().Is(tftypes.Object{}) {
		return path.NewErrorf("can't reflect %s into a struct, must be an object", object.Type().String())
	}

	// collect a map of fields that are in the object passed in
	var objectFields map[string]tftypes.Value
	err := object.As(&objectFields)
	if err != nil {
		return path.NewErrorf("unexpected error converting object: %w", err)
	}

	// collect a map of fields that are defined in the tags of the struct
	// passed in
	targetFields, err := getStructTags(ctx, target, path)
	if err != nil {
		return fmt.Errorf("error retrieving field names from struct tags: %w", err)
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
		return path.NewErrorf("mismatch between struct and object: %s", strings.Join(missing, " "))
	}

	// now that we know they match perfectly, fill the struct with the
	// values in the object
	structValue := trueReflectValue(target)
	for field, structFieldPos := range targetFields {
		structField := structValue.Field(structFieldPos)
		err := Into(ctx, objectFields[field], structField, opts, path.WithAttributeName(field))
		if err != nil {
			return err
		}
	}
	return nil
}

func reflectObjectOutOfStruct(ctx context.Context, val interface{}, opts OutOfOptions, path *tftypes.AttributePath) (attr.Value, attr.ObjectType, error) {
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

		fieldValue := trueReflectValue(val).Field(i).Interface()

		var attrVal attr.Value
		var attrType attr.Type
		var err error
		switch field.Type.Kind() {
		case reflect.String:
			attrVal, attrType, err = reflectOutOfString(ctx, fieldValue.(string), opts, path)
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
