package fromproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
)

// PrivateData returns the privatestate.Data for []byte.
func PrivateData(ctx context.Context, input []byte) (privatestate.Data, diag.Diagnostics) {
	var diags diag.Diagnostics

	output, err := privatestate.NewData(ctx, input)
	if err != nil {
		diags.AddError(
			"Unable to Create Private Data",
			"An unexpected error was encountered when creating the private data. "+
				"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
				"Please report this to the provider developer:\n\n"+
				"Private data create error.",
		)

		return privatestate.Data{}, diags
	}

	return output, nil
}
