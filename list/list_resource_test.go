// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package list_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-cty/cty/msgpack"
	"github.com/hashicorp/terraform-plugin-framework/hcl2shim"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	sdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tsdk "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

func TestListResultToResourceData(t *testing.T) {
	t.Parallel()

	// 1: we have a resource type defined in SDKv2
	sdkResource := sdk.Resource{
		Schema: map[string]*sdk.Schema{
			"id": &sdk.Schema{
				Type: sdk.TypeString,
			},
			"name": &sdk.Schema{
				Type: sdk.TypeString,
			},
		},
	}

	// 2: from the resource type, we can obtain an initialized ResourceData value
	d := sdkResource.Data(&tsdk.InstanceState{ID: "#groot"})

	// 3: the initialized ResourceData value is schema-aware
	if err := d.Set("name", "Groot"); err != nil {
		t.Fatalf("Error setting `name`: %v", err)
	}

	if err := d.Set("nom", "groot"); err == nil {
		t.Fatal("False negative outcome: `nom` is not a schema attribute")
	}

	displayName := "I am Groot"

	// 4: mimic SDK GRPCProviderServer.ReadResource ResourceData -> MsgPack
	state := d.State()
	if state == nil {
		t.Fatal("Expected state to be non-nil")
	}

	schemaBlock := sdkResource.CoreConfigSchema()
	if schemaBlock == nil {
		t.Fatal("Expected schemaBlock to be non-nil")
	}

	// Copied hcl2shim wholesale for purposes of making the test pass
	newStateVal, err := hcl2shim.HCL2ValueFromFlatmap(state.Attributes, schemaBlock.ImpliedType())
	if err != nil {
		t.Fatalf("Error converting state attributes to HCL2 value: %v", err)
	}

	// newStateVal = normalizeNullValues(newStateVal, stateVal, false)

	pack, err := msgpack.Marshal(newStateVal, schemaBlock.ImpliedType())
	if err != nil {
		t.Fatalf("Error marshaling new state value to MsgPack: %v", err)
	}

	fmt.Printf("MsgPack: %s\n", pack)

	// 5: construct a tfprotov5.ListResourceResult
	listResult := tfprotov5.ListResourceResult{}
	listResult.Resource = &tfprotov5.DynamicValue{MsgPack: pack}
	listResult.DisplayName = displayName
}
