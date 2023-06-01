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
	_ xattr.TypeWithValidate = BoolTypeWithValidateError{}
	_ xattr.TypeWithValidate = BoolTypeWithValidateWarning{}
)

type BoolTypeWithValidateError struct {
	BoolType
}

func (b BoolTypeWithValidateError) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	res, err := b.BoolType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	newBool, ok := res.(Bool)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", res)
	}
	newBool.CreatedBy = b
	return newBool, nil
}

type BoolTypeWithValidateWarning struct {
	BoolType
}

func (b BoolTypeWithValidateWarning) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	res, err := b.BoolType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	newBool, ok := res.(Bool)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", res)
	}
	newBool.CreatedBy = b
	return newBool, nil
}

func (t BoolTypeWithValidateError) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	return diag.Diagnostics{TestErrorDiagnostic(path)}
}

func (t BoolTypeWithValidateWarning) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	return diag.Diagnostics{TestWarningDiagnostic(path)}
}
