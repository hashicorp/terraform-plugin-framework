// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package refinement

import "fmt"

// CollectionLengthUpperBound represents an unknown value refinement which indicates the length of the final collection value will be
// at most the specified int64 value. This refinement can only be applied to types.List, types.Map, and types.Set.
type CollectionLengthUpperBound struct {
	value int64
}

func (n CollectionLengthUpperBound) Equal(other Refinement) bool {
	otherVal, ok := other.(CollectionLengthUpperBound)
	if !ok {
		return false
	}

	return n.UpperBound() == otherVal.UpperBound()
}

func (n CollectionLengthUpperBound) String() string {
	return fmt.Sprintf("length upper bound = %d", n.UpperBound())
}

// UpperBound returns the int64 value that the final value's collection length will be at most.
func (n CollectionLengthUpperBound) UpperBound() int64 {
	return n.value
}

func (n CollectionLengthUpperBound) unimplementable() {}

// NewCollectionLengthUpperBound returns the CollectionLengthUpperBound unknown value refinement which indicates the length of the final
// collection value will be at most the specified int64 value. This refinement can only be applied to types.List, types.Map, and types.Set.
func NewCollectionLengthUpperBound(value int64) Refinement {
	return CollectionLengthUpperBound{
		value: value,
	}
}
