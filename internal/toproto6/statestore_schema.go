// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package toproto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	stateschema "github.com/hashicorp/terraform-plugin-framework/statestore/schema"
)

// StateStoreSchema returns the *tfprotov6.StateStoreSchema equivalent of a StateStoreSchema.
func StateStoreSchema(ctx context.Context, s stateschema.Schema) (*tfprotov6.StateStoreSchema, error) {
	configSchema, err := Schema(ctx, s)
	if err != nil {
		return nil, err
	}

	result := &tfprotov6.StateStoreSchema{
		Schema: configSchema,
	}

	return result, nil
}
