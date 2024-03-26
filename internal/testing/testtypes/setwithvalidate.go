// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testtypes

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	//nolint:staticcheck // xattr.TypeWithValidate is deprecated, but we still need to support it.
	_ xattr.TypeWithValidate = SetTypeWithValidateError{}
	//nolint:staticcheck // xattr.TypeWithValidate is deprecated, but we still need to support it.
	_ xattr.TypeWithValidate = SetTypeWithValidateWarning{}
)

type SetTypeWithValidateError struct {
	types.SetType
}

type SetTypeWithValidateWarning struct {
	types.SetType
}

func (t SetTypeWithValidateError) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	return diag.Diagnostics{TestErrorDiagnostic(path)}
}

func (t SetTypeWithValidateWarning) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	return diag.Diagnostics{TestWarningDiagnostic(path)}
}
