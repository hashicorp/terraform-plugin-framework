// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
)

// Function returns the *tfprotov5.Function for a function.Definition.
func Function(ctx context.Context, fw function.Definition) *tfprotov5.Function {
	proto := &tfprotov5.Function{
		DeprecationMessage: fw.DeprecationMessage,
		Parameters:         make([]*tfprotov5.FunctionParameter, 0, len(fw.Parameters)),
		Return:             FunctionReturn(ctx, fw.Return),
		Summary:            fw.Summary,
	}

	if fw.MarkdownDescription != "" {
		proto.Description = fw.MarkdownDescription
		proto.DescriptionKind = tfprotov5.StringKindMarkdown
	} else if fw.Description != "" {
		proto.Description = fw.Description
		proto.DescriptionKind = tfprotov5.StringKindPlain
	}

	for i, fwParameter := range fw.Parameters {
		proto.Parameters = append(proto.Parameters, FunctionParameter(ctx, fw, fwParameter, i))
	}

	if fw.VariadicParameter != nil {
		proto.VariadicParameter = FunctionParameter(ctx, fw, fw.VariadicParameter, len(fw.Parameters)+1)
	}

	return proto
}

// FunctionParameter returns the *tfprotov5.FunctionParameter for a
// function.Parameter.
func FunctionParameter(ctx context.Context, def function.Definition, param function.Parameter, position int) *tfprotov5.FunctionParameter {
	if param == nil {
		return nil
	}

	// TODO: what should we do with the diags? This should never happen
	name, _ := def.ParameterName(ctx, position)

	proto := &tfprotov5.FunctionParameter{
		AllowNullValue:     param.GetAllowNullValue(),
		AllowUnknownValues: param.GetAllowUnknownValues(),
		Name:               name,
		Type:               param.GetType().TerraformType(ctx),
	}

	if param.GetMarkdownDescription() != "" {
		proto.Description = param.GetMarkdownDescription()
		proto.DescriptionKind = tfprotov5.StringKindMarkdown
	} else if param.GetDescription() != "" {
		proto.Description = param.GetDescription()
		proto.DescriptionKind = tfprotov5.StringKindPlain
	}

	return proto
}

// FunctionMetadata returns the tfprotov5.FunctionMetadata for a
// fwserver.FunctionMetadata.
func FunctionMetadata(ctx context.Context, fw fwserver.FunctionMetadata) tfprotov5.FunctionMetadata {
	proto := tfprotov5.FunctionMetadata{
		Name: fw.Name,
	}

	return proto
}

// FunctionReturn returns the *tfprotov5.FunctionReturn for a
// function.Return.
func FunctionReturn(ctx context.Context, fw function.Return) *tfprotov5.FunctionReturn {
	if fw == nil {
		return nil
	}

	proto := &tfprotov5.FunctionReturn{
		Type: fw.GetType().TerraformType(ctx),
	}

	return proto
}

// FunctionResultData returns the *tfprotov5.DynamicValue for a given
// function.ResultData.
func FunctionResultData(ctx context.Context, data function.ResultData) (*tfprotov5.DynamicValue, *function.FuncError) {
	attrValue := data.Value()

	if attrValue == nil {
		return nil, nil
	}

	tfType := attrValue.Type(ctx).TerraformType(ctx)
	tfValue, err := attrValue.ToTerraformValue(ctx)

	if err != nil {
		msg := "Unable to Convert Function Result Data: An unexpected error was encountered when converting the function result data to the protocol type. " +
			"Please report this to the provider developer:\n\n" +
			"Unable to convert framework type to tftypes: " + err.Error()

		return nil, function.NewFuncError(msg)
	}

	dynamicValue, err := tfprotov5.NewDynamicValue(tfType, tfValue)

	if err != nil {
		msg := "Unable to Convert Function Result Data: An unexpected error was encountered when converting the function result data to the protocol type. " +
			"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n" +
			"Unable to create DynamicValue: " + err.Error()

		return nil, function.NewFuncError(msg)
	}

	return &dynamicValue, nil
}
