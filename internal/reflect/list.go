package reflect

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// build a slice of elements, matching the type of `target`, and fill it with
// the data in `val`.
func List(ctx context.Context, val attr.Value, target reflect.Value, opts Options, path *tftypes.AttributePath) (reflect.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	tfVal, err := val.ToTerraformValue(ctx)
	if err != nil {
		// TODO: handle error
	}

	// this only works with slices, so check that out first
	if target.Kind() != reflect.Slice {
		diags.Append(diag.WithPath(path, DiagIntoIncompatibleType{
			Val:        tfVal,
			TargetType: target.Type(),
			Err:        fmt.Errorf("expected a slice type, got %s", target.Type()),
		}))
		return target, diags
	}

	elemValuer, ok := val.(attr.ValueWithElements)
	if !ok {
		// TODO: handle error
	}
	values := elemValuer.Elements(ctx)

	// we need to know the type the slice is wrapping
	elemType := target.Type().Elem()

	// we want an empty version of the slice
	slice := reflect.MakeSlice(target.Type(), 0, len(values))

	// go over each of the values passed in, create a Go value of the right
	// type for them, and add it to our new slice
	for pos, value := range values {
		// create a new Go value of the type that can go in the slice
		targetValue := reflect.Zero(elemType)

		// update our path so we can have nice errors
		valPath := path.WithElementKeyInt(pos)

		// reflect the value into our new target
		val, valDiags := BuildValue(ctx, value, targetValue, opts, valPath)
		diags.Append(valDiags...)

		if diags.HasError() {
			return target, diags
		}

		// add the new target to our slice
		slice = reflect.Append(slice, val)
	}

	return slice, diags
}
