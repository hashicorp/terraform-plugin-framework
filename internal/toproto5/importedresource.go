package toproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
)

// ImportedResource returns the *tfprotov5.ImportedResource equivalent of a
// *fwserver.ImportedResource.
func ImportedResource(ctx context.Context, fw *fwserver.ImportedResource) (*tfprotov5.ImportedResource, diag.Diagnostics) {
	if fw == nil {
		return nil, nil
	}

	proto5 := &tfprotov5.ImportedResource{
		Private:  fw.Private,
		TypeName: fw.TypeName,
	}

	state, diags := State(ctx, &fw.State)

	proto5.State = state

	return proto5, diags
}
