// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package attr

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types/refinement"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

const (
	// UnknownValueString should be returned by Value.String() implementations,
	// when Value.IsUnknown() returns true.
	UnknownValueString = "<unknown>"

	// NullValueString should be returned by Value.String() implementations
	// when Value.IsNull() returns true.
	NullValueString = "<null>"

	// UnsetValueString should be returned by Value.String() implementations
	// when Value does not contain sufficient information to display to users.
	//
	// This is primarily used for invalid Dynamic Value implementations.
	UnsetValueString = "<unset>"
)

// Value defines an interface for describing data associated with an attribute.
// Values allow provider developers to specify data in a convenient format, and
// have it transparently be converted to formats Terraform understands.
type Value interface {
	// Type returns the Type that created the Value.
	Type(context.Context) Type

	// ToTerraformValue returns the data contained in the Value as
	// a tftypes.Value.
	ToTerraformValue(context.Context) (tftypes.Value, error)

	// Equal should return true if the Value is considered type and data
	// value equivalent to the Value passed as an argument.
	//
	// Most types should verify the associated Type is exactly equal to prevent
	// potential data consistency issues. For example:
	//
	//  - basetypes.Number is inequal to basetypes.Int64 or basetypes.Float64
	//  - basetypes.String is inequal to a custom Go type that embeds it
	//
	// Additionally, most types should verify that known values are compared
	// to comply with Terraform's data consistency rules. For example:
	//
	//  - In a list, element order is significant
	//  - In a string, runes are compared byte-wise (e.g. whitespace is
	//    significant in JSON-encoded strings)
	//
	Equal(Value) bool

	// IsNull returns true if the Value is not set, or is explicitly set to null.
	IsNull() bool

	// IsUnknown returns true if the value is not yet known.
	IsUnknown() bool

	// String returns a summary representation of either the underlying Value,
	// or UnknownValueString (`<unknown>`) when IsUnknown() returns true,
	// or NullValueString (`<null>`) when IsNull() return true.
	//
	// This is an intentionally lossy representation, that are best suited for
	// logging and error reporting, as they are not protected by
	// compatibility guarantees within the framework.
	String() string
}

// ValueWithNotNullRefinement defines an interface describing a Value that can contain
// a refinement that indicates the Value is unknown, but will not be null once it becomes known.
//
// This interface is implemented by all base value types except for DynamicValue, as dynamic types
// in Terraform don't support value refinements.
type ValueWithNotNullRefinement interface {
	Value

	// NotNullRefinement returns a value refinement, if one exists, that indicates an unknown value
	// will not be null once it becomes known.
	NotNullRefinement() *refinement.NotNull
}

// TODO: Should we add interfaces for all the other refinements retrieval? Even though we don't need them ATM?
