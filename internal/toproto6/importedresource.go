package toproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// ImportedResource returns the *tfprotov6.ImportedResource equivalent of a
// *fwserver.ImportedResource.
func ImportedResource(ctx context.Context, fw *fwserver.ImportedResource) (*tfprotov6.ImportedResource, diag.Diagnostics) {
	if fw == nil {
		return nil, nil
	}

	proto6 := &tfprotov6.ImportedResource{
		Private:  fw.Private,
		TypeName: fw.TypeName,
	}

	state, diags := State(ctx, &fw.State)

	proto6.State = state

	return proto6, diags
}
