// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource_test

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type ComputeInstance struct {
}

type ComputeInstanceWithValidateConfig struct {
	ComputeInstance
}

type ComputeInstanceWithConfigValidators struct {
	ComputeInstance
}

func (c *ComputeInstance) Configure(_ context.Context, _ resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	panic("not implemented")
}

func (c *ComputeInstance) ListSchema(_ context.Context, _ resource.SchemaRequest, _ resource.SchemaResponse) {
	panic("not implemented")
}

func (c *ComputeInstance) ListResources(_ context.Context, _ resource.ListRequest, _ resource.ListResponse) {
	panic("not implemented")
}

func (c *ComputeInstance) Metadata(_ context.Context, _ resource.MetadataRequest, _ *resource.MetadataResponse) {
	panic("not implemented")
}

func (c *ComputeInstance) Schema(_ context.Context, _ resource.SchemaRequest, _ *resource.SchemaResponse) {
	panic("not implemented")
}

func (c *ComputeInstance) Create(_ context.Context, _ resource.CreateRequest, _ *resource.CreateResponse) {
	panic("not implemented")
}

func (c *ComputeInstance) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
	panic("not implemented")
}

func (c *ComputeInstance) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	panic("not implemented")
}

func (c *ComputeInstance) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	panic("not implemented")
}

func (c *ComputeInstanceWithValidateConfig) ValidateListConfig(_ context.Context, _ resource.ValidateListConfigRequest, _ *resource.ValidateListConfigResponse) {
	panic("not implemented")
}

func (c *ComputeInstanceWithConfigValidators) ListConfigValidators(_ context.Context) []resource.ListConfigValidator {
	panic("not implemented")
}

// ExampleResource_listable demonstrates a resource.Resource that implements
// resource.List interfaces.
func ExampleResource_listable() {
	var _ resource.List = &ComputeInstance{}
	var _ resource.ListWithConfigure = &ComputeInstance{}
	var _ resource.ListWithValidateConfig = &ComputeInstanceWithValidateConfig{}
	var _ resource.ListWithConfigValidators = &ComputeInstanceWithConfigValidators{}

	var _ resource.Resource = &ComputeInstance{}
	var _ resource.ResourceWithConfigure = &ComputeInstance{}

	// Output:
}
