package tfsdk

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/totftypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Config represents a Terraform config.
type Config struct {
	Raw    tftypes.Value
	Schema Schema
}

// Get populates the struct passed as `target` with the entire config.
func (c Config) Get(ctx context.Context, target interface{}) diag.Diagnostics {
	return c.data().Get(ctx, target)
}

// GetAttribute retrieves the attribute found at `path` and populates the
// `target` with the value.
func (c Config) GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics {
	ctx = logging.FrameworkWithAttributePath(ctx, path.String())

	attrValue, diags := c.getAttributeValue(ctx, path)

	if diags.HasError() {
		return diags
	}

	if attrValue == nil {
		diags.AddAttributeError(
			path,
			"Config Read Error",
			"An unexpected error was encountered trying to read an attribute from the configuration. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				"Missing attribute value, however no error was returned. Preventing the panic from this situation.",
		)
		return diags
	}

	valueAsDiags := ValueAs(ctx, attrValue, target)

	// ValueAs does not have path information for its Diagnostics.
	for idx, valueAsDiag := range valueAsDiags {
		valueAsDiags[idx] = diag.WithPath(path, valueAsDiag)
	}

	diags.Append(valueAsDiags...)

	return diags
}

// PathMatches returns all matching path.Paths from the given path.Expression.
//
// If a parent path is null or unknown, which would prevent a full expression
// from matching, the parent path is returned rather than no match to prevent
// false positives.
func (c Config) PathMatches(ctx context.Context, pathExpr path.Expression) (path.Paths, diag.Diagnostics) {
	return c.data().PathMatches(ctx, pathExpr)
}

func (c Config) data() fwschemadata.Data {
	return fwschemadata.Data{
		Schema:         c.Schema,
		TerraformValue: c.Raw,
	}
}

// getAttributeValue retrieves the attribute found at `path` and returns it as an
// attr.Value. Consumers should assert the type of the returned value with the
// desired attr.Type.
func (c Config) getAttributeValue(ctx context.Context, path path.Path) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	tftypesPath, tftypesPathDiags := totftypes.AttributePath(ctx, path)

	diags.Append(tftypesPathDiags...)

	if diags.HasError() {
		return nil, diags
	}

	attrType, err := c.Schema.AttributeTypeAtPath(tftypesPath)
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

	tfValue, err := c.data().TerraformValueAtTerraformPath(ctx, tftypesPath)

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

	if attrTypeWithValidate, ok := attrType.(xattr.TypeWithValidate); ok {
		logging.FrameworkTrace(ctx, "Type implements TypeWithValidate")
		logging.FrameworkDebug(ctx, "Calling provider defined Type Validate")
		diags.Append(attrTypeWithValidate.Validate(ctx, tfValue, path)...)
		logging.FrameworkDebug(ctx, "Called provider defined Type Validate")

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
