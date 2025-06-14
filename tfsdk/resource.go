// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Resource represents a Terraform resource.
type Resource struct {
	Raw    tftypes.Value
	Schema fwschema.Schema
}

// Get populates the struct passed as `target` with the resource.
func (r Resource) Get(ctx context.Context, target interface{}) diag.Diagnostics {
	return r.data().Get(ctx, target)
}

// GetAttribute retrieves the attribute or block found at `path` and populates
// the `target` with the value. This method is intended for top level schema
// attributes or blocks. Use `types` package methods or custom types to step
// into collections.
//
// Attributes or elements under null or unknown collections return null
// values, however this behavior is not protected by compatibility promises.
func (r Resource) GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics {
	return r.data().GetAtPath(ctx, path, target)
}

// PathMatches returns all matching path.Paths from the given path.Expression.
//
// If a parent path is null or unknown, which would prevent a full expression
// from matching, the parent path is returned rather than no match to prevent
// false positives.
func (r Resource) PathMatches(ctx context.Context, pathExpr path.Expression) (path.Paths, diag.Diagnostics) {
	return r.data().PathMatches(ctx, pathExpr)
}

// Set populates the entire identity using the supplied Go value. The value `val`
// should be a struct whose values have one of the attr.Value types. Each field
// must be tagged with the corresponding schema field.
func (r *Resource) Set(ctx context.Context, val interface{}) diag.Diagnostics {
	data := r.data()
	diags := data.Set(ctx, val)

	if diags.HasError() {
		return diags
	}

	r.Raw = data.TerraformValue

	return diags
}

func (r Resource) data() fwschemadata.Data {
	return fwschemadata.Data{
		Description:    fwschemadata.DataDescriptionResource,
		Schema:         r.Schema,
		TerraformValue: r.Raw,
	}
}
