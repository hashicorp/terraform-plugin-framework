package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ attr.TypeWithValidate = SetTypeWithValidateError{}
	_ attr.TypeWithValidate = SetTypeWithValidateWarning{}
)

type SetTypeWithValidateError struct {
	types.SetType
}

type SetTypeWithValidateWarning struct {
	types.SetType
}

func (t SetTypeWithValidateError) Validate(ctx context.Context, in tftypes.Value, path *tftypes.AttributePath) diag.Diagnostics {
	return diag.Diagnostics{TestErrorDiagnostic(path)}
}

func (t SetTypeWithValidateWarning) Validate(ctx context.Context, in tftypes.Value, path *tftypes.AttributePath) diag.Diagnostics {
	return diag.Diagnostics{TestWarningDiagnostic(path)}
}
