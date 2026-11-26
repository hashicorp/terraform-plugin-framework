// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package toproto6

import "github.com/hashicorp/terraform-plugin-go/tftypes"

// schemaError records names in reverse order as a conversion error unwinds.
type schemaError struct {
	message string
	steps   []string
}

func (e *schemaError) withAttributeName(name string) *schemaError {
	e.steps = append(e.steps, name)
	return e
}

func (e *schemaError) attributePathError(base *tftypes.AttributePath) error {
	for i := len(e.steps) - 1; i >= 0; i-- {
		base = base.WithAttributeName(e.steps[i])
	}
	return base.NewErrorf("%s", e.message)
}
