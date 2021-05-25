package tfsdk

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var attributeValueReflectType = reflect.TypeOf(new(attr.Value)).Elem()

type State struct {
	Raw    tftypes.Value
	Schema schema.Schema
}

func isValidFieldName(name string) bool {
	re := regexp.MustCompile("^[a-z][a-z0-9_]*$")
	return re.MatchString(name)
}

func (s State) As(ctx context.Context, in interface{}) error {
	reflectValue := reflect.ValueOf(in)
	reflectType := reflect.TypeOf(in)
	reflectKind := reflectType.Kind()
	for reflectKind == reflect.Interface || reflectKind == reflect.Ptr {
		reflectValue = reflectValue.Elem()
		reflectType = reflectValue.Type()
		reflectKind = reflectType.Kind()
	}

	if reflectKind != reflect.Struct {
		return fmt.Errorf("can only pass structs to As, can't use %s", reflectKind)
	}

	structFields := map[string]int{}
	for i := 0; i < reflectType.NumField(); i++ {
		field := reflectType.Field(i)
		if field.PkgPath != "" {
			// skip unexported fields
			continue
		}
		tag := field.Tag.Get(`tfsdk`)
		if tag == "-" {
			// skip explicitly skipped fields
			continue
		}
		if tag == "" {
			return fmt.Errorf("Need a tfsdk tag on %s to use As", field.Name)
		}
		if !isValidFieldName(tag) {
			return fmt.Errorf("Can't use %q as a field name, must only contain a-z (lowercase), underscores, and numbers, and must start with a letter.", tag)
		}
		if other, ok := structFields[tag]; ok {
			return fmt.Errorf("Can't use %s as a field name for both %s and %s", tag, reflectType.Field(other).Name, field.Name)
		}
		structFields[tag] = i
	}

	var raw map[string]tftypes.Value
	err := s.Raw.As(&raw)
	if err != nil {
		return fmt.Errorf("error asserting type of state: %w", err)
	}

	var stateMissing []string
	var structMissing []string
	for k := range structFields {
		if _, ok := raw[k]; !ok {
			stateMissing = append(stateMissing, k)
		}
	}
	for k := range raw {
		if _, ok := raw[k]; !ok {
			structMissing = append(structMissing, k)
		}
	}
	if len(stateMissing) > 0 || len(structMissing) > 0 {
		var missing []string
		if len(stateMissing) > 0 {
			var fields string
			if len(stateMissing) == 1 {
				fields = stateMissing[0]
			} else if len(stateMissing) == 2 {
				fields = strings.Join(stateMissing, " and ")
			} else {
				stateMissing[len(stateMissing)-1] = "and " + stateMissing[len(stateMissing)-1]
				fields = strings.Join(stateMissing, ", ")
			}
			missing = append(missing, fmt.Sprintf("Struct defines fields (%s) that weren't included in the request.", fields))
		}
		if len(structMissing) > 0 {
			var fields string
			if len(structMissing) == 1 {
				fields = structMissing[0]
			} else if len(structMissing) == 2 {
				fields = strings.Join(structMissing, " and ")
			} else {
				structMissing[len(structMissing)-1] = "and " + structMissing[len(structMissing)-1]
				fields = strings.Join(structMissing, ", ")
			}
			missing = append(missing, fmt.Sprintf("Struct defines fields (%s) that weren't included in the request.", fields))
		}
		return fmt.Errorf("Invalid struct definition for this request: " + strings.Join(missing, " "))
	}
	for tag, i := range structFields {
		field := reflectType.Field(i)
		if !field.Type.Implements(attributeValueReflectType) {
			return fmt.Errorf("%s doesn't fill the attr.Value interface", field.Name)
		}
		fieldValue := reflectValue.Field(i)
		if !fieldValue.CanSet() {
			return fmt.Errorf("can't set %s", field.Name)
		}

		// find out how to instantiate new value of that type
		// pull the attr.Type out of the schema
		schemaAttr, ok := s.Schema.Attributes[tag]
		if !ok {
			return fmt.Errorf("Couldn't find a schema for %s", tag)
		}
		attrValue, err := schemaAttr.Type.ValueFromTerraform(ctx, raw[tag])
		if err != nil {
			return fmt.Errorf("Error converting %q from state: %w", tag, err)
		}

		newValue := reflect.ValueOf(attrValue)
		if !newValue.Type().AssignableTo(field.Type) {
			return fmt.Errorf("can't assign %s to %s for %s", newValue.Type().Name(), field.Type.Name(), field.Name)
		}

		fieldValue.Set(newValue)
	}
	return nil
}
