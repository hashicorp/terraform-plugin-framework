// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types/validation"
)

type BoolTypeWithValidateParameterError struct {
	BoolType
}

func (t BoolTypeWithValidateParameterError) Equal(o attr.Type) bool {
	other, ok := o.(BoolTypeWithValidateParameterError)
	if !ok {
		return false
	}
	return t == other
}

func (t BoolTypeWithValidateParameterError) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := t.BoolType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	newBool, ok := val.(Bool)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", val)
	}

	newBool.CreatedBy = t

	return BoolValueWithValidateParameterError{
		newBool,
	}, nil
}

var _ validation.ValidateableParameter = BoolValueWithValidateParameterError{}

type BoolValueWithValidateParameterError struct {
	Bool
}

func (v BoolValueWithValidateParameterError) ValidateParameter(ctx context.Context, req validation.ValidateParameterRequest, resp *validation.ValidateParameterResponse) {
	resp.Error = function.NewArgumentFuncError(req.Position, "This is a function error")
}
