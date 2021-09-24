package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ attr.TypeWithValidate = MapTypeWithValidateError{}
	_ attr.TypeWithValidate = MapTypeWithValidateWarning{}
)

type MapTypeWithValidateError struct {
	types.MapType
}

type MapTypeWithValidateWarning struct {
	types.MapType
}

func (t MapTypeWithValidateError) Validate(ctx context.Context, in tftypes.Value, path *tftypes.AttributePath) diag.Diagnostics {
	return diag.Diagnostics{TestErrorDiagnostic(path)}
}

func (t MapTypeWithValidateWarning) Validate(ctx context.Context, in tftypes.Value, path *tftypes.AttributePath) diag.Diagnostics {
	return diag.Diagnostics{TestWarningDiagnostic(path)}
}
