// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

// Function returns the Function for a given name.
func (s *Server) Function(ctx context.Context, name string) (function.Function, diag.Diagnostics) {
	functionFuncs, diags := s.FunctionFuncs(ctx)

	functionFunc, ok := functionFuncs[name]

	if !ok {
		diags.AddError(
			"Function Not Found",
			fmt.Sprintf("No function named %q was found in the provider.", name),
		)

		return nil, diags
	}

	return functionFunc(), diags
}

// FunctionDefinition returns the Function Definition for the given name and
// caches the result for later Function operations.
func (s *Server) FunctionDefinition(ctx context.Context, name string) (function.Definition, diag.Diagnostics) {
	s.functionDefinitionsMutex.RLock()
	functionDefinition, ok := s.functionDefinitions[name]
	s.functionDefinitionsMutex.RUnlock()

	if ok {
		return functionDefinition, nil
	}

	var diags diag.Diagnostics

	functionImpl, functionDiags := s.Function(ctx, name)

	diags.Append(functionDiags...)

	if diags.HasError() {
		return function.Definition{}, diags
	}

	definitionReq := function.DefinitionRequest{}
	definitionResp := function.DefinitionResponse{}

	logging.FrameworkTrace(ctx, "Calling provider defined Function Definition method", map[string]interface{}{logging.KeyFunctionName: name})
	functionImpl.Definition(ctx, definitionReq, &definitionResp)
	logging.FrameworkTrace(ctx, "Called provider defined Function Definition method", map[string]interface{}{logging.KeyFunctionName: name})

	diags.Append(definitionResp.Diagnostics...)

	if diags.HasError() {
		return definitionResp.Definition, diags
	}

	s.functionDefinitionsMutex.Lock()

	if s.functionDefinitions == nil {
		s.functionDefinitions = make(map[string]function.Definition)
	}

	s.functionDefinitions[name] = definitionResp.Definition

	s.functionDefinitionsMutex.Unlock()

	return definitionResp.Definition, diags
}

// FunctionDefinitions returns a map of Function Definitions for the
// GetProviderSchema RPC without caching since not all definitions are
// guaranteed to be necessary for later provider operations. The definition
// implementations are also validated.
func (s *Server) FunctionDefinitions(ctx context.Context) (map[string]function.Definition, diag.Diagnostics) {
	functionDefinitions := make(map[string]function.Definition)

	functionFuncs, diags := s.FunctionFuncs(ctx)

	for name, functionFunc := range functionFuncs {
		functionImpl := functionFunc()

		definitionReq := function.DefinitionRequest{}
		definitionResp := function.DefinitionResponse{}

		logging.FrameworkTrace(ctx, "Calling provider defined Function Definition", map[string]interface{}{logging.KeyFunctionName: name})
		functionImpl.Definition(ctx, definitionReq, &definitionResp)
		logging.FrameworkTrace(ctx, "Called provider defined Function Definition", map[string]interface{}{logging.KeyFunctionName: name})

		diags.Append(definitionResp.Diagnostics...)

		if definitionResp.Diagnostics.HasError() {
			continue
		}

		validateDiags := definitionResp.Definition.ValidateImplementation(ctx)

		diags.Append(validateDiags...)

		if validateDiags.HasError() {
			continue
		}

		functionDefinitions[name] = definitionResp.Definition
	}

	return functionDefinitions, diags
}

// FunctionFuncs returns a map of Function functions. The results are cached
// on first use.
func (s *Server) FunctionFuncs(ctx context.Context) (map[string]func() function.Function, diag.Diagnostics) {
	logging.FrameworkTrace(ctx, "Checking FunctionTypes lock")
	s.functionFuncsMutex.Lock()
	defer s.functionFuncsMutex.Unlock()

	if s.functionFuncs != nil {
		return s.functionFuncs, s.functionFuncsDiags
	}

	s.functionFuncs = make(map[string]func() function.Function)

	provider, ok := s.Provider.(provider.ProviderWithFunctions)

	if !ok {
		// Only function-specific RPCs should return diagnostics about the
		// provider not implementing functions or missing functions.
		return s.functionFuncs, s.functionFuncsDiags
	}

	logging.FrameworkTrace(ctx, "Calling provider defined Provider Functions")
	functionFuncs := provider.Functions(ctx)
	logging.FrameworkTrace(ctx, "Called provider defined Provider Functions")

	for _, functionFunc := range functionFuncs {
		functionImpl := functionFunc()

		metadataReq := function.MetadataRequest{}
		metadataResp := function.MetadataResponse{}

		functionImpl.Metadata(ctx, metadataReq, &metadataResp)

		if metadataResp.Name == "" {
			s.functionFuncsDiags.AddError(
				"Function Name Missing",
				fmt.Sprintf("The %T Function returned an empty string from the Metadata method. ", functionImpl)+
					"This is always an issue with the provider and should be reported to the provider developers.",
			)
			continue
		}

		logging.FrameworkTrace(ctx, "Found function", map[string]interface{}{logging.KeyFunctionName: metadataResp.Name})

		if _, ok := s.functionFuncs[metadataResp.Name]; ok {
			s.functionFuncsDiags.AddError(
				"Duplicate Function Name Defined",
				fmt.Sprintf("The %s function name was returned for multiple functions. ", metadataResp.Name)+
					"Function names must be unique. "+
					"This is always an issue with the provider and should be reported to the provider developers.",
			)
			continue
		}

		s.functionFuncs[metadataResp.Name] = functionFunc
	}

	return s.functionFuncs, s.functionFuncsDiags
}

// FunctionMetadatas returns a slice of FunctionMetadata for the GetMetadata
// RPC.
func (s *Server) FunctionMetadatas(ctx context.Context) ([]FunctionMetadata, diag.Diagnostics) {
	functionFuncs, diags := s.FunctionFuncs(ctx)

	functionMetadatas := make([]FunctionMetadata, 0, len(functionFuncs))

	for name := range functionFuncs {
		functionMetadatas = append(functionMetadatas, FunctionMetadata{
			Name: name,
		})
	}

	return functionMetadatas, diags
}
