// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package toproto6

import (
	"context"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// IdentitySchema returns the *tfprotov6.ResourceIdentitySchema equivalent of a Schema.
func IdentitySchema(ctx context.Context, s fwschema.Schema) (*tfprotov6.ResourceIdentitySchema, error) {
	if s == nil {
		return nil, nil
	}

	result := &tfprotov6.ResourceIdentitySchema{
		Version: s.GetVersion(),
	}

	attrs := make([]*tfprotov6.ResourceIdentitySchemaAttribute, 0)

	for name, attr := range s.GetAttributes() {
		a, err := IdentitySchemaAttribute(ctx, name, tftypes.NewAttributePath().WithAttributeName(name), attr)

		if err != nil {
			return nil, err
		}

		attrs = append(attrs, a)
	}

	slices.SortFunc(attrs, func(a, b *tfprotov6.ResourceIdentitySchemaAttribute) int {
		if a == nil {
			if b == nil {
				return 0
			}

			return -1
		}

		if b == nil {
			return 1
		}

		return strings.Compare(a.Name, b.Name)
	})

	result.IdentityAttributes = attrs

	return result, nil
}
