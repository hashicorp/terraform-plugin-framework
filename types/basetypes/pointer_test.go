// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package basetypes

func pointer[T any](value T) *T {
	return &value
}
