package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// State represents a Terraform state.
type State[T any] struct {
	Raw    tftypes.Value
	Schema fwschema.Schema
}

// Get populates the struct passed as `target` with the entire state.
func (s State[T]) Get(ctx context.Context) (T, diag.Diagnostics) {
	var target T
	return target, s.data().Get(ctx, &target)
}

// GetAttribute retrieves the attribute or block found at `path` and populates
// the `target` with the value. This method is intended for top level schema
// attributes or blocks. Use `types` package methods or custom types to step
// into collections.
//
// Attributes or elements under null or unknown collections return null
// values, however this behavior is not protected by compatibility promises.
func (s State[T]) GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics {
	return s.data().GetAtPath(ctx, path, target)
}

// PathMatches returns all matching path.Paths from the given path.Expression.
//
// If a parent path is null or unknown, which would prevent a full expression
// from matching, the parent path is returned rather than no match to prevent
// false positives.
func (s State[T]) PathMatches(ctx context.Context, pathExpr path.Expression) (path.Paths, diag.Diagnostics) {
	return s.data().PathMatches(ctx, pathExpr)
}

// Set populates the entire state using the supplied Go value. The value `val`
// should be a struct whose values have one of the attr.Value types. Each field
// must be tagged with the corresponding schema field.
func (s *State[T]) Set(ctx context.Context, val T) diag.Diagnostics {
	data := s.data()
	diags := data.Set(ctx, val)

	if diags.HasError() {
		return diags
	}

	s.Raw = data.TerraformValue

	return diags
}

// SetAttribute sets the attribute at `path` using the supplied Go value.
//
// The attribute path and value must be valid with the current schema. If the
// attribute path already has a value, it will be overwritten. If the attribute
// path does not have a value, it will be added, including any parent attribute
// paths as necessary.
//
// The value must not be an untyped nil. Use a typed nil or types package null
// value function instead. For example with a types.StringType attribute,
// use (*string)(nil) or types.StringNull().
//
// Lists can only have the next element added according to the current length.
func (s *State[T]) SetAttribute(ctx context.Context, path path.Path, val interface{}) diag.Diagnostics {
	data := s.data()
	diags := data.SetAtPath(ctx, path, val)

	if diags.HasError() {
		return diags
	}

	s.Raw = data.TerraformValue

	return diags
}

// RemoveResource removes the entire resource from state.
//
// If a Resource type Delete method is completed without error, this is
// automatically called on the DeleteResourceResponse.State.
func (s *State[T]) RemoveResource(ctx context.Context) {
	s.Raw = tftypes.NewValue(s.Schema.Type().TerraformType(ctx), nil)
}

func (s State[T]) data() fwschemadata.Data {
	return fwschemadata.Data{
		Description:    fwschemadata.DataDescriptionState,
		Schema:         s.Schema,
		TerraformValue: s.Raw,
	}
}
