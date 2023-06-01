// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

func pointer[T any](value T) *T {
	return &value
}
