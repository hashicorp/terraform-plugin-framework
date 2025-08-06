// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

const (
	// ExecutionOrderInvalid is used to indicate an invalid [ExecutionOrder].
	// Provider developers should not use it.
	ExecutionOrderInvalid ExecutionOrder = 0

	// ExecutionOrderBefore is used to indicate that the action must be invoked before it's
	// linked resource's plan/apply.
	ExecutionOrderBefore ExecutionOrder = 1

	// ExecutionOrderAfter is used to indicate that the action must be invoked after it's
	// linked resource's plan/apply.
	ExecutionOrderAfter ExecutionOrder = 2
)

// ExecutionOrder is an enum that represents when an action is invoked relative to it's linked resource.
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
