// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
)

type StringTypeWithValidateParameterError struct {
	StringType
}

func (t StringTypeWithValidateParameterError) Equal(o attr.Type) bool {
	other, ok := o.(StringTypeWithValidateParameterError)
	if !ok {
		return false
	}
	return t == other
}

func (t StringTypeWithValidateParameterError) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	val, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	newString, ok := val.(String)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", val)
	}

	newString.CreatedBy = t

	return StringValueWithValidateParameterError{
		InternalString: newString,
	}, nil
}

var _ function.ValidateableParameter = StringValueWithValidateParameterError{}

type StringValueWithValidateParameterError struct {
	InternalString String
}

func (v StringValueWithValidateParameterError) Type(ctx context.Context) attr.Type {
	return v.InternalString.Type(ctx)
}

func (v StringValueWithValidateParameterError) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return v.InternalString.ToTerraformValue(ctx)
}

func (v StringValueWithValidateParameterError) Equal(value attr.Value) bool {
	other, ok := value.(StringValueWithValidateParameterError)

	if !ok {
		return false
	}

	return v == other
}

func (v StringValueWithValidateParameterError) IsNull() bool {
	return v.InternalString.IsNull()
}

func (v StringValueWithValidateParameterError) IsUnknown() bool {
	return v.InternalString.IsUnknown()
}

func (v StringValueWithValidateParameterError) String() string {
	return v.InternalString.String()
}

func (v StringValueWithValidateParameterError) ValidateParameter(ctx context.Context, req function.ValidateParameterRequest, resp *function.ValidateParameterResponse) {
	resp.Error = function.NewArgumentFuncError(req.Position, "This is a function error")
}
