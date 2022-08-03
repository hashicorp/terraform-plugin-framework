package fromproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
)

// PrivateData returns the privatestate.Data for []byte.
func PrivateData(ctx context.Context, input []byte) (*privatestate.Data, diag.Diagnostics) {
	output, diags := privatestate.NewData(ctx, input)
	if diags.HasError() {
		return nil, diags
	}

	return output, nil
}
