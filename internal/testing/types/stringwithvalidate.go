package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ attr.TypeWithValidate = StringTypeWithValidateError{}
	_ attr.TypeWithValidate = StringTypeWithValidateWarning{}
)

type StringTypeWithValidateError struct {
	StringType
}

type StringTypeWithValidateWarning struct {
	StringType
}

func (t StringTypeWithValidateError) Validate(ctx context.Context, in tftypes.Value) diag.Diagnostics {
	return diag.Diagnostics{TestErrorDiagnostic}
}

func (t StringTypeWithValidateWarning) Validate(ctx context.Context, in tftypes.Value) diag.Diagnostics {
	return diag.Diagnostics{TestWarningDiagnostic}
}
