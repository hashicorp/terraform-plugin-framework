package reflect

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func reflectObjectIntoStruct(ctx context.Context, object tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) error {
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
		log.Println("reflecting", objectFields[field], "into", structField.Type())
		err := into(ctx, objectFields[field], structField, opts, path.WithAttributeName(field))
		if err != nil {
			return err
		}
	}
	return nil
}
