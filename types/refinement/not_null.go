// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package refinement

// NotNull represents an unknown value refinement that indicates the final value will not be null. This refinement
// can be applied to a value of any type (excluding types.Dynamic).
type NotNull struct{}

func (n NotNull) Equal(other Refinement) bool {
	_, refnMatches := other.(NotNull)
	return refnMatches
}

func (n NotNull) String() string {
	return "not null"
}

func (n NotNull) unimplementable() {}

// NewNotNull returns the NotNull unknown value refinement that indicates the final value will not be null. This refinement
// can be applied to a value of any type (excluding types.Dynamic).
func NewNotNull() Refinement {
	return NotNull{}
}
