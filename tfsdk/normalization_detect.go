package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// undoNormalizationDifferences looks for situations where attribute values
// in the new object differ from the corresponding attributes in the
// old object only to the extent that the schema types still consider
// the values to be equivalent, and constructs a resulting object which
// prefers to keep the values from the configuration in that case.
//
// The "old" object typically represents prior state, while the "new"
// object represents a new value that would eventually supersede it, such as
// the result of refreshing or the "proposed new value" provided by Terraform
// Core during planning.
//
// This compromise allows us to respect the higher-level definitions of
// equality implemented by schema types while also remaining compatible with
// Terraform Core's stricter definition of equality. The configuration value
// must "win" in such cases, rather than the proposed (refreshed or planned)
// value, so that a reference to an attribute from elsewhere in the Terraform
// module will see exactly the value the module author wrote, rather than the
// modified value the provider or remote system generated.
func undoNormalizationDifferences(ctx context.Context, old, new tftypes.Value, schema Schema) tftypes.Value {
	if old.IsNull() || !old.IsKnown() {
		// No old values, so we'll just keep the new value exactly as-is.
		return new
	}

	// We use "new" as the basis of our Transform here because if there's
	// anything additional in "new" that wasn't present in "old" then we
	// want to retain that unique portion of "new" verbatim, rather than
	// discarding it to match the overall shape of "old".
	result, _ := tftypes.Transform(new, func(path *tftypes.AttributePath, val tftypes.Value) (tftypes.Value, error) {
		attribute, err := schema.AttributeAtPath(path)
		if err != nil {
			// Equality rules are only defined for entire attributes, so
			// we have nothing to change if we're not referring to exactly
			// an attribute.
			return val, nil
		}

		ty := attribute.Type
		if ty == nil {
			// Equality rules are only for leaf attributes, which can define
			// a custom type with its own equality rule.
			return val, nil
		}

		oldValRaw, _, err := tftypes.WalkAttributePath(old, path)
		if err != nil {
			// If the old value doesn't even have this attribute then
			// we'll just keep the new one.
			return val, nil
		}
		oldVal := oldValRaw.(tftypes.Value)

		// If we finally get here then that means we _do_ have corresponding
		// old and new values for this particular attribute, and so we might
		// possibly choose to keep oldVal instead of val.
		schemaVal, err := ty.ValueFromTerraform(ctx, val)
		if err != nil {
			// The new value isn't valid for its defined type? Weird.
			return val, nil
		}

		schemaOldVal, err := ty.ValueFromTerraform(ctx, oldVal)
		if err != nil {
			// The old value isn't valid for its defined type? Also weird.
			return val, nil
		}

		// Finally, if the schema-level type considers these values to be
		// equal then this is the one case where we will keep the old value.
		if schemaOldVal.Equal(schemaVal) {
			return oldVal, nil
		}
		return val, nil
	})
	return result
}
