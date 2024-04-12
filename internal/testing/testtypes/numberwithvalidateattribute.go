// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
)

type NumberTypeWithValidateAttributeError struct {
	NumberType
}

func (t NumberTypeWithValidateAttributeError) Equal(o attr.Type) bool {
	other, ok := o.(NumberTypeWithValidateAttributeError)
	if !ok {
		return false
	}
	return t == other
}

func (t NumberTypeWithValidateAttributeError) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := t.NumberType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	newNumber, ok := val.(Number)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", val)
	}

	newNumber.CreatedBy = t

	return NumberValueWithValidateAttributeError{
		InternalNumber: newNumber,
	}, nil
}

var _ xattr.ValidateableAttribute = NumberValueWithValidateAttributeError{}

type NumberValueWithValidateAttributeError struct {
	InternalNumber Number
}

func (v NumberValueWithValidateAttributeError) Type(ctx context.Context) attr.Type {
	return v.InternalNumber.Type(ctx)
}

func (v NumberValueWithValidateAttributeError) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return v.InternalNumber.ToTerraformValue(ctx)
}

func (v NumberValueWithValidateAttributeError) Equal(value attr.Value) bool {
	other, ok := value.(NumberValueWithValidateAttributeError)

	if !ok {
		return false
	}

	return v == other
}

func (v NumberValueWithValidateAttributeError) IsNull() bool {
	return v.InternalNumber.IsNull()
}

func (v NumberValueWithValidateAttributeError) IsUnknown() bool {
	return v.InternalNumber.IsUnknown()
}

func (v NumberValueWithValidateAttributeError) String() string {
	return v.InternalNumber.String()
}

func (v NumberValueWithValidateAttributeError) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	resp.Diagnostics.Append(TestErrorDiagnostic(req.Path))
}

type NumberTypeWithValidateAttributeWarning struct {
	NumberType
}

func (t NumberTypeWithValidateAttributeWarning) Equal(o attr.Type) bool {
	other, ok := o.(NumberTypeWithValidateAttributeWarning)
	if !ok {
		return false
	}
	return t == other
}

func (t NumberTypeWithValidateAttributeWarning) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := t.NumberType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	newNumber, ok := val.(Number)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", val)
	}

	newNumber.CreatedBy = t

	return NumberValueWithValidateAttributeWarning{
		InternalNumber: newNumber,
	}, nil
}

var _ xattr.ValidateableAttribute = NumberValueWithValidateAttributeWarning{}

type NumberValueWithValidateAttributeWarning struct {
	InternalNumber Number
}

func (v NumberValueWithValidateAttributeWarning) Type(ctx context.Context) attr.Type {
	return v.InternalNumber.Type(ctx)
}

func (v NumberValueWithValidateAttributeWarning) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return v.InternalNumber.ToTerraformValue(ctx)
}

func (v NumberValueWithValidateAttributeWarning) Equal(value attr.Value) bool {
	other, ok := value.(NumberValueWithValidateAttributeWarning)

	if !ok {
		return false
	}

	return v.InternalNumber.Number.Equal(other.InternalNumber.Number)
}

func (v NumberValueWithValidateAttributeWarning) IsNull() bool {
	return v.InternalNumber.IsNull()
}

func (v NumberValueWithValidateAttributeWarning) IsUnknown() bool {
	return v.InternalNumber.IsUnknown()
}

func (v NumberValueWithValidateAttributeWarning) String() string {
	return v.InternalNumber.String()
}

func (v NumberValueWithValidateAttributeWarning) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	resp.Diagnostics.Append(TestWarningDiagnostic(req.Path))
}
