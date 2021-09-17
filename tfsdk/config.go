package tfsdk

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Config represents a Terraform config.
type Config struct {
	Raw    tftypes.Value
	Schema Schema
}

// Get populates the struct passed as `target` with the entire config.
func (c Config) Get(ctx context.Context, target interface{}) diag.Diagnostics {
	return reflect.Into(ctx, c.Schema.AttributeType(), c.Raw, target, reflect.Options{})
}

// GetAttribute retrieves the attribute found at `path` and populates the
// `target` with the value.
func (c Config) GetAttribute(ctx context.Context, path *tftypes.AttributePath, target interface{}) diag.Diagnostics {
	attrValue, diags := c.getAttributeValue(ctx, path)

	if attrValue == nil {
		return diags
	}

	valueAsDiags := ValueAs(ctx, attrValue, target)

	// ValueAs does not have path information for its Diagnostics.
	for idx, valueAsDiag := range valueAsDiags {
		if valueAsDiag.Severity() == diag.SeverityError {
			valueAsDiags[idx] = diag.NewAttributeErrorDiagnostic(
				path,
				valueAsDiag.Summary(),
				valueAsDiag.Detail(),
			)
		} else if valueAsDiag.Severity() == diag.SeverityWarning {
			valueAsDiags[idx] = diag.NewAttributeWarningDiagnostic(
				path,
				valueAsDiag.Summary(),
				valueAsDiag.Detail(),
			)
		}
	}

	diags.Append(valueAsDiags...)

	return diags
}

// getAttribute retrieves the attribute found at `path` and returns it as an
// attr.Value. Consumers should assert the type of the returned value with the
// desired attr.Type.
func (c Config) getAttributeValue(ctx context.Context, path *tftypes.AttributePath) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	attrType, err := c.Schema.AttributeTypeAtPath(path)
	if err != nil {
		diags.AddAttributeError(
			path,
			"Configuration Read Error",
			"An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	// if the whole config is nil, the value of a valid attribute is also
	// nil
	if c.Raw.IsNull() {
		return nil, nil
	}

	tfValue, err := c.terraformValueAtPath(path)

	// Ignoring ErrInvalidStep will allow this method to return a null value of the type.
	if err != nil && !errors.Is(err, tftypes.ErrInvalidStep) {
		diags.AddAttributeError(
			path,
			"Configuration Read Error",
			"An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	// TODO: If ErrInvalidStep, check parent paths for unknown value.
	//       If found, convert this value to an unknown value.
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/186

	if attrTypeWithValidate, ok := attrType.(attr.TypeWithValidate); ok {
		diags.Append(attrTypeWithValidate.Validate(ctx, tfValue, path)...)

		if diags.HasError() {
			return nil, diags
		}
	}

	attrValue, err := attrType.ValueFromTerraform(ctx, tfValue)

	if err != nil {
		diags.AddAttributeError(
			path,
			"Configuration Read Error",
			"An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	return attrValue, diags
}

func (c Config) terraformValueAtPath(path *tftypes.AttributePath) (tftypes.Value, error) {
	rawValue, remaining, err := tftypes.WalkAttributePath(c.Raw, path)
	if err != nil {
		return tftypes.Value{}, fmt.Errorf("%v still remains in the path: %w", remaining, err)
	}
	attrValue, ok := rawValue.(tftypes.Value)
	if !ok {
		return tftypes.Value{}, fmt.Errorf("got non-tftypes.Value result %v", rawValue)
	}
	return attrValue, err
}
