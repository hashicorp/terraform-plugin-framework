// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

// Action returns the Action for a given type name.
func (s *Server) Action(ctx context.Context, typeName string) (action.Action, diag.Diagnostics) {
	actionFuncs, diags := s.ActionFuncs(ctx)

	actionFunc, ok := actionFuncs[typeName]

	if !ok {
		diags.AddError(
			"Action Type Not Found",
			fmt.Sprintf("No action type named %q was found in the provider.", typeName),
		)

		return nil, diags
	}

	return actionFunc(), diags
}

// ActionFuncs returns a map of Action functions. The results are cached
// on first use.
func (s *Server) ActionFuncs(ctx context.Context) (map[string]func() action.Action, diag.Diagnostics) {
	logging.FrameworkTrace(ctx, "Checking ActionFuncs lock")
	s.actionFuncsMutex.Lock()
	defer s.actionFuncsMutex.Unlock()

	if s.actionFuncs != nil {
		return s.actionFuncs, s.actionFuncsDiags
	}

	providerTypeName := s.ProviderTypeName(ctx)
	s.actionFuncs = make(map[string]func() action.Action)

	actionProvider, ok := s.Provider.(provider.ProviderWithActions)

	if !ok {
		// Only action resource specific RPCs should return diagnostics about the
		// provider not implementing action resources or missing action resources.
		return s.actionFuncs, s.actionFuncsDiags
	}

	logging.FrameworkTrace(ctx, "Calling provider defined Provider Actions")
	actionFuncsSlice := actionProvider.Actions(ctx)
	logging.FrameworkTrace(ctx, "Called provider defined Provider Actions")

	for _, actionFunc := range actionFuncsSlice {
		resAction := actionFunc()

		actionTypeNameReq := action.MetadataRequest{
			ProviderTypeName: providerTypeName,
		}
		actionTypeNameResp := action.MetadataResponse{}

		resAction.Metadata(ctx, actionTypeNameReq, &actionTypeNameResp)

		if actionTypeNameResp.TypeName == "" {
			s.actionFuncsDiags.AddError(
				"Action Resource Type Name Missing",
				fmt.Sprintf("The %T Action returned an empty string from the Metadata method. ", resAction)+
					"This is always an issue with the provider and should be reported to the provider developers.",
			)
			continue
		}

		logging.FrameworkTrace(ctx, "Found action resource type", map[string]interface{}{logging.KeyActionType: actionTypeNameResp.TypeName})

		if _, ok := s.actionFuncs[actionTypeNameResp.TypeName]; ok {
			s.actionFuncsDiags.AddError(
				"Duplicate Action Resource Type Defined",
				fmt.Sprintf("The %s action resource type name was returned for multiple action resources. ", actionTypeNameResp.TypeName)+
					"Action resource type names must be unique. "+
					"This is always an issue with the provider and should be reported to the provider developers.",
			)
			continue
		}

		s.actionFuncs[actionTypeNameResp.TypeName] = actionFunc
	}

	return s.actionFuncs, s.actionFuncsDiags
}

// ActionMetadatas returns a slice of ActionMetadata for the GetMetadata
// RPC.
func (s *Server) ActionMetadatas(ctx context.Context) ([]ActionMetadata, diag.Diagnostics) {
	actionFuncs, diags := s.ActionFuncs(ctx)

	actionMetadatas := make([]ActionMetadata, 0, len(actionFuncs))

	for typeName := range actionFuncs {
		actionMetadatas = append(actionMetadatas, ActionMetadata{
			TypeName: typeName,
		})
	}

	return actionMetadatas, diags
}

