package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ attr.TypeWithValidate = ListTypeWithValidateError{}
	_ attr.TypeWithValidate = ListTypeWithValidateWarning{}
)

type ListTypeWithValidateError struct {
	types.ListType
}

type ListTypeWithValidateWarning struct {
	types.ListType
}

func (t ListTypeWithValidateError) Validate(ctx context.Context, in tftypes.Value, path *tftypes.AttributePath) diag.Diagnostics {
	return diag.Diagnostics{TestErrorDiagnostic(path)}
}

func (t ListTypeWithValidateWarning) Validate(ctx context.Context, in tftypes.Value, path *tftypes.AttributePath) diag.Diagnostics {
	return diag.Diagnostics{TestWarningDiagnostic(path)}
}
