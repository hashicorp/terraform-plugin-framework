// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package proto5server

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerListResource(t *testing.T) {
	t.Parallel()

	type ThingResourceIdentity struct {
		Id string `tfsdk:"id"`
	}

	type ThingResource struct {
		// TODO: how do we feel about this?
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

					result := req.ToListResult(ctx, resources[name].ThingResourceIdentity, resources[name], name)
					results = append(results, result)
				}
				resp.Results = slices.Values(results)
			},
			MetadataMethod: func(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
				resp.TypeName = "test_resource"
			},
		}
	}

	listResourceThatDoesNotPopulateResource := func() list.ListResource {
		r, ok := listResource().(*testprovider.ListResource)
		if !ok {
			t.Fatal("listResourceThatDoesNotPopulateResource must be a testprovider.ListResource")
		}

		r.ListMethod = func(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
			result := req.ToListResult(ctx, resources["plateau"].ThingResourceIdentity, nil, "plateau")

			resp.Results = slices.Values([]list.ListResult{result})
		}

		return r
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
		"result-with-include-resource-warning": {
			server: server(listResourceThatDoesNotPopulateResource, managedResource),
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
					Diagnostics: []*tfprotov5.Diagnostic{
						{
							Severity: tfprotov5.DiagnosticSeverityWarning,
							Summary:  "Incomplete List Result",
							Detail:   "The provider did not populate the Resource field in the ListResourceResult. This may be due to the provider not supporting this functionality or an error in the provider's implementation.",
						},
					},
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
