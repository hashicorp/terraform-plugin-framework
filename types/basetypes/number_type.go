// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfrefinement "github.com/hashicorp/terraform-plugin-go/tftypes/refinement"
)

// NumberTypable extends attr.Type for number types.
// Implement this interface to create a custom NumberType type.
type NumberTypable interface {
	attr.Type

	// ValueFromNumber should convert the Number to a NumberValuable type.
	ValueFromNumber(context.Context, NumberValue) (NumberValuable, diag.Diagnostics)
}

var _ NumberTypable = NumberType{}

// NumberType is the base framework type for a floating point number.
// NumberValue is the associated value type.
type NumberType struct{}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the
// type.
func (t NumberType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("cannot apply AttributePathStep %T to %s", step, t.String())
}

// Equal returns true if the given type is equivalent.
func (t NumberType) Equal(o attr.Type) bool {
	_, ok := o.(NumberType)

	return ok
}

// String returns a human readable string of the type name.
func (t NumberType) String() string {
	return "basetypes.NumberType"
}

// TerraformType returns the tftypes.Type that should be used to represent this
// framework type.
func (t NumberType) TerraformType(_ context.Context) tftypes.Type {
	return tftypes.Number
}

// ValueFromNumber returns a NumberValuable type given a NumberValue.
func (t NumberType) ValueFromNumber(_ context.Context, v NumberValue) (NumberValuable, diag.Diagnostics) {
	return v, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.  This is meant to
// convert the tftypes.Value into a more convenient Go type for the provider to
// consume the data with.
func (t NumberType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		unknownVal := NewNumberUnknown()
		refinements := in.Refinements()

		if len(refinements) == 0 {
			return unknownVal, nil
		}

		for _, refn := range refinements {
			switch refnVal := refn.(type) {
			case tfrefinement.Nullness:
				if !refnVal.Nullness() {
					unknownVal = unknownVal.RefineAsNotNull()
				} else {
					// This scenario shouldn't occur, as Terraform should have already collapsed an
					// unknown value with a definitely null refinement into a known null value. However,
					// the protocol encoding does support this refinement value, so we'll also just collapse
					// it into a known null value here.
					return NewNumberNull(), nil
				}
			case tfrefinement.NumberLowerBound:
				unknownVal = unknownVal.RefineWithLowerBound(refnVal.LowerBound(), refnVal.IsInclusive())
			case tfrefinement.NumberUpperBound:
				unknownVal = unknownVal.RefineWithUpperBound(refnVal.UpperBound(), refnVal.IsInclusive())
			}
		}

		return unknownVal, nil
	}

	if in.IsNull() {
		return NewNumberNull(), nil
	}

	n := big.NewFloat(0)

	err := in.As(&n)

	if err != nil {
		return nil, err
	}

	return NewNumberValue(n), nil
}

// ValueType returns the Value type.
func (t NumberType) ValueType(_ context.Context) attr.Value {
	// This Value does not need to be valid.
	return NumberValue{}
}
