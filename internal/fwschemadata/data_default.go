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

		attrAtPath, attrAtPathDiags := d.Schema.AttributeAtPath(context.Background(), fwPath)

		diags.Append(attrAtPathDiags...)

		// Do not transform if schema attribute path cannot be retrieved.
		if attrAtPathDiags.HasError() {
			return tfTypeValue, nil
		}

		switch attrAtPath.(type) {
		case fwschema.AttributeWithBoolDefaultValue:
			attribWithBoolDefaultValue := attrAtPath.(fwschema.AttributeWithBoolDefaultValue)
			resp := defaults.BoolResponse{}
			attribWithBoolDefaultValue.DefaultValue().DefaultBool(ctx, defaults.BoolRequest{}, &resp)

			return tftypes.NewValue(tfTypeValue.Type(), resp.PlanValue.ValueBool()), nil
		}

		return tfTypeValue, nil
	})

	return diags
}
