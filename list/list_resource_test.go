// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package list_test

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type ComputeInstanceResource struct {
	NoOpListResource
	NoOpResource
}

type ComputeInstanceWithValidateListResourceConfig struct {
	ComputeInstanceResource
}

type ComputeInstanceWithListResourceConfigValidators struct {
	ComputeInstanceResource
}

func (c *ComputeInstanceResource) Configure(_ context.Context, _ resource.ConfigureRequest, _ *resource.ConfigureResponse) {
}

func (c *ComputeInstanceResource) Metadata(_ context.Context, _ resource.MetadataRequest, _ *resource.MetadataResponse) {
}

func (c *ComputeInstanceWithValidateListResourceConfig) ValidateListResourceConfig(_ context.Context, _ list.ValidateConfigRequest, _ *list.ValidateConfigResponse) {
}

func (c *ComputeInstanceWithListResourceConfigValidators) ListResourceConfigValidators(_ context.Context) []list.ConfigValidator {
	return nil
}

// ExampleResource_listable demonstrates a resource.Resource that implements
// list.ListResource interfaces.
func ExampleResource_listable() {
	var _ list.ListResource = &ComputeInstanceResource{}
	var _ list.ListResourceWithConfigure = &ComputeInstanceResource{}
	var _ list.ListResourceWithValidateConfig = &ComputeInstanceWithValidateListResourceConfig{}
	var _ list.ListResourceWithConfigValidators = &ComputeInstanceWithListResourceConfigValidators{}

	var _ resource.Resource = &ComputeInstanceResource{}
	var _ resource.ResourceWithConfigure = &ComputeInstanceResource{}

	// Output:
}
