package proto5server

import (
	"context"
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/go-cty/cty/msgpack"
	"github.com/hashicorp/terraform-plugin-framework/hcl2shim"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	sdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	terraformsdk "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// a resource type defined in SDKv2
var sdkResource sdk.Resource = sdk.Resource{
	Schema: map[string]*sdk.Schema{
		"id": &sdk.Schema{
			Type: sdk.TypeString,
		},
		"name": &sdk.Schema{
			Type: sdk.TypeString,
		},
	},
}

func diagnosticResult(format string, args ...any) tfprotov5.ListResourceResult {
	return tfprotov5.ListResourceResult{
		Diagnostics: []*tfprotov5.Diagnostic{
			{
				Summary: fmt.Sprintf(format, args...),
			},
		},
	}

}
func listFunc(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	sdkResource, ok := SDKResourceFromContext(ctx)
	if !ok {
		return
	}

	stream.Proto5Results = func(push func(tfprotov5.ListResourceResult) bool) {
		// From the resource type, we can obtain an initialized ResourceData value
		d := sdkResource.Data(&terraformsdk.InstanceState{ID: "#groot"})

		// The initialized ResourceData value is schema-aware
		if err := d.Set("name", "Groot"); err != nil {
			push(diagnosticResult("Error setting `name`: %v", err))
			return
		}

		if err := d.Set("nom", "groot"); err == nil {
			push(diagnosticResult("False negative outcome: `nom` is not a schema attribute"))
			return
		}

		displayName := "I am Groot"

		// Mimic SDK GRPCProviderServer.ReadResource ResourceData -> MsgPack
		state := d.State()
		if state == nil {
			push(diagnosticResult("Expected state to be non-nil"))
			return
		}

		schemaBlock := sdkResource.CoreConfigSchema()
		if schemaBlock == nil {
			push(diagnosticResult("Expected schemaBlock to be non-nil"))
			return
		}

		// We've copied hcl2shim wholesale for purposes of making the test pass
		newStateVal, err := hcl2shim.HCL2ValueFromFlatmap(state.Attributes, schemaBlock.ImpliedType())
		if err != nil {
			push(diagnosticResult("Error converting state attributes to HCL2 value: %v", err))
			return
		}

		// Think about this later
		// newStateVal = normalizeNullValues(newStateVal, stateVal, false)

		pack, err := msgpack.Marshal(newStateVal, schemaBlock.ImpliedType())
		if err != nil {
			push(diagnosticResult("Error marshaling new state value to MsgPack: %v", err))
			return
		}

		fmt.Printf("MsgPack: %s\n", pack)

		// Construct a tfprotov5.ListResourceResult
		listResult := tfprotov5.ListResourceResult{}
		listResult.Resource = &tfprotov5.DynamicValue{MsgPack: pack}
		listResult.DisplayName = displayName

		if !push(listResult) {
			return
		}
	}
}

func TestServerListResourceProto5ToProto5(t *testing.T) {
	t.Parallel()

	server := func(listResource func() list.ListResource) *Server {
		return &Server{
			FrameworkServer: fwserver.Server{
				Provider: &testprovider.Provider{
					ListResourcesMethod: func(ctx context.Context) []func() list.ListResource {
						return []func() list.ListResource{listResource}
					},
				},
			},
		}
	}

	listResource := func() list.ListResource {
		return &testprovider.ListResource{
			ListMethod: listFunc,
			MetadataMethod: func(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
				resp.TypeName = "test_resource"
			},
		}
	}
	aServer := server(listResource)

	ctx := context.Background()
	ctx = NewContextWithSDKResource(ctx, &sdkResource)
	req := &tfprotov5.ListResourceRequest{
		TypeName: "test_resource",
	}

	stream, err := aServer.ListResource(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error returned from ListResource: %v", err)
	}

	values := slices.Collect(stream.Results)
	if len(values) > 0 {
		if len(values[0].Diagnostics) > 0 {
			for _, diag := range values[0].Diagnostics {
				t.Logf("unexpected diagnostic returned from ListResource: %v", diag)
			}
			t.FailNow()
		}
	}

	if len(values) != 1 {
		t.Fatalf("expected 1 list result; got %d list results", len(values))
	}

	value := values[0]
	if value.DisplayName != "I am Groot" {
		t.Fatalf("I am not Groot")
	}
}
