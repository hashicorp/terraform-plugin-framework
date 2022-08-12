package tfsdk

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/internal/totftypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// State represents a Terraform state.
type State struct {
	Raw    tftypes.Value
	Schema Schema
}

// Get populates the struct passed as `target` with the entire state.
func (s State) Get(ctx context.Context, target interface{}) diag.Diagnostics {
	return s.data().Get(ctx, target)
}

// GetAttribute retrieves the attribute found at `path` and populates the
// `target` with the value.
func (s State) GetAttribute(ctx context.Context, path path.Path, target interface{}) diag.Diagnostics {
	ctx = logging.FrameworkWithAttributePath(ctx, path.String())

	attrValue, diags := s.getAttributeValue(ctx, path)

	if diags.HasError() {
		return diags
	}

	if attrValue == nil {
		diags.AddAttributeError(
			path,
			"State Read Error",
			"An unexpected error was encountered trying to read an attribute from the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				"Missing attribute value, however no error was returned. Preventing the panic from this situation.",
		)
		return diags
	}

	valueAsDiags := ValueAs(ctx, attrValue, target)

	// ValueAs does not have path information for its Diagnostics.
	for idx, valueAsDiag := range valueAsDiags {
		valueAsDiags[idx] = diag.WithPath(path, valueAsDiag)
	}

	diags.Append(valueAsDiags...)

	return diags
}

