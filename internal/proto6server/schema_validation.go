package proto6server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ValidateSchemaRequest repesents a request for validating a Schema.
type ValidateSchemaRequest struct {
	// Config contains the entire configuration of the data source, provider, or resource.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	Config tfsdk.Config
}

// ValidateSchemaResponse represents a response to a
// ValidateSchemaRequest.
type ValidateSchemaResponse struct {
	// Diagnostics report errors or warnings related to validating the schema.
	// An empty slice indicates success, with no warnings or errors generated.
	Diagnostics diag.Diagnostics
}

// SchemaValidate performs all Attribute and Block validation.
//
// TODO: Clean up this abstraction back into an internal Schema type method.
// The extra Schema parameter is a carry-over of creating the proto6server
// package from the tfsdk package and not wanting to export the method.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/215
func SchemaValidate(ctx context.Context, s tfsdk.Schema, req ValidateSchemaRequest, resp *ValidateSchemaResponse) {
	for name, attribute := range s.Attributes {

		attributeReq := tfsdk.ValidateAttributeRequest{
			AttributePath: tftypes.NewAttributePath().WithAttributeName(name),
			Config:        req.Config,
		}
		attributeResp := &tfsdk.ValidateAttributeResponse{
			Diagnostics: resp.Diagnostics,
		}

		AttributeValidate(ctx, attribute, attributeReq, attributeResp)

		resp.Diagnostics = attributeResp.Diagnostics
	}

	//nolint:staticcheck // Block support is required within the framework.
	for name, block := range s.Blocks {
		attributeReq := tfsdk.ValidateAttributeRequest{
			AttributePath: tftypes.NewAttributePath().WithAttributeName(name),
			Config:        req.Config,
		}
		attributeResp := &tfsdk.ValidateAttributeResponse{
			Diagnostics: resp.Diagnostics,
		}

		BlockValidate(ctx, block, attributeReq, attributeResp)

		resp.Diagnostics = attributeResp.Diagnostics
	}

	if s.DeprecationMessage != "" {
		resp.Diagnostics.AddWarning(
			"Deprecated",
			s.DeprecationMessage,
		)
	}
}
