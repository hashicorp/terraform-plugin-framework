// Copyright IBM Corp. 2021, 2025
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
	_ xattr.TypeWithValidate = MapTypeWithValidateError{}
	//nolint:staticcheck // xattr.TypeWithValidate is deprecated, but we still need to support it.
	_ xattr.TypeWithValidate = MapTypeWithValidateWarning{}
)

type MapTypeWithValidateError struct {
	types.MapType
}

type MapTypeWithValidateWarning struct {
	types.MapType
}

func (t MapTypeWithValidateError) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	return diag.Diagnostics{TestErrorDiagnostic(path)}
}

func (t MapTypeWithValidateWarning) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	return diag.Diagnostics{TestWarningDiagnostic(path)}
}
