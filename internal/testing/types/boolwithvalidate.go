package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ attr.TypeWithValidate = BoolTypeWithValidateError{}
	_ attr.TypeWithValidate = BoolTypeWithValidateWarning{}
)

type BoolTypeWithValidateError struct {
	BoolType
}

type BoolTypeWithValidateWarning struct {
	BoolType
}

func (t BoolTypeWithValidateError) Validate(ctx context.Context, in tftypes.Value) []*tfprotov6.Diagnostic {
	return []*tfprotov6.Diagnostic{TestErrorDiagnostic}
}

func (t BoolTypeWithValidateWarning) Validate(ctx context.Context, in tftypes.Value) []*tfprotov6.Diagnostic {
	return []*tfprotov6.Diagnostic{TestWarningDiagnostic}
}
