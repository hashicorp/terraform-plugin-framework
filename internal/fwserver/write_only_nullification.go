// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func NullifyWriteOnlyAttributes(ctx context.Context, resourceSchema fwschema.Schema) func(*tftypes.AttributePath, tftypes.Value) (tftypes.Value, error) {
	return func(path *tftypes.AttributePath, val tftypes.Value) (tftypes.Value, error) {
		ctx = logging.FrameworkWithAttributePath(ctx, path.String())

		// we are only modifying attributes, not the entire resource
		if len(path.Steps()) < 1 {
			return val, nil
		}

		attribute, err := resourceSchema.AttributeAtTerraformPath(ctx, path)

		if err != nil {
			if errors.Is(err, fwschema.ErrPathInsideAtomicAttribute) {
				// ignore attributes/elements inside schema.Attributes, they have no schema of their own
				logging.FrameworkTrace(ctx, "attribute is a non-schema attribute, not nullifying")
				return val, nil
			}

			if errors.Is(err, fwschema.ErrPathIsBlock) {
				// ignore blocks, they do not have a writeOnly field
				logging.FrameworkTrace(ctx, "attribute is a block, not nullifying")
				return val, nil
			}

			if errors.Is(err, fwschema.ErrPathInsideDynamicAttribute) {
				// ignore attributes/elements inside schema.DynamicAttribute, they have no schema of their own
				logging.FrameworkTrace(ctx, "attribute is inside of a dynamic attribute, not nullifying")
				return val, nil
			}

			logging.FrameworkError(ctx, "couldn't find attribute in resource schema")

			return tftypes.Value{}, fmt.Errorf("couldn't find attribute in resource schema: %w", err)
		}

		// Value type from new state to create null with
		newValueType := val.Type()

		// If the attribute is dynamic set the new value type to DynamicPseudoType
		// instead of the underlying concrete type
		// TODO: verify if this is the correct behavior once Terraform Core implementation is complete
		_, isDynamic := attribute.GetType().(basetypes.DynamicTypable)
		if isDynamic {
			newValueType = tftypes.DynamicPseudoType
		}

		if attribute.IsWriteOnly() && !val.IsNull() {
			logging.FrameworkDebug(ctx, "Nullifying write-only attribute in the newState")

			return tftypes.NewValue(newValueType, nil), nil
		}

		return val, nil
	}
}