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

type StringTypeWithValidateAttributeError struct {
	StringType
}

func (t StringTypeWithValidateAttributeError) Equal(o attr.Type) bool {
	other, ok := o.(StringTypeWithValidateAttributeError)
	if !ok {
		return false
	}
	return t == other
}

func (t StringTypeWithValidateAttributeError) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	newString, ok := val.(String)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", val)
	}

	newString.CreatedBy = t

	return StringValueWithValidateAttributeError{
		InternalString: newString,
	}, nil
}

var _ xattr.ValidateableAttribute = StringValueWithValidateAttributeError{}

type StringValueWithValidateAttributeError struct {
	InternalString String
}

func (v StringValueWithValidateAttributeError) Type(ctx context.Context) attr.Type {
	return v.InternalString.Type(ctx)
}

func (v StringValueWithValidateAttributeError) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return v.InternalString.ToTerraformValue(ctx)
}

func (v StringValueWithValidateAttributeError) Equal(value attr.Value) bool {
	other, ok := value.(StringValueWithValidateAttributeError)

	if !ok {
		return false
	}

	return v == other
}

func (v StringValueWithValidateAttributeError) IsNull() bool {
	return v.InternalString.IsNull()
}

func (v StringValueWithValidateAttributeError) IsUnknown() bool {
	return v.InternalString.IsUnknown()
}

func (v StringValueWithValidateAttributeError) IsFullyNullableKnown() bool {
	return v.InternalString.IsFullyNullableKnown()
}

func (v StringValueWithValidateAttributeError) String() string {
	return v.InternalString.String()
}

func (v StringValueWithValidateAttributeError) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	resp.Diagnostics.Append(TestErrorDiagnostic(req.Path))
}

type StringTypeWithValidateAttributeWarning struct {
	StringType
}

func (t StringTypeWithValidateAttributeWarning) Equal(o attr.Type) bool {
	other, ok := o.(StringTypeWithValidateAttributeWarning)
	if !ok {
		return false
	}
	return t == other
}

func (t StringTypeWithValidateAttributeWarning) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	newString, ok := val.(String)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", val)
	}

	newString.CreatedBy = t

	return StringValueWithValidateAttributeWarning{
		InternalString: newString,
	}, nil
}

var _ xattr.ValidateableAttribute = StringValueWithValidateAttributeWarning{}

type StringValueWithValidateAttributeWarning struct {
	InternalString String
}

func (v StringValueWithValidateAttributeWarning) Type(ctx context.Context) attr.Type {
	return v.InternalString.Type(ctx)
}

func (v StringValueWithValidateAttributeWarning) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return v.InternalString.ToTerraformValue(ctx)
}

func (v StringValueWithValidateAttributeWarning) Equal(value attr.Value) bool {
	other, ok := value.(StringValueWithValidateAttributeWarning)

	if !ok {
		return false
	}

	return v == other
}

func (v StringValueWithValidateAttributeWarning) IsNull() bool {
	return v.InternalString.IsNull()
}

func (v StringValueWithValidateAttributeWarning) IsUnknown() bool {
	return v.InternalString.IsUnknown()
}

func (v StringValueWithValidateAttributeWarning) IsFullyNullableKnown() bool {
	return v.InternalString.IsFullyNullableKnown()
}

func (v StringValueWithValidateAttributeWarning) String() string {
	return v.InternalString.String()
}

func (v StringValueWithValidateAttributeWarning) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	resp.Diagnostics.Append(TestWarningDiagnostic(req.Path))
}
