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
var _ Return = ObjectReturn{}

// ObjectReturn represents a function return that is mapping of defined
// attribute names to values. When setting the value for this return, use
// [types.Object] or a compatible Go struct as the value type unless the
// CustomType field is set. The AttributeTypes field must be set.
type ObjectReturn struct {
	// AttributeTypes is the mapping of underlying attribute names to attribute
	// types. This field must be set.
	AttributeTypes map[string]attr.Type

	// CustomType enables the use of a custom data type in place of the
	// default [basetypes.ObjectType]. When setting data, the
	// [basetypes.ObjectValuable] implementation associated with this custom
	// type must be used in place of [types.Object].
	CustomType basetypes.ObjectTypable
}

// GetType returns the return data type.
func (r ObjectReturn) GetType() attr.Type {
	if r.CustomType != nil {
		return r.CustomType
	}

	return basetypes.ObjectType{
		AttrTypes: r.AttributeTypes,
	}
}

// NewResultData returns a new result data based on the type.
func (r ObjectReturn) NewResultData(ctx context.Context) (ResultData, diag.Diagnostics) {
	value := basetypes.NewObjectUnknown(r.AttributeTypes)

	if r.CustomType == nil {
		return NewResultData(value), nil
	}

	valuable, diags := r.CustomType.ValueFromObject(ctx, value)

	return NewResultData(valuable), diags
}
