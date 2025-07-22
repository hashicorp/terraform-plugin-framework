// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

// TODO:Actions: docs
const (
	ExecutionOrderInvalid ExecutionOrder = 0

	ExecutionOrderBefore ExecutionOrder = 1

	ExecutionOrderAfter ExecutionOrder = 2
)

type ExecutionOrder int32

func (d ExecutionOrder) String() string {
	switch d {
	case 0:
		return "Invalid"
	case 1:
		return "Before"
	case 2:
		return "After"
	}
	return "Unknown"
}
