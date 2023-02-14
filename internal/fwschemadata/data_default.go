package fwschemadata

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromtftypes"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
)

// TransformPlanDefaults walks the schema and applies schema defined default values
// when the source Data type contains a null value at the same path.
// values. The reverse conversion is ReifyNullCollectionBlocks.
func (d *Data) TransformPlanDefaults(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics

	// Errors are handled as richer diag.Diagnostics instead.
	d.TerraformValue, _ = tftypes.Transform(d.TerraformValue, func(tfTypePath *tftypes.AttributePath, tfTypeValue tftypes.Value) (tftypes.Value, error) {
		// Do not transform if value is not null.
		if !tfTypeValue.IsNull() {
			return tfTypeValue, nil
		}

		fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, tfTypePath, d.Schema)

		diags.Append(fwPathDiags...)

		// Do not transform if path cannot be converted.
		// Checking against fwPathDiags will capture all errors.
		if fwPathDiags.HasError() {
			return tfTypeValue, nil
		}

		attrAtPath, attrAtPathDiags := d.Schema.AttributeAtPath(context.Background(), fwPath)

		diags.Append(attrAtPathDiags...)

		// Do not transform if schema attribute path cannot be retrieved.
		// Checking against fwPathDiags will capture all errors.
		if attrAtPathDiags.HasError() {
			return tfTypeValue, nil
		}

		switch attrAtPath.(type) {
		case fwschema.AttributeWithBoolDefaultValue:
			attribWithBoolDefaultValue := attrAtPath.(fwschema.AttributeWithBoolDefaultValue)
			defVal := attribWithBoolDefaultValue.DefaultValue()

			resp := defaults.BoolResponse{}
			defVal.DefaultBool(ctx, defaults.BoolRequest{}, &resp)

			return tftypes.NewValue(tfTypeValue.Type(), resp.PlanValue.ValueBool()), nil
		}

		return tfTypeValue, nil
	})

	return diags
}
