// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ DynamicValuable = DynamicValue{}
)

// DynamicValuable extends attr.Value for dynamic value types. Implement this interface
// to create a custom Dynamic value type.
type DynamicValuable interface {
	attr.Value

	// ToDynamicValue should convert the value type to a DynamicValue.
	ToDynamicValue(context.Context) (DynamicValue, diag.Diagnostics)
}

// DynamicValuableWithSemanticEquals extends DynamicValuable with semantic equality logic.
type DynamicValuableWithSemanticEquals interface {
	DynamicValuable

	// DynamicSemanticEquals should return true if the given value is
	// semantically equal to the current value. This logic is used to prevent
	// Terraform data consistency errors and resource drift where a value change
	// may have inconsequential differences.
	//
	// Only known values are compared with this method as changing a value's
	// state implicitly represents a different value.
	DynamicSemanticEquals(context.Context, DynamicValuable) (bool, diag.Diagnostics)
}

// TODO: doc
func NewDynamicValue(value attr.Value) DynamicValue {
	// TODO: validate that a known value is passed here?
	// TODO: validate that DynamicValue is NOT passed here?
	// 		- Treat like the object/list/map/set creation functions and return an error?
	// 		- If value == DynamicValue, throw error
	// 		- Introduce *Must function?
	return DynamicValue{
		value: value,
		state: attr.ValueStateKnown,
	}
}

// TODO: doc
func NewDynamicNull() DynamicValue {
	return DynamicValue{
		state: attr.ValueStateNull,
	}
}

// TODO: doc
func NewDynamicUnknown() DynamicValue {
	return DynamicValue{
		state: attr.ValueStateUnknown,
	}
}

// TODO: docs
type DynamicValue struct {
	// TODO: doc
	value attr.Value

	// state represents whether the value is null, unknown, or known. The
	// zero-value is null.
	state attr.ValueState
}

// Type returns DynamicType.
func (v DynamicValue) Type(ctx context.Context) attr.Type {
	// TODO: implement
	return DynamicType{}
}

// ToTerraformValue returns the equivalent tftypes.Value for the DynamicValue.
func (v DynamicValue) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	// TODO: should we check for a nil `v.value`?
	switch v.state {
	case attr.ValueStateKnown:
		return v.value.ToTerraformValue(ctx)
	case attr.ValueStateNull:
		return tftypes.NewValue(tftypes.DynamicPseudoType, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(tftypes.DynamicPseudoType, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Dynamic state in ToTerraformValue: %s", v.state))
	}
}

// Equal returns true if the given attr.Value is also a DynamicValue and contains an equal underlying value as defined by its Equal method.
func (v DynamicValue) Equal(o attr.Value) bool {
	other, ok := o.(DynamicValue)
	if !ok {
		return false
	}

	if v.state != other.state {
		return false
	}

	if v.state != attr.ValueStateKnown {
		return true
	}

	// TODO: should we check for a nil `v.value`?
	return v.value.Equal(other.value)
}

// IsNull returns true if the underlying value in the DynamicValue represents a null value.
func (v DynamicValue) IsNull() bool {
	return v.state == attr.ValueStateNull
}

// IsUnknown returns true if the underlying value in the DynamicValue represents an unknown value.
func (v DynamicValue) IsUnknown() bool {
	return v.state == attr.ValueStateUnknown
}

// String returns a human-readable representation of the DynamicValue. The string returned here is not protected by any compatibility guarantees,
// and is intended for logging and error reporting.
func (v DynamicValue) String() string {
	if v.IsUnknown() {
		return attr.UnknownValueString
	}

	if v.IsNull() {
		return attr.NullValueString
	}

	// TODO: should we check for a nil `v.value`?
	return v.value.String()
}

// ToDynamicValue returns DynamicValue.
func (v DynamicValue) ToDynamicValue(ctx context.Context) (DynamicValue, diag.Diagnostics) {
	return v, nil
}

// UnderlyingValue returns the underlying value in the DynamicValue.
// TODO: document that it will be nil if no underlying type or value
func (v DynamicValue) UnderlyingValue() attr.Value {
	return v.value
}
