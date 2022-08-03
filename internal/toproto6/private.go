package toproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
)

// PrivateData returns []byte from privateState.Data.
func PrivateData(ctx context.Context, input *privatestate.Data) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	output, err := input.Bytes(ctx)
	if err != nil {
		diags.AddError(
			"Unable to Convert Private Data",
			"An unexpected error was encountered when converting the private data. "+
				"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
				"Please report this to the provider developer:\n\n"+
				"Private data convert error.",
		)

		return nil, diags
	}

	return output, nil
}
