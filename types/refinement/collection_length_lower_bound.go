// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package refinement

import "fmt"

// CollectionLengthLowerBound represents an unknown value refinement which indicates the length of the final collection value will be
// at least the specified int64 value. This refinement can only be applied to types.List, types.Map, and types.Set.
type CollectionLengthLowerBound struct {
	value int64
}

func (n CollectionLengthLowerBound) Equal(other Refinement) bool {
	otherVal, ok := other.(CollectionLengthLowerBound)
	if !ok {
		return false
	}

	return n.LowerBound() == otherVal.LowerBound()
}

func (n CollectionLengthLowerBound) String() string {
	return fmt.Sprintf("length lower bound = %d", n.LowerBound())
}

// LowerBound returns the int64 value that the final value's collection length will be at least.
func (n CollectionLengthLowerBound) LowerBound() int64 {
	return n.value
}

func (n CollectionLengthLowerBound) unimplementable() {}

// NewCollectionLengthLowerBound returns the CollectionLengthLowerBound unknown value refinement which indicates the length of the final
// collection value will be at least the specified int64 value. This refinement can only be applied to types.List, types.Map, and types.Set.
func NewCollectionLengthLowerBound(value int64) Refinement {
	return CollectionLengthLowerBound{
		value: value,
	}
}
