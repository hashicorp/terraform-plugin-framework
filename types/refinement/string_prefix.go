// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package refinement

import "fmt"

// StringPrefix represents an unknown value refinement that indicates the final value will be prefixed with the specified string value.
// String prefixes that exceed 256 characters in length will be truncated and empty string prefixes will not be encoded. This refinement can
// only be applied to the String type.
type StringPrefix struct {
	value string
}

func (s StringPrefix) Equal(other Refinement) bool {
	otherVal, ok := other.(StringPrefix)
	if !ok {
		return false
	}

	return s.PrefixValue() == otherVal.PrefixValue()
}

func (s StringPrefix) String() string {
	return fmt.Sprintf("prefix = %q", s.PrefixValue())
}

// PrefixValue returns the string value that the final value will be prefixed with.
func (s StringPrefix) PrefixValue() string {
	return s.value
}

func (s StringPrefix) unimplementable() {}

// NewStringPrefix returns the StringPrefix unknown value refinement that indicates the final value will be prefixed with the specified
// string value. String prefixes that exceed 256 characters in length will be truncated and empty string prefixes will not be encoded. This
// refinement can only be applied to the String type.
func NewStringPrefix(value string) Refinement {
	return StringPrefix{
		value: value,
	}
}
