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

type MapTypeWithValidateAttributeError struct {
	types.MapType
}

func (t MapTypeWithValidateAttributeError) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := t.MapType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	m, ok := val.(types.Map)
	if !ok {
		return nil, fmt.Errorf("cannot assert %T as types.Map", val)
	}

	return MapValueWithValidateAttributeError{
		m,
	}, nil
}

var _ xattr.ValidateableAttribute = MapValueWithValidateAttributeError{}

type MapValueWithValidateAttributeError struct {
	types.Map
}

func (v MapValueWithValidateAttributeError) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	resp.Diagnostics.Append(TestErrorDiagnostic(req.Path))
}

type MapTypeWithValidateAttributeWarning struct {
	types.MapType
}

func (t MapTypeWithValidateAttributeWarning) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := t.MapType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	m, ok := val.(types.Map)
	if !ok {
		return nil, fmt.Errorf("cannot assert %T as types.Map", val)
	}

	return MapValueWithValidateAttributeWarning{
		m,
	}, nil
}

var _ xattr.ValidateableAttribute = MapValueWithValidateAttributeWarning{}

type MapValueWithValidateAttributeWarning struct {
	types.Map
}

func (v MapValueWithValidateAttributeWarning) Equal(o attr.Value) bool {
	other, ok := o.(MapValueWithValidateAttributeWarning)

	if !ok {
		return false
	}

	return v.Map.Equal(other.Map)
}

func (v MapValueWithValidateAttributeWarning) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	resp.Diagnostics.Append(TestWarningDiagnostic(req.Path))
}