// ActionSchema returns the Action Schema for the given type name and
// caches the result for later Action operations.
func (s *Server) ActionSchema(ctx context.Context, typeName string) (fwschema.Schema, diag.Diagnostics) {
	s.actionSchemasMutex.RLock()
	actionSchema, ok := s.actionSchemas[typeName]
	s.actionSchemasMutex.RUnlock()

	if ok {
		return actionSchema, nil
	}

	var diags diag.Diagnostics

	resAction, actionDiags := s.Action(ctx, typeName)

	diags.Append(actionDiags...)

	if diags.HasError() {
		return nil, diags
	}

	schemaReq := action.SchemaRequest{}
	schemaResp := action.SchemaResponse{}

	logging.FrameworkTrace(ctx, "Calling provider defined Action Schema method", map[string]interface{}{logging.KeyActionType: typeName})
	resAction.Schema(ctx, schemaReq, &schemaResp)
	logging.FrameworkTrace(ctx, "Called provider defined Action Schema method", map[string]interface{}{logging.KeyActionType: typeName})

	diags.Append(schemaResp.Diagnostics...)

	if diags.HasError() {
		return schemaResp.Schema, diags
	}

	s.actionSchemasMutex.Lock()

	if s.actionSchemas == nil {
		s.actionSchemas = make(map[string]fwschema.Schema)
	}

	s.actionSchemas[typeName] = schemaResp.Schema

	s.actionSchemasMutex.Unlock()

	return schemaResp.Schema, diags
}

// ActionLinkedResources returns the Action LinkedResources for the given type name and
// caches the result for later Action operations.
func (s *Server) ActionLinkedResources(ctx context.Context, typeName string) (action.LinkedResources, diag.Diagnostics) {
	s.actionLinkedResourcesMutex.RLock()
	actionLinkedResources, ok := s.actionLinkedResources[typeName]
	s.actionLinkedResourcesMutex.RUnlock()

	if ok {
		return actionLinkedResources, nil
	}

	var diags diag.Diagnostics

	resAction, actionDiags := s.Action(ctx, typeName)

	diags.Append(actionDiags...)

	if diags.HasError() {
		return nil, diags
	}

	schemaReq := action.SchemaRequest{}
	schemaResp := action.SchemaResponse{}

	logging.FrameworkTrace(ctx, "Calling provider defined Action Schema method", map[string]interface{}{logging.KeyActionType: typeName})
	resAction.Schema(ctx, schemaReq, &schemaResp)
	logging.FrameworkTrace(ctx, "Called provider defined Action Schema method", map[string]interface{}{logging.KeyActionType: typeName})

	diags.Append(schemaResp.Diagnostics...)

	if diags.HasError() {
		return schemaResp.LinkedResources, diags
	}

	s.actionLinkedResourcesMutex.Lock()

	if s.actionLinkedResources == nil {
		s.actionLinkedResources = make(map[string]action.LinkedResources)
	}

	s.actionLinkedResources[typeName] = schemaResp.LinkedResources

	s.actionLinkedResourcesMutex.Unlock()

	return schemaResp.LinkedResources, diags
}

// ActionSchemas returns a map of Action Schemas for the
// GetProviderSchema RPC without caching since not all schemas are guaranteed to
// be necessary for later provider operations. The schema implementations are
// also validated.
func (s *Server) ActionSchemas(ctx context.Context) (map[string]fwschema.Schema, map[string]action.LinkedResources, diag.Diagnostics) {
	actionSchemas := make(map[string]fwschema.Schema)
	actionLinkedResources := make(map[string]action.LinkedResources)

	actionFuncs, diags := s.ActionFuncs(ctx)

	for typeName, actionFunc := range actionFuncs {
		resAction := actionFunc()

		schemaReq := action.SchemaRequest{}
		schemaResp := action.SchemaResponse{}

		logging.FrameworkTrace(ctx, "Calling provider defined Action Schema", map[string]interface{}{logging.KeyActionType: typeName})
		resAction.Schema(ctx, schemaReq, &schemaResp)
		logging.FrameworkTrace(ctx, "Called provider defined Action Schema", map[string]interface{}{logging.KeyActionType: typeName})

		diags.Append(schemaResp.Diagnostics...)

		if schemaResp.Diagnostics.HasError() {
			continue
		}

		validateDiags := schemaResp.Schema.ValidateImplementation(ctx)

		diags.Append(validateDiags...)

		if validateDiags.HasError() {
			continue
		}

		actionSchemas[typeName] = schemaResp.Schema
		actionLinkedResources[typeName] = schemaResp.LinkedResources
	}

	return actionSchemas, actionLinkedResources, diags
}
