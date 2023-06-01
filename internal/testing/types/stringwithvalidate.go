// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ xattr.TypeWithValidate = StringTypeWithValidateError{}
	_ xattr.TypeWithValidate = StringTypeWithValidateWarning{}
)

type StringTypeWithValidateError struct {
	StringType
}

func (t StringTypeWithValidateError) Equal(o attr.Type) bool {
	other, ok := o.(StringTypeWithValidateError)
	if !ok {
		return false
	}
	return t == other
}

func (s StringTypeWithValidateError) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	res, err := s.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	newString, ok := res.(String)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", res)
	}
	newString.CreatedBy = s
	return newString, nil
}

type StringTypeWithValidateWarning struct {
	StringType
}

func (t StringTypeWithValidateError) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	return diag.Diagnostics{TestErrorDiagnostic(path)}
}

func (t StringTypeWithValidateWarning) Equal(o attr.Type) bool {
	other, ok := o.(StringTypeWithValidateWarning)
	if !ok {
		return false
	}
	return t == other
}

func (s StringTypeWithValidateWarning) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	res, err := s.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	newString, ok := res.(String)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", res)
	}
	newString.CreatedBy = s
	return newString, nil
}

func (t StringTypeWithValidateWarning) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	return diag.Diagnostics{TestWarningDiagnostic(path)}
}
