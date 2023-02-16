package fwschemadata

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromtftypes"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
)

// TransformDefaults walks the schema and applies schema defined default values
// when the rawConfig Data type contains a null value at the same path.
func (d *Data) TransformDefaults(ctx context.Context, configRaw tftypes.Value) diag.Diagnostics {
	var diags diag.Diagnostics

	configData := Data{
		Description:    DataDescriptionConfiguration,
		Schema:         d.Schema,
		TerraformValue: configRaw,
	}

	// Errors are handled as richer diag.Diagnostics instead.
	d.TerraformValue, _ = tftypes.Transform(d.TerraformValue, func(tfTypePath *tftypes.AttributePath, tfTypeValue tftypes.Value) (tftypes.Value, error) {
		fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, tfTypePath, d.Schema)

		diags.Append(fwPathDiags...)

		// Do not transform if path cannot be converted.
		// Checking against fwPathDiags will capture all errors.
		if fwPathDiags.HasError() {
			return tfTypeValue, nil
		}

		configValue, configValueDiags := configData.ValueAtPath(ctx, fwPath)

		diags.Append(configValueDiags...)

		// Do not transform if rawConfig value cannot be retrieved.
		if configValueDiags.HasError() {
			return tfTypeValue, nil
		}

		// Do not transform if rawConfig value is not null.
		if !configValue.IsNull() {
			return tfTypeValue, nil
		}

		// TODO: Handle blocks
		attrAtPath, attrAtPathDiags := d.Schema.AttributeAtPath(ctx, fwPath)

		diags.Append(attrAtPathDiags...)

		// Do not transform if schema attribute path cannot be retrieved.
		if attrAtPathDiags.HasError() {
			return tfTypeValue, nil
		}

		switch a := attrAtPath.(type) {
		case fwschema.AttributeWithBoolDefaultValue:
			defaultValue := a.DefaultValue()
			if defaultValue != nil {
				resp := defaults.BoolResponse{}
				defaultValue.DefaultBool(ctx, defaults.BoolRequest{}, &resp)
				return tftypes.NewValue(tfTypeValue.Type(), resp.PlanValue.ValueBool()), nil
			}
		case fwschema.AttributeWithFloat64DefaultValue:
			defaultValue := a.DefaultValue()
			if defaultValue != nil {
				resp := defaults.Float64Response{}
				defaultValue.DefaultFloat64(ctx, defaults.Float64Request{}, &resp)
				return tftypes.NewValue(tfTypeValue.Type(), resp.PlanValue.ValueFloat64()), nil
			}
		case fwschema.AttributeWithInt64DefaultValue:
			defaultValue := a.DefaultValue()
			if defaultValue != nil {
				resp := defaults.Int64Response{}
				defaultValue.DefaultInt64(ctx, defaults.Int64Request{}, &resp)
				return tftypes.NewValue(tfTypeValue.Type(), resp.PlanValue.ValueInt64()), nil
			}
		case fwschema.AttributeWithNumberDefaultValue:
			defaultValue := a.DefaultValue()
			if defaultValue != nil {
				resp := defaults.NumberResponse{}
				defaultValue.DefaultNumber(ctx, defaults.NumberRequest{}, &resp)
				return tftypes.NewValue(tfTypeValue.Type(), resp.PlanValue.ValueBigFloat()), nil
			}
		case fwschema.AttributeWithStringDefaultValue:
			defaultValue := a.DefaultValue()
			if defaultValue != nil {
				resp := defaults.StringResponse{}
				defaultValue.DefaultString(ctx, defaults.StringRequest{}, &resp)
				return tftypes.NewValue(tfTypeValue.Type(), resp.PlanValue.ValueString()), nil
			}
		}

		return tfTypeValue, nil
	})

	return diags
}
