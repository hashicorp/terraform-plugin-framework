// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto6server

import (
	"context"
	"fmt"

	actionschema "github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
)

// LinkedResourceSchemas returns the linked resource schemas for the given Action Schema. Linked resource schemas
// are either retrieved from the provider server or converted from the action schema definition.
func (s *Server) LinkedResourceSchemas(ctx context.Context, actionSchema actionschema.SchemaType) ([]fwschema.Schema, []fwschema.Schema, diag.Diagnostics) {
	allDiags := make(diag.Diagnostics, 0)
	lrSchemas := make([]fwschema.Schema, 0)
	lrIdentitySchemas := make([]fwschema.Schema, 0)

	for _, lrType := range actionSchema.LinkedResourceTypes() {
		switch lrType := lrType.(type) {
		case actionschema.RawV5LinkedResource:
			allDiags.AddError(
				"Invalid Linked Resource Schema",
				fmt.Sprintf("An unexpected error was encountered when converting %[1]q linked resource schema from the protocol type. "+
					"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
					"Please report this to the provider developer:\n\n"+
					"The %[1]q linked resource is a protocol v5 resource but the provider is being served using protocol v6.", lrType.GetTypeName()),
			)

			return nil, nil, allDiags
		case actionschema.RawV6LinkedResource:
			// Raw linked resources are not stored on this provider server, so we retrieve the schemas from the
			// action definition directly and convert them to framework schemas.
			lrSchema, err := fromproto6.ResourceSchema(ctx, lrType.GetSchema())
			if err != nil {
				allDiags.AddError(
					"Invalid Linked Resource Schema",
					fmt.Sprintf("An unexpected error was encountered when converting %q linked resource schema from the protocol type. "+
						"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n%s", lrType.GetTypeName(), err.Error()),
				)

				return nil, nil, allDiags
			}
			lrSchemas = append(lrSchemas, lrSchema)

			lrIdentitySchema, err := fromproto6.IdentitySchema(ctx, lrType.GetIdentitySchema())
			if err != nil {
				allDiags.AddError(
					"Invalid Linked Resource Schema",
					fmt.Sprintf("An unexpected error was encountered when converting %q linked resource identity schema from the protocol type. "+
						"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n%s", lrType.GetTypeName(), err.Error()),
				)

				return nil, nil, allDiags
			}
			lrIdentitySchemas = append(lrIdentitySchemas, lrIdentitySchema)
		default:
			// Any other linked resource type should be stored on the same provider server as the action,
			// so we can just retrieve it via the type name.
			lrSchema, diags := s.FrameworkServer.ResourceSchema(ctx, lrType.GetTypeName())
			if diags.HasError() {
				allDiags.AddError(
					"Invalid Linked Resource Schema",
					fmt.Sprintf("An unexpected error was encountered when converting %[1]q linked resource data from the protocol type. "+
						"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"The %[1]q linked resource was not found on the provider server.", lrType.GetTypeName()),
				)

				return nil, nil, allDiags
			}
			lrSchemas = append(lrSchemas, lrSchema)

			lrIdentitySchema, diags := s.FrameworkServer.ResourceIdentitySchema(ctx, lrType.GetTypeName())
			allDiags.Append(diags...)
			if allDiags.HasError() {
				// If the resource is found, the identity schema will only return a diagnostic if the provider implementation
				// returns an error from (resource.Resource).IdentitySchema method.
				return nil, nil, allDiags
			}
			lrIdentitySchemas = append(lrIdentitySchemas, lrIdentitySchema)
		}
	}

	return lrSchemas, lrIdentitySchemas, allDiags
}
