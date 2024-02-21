// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisifies the desired interfaces.
var _ Return = SetReturn{}

// SetReturn represents a function return that is an unordered collection of a
// single element type. Either the ElementType or CustomType field must be set.
//
// When setting the value for this return:
//
//   - If CustomType is set, use its associated value type.
//   - Otherwise, use [types.Set] or a Go slice value type compatible with the
//     element type.
type SetReturn struct {
	// ElementType is the type for all elements of the set. This field must be
	// set.
	ElementType attr.Type

	// CustomType enables the use of a custom data type in place of the
	// default [basetypes.SetType]. When setting data, the
	// [basetypes.SetValuable] implementation associated with this custom
	// type must be used in place of [types.Set].
	CustomType basetypes.SetTypable
}

// GetType returns the return data type.
func (r SetReturn) GetType() attr.Type {
	if r.CustomType != nil {
		return r.CustomType
	}

	return basetypes.SetType{
		ElemType: r.ElementType,
	}
}

// NewResultData returns a new result data based on the type.
func (r SetReturn) NewResultData(ctx context.Context) (ResultData, diag.Diagnostics) {
	value := basetypes.NewSetUnknown(r.ElementType)

	if r.CustomType == nil {
		return NewResultData(value), nil
	}

	valuable, diags := r.CustomType.ValueFromSet(ctx, value)

	return NewResultData(valuable), diags
}
