// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package list_test

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type ComputeInstance struct {
}

type ComputeInstanceWithValidateListResourceConfig struct {
	ComputeInstance
}

type ComputeInstanceWithListResourceConfigValidators struct {
	ComputeInstance
}

func (c *ComputeInstance) Configure(_ context.Context, _ resource.ConfigureRequest, _ *resource.ConfigureResponse) {
}

func (c *ComputeInstance) ListResourceConfigSchema(_ context.Context, _ resource.SchemaRequest, _ resource.SchemaResponse) {
}

func (c *ComputeInstance) ListResource(_ context.Context, _ list.ListResourceRequest, _ list.ListResourceResponse) {
}

func (c *ComputeInstance) Metadata(_ context.Context, _ resource.MetadataRequest, _ *resource.MetadataResponse) {
}

func (c *ComputeInstance) Schema(_ context.Context, _ resource.SchemaRequest, _ *resource.SchemaResponse) {
}

func (c *ComputeInstance) Create(_ context.Context, _ resource.CreateRequest, _ *resource.CreateResponse) {
}

func (c *ComputeInstance) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
}

func (c *ComputeInstance) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
}

func (c *ComputeInstance) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (c *ComputeInstanceWithValidateListResourceConfig) ValidateListResourceConfig(_ context.Context, _ list.ValidateConfigRequest, _ *list.ValidateConfigResponse) {
}

func (c *ComputeInstanceWithListResourceConfigValidators) ListResourceConfigValidators(_ context.Context) []list.ConfigValidator {
	return nil
}

// ExampleResource_listable demonstrates a resource.Resource that implements
// list.ListResource interfaces.
func ExampleResource_listable() {
	var _ list.ListResource = &ComputeInstance{}
	var _ list.ListResourceWithConfigure = &ComputeInstance{}
	var _ list.ListResourceWithValidateConfig = &ComputeInstanceWithValidateListResourceConfig{}
	var _ list.ListResourceWithConfigValidators = &ComputeInstanceWithListResourceConfigValidators{}

	var _ resource.Resource = &ComputeInstance{}
	var _ resource.ResourceWithConfigure = &ComputeInstance{}

	// Output:
}
