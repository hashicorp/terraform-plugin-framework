// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	//nolint:staticcheck // xattr.TypeWithValidate is deprecated, but we still need to support it.
	_ xattr.TypeWithValidate = NumberTypeWithValidateError{}
	//nolint:staticcheck // xattr.TypeWithValidate is deprecated, but we still need to support it.
	_ xattr.TypeWithValidate = NumberTypeWithValidateWarning{}
)

type NumberTypeWithValidateError struct {
	NumberType
}

type NumberTypeWithValidateWarning struct {
	NumberType
}

func (t NumberTypeWithValidateError) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	return diag.Diagnostics{TestErrorDiagnostic(path)}
}

func (t NumberTypeWithValidateWarning) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	return diag.Diagnostics{TestWarningDiagnostic(path)}
}

func (n NumberTypeWithValidateError) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	res, err := n.NumberType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	newNumber, ok := res.(Number)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", res)
	}
	newNumber.CreatedBy = n
	return newNumber, nil
}

func (n NumberTypeWithValidateWarning) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	res, err := n.NumberType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	newNumber, ok := res.(Number)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", res)
	}
	newNumber.CreatedBy = n
	return newNumber, nil
}
