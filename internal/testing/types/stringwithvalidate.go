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

func (t StringTypeWithValidateError) Equal(o attr.Type) bool {
	other, ok := o.(StringTypeWithValidateError)
	if !ok {
		return false
	}
	return t == other
}

func (s StringTypeWithValidateError) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	res, err := s.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	newString := res.(String)
	newString.CreatedBy = s
	return newString, nil
}

type StringTypeWithValidateWarning struct {
	StringType
}

func (t StringTypeWithValidateError) Validate(ctx context.Context, in tftypes.Value, path *tftypes.AttributePath) diag.Diagnostics {
	return diag.Diagnostics{TestErrorDiagnostic(path)}
}

func (t StringTypeWithValidateWarning) Equal(o attr.Type) bool {
	other, ok := o.(StringTypeWithValidateWarning)
	if !ok {
		return false
	}
	return t == other
}

func (s StringTypeWithValidateWarning) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	res, err := s.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	newString := res.(String)
	newString.CreatedBy = s
	return newString, nil
}

func (t StringTypeWithValidateWarning) Validate(ctx context.Context, in tftypes.Value, path *tftypes.AttributePath) diag.Diagnostics {
	return diag.Diagnostics{TestWarningDiagnostic(path)}
}
