// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

var _ function.Function = &Function{}

// Declarative function.Function for unit testing.
type Function struct {
	// Function interface methods
	DefinitionMethod func(context.Context, function.DefinitionRequest, *function.DefinitionResponse)
	MetadataMethod   func(context.Context, function.MetadataRequest, *function.MetadataResponse)
	RunMethod        func(context.Context, function.RunRequest, *function.RunResponse)
}

// Definition satisfies the function.Function interface.
func (d *Function) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	if d.DefinitionMethod == nil {
		return
	}

	d.DefinitionMethod(ctx, req, resp)
}

// Metadata satisfies the function.Function interface.
func (d *Function) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	if d.MetadataMethod == nil {
		return
	}

	d.MetadataMethod(ctx, req, resp)
}

// Run satisfies the function.Function interface.
func (d *Function) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	if d.RunMethod == nil {
		return
	}

	d.RunMethod(ctx, req, resp)
}
