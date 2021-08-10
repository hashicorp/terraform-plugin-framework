package reflect

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/diagnostics"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Map creates a map value that matches the type of `target`, and populates it
// with the contents of `val`.
func Map(ctx context.Context, typ attr.Type, val tftypes.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, []*tfprotov6.Diagnostic) {
	var diags []*tfprotov6.Diagnostic
	underlyingValue := trueReflectValue(target)

	// this only works with maps, so check that out first
	if underlyingValue.Kind() != reflect.Map {
		err := fmt.Errorf("expected a map type, got %s", target.Type())
		return target, append(diags, &tfprotov6.Diagnostic{
			Severity:  tfprotov6.DiagnosticSeverityError,
			Summary:   "Value Conversion Error",
			Detail:    "An unexpected error was encountered trying to convert to map value. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			Attribute: path,
		})
	}
	if !val.Type().Is(tftypes.Map{}) {
		err := fmt.Errorf("cannot reflect %s into a map, must be a map", val.Type().String())
		return target, append(diags, &tfprotov6.Diagnostic{
			Severity:  tfprotov6.DiagnosticSeverityError,
			Summary:   "Value Conversion Error",
			Detail:    "An unexpected error was encountered trying to convert to map value. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			Attribute: path,
		})
	}
	elemTyper, ok := typ.(attr.TypeWithElementType)
	if !ok {
		err := fmt.Errorf("cannot reflect map using type information provided by %T, %T must be an attr.TypeWithElementType", typ, typ)
		return target, append(diags, &tfprotov6.Diagnostic{
			Severity:  tfprotov6.DiagnosticSeverityError,
			Summary:   "Value Conversion Error",
			Detail:    "An unexpected error was encountered trying to convert to map value. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			Attribute: path,
		})
	}

	// we need our value to become a map of values so we can iterate over
	// them and handle them individually
	values := map[string]tftypes.Value{}
	err := val.As(&values)
	if err != nil {
		return target, append(diags, &tfprotov6.Diagnostic{
			Severity:  tfprotov6.DiagnosticSeverityError,
			Summary:   "Value Conversion Error",
			Detail:    "An unexpected error was encountered trying to convert to map value. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			Attribute: path,
		})
	}

	// we need to know the type the slice is wrapping
	elemType := underlyingValue.Type().Elem()
	elemAttrType := elemTyper.ElementType()

	// we want an empty version of the map
	m := reflect.MakeMapWithSize(underlyingValue.Type(), len(values))

	// go over each of the values passed in, create a Go value of the right
	// type for them, and add it to our new map
	for key, value := range values {
		// create a new Go value of the type that can go in the map
		targetValue := reflect.Zero(elemType)

		// update our path so we can have nice errors
		path := path.WithElementKeyString(key)

		// reflect the value into our new target
		result, elemDiags := BuildValue(ctx, elemAttrType, value, targetValue, opts, path)
		diags = append(diags, elemDiags...)

		if diagnostics.DiagsHasErrors(diags) {
			return target, diags
		}

		m.SetMapIndex(reflect.ValueOf(key), result)
	}

	return m, diags
}

// FromMap returns an attr.Value representing the data contained in `val`.
// `val` must be a map type with keys that are a string type. The attr.Value
// will be of the type produced by `typ`.
//
// It is meant to be called through OutOf, not directly.
func FromMap(ctx context.Context, typ attr.TypeWithElementType, val reflect.Value, path *tftypes.AttributePath) (attr.Value, []*tfprotov6.Diagnostic) {
	var diags []*tfprotov6.Diagnostic
	tfType := typ.TerraformType(ctx)

	if val.IsNil() {
		tfVal := tftypes.NewValue(tfType, nil)

		if typeWithValidate, ok := typ.(attr.TypeWithValidate); ok {
			diags = append(diags, typeWithValidate.Validate(ctx, tfVal)...)

			if diagnostics.DiagsHasErrors(diags) {
				return nil, diags
			}
		}

		attrVal, err := typ.ValueFromTerraform(ctx, tfVal)

		if err != nil {
			return nil, append(diags, &tfprotov6.Diagnostic{
				Severity:  tfprotov6.DiagnosticSeverityError,
				Summary:   "Value Conversion Error",
				Detail:    "An unexpected error was encountered trying to convert from map value. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
				Attribute: path,
			})
		}

		return attrVal, nil
	}

	elemType := typ.ElementType()
	tfElems := map[string]tftypes.Value{}
	for _, key := range val.MapKeys() {
		if key.Kind() != reflect.String {
			err := fmt.Errorf("map keys must be strings, got %s", key.Type())
			return nil, append(diags, &tfprotov6.Diagnostic{
				Severity:  tfprotov6.DiagnosticSeverityError,
				Summary:   "Value Conversion Error",
				Detail:    "An unexpected error was encountered trying to convert into a Terraform value. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
				Attribute: path,
			})
		}
		val, valDiags := FromValue(ctx, elemType, val.MapIndex(key).Interface(), path.WithElementKeyString(key.String()))
		diags = append(diags, valDiags...)

		if diagnostics.DiagsHasErrors(diags) {
			return nil, diags
		}
		tfVal, err := val.ToTerraformValue(ctx)
		if err != nil {
			return nil, append(diags, toTerraformValueErrorDiag(err, path))
		}

		tfElemType := elemType.TerraformType(ctx)
		err = tftypes.ValidateValue(tfElemType, tfVal)

		if err != nil {
			return nil, append(diags, validateValueErrorDiag(err, path))
		}

		tfElemVal := tftypes.NewValue(tfElemType, tfVal)

		if typeWithValidate, ok := typ.(attr.TypeWithValidate); ok {
			diags = append(diags, typeWithValidate.Validate(ctx, tfElemVal)...)

			if diagnostics.DiagsHasErrors(diags) {
				return nil, diags
			}
		}

		tfElems[key.String()] = tfElemVal
	}

	err := tftypes.ValidateValue(tfType, tfElems)
	if err != nil {
		return nil, append(diags, validateValueErrorDiag(err, path))
	}

	tfVal := tftypes.NewValue(tfType, tfElems)

	if typeWithValidate, ok := typ.(attr.TypeWithValidate); ok {
		diags = append(diags, typeWithValidate.Validate(ctx, tfVal)...)

		if diagnostics.DiagsHasErrors(diags) {
			return nil, diags
		}
	}

	attrVal, err := typ.ValueFromTerraform(ctx, tfVal)

	if err != nil {
		return nil, append(diags, &tfprotov6.Diagnostic{
			Severity:  tfprotov6.DiagnosticSeverityError,
			Summary:   "Value Conversion Error",
			Detail:    "An unexpected error was encountered trying to convert to map value. This is always an error in the provider. Please report the following to the provider developer:\n\n" + err.Error(),
			Attribute: path,
		})
	}

	return attrVal, diags
}
