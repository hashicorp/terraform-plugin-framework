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

type ListTypeWithValidateAttributeError struct {
	types.ListType
}

func (t ListTypeWithValidateAttributeError) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := t.ListType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	list, ok := val.(types.List)
	if !ok {
		return nil, fmt.Errorf("cannot assert %T as types.List", val)
	}

	return ListValueWithValidateAttributeError{
		list,
	}, nil
}

var _ xattr.ValidateableAttribute = ListValueWithValidateAttributeError{}

type ListValueWithValidateAttributeError struct {
	types.List
}

func (v ListValueWithValidateAttributeError) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	resp.Diagnostics.Append(TestErrorDiagnostic(req.Path))
}

type ListTypeWithValidateAttributeWarning struct {
	types.ListType
}

func (t ListTypeWithValidateAttributeWarning) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := t.ListType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	list, ok := val.(types.List)
	if !ok {
		return nil, fmt.Errorf("cannot assert %T as types.List", val)
	}

	return ListValueWithValidateAttributeWarning{
		list,
	}, nil
}

var _ xattr.ValidateableAttribute = ListValueWithValidateAttributeWarning{}

type ListValueWithValidateAttributeWarning struct {
	types.List
}

func (v ListValueWithValidateAttributeWarning) Equal(o attr.Value) bool {
	other, ok := o.(ListValueWithValidateAttributeWarning)

	if !ok {
		return false
	}

	return v.List.Equal(other.List)
}

func (v ListValueWithValidateAttributeWarning) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	resp.Diagnostics.Append(TestWarningDiagnostic(req.Path))
}
