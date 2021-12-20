package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attrpath"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ attr.TypeWithValidate = BoolTypeWithValidateError{}
	_ attr.TypeWithValidate = BoolTypeWithValidateWarning{}
)

type BoolTypeWithValidateError struct {
	BoolType
}

func (b BoolTypeWithValidateError) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	res, err := b.BoolType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	newBool := res.(Bool)
	newBool.CreatedBy = b
	return newBool, nil
}

type BoolTypeWithValidateWarning struct {
	BoolType
}

func (b BoolTypeWithValidateWarning) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	res, err := b.BoolType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}
	newBool := res.(Bool)
	newBool.CreatedBy = b
	return newBool, nil
}

func (t BoolTypeWithValidateError) Validate(ctx context.Context, in tftypes.Value, path attrpath.Path) diag.Diagnostics {
	return diag.Diagnostics{TestErrorDiagnostic(path)}
}

func (t BoolTypeWithValidateWarning) Validate(ctx context.Context, in tftypes.Value, path attrpath.Path) diag.Diagnostics {
	return diag.Diagnostics{TestWarningDiagnostic(path)}
}
