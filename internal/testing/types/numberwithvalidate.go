package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ attr.TypeWithValidate = NumberTypeWithValidateError{}
	_ attr.TypeWithValidate = NumberTypeWithValidateWarning{}
)

type NumberTypeWithValidateError struct {
	NumberType
}

type NumberTypeWithValidateWarning struct {
	NumberType
}

func (t NumberTypeWithValidateError) Validate(ctx context.Context, in tftypes.Value) []*tfprotov6.Diagnostic {
	return []*tfprotov6.Diagnostic{TestErrorDiagnostic}
}

func (t NumberTypeWithValidateWarning) Validate(ctx context.Context, in tftypes.Value) []*tfprotov6.Diagnostic {
	return []*tfprotov6.Diagnostic{TestWarningDiagnostic}
}
