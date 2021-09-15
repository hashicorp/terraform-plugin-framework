package tfsdk

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// importedResource represents a resource that was imported.
//
// This type is not exported as the framework import implementation is
// currently designed for the most common use case of single resource import.
type importedResource struct {
	Private  []byte
	State    State
	TypeName string
}

func (r importedResource) toTfprotov6(ctx context.Context) (*tfprotov6.ImportedResource, diag.Diagnostics) {
	var diags diag.Diagnostics
	irProto6 := &tfprotov6.ImportedResource{
		Private:  r.Private,
		TypeName: r.TypeName,
	}

	stateProto6, err := tfprotov6.NewDynamicValue(r.State.Schema.TerraformType(ctx), r.State.Raw)

	if err != nil {
		diags.AddError(
			"Error converting imported resource response",
			"An unexpected error was encountered when converting the imported resource response to a usable type. This is always a problem with the provider. Please give the following information to the provider developer:\n\n"+err.Error(),
		)
		return nil, diags
	}

	irProto6.State = &stateProto6

	return irProto6, diags
}

// importResourceStateResponse is a thin abstraction to allow native Diagnostics usage
type importResourceStateResponse struct {
	Diagnostics       diag.Diagnostics
	ImportedResources []importedResource
}

func (r importResourceStateResponse) toTfprotov6(ctx context.Context) *tfprotov6.ImportResourceStateResponse {
	resp := &tfprotov6.ImportResourceStateResponse{
		Diagnostics: r.Diagnostics.ToTfprotov6Diagnostics(),
	}

	for _, ir := range r.ImportedResources {
		irProto6, diags := ir.toTfprotov6(ctx)
		resp.Diagnostics = append(resp.Diagnostics, diags.ToTfprotov6Diagnostics()...)
		if diags.HasError() {
			continue
		}
		resp.ImportedResources = append(resp.ImportedResources, irProto6)
	}

	return resp
}

func (s *server) importResourceState(ctx context.Context, req *tfprotov6.ImportResourceStateRequest, resp *importResourceStateResponse) {
	resourceType, diags := s.getResourceType(ctx, req.TypeName)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resourceSchema, diags := resourceType.GetSchema(ctx)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	resource, diags := resourceType.NewResource(ctx, s.p)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	emptyState := tftypes.NewValue(resourceSchema.TerraformType(ctx), nil)
	importReq := ImportResourceStateRequest{
		ID: req.ID,
	}
	importResp := ImportResourceStateResponse{
		State: State{
			Raw:    emptyState,
			Schema: resourceSchema,
		},
	}

	resource.ImportState(ctx, importReq, &importResp)
	resp.Diagnostics.Append(importResp.Diagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	if importResp.State.Raw.Equal(emptyState) {
		resp.Diagnostics.AddError(
			"Missing Resource Import State",
			"An unexpected error was encountered when importing the resource. This is always a problem with the provider. Please give the following information to the provider developer:\n\n"+
				"Resource ImportState method returned no State in response. If import is intentionally not supported, call the ResourceImportStateNotImplemented() function or return an error.",
		)
		return
	}

	resp.ImportedResources = []importedResource{
		{
			State:    importResp.State,
			TypeName: req.TypeName,
		},
	}
}

// ImportResourceState satisfies the tfprotov6.ProviderServer interface.
func (s *server) ImportResourceState(ctx context.Context, req *tfprotov6.ImportResourceStateRequest) (*tfprotov6.ImportResourceStateResponse, error) {
	ctx = s.registerContext(ctx)
	resp := &importResourceStateResponse{}

	s.importResourceState(ctx, req, resp)

	return resp.toTfprotov6(ctx), nil
}
