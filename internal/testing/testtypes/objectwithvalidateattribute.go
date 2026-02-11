// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ObjectTypeWithValidateAttributeError struct {
	types.ObjectType
}

func (t ObjectTypeWithValidateAttributeError) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := t.ObjectType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	list, ok := val.(types.Object)
	if !ok {
		return nil, fmt.Errorf("cannot assert %T as types.Object", val)
	}

	return ObjectValueWithValidateAttributeError{
		list,
	}, nil
}

var _ xattr.ValidateableAttribute = ObjectValueWithValidateAttributeError{}

type ObjectValueWithValidateAttributeError struct {
	types.Object
}

func (v ObjectValueWithValidateAttributeError) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	resp.Diagnostics.Append(TestErrorDiagnostic(req.Path))
}

type ObjectTypeWithValidateAttributeWarning struct {
	types.ObjectType
}

func (t ObjectTypeWithValidateAttributeWarning) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := t.ObjectType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	list, ok := val.(types.Object)
	if !ok {
		return nil, fmt.Errorf("cannot assert %T as types.Object", val)
	}

	return ObjectValueWithValidateAttributeWarning{
		list,
	}, nil
}

var _ xattr.ValidateableAttribute = ObjectValueWithValidateAttributeWarning{}

type ObjectValueWithValidateAttributeWarning struct {
	types.Object
}

func (v ObjectValueWithValidateAttributeWarning) Equal(o attr.Value) bool {
	other, ok := o.(ObjectValueWithValidateAttributeWarning)

	if !ok {
		return false
	}

	return v.Object.Equal(other.Object)
}

func (v ObjectValueWithValidateAttributeWarning) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	resp.Diagnostics.Append(TestWarningDiagnostic(req.Path))
}
