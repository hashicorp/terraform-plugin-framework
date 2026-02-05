// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

func pointer[T any](value T) *T {
	return &value
}
