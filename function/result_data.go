// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwreflect "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// ResultData is the response data sent to Terraform for a single function call.
// Use the Set method in the Function type Run method to set the result data.
//
// For unit testing, use the NewResultData function to manually create the data
// for comparison.
type ResultData struct {
	value attr.Value
}

// Equal returns true if the value is equivalent.
func (d ResultData) Equal(o ResultData) bool {
	if d.value == nil {
		return o.value == nil
	}

	return d.value.Equal(o.value)
}

// Set saves the result data. The value type must be acceptable for the data
// type in the result definition.
func (d *ResultData) Set(ctx context.Context, value any) diag.Diagnostics {
	var diags diag.Diagnostics

	reflectValue, reflectDiags := fwreflect.FromValue(ctx, d.value.Type(ctx), value, path.Empty())

	diags.Append(reflectDiags...)

	if diags.HasError() {
		return diags
	}

	d.value = reflectValue

	return diags
}

// Value returns the saved value.
func (d ResultData) Value() attr.Value {
	return d.value
}

// NewResultData creates a ResultData. This is only necessary for unit testing
// as the framework automatically creates this data for the Function type Run
// method.
func NewResultData(value attr.Value) ResultData {
	return ResultData{
		value: value,
	}
}
