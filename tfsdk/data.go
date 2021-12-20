package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type ReadOnlyData struct {
	Values types.Object
}

type GetDataOption struct{} // TODO: specify options for calling Get

type GetAttributeDataOption struct{} // TODO: specify options for calling GetAttribute

func (r ReadOnlyData) Get(ctx context.Context, target interface{}, options ...GetDataOption) diag.Diagnostics {
	var reflectOpts reflect.Options
	/*for _, option := range options {
		// TODO: apply to reflectOpts
	}*/
	return reflect.Into(ctx, r.Values, target, reflectOpts)
}

func (r ReadOnlyData) GetAttribute(ctx context.Context, path *tftypes.AttributePath, target interface{}, options ...GetAttributeDataOption) diag.Diagnostics {
	attrValue, diags := getAttributeValue(ctx, r.Values, path)
	if diags.HasError() {
		return diags
	}

	// TODO: handle attrValue == nil to avoid a panic

	valueAsDiags := ValueAs(ctx, attrValue, target)

	// ValueAs does not have path information for its Diagnostics
	for pos, valueAsDiag := range valueAsDiags {
		valueAsDiags[pos] = diag.WithPath(path, valueAsDiag)
	}

	diags.Append(valueAsDiags...)
	return diags
}

func getAttributeValue(ctx context.Context, data types.Object, path *tftypes.AttributePath) (attr.Value, diag.Diagnostics) {
	// TODO: need to walk into data to retrieve whatever's at path
	return nil, nil
}

type Data struct {
	ReadOnlyData
}

type SetDataOption struct{} // TODO: specify options for calling Set

type SetAttributeDataOption struct{} // TODO: specify options for calling SetAttribute

func (d *Data) SetAttribute(ctx context.Context, path *tftypes.AttributePath, val interface{}, options ...SetAttributeDataOption) diag.Diagnostics {
	attrValue, diags := getAttributeValue(ctx, d.Values, path)
	if diags.HasError() {
		return diags
	}

	newVal, newValDiags := reflect.FromValue(ctx, attrValue.Type(ctx), val, path)
	diags.Append(newValDiags...)
	if diags.HasError() {
		return diags
	}

	if attrTypeWithValidate, ok := attrValue.Type(ctx).(attr.TypeWithValidate); ok {
		tfVal, err := newVal.ToTerraformValue(ctx)
		if err != nil {
			err = fmt.Errorf("error running ToTerraformValue on new state value: %w", err)
			diags.AddAttributeError(
				path,
				"State Write Error",
				"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			)
			return diags
		}
		diags.Append(attrTypeWithValidate.Validate(ctx, tfVal, path)...)

		if diags.HasError() {
			return diags
		}
	}

	res, ds := setAttributeValue(ctx, path, d.Values, attrValue)
	diags.Append(ds...)
	if diags.HasError() {
		return diags
	}
	d.ReadOnlyData = ReadOnlyData{Values: res}

	return diags
}

func setAttributeValue(ctx context.Context, path *tftypes.AttributePath, object types.Object, val attr.Value) (types.Object, diag.Diagnostics) {
	// TODO: update object with val at path
	return types.Object{}, nil
}

func (d *Data) Set(ctx context.Context, val interface{}, options ...SetDataOption) diag.Diagnostics {
	return nil
}

func (d *Data) RemoveResource(ctx context.Context) {
	d.ReadOnlyData = ReadOnlyData{
		Values: types.Object{
			Null:      true,
			AttrTypes: d.ReadOnlyData.Values.AttrTypes,
		},
	}
}

func objectFromSchemaAndTerraformValue(ctx context.Context, schema Schema, val tftypes.Value) (types.Object, diag.Diagnostics) {
	if !val.Type().UsableAs(schema.TerraformType(ctx)) {
		// TODO: return error
	}

	res, err := schema.AttributeType().ValueFromTerraform(ctx, val)
	if err != nil {
		// TODO: return error
	}

	return res.(types.Object), nil
}
