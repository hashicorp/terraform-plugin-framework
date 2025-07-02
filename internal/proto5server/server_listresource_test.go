// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/go-cty/cty/msgpack"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/hcl2shim"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	sdk "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	terraformsdk "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestServerListResource(t *testing.T) {
	t.Parallel()

	type ThingResourceIdentity struct {
		Id string `tfsdk:"id"`
	}

	type ThingResource struct {
		ThingResourceIdentity
		Name string `tfsdk:"name"`
	}

	resources := map[string]ThingResource{}
	expectedResources := map[string]*tfprotov5.DynamicValue{}
	expectedResourceIdentities := map[string]*tfprotov5.ResourceIdentityData{}

	examples := []string{"bookbag", "bookshelf", "bookworm", "plateau", "platinum", "platypus"}
	for _, example := range examples {
		id := "id-" + example
		resources[example] = ThingResource{Name: example, ThingResourceIdentity: ThingResourceIdentity{Id: id}}

		expectedResources[example] = testNewDynamicValue(t, tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"id":   tftypes.String,
				"name": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"id":   tftypes.NewValue(tftypes.String, id),
			"name": tftypes.NewValue(tftypes.String, example),
		})

		expectedResourceIdentities[example] = &tfprotov5.ResourceIdentityData{
			IdentityData: testNewDynamicValue(t, tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"id": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, id),
			}),
		}
	}

	listResourceType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"filter": tftypes.String,
		},
	}

	type listConfig struct {
		Filter string `tfsdk:"filter"`
	}

	plat := testNewDynamicValue(t, listResourceType, map[string]tftypes.Value{
		"filter": tftypes.NewValue(tftypes.String, "plat"),
	})

	plateau := testNewDynamicValue(t, listResourceType, map[string]tftypes.Value{
		"filter": tftypes.NewValue(tftypes.String, "plateau"),
	})

	listResource := func() list.ListResource {
		return &testprovider.ListResource{
			ListResourceConfigSchemaMethod: func(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
				resp.Schema = listschema.Schema{
					Attributes: map[string]listschema.Attribute{
						"filter": listschema.StringAttribute{},
					},
				}
			},
			ListMethod: func(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
				var config listConfig
				diags := req.Config.Get(ctx, &config)
				if len(diags) > 0 {
					t.Fatalf("unexpected diagnostics: %s", diags)
				}

				results := []list.ListResult{}
				for name := range resources {
					if !strings.HasPrefix(name, config.Filter) {
						continue
					}

					result := req.NewListResult()
					result.DisplayName = name

					diags = result.Identity.Set(ctx, resources[name].ThingResourceIdentity)
					result.Diagnostics.Append(diags...)

					diags = result.Resource.Set(ctx, resources[name])
					result.Diagnostics.Append(diags...)

					results = append(results, result)
				}
				resp.Results = slices.Values(results)
			},
			MetadataMethod: func(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
				resp.TypeName = "test_resource"
			},
		}
	}

	managedResource := func() resource.Resource {
		return &testprovider.ResourceWithIdentity{
			IdentitySchemaMethod: func(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
				resp.IdentitySchema = identityschema.Schema{
					Attributes: map[string]identityschema.Attribute{
						"id": identityschema.StringAttribute{},
					},
				}
			},
			Resource: &testprovider.Resource{
				MetadataMethod: func(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
					resp.TypeName = "test_resource"
				},
				SchemaMethod: func(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
					resp.Schema = resourceschema.Schema{
						Attributes: map[string]resourceschema.Attribute{
							"id":   resourceschema.StringAttribute{},
							"name": resourceschema.StringAttribute{},
						},
					}
				},
			},
		}
	}

	server := func(listResource func() list.ListResource, managedResource func() resource.Resource) *Server {
		return &Server{
			FrameworkServer: fwserver.Server{
				Provider: &testprovider.Provider{
					ListResourcesMethod: func(ctx context.Context) []func() list.ListResource {
						return []func() list.ListResource{
							listResource,
						}
					},
					ResourcesMethod: func(ctx context.Context) []func() resource.Resource {
						return []func() resource.Resource{
							managedResource,
						}
					},
				},
			},
		}
	}

	happyServer := server(listResource, managedResource)

	testCases := map[string]struct {
		server              *Server
		request             *tfprotov5.ListResourceRequest
		expectedError       error
		expectedDiagnostics diag.Diagnostics
		expectedResults     []tfprotov5.ListResourceResult
	}{
		"error-on-unknown-list-resource-type": {
			server: &Server{
				FrameworkServer: fwserver.Server{
					Provider: &testprovider.Provider{
						ListResourcesMethod: func(ctx context.Context) []func() list.ListResource {
							return []func() list.ListResource{}
						},
					},
				},
			},
			request: &tfprotov5.ListResourceRequest{
				TypeName: "test_resource",
				Config:   plat,
			},
			expectedDiagnostics: diag.Diagnostics{},
			expectedResults: []tfprotov5.ListResourceResult{
				{
					Diagnostics: []*tfprotov5.Diagnostic{
						{
							Severity: tfprotov5.DiagnosticSeverityError,
							Summary:  "List Resource Type Not Found",
						},
					},
				},
			},
		},
		"result": {
			server: happyServer,
			request: &tfprotov5.ListResourceRequest{
				TypeName: "test_resource",
				Config:   plat,
			},
			expectedError:       nil,
			expectedDiagnostics: diag.Diagnostics{},
			expectedResults: []tfprotov5.ListResourceResult{
				{
					DisplayName: "plateau",
					Identity:    expectedResourceIdentities["plateau"],
				},
				{
					DisplayName: "platinum",
					Identity:    expectedResourceIdentities["platinum"],
				},
				{
					DisplayName: "platypus",
					Identity:    expectedResourceIdentities["platypus"],
				},
			},
		},
		"result-with-include-resource": {
			server: happyServer,
			request: &tfprotov5.ListResourceRequest{
				TypeName:        "test_resource",
				Config:          plateau,
				IncludeResource: true,
			},
			expectedError:       nil,
			expectedDiagnostics: diag.Diagnostics{},
			expectedResults: []tfprotov5.ListResourceResult{
				{
					DisplayName: "plateau",
					Identity:    expectedResourceIdentities["plateau"],
					Resource:    expectedResources["plateau"],
				},
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			metadataResponse := &fwserver.GetMetadataResponse{}
			testCase.server.FrameworkServer.GetMetadata(context.Background(), &fwserver.GetMetadataRequest{}, metadataResponse)

			if diff := cmp.Diff(metadataResponse.Diagnostics, diag.Diagnostics{}); diff != "" {
				t.Fatalf("unexpected metadata diagnostics difference: got %s\nwanted %s", metadataResponse.Diagnostics, diag.Diagnostics{})
			}

			got, err := testCase.server.ListResource(context.Background(), testCase.request)

			if diff := cmp.Diff(testCase.expectedError, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("unexpected error difference: %s", diff)
			}

			sortResults := cmpopts.SortSlices(func(a, b tfprotov5.ListResourceResult) bool {
				return a.DisplayName < b.DisplayName
			})
			opts := []cmp.Option{
				sortResults,
				cmpopts.EquateEmpty(),
				cmpopts.IgnoreFields(tfprotov5.Diagnostic{}, "Detail"),
			}
			if diff := cmp.Diff(testCase.expectedResults, slices.Collect(got.Results), opts...); diff != "" {
				t.Errorf("unexpected results difference: %s", diff)
			}
		})
	}
}

type SDKContext string

var SDKResource SDKContext = "sdk_resource"

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

func listFunc(ctx context.Context, req list.ListRequest, stream *list.ListResultsStream) {
	panic("hats")
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
	ctx = context.WithValue(ctx, SDKResource, sdkResource)
	req := &tfprotov5.ListResourceRequest{}

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

	// 2: from the resource type, we can obtain an initialized ResourceData value
	d := sdkResource.Data(&terraformsdk.InstanceState{ID: "#groot"})

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
