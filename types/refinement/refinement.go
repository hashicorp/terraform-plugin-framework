// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package refinement

import (
	"fmt"
	"sort"
	"strings"
)

type Key int64

func (k Key) String() string {
	switch k {
	case KeyNotNull:
		return "not_null"
	case KeyStringPrefix:
		return "string_prefix"
	case KeyNumberLowerBound:
		return "number_lower_bound"
	case KeyNumberUpperBound:
		return "number_upper_bound"
	case KeyCollectionLengthLowerBound:
		return "collection_length_lower_bound"
	case KeyCollectionLengthUpperBound:
		return "collection_length_upper_bound"
	default:
		return fmt.Sprintf("unsupported refinement: %d", k)
	}
}

const (
	// KeyNotNull represents a refinement that specifies that the final value will not be null.
	//
	// This refinement is relevant for all types except types.Dynamic.
	//
	// MAINTAINER NOTE: This is named slightly different from the terraform-plugin-go `Nullness` refinement it maps to.
	// This is done because framework only support nullness refinements that indicate an unknown value is definitely not null.
	// Values that are definitely null should be represented as a known null value instead.
	KeyNotNull = Key(1)

	// KeyStringPrefix represents a refinement that specifies a known prefix of a final string value.
	//
	// This refinement is only relevant for types.String.
	KeyStringPrefix = Key(2)

	// KeyNumberLowerBound represents a refinement that specifies the lower bound of possible values for a final number value.
	// The refinement data contains a boolean which indicates whether the bound is inclusive.
	//
	// This refinement is relevant for types.Int32, types.Int64, types.Float32, types.Float64, and types.Number.
	//
	// This Key is abstracted by the following refinements:
	//  - Int64LowerBound
	//  - Int32LowerBound
	//  - Float64LowerBound
	//  - Float32LowerBound
	//  - NumberLowerBound
	KeyNumberLowerBound = Key(3)

	// KeyNumberUpperBound represents a refinement that specifies the upper bound of possible values for a final number value.
	// The refinement data contains a boolean which indicates whether the bound is inclusive.
	//
	// This refinement is relevant for types.Int32, types.Int64, types.Float32, types.Float64, and types.Number.
	//
	// This Key is abstracted by the following refinements:
	//  - Int64UpperBound
	//  - Int32UpperBound
	//  - Float64UpperBound
	//  - Float32UpperBound
	//  - NumberUpperBound
	KeyNumberUpperBound = Key(4)

	// KeyCollectionLengthLowerBound represents a refinement that specifies the lower bound of possible length for a final collection value.
	//
	// This refinement is only relevant for types.List, types.Set, and types.Map.
	KeyCollectionLengthLowerBound = Key(5)

	// KeyCollectionLengthUpperBound represents a refinement that specifies the upper bound of possible length for a final collection value.
	//
	// This refinement is only relevant for types.List, types.Set, and types.Map.
	KeyCollectionLengthUpperBound = Key(6)
)

// Refinement represents an unknown value refinement with data constraints relevant to the final value. This interface can be asserted further
// with the associated structs in the `refinement` package to extract underlying refinement data.
type Refinement interface {
	// Equal should return true if the Refinement is considered equivalent to the
	// Refinement passed as an argument.
	Equal(Refinement) bool

	// String should return a human-friendly version of the Refinement.
	String() string

	unimplementable() // prevents external implementations, all refinements are defined in the Terraform/HCL type system go-cty.
}

// Refinements represents a map of unknown value refinement data.
type Refinements map[Key]Refinement

func (r Refinements) Equal(other Refinements) bool {
	if len(r) != len(other) {
		return false
	}

	for key, refnVal := range r {
		otherRefnVal, ok := other[key]
		if !ok {
			// Didn't find a refinement at the same key
			return false
		}

		if !refnVal.Equal(otherRefnVal) {
			// Refinement data is not equal
			return false
		}
	}

	return true
}
func (r Refinements) String() string {
	var res strings.Builder

	keys := make([]Key, 0, len(r))
	for k := range r {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(a, b int) bool { return keys[a] < keys[b] })
	for pos, key := range keys {
		if pos != 0 {
			res.WriteString(", ")
		}
		res.WriteString(r[key].String())
	}

	return res.String()
}
