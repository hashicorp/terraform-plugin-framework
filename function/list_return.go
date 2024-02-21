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
var _ Return = ListReturn{}

// ListReturn represents a function return that is an ordered collection of a
// single element type. Either the ElementType or CustomType field must be set.
//
// When setting the value for this return:
//
//   - If CustomType is set, use its associated value type.
//   - Otherwise, use [types.List] or a Go slice value type compatible with the
//     element type.
type ListReturn struct {
	// ElementType is the type for all elements of the list. This field must be
	// set.
	ElementType attr.Type

	// CustomType enables the use of a custom data type in place of the
	// default [basetypes.ListType]. When setting data, the
	// [basetypes.ListValuable] implementation associated with this custom
	// type must be used in place of [types.List].
	CustomType basetypes.ListTypable
}

// GetType returns the return data type.
func (r ListReturn) GetType() attr.Type {
	if r.CustomType != nil {
		return r.CustomType
	}

	return basetypes.ListType{
		ElemType: r.ElementType,
	}
}

// NewResultData returns a new result data based on the type.
func (r ListReturn) NewResultData(ctx context.Context) (ResultData, diag.Diagnostics) {
	value := basetypes.NewListUnknown(r.ElementType)

	if r.CustomType == nil {
		return NewResultData(value), nil
	}

	valuable, diags := r.CustomType.ValueFromList(ctx, value)

	return NewResultData(valuable), diags
}
