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

type BoolTypeWithValidateAttributeError struct {
	BoolType
}

func (t BoolTypeWithValidateAttributeError) Equal(o attr.Type) bool {
	other, ok := o.(BoolTypeWithValidateAttributeError)
	if !ok {
		return false
	}
	return t == other
}

func (t BoolTypeWithValidateAttributeError) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := t.BoolType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	newBool, ok := val.(Bool)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", val)
	}

	newBool.CreatedBy = t

	return BoolValueWithValidateAttributeError{
		newBool,
	}, nil
}

var _ xattr.ValidateableAttribute = BoolValueWithValidateAttributeError{}

type BoolValueWithValidateAttributeError struct {
	Bool
}

func (v BoolValueWithValidateAttributeError) Equal(o attr.Value) bool {
	ob, ok := o.(BoolValueWithValidateAttributeError)

	if !ok {
		return false
	}

	return v.Bool.Equal(ob.Bool)
}

func (v BoolValueWithValidateAttributeError) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	resp.Diagnostics.Append(TestErrorDiagnostic(req.Path))
}
