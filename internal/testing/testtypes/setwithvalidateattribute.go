// Copyright IBM Corp. 2021, 2025
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

type SetTypeWithValidateAttributeError struct {
	types.SetType
}

func (t SetTypeWithValidateAttributeError) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := t.SetType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	set, ok := val.(types.Set)
	if !ok {
		return nil, fmt.Errorf("cannot assert %T as types.Set", val)
	}

	return SetValueWithValidateAttributeError{
		set,
	}, nil
}

var _ xattr.ValidateableAttribute = SetValueWithValidateAttributeError{}

type SetValueWithValidateAttributeError struct {
	types.Set
}

func (v SetValueWithValidateAttributeError) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	resp.Diagnostics.Append(TestErrorDiagnostic(req.Path))
}

type SetTypeWithValidateAttributeWarning struct {
	types.SetType
}

func (t SetTypeWithValidateAttributeWarning) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := t.SetType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	set, ok := val.(types.Set)
	if !ok {
		return nil, fmt.Errorf("cannot assert %T as types.Set", val)
	}

	return SetValueWithValidateAttributeWarning{
		set,
	}, nil
}

var _ xattr.ValidateableAttribute = SetValueWithValidateAttributeWarning{}

type SetValueWithValidateAttributeWarning struct {
	types.Set
}

func (v SetValueWithValidateAttributeWarning) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	resp.Diagnostics.Append(TestWarningDiagnostic(req.Path))
}