// getAttributeValue retrieves the attribute found at `path` and returns it as an
// attr.Value. Consumers should assert the type of the returned value with the
// desired attr.Type.
func (s State) getAttributeValue(ctx context.Context, path path.Path) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	tftypesPath, tftypesPathDiags := totftypes.AttributePath(ctx, path)

	diags.Append(tftypesPathDiags...)

	if diags.HasError() {
		return nil, diags
	}

	attrType, err := s.Schema.AttributeTypeAtPath(tftypesPath)
	if err != nil {
		err = fmt.Errorf("error getting attribute type in schema: %w", err)
		diags.AddAttributeError(
			path,
			"State Read Error",
			"An unexpected error was encountered trying to read an attribute from the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	// if the whole state is nil, the value of a valid attribute is also nil
	if s.Raw.IsNull() {
		return nil, nil
	}

	tfValue, err := s.data().TerraformValueAtTerraformPath(ctx, tftypesPath)

	// Ignoring ErrInvalidStep will allow this method to return a null value of the type.
	if err != nil && !errors.Is(err, tftypes.ErrInvalidStep) {
		diags.AddAttributeError(
			path,
			"State Read Error",
			"An unexpected error was encountered trying to read an attribute from the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	// TODO: If ErrInvalidStep, check parent paths for unknown value.
	//       If found, convert this value to an unknown value.
	// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/186

	if attrTypeWithValidate, ok := attrType.(xattr.TypeWithValidate); ok {
		logging.FrameworkTrace(ctx, "Type implements TypeWithValidate")
		logging.FrameworkDebug(ctx, "Calling provider defined Type Validate")
		diags.Append(attrTypeWithValidate.Validate(ctx, tfValue, path)...)
		logging.FrameworkDebug(ctx, "Called provider defined Type Validate")

		if diags.HasError() {
			return nil, diags
		}
	}

	attrValue, err := attrType.ValueFromTerraform(ctx, tfValue)

	if err != nil {
		diags.AddAttributeError(
			path,
			"State Read Error",
			"An unexpected error was encountered trying to read an attribute from the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	return attrValue, diags
}

// PathMatches returns all matching path.Paths from the given path.Expression.
//
// If a parent path is null or unknown, which would prevent a full expression
// from matching, the parent path is returned rather than no match to prevent
// false positives.
func (s State) PathMatches(ctx context.Context, pathExpr path.Expression) (path.Paths, diag.Diagnostics) {
	return s.data().PathMatches(ctx, pathExpr)
}

// Set populates the entire state using the supplied Go value. The value `val`
// should be a struct whose values have one of the attr.Value types. Each field
// must be tagged with the corresponding schema field.
func (s *State) Set(ctx context.Context, val interface{}) diag.Diagnostics {
	if val == nil {
		err := fmt.Errorf("cannot set nil as entire state; to remove a resource from state, call State.RemoveResource, instead")
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"State Read Error",
				"An unexpected error was encountered trying to write the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			),
		}
	}
	newStateAttrValue, diags := reflect.FromValue(ctx, s.Schema.Type(), val, path.Empty())
	if diags.HasError() {
		return diags
	}

	newState, err := newStateAttrValue.ToTerraformValue(ctx)
	if err != nil {
		err = fmt.Errorf("error running ToTerraformValue on state: %w", err)
		diags.AddError(
			"State Write Error",
			"An unexpected error was encountered trying to write the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	s.Raw = newState
	return diags
}

// SetAttribute sets the attribute at `path` using the supplied Go value.
//
// The attribute path and value must be valid with the current schema. If the
// attribute path already has a value, it will be overwritten. If the attribute
// path does not have a value, it will be added, including any parent attribute
// paths as necessary.
//
// Lists can only have the next element added according to the current length.
func (s *State) SetAttribute(ctx context.Context, path path.Path, val interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	ctx = logging.FrameworkWithAttributePath(ctx, path.String())

	tftypesPath, tftypesPathDiags := totftypes.AttributePath(ctx, path)

	diags.Append(tftypesPathDiags...)

	if diags.HasError() {
		return diags
	}

	attrType, err := s.Schema.AttributeTypeAtPath(tftypesPath)
	if err != nil {
		err = fmt.Errorf("error getting attribute type in schema: %w", err)
		diags.AddAttributeError(
			path,
			"State Write Error",
			"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	newVal, newValDiags := reflect.FromValue(ctx, attrType, val, path)
	diags.Append(newValDiags...)

	if diags.HasError() {
		return diags
	}

	tfVal, err := newVal.ToTerraformValue(ctx)
	if err != nil {
		err = fmt.Errorf("error running ToTerraformValue on new state value: %w", err)
		diags.AddAttributeError(
			path,
			"State Write Error",
			"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	if attrTypeWithValidate, ok := attrType.(xattr.TypeWithValidate); ok {
		logging.FrameworkTrace(ctx, "Type implements TypeWithValidate")
		logging.FrameworkDebug(ctx, "Calling provider defined Type Validate")
		diags.Append(attrTypeWithValidate.Validate(ctx, tfVal, path)...)
		logging.FrameworkDebug(ctx, "Called provider defined Type Validate")

		if diags.HasError() {
			return diags
		}
	}

	transformFunc, transformFuncDiags := s.setAttributeTransformFunc(ctx, path, tfVal, nil)
	diags.Append(transformFuncDiags...)

	if diags.HasError() {
		return diags
	}

	s.Raw, err = tftypes.Transform(s.Raw, transformFunc)
	if err != nil {
		err = fmt.Errorf("Cannot transform state: %w", err)
		diags.AddAttributeError(
			path,
			"State Write Error",
			"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return diags
	}

	return diags
}

// setAttributeTransformFunc recursively creates a value based on the current
// Plan values along the path. If the value at the path does not yet exist,
// this will perform recursion to add the child value to a parent value,
// creating the parent value if necessary.
func (s State) setAttributeTransformFunc(ctx context.Context, path path.Path, tfVal tftypes.Value, diags diag.Diagnostics) (func(*tftypes.AttributePath, tftypes.Value) (tftypes.Value, error), diag.Diagnostics) {
	exists, pathExistsDiags := s.data().PathExists(ctx, path)
	diags.Append(pathExistsDiags...)

	if diags.HasError() {
		return nil, diags
	}

	tftypesPath, tftypesPathDiags := totftypes.AttributePath(ctx, path)

	diags.Append(tftypesPathDiags...)

	if diags.HasError() {
		return nil, diags
	}

	if exists {
		// Overwrite existing value
		return func(p *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
			if p.Equal(tftypesPath) {
				return tfVal, nil
			}
			return v, nil
		}, diags
	}

	parentPath := path.ParentPath()
	parentTftypesPath := tftypesPath.WithoutLastStep()
	parentAttrType, err := s.Schema.AttributeTypeAtPath(parentTftypesPath)

	if err != nil {
		err = fmt.Errorf("error getting parent attribute type in schema: %w", err)
		diags.AddAttributeError(
			parentPath,
			"State Write Error",
			"An unexpected error was encountered trying to write an attribute to the state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	parentValue, err := s.data().TerraformValueAtTerraformPath(ctx, parentTftypesPath)

	if err != nil && !errors.Is(err, tftypes.ErrInvalidStep) {
		diags.AddAttributeError(
			parentPath,
			"Plan Read Error",
			"An unexpected error was encountered trying to read an attribute from the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	if parentValue.IsNull() || !parentValue.IsKnown() {
		// TODO: This will break when DynamicPsuedoType is introduced.
		// tftypes.Type should implement AttributePathStepper, but it currently does not.
		// When it does, we should use: tftypes.WalkAttributePath(s.Raw.Type(), parentPath)
		// Reference: https://github.com/hashicorp/terraform-plugin-go/issues/110
		parentType := parentAttrType.TerraformType(ctx)
		var childValue interface{}

		if !parentValue.IsKnown() {
			childValue = tftypes.UnknownValue
		}

		var parentValueDiags diag.Diagnostics
		parentValue, parentValueDiags = createParentValue(ctx, parentPath, parentType, childValue)
		diags.Append(parentValueDiags...)

		if diags.HasError() {
			return nil, diags
		}
	}

	var childValueDiags diag.Diagnostics
	childStep, _ := path.Steps().LastStep()
	parentValue, childValueDiags = upsertChildValue(ctx, parentPath, parentValue, childStep, tfVal)
	diags.Append(childValueDiags...)

	if diags.HasError() {
		return nil, diags
	}

	if attrTypeWithValidate, ok := parentAttrType.(xattr.TypeWithValidate); ok {
		logging.FrameworkTrace(ctx, "Type implements TypeWithValidate")
		logging.FrameworkDebug(ctx, "Calling provider defined Type Validate")
		diags.Append(attrTypeWithValidate.Validate(ctx, parentValue, parentPath)...)
		logging.FrameworkDebug(ctx, "Called provider defined Type Validate")

		if diags.HasError() {
			return nil, diags
		}
	}

	return s.setAttributeTransformFunc(ctx, parentPath, parentValue, diags)
}

// RemoveResource removes the entire resource from state.
//
// If a Resource type Delete method is completed without error, this is
// automatically called on the DeleteResourceResponse.State.
func (s *State) RemoveResource(ctx context.Context) {
	s.Raw = tftypes.NewValue(s.Schema.Type().TerraformType(ctx), nil)
}

func (s State) data() fwschemadata.Data {
	return fwschemadata.Data{
		Schema:         s.Schema,
		TerraformValue: s.Raw,
	}
}
