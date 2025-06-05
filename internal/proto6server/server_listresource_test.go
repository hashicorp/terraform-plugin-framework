package proto6server

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
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerListResource(t *testing.T) {
	t.Parallel()

	type ThingResource struct {
		Name string `tfsdk:"name"`
	}

	type ThingResourceIdentity struct {
		Name string `tfsdk:"name"`
	}

	resources := map[string]ThingResource{}
	resourceIdentities := map[string]ThingResourceIdentity{}
	expectedResources := map[string]*tfprotov6.DynamicValue{}
	expectedResourceIdentities := map[string]*tfprotov6.ResourceIdentityData{}

	examples := []string{"bookbag", "bookshelf", "bookworm", "plateau", "platinum", "platypus"}
	for _, example := range examples {
		resources[example] = ThingResource{Name: example}
		resourceIdentities[example] = ThingResourceIdentity{Name: example}

		expectedResources[example] = testNewDynamicValue(t, tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"name": tftypes.String,
			},
		}, map[string]tftypes.Value{
			"name": tftypes.NewValue(tftypes.String, example),
		})

		expectedResourceIdentities[example] = &tfprotov6.ResourceIdentityData{
			IdentityData: testNewDynamicValue(t, tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, example),
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
				results := []list.ListResult{}
				var config listConfig
				diags := req.Config.Get(ctx, &config)
				if len(diags) > 0 {
					t.Fatalf("unexpected diagnostics: %s", diags)
				}

				for _, name := range []string{"plateau", "platinum", "platypus"} {
					if !strings.HasPrefix(name, config.Filter) {
						continue
					}

					result := list.ListResult{}
					identity, diags := req.ToIdentity(ctx, resourceIdentities[name])
					if diags.HasError() {
						result.Diagnostics = diags
						results = append(results, result)
						continue
					}

					resource, diags := req.ToResource(ctx, resources[name])
					if diags.HasError() {
						result.Diagnostics = diags
						results = append(results, result)
						continue
					}

					result.DisplayName = name
					result.Identity = identity
					result.Resource = resource // maybe only include if request says so
					result.Diagnostics = diags
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
						"name": identityschema.StringAttribute{},
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
							"name": resourceschema.StringAttribute{},
						},
					}
				},
			},
		}
	}

	happyServer := &Server{
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

	testCases := map[string]struct {
		server              *Server
		request             *tfprotov6.ListResourceRequest
		expectedError       error
		expectedDiagnostics diag.Diagnostics
		expectedResults     []tfprotov6.ListResourceResult
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
			request: &tfprotov6.ListResourceRequest{
				TypeName: "test_resource",
				Config:   plat,
			},
			expectedError:       &fwserver.ListResourceTypeNotFoundError{TypeName: "test_resource"},
			expectedDiagnostics: diag.Diagnostics{},
			expectedResults:     []tfprotov6.ListResourceResult{},
		},
		"result": {
			server: happyServer,
			request: &tfprotov6.ListResourceRequest{
				TypeName: "test_resource",
				Config:   plat,
			},
			expectedError:       nil,
			expectedDiagnostics: diag.Diagnostics{},
			expectedResults: []tfprotov6.ListResourceResult{
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
			request: &tfprotov6.ListResourceRequest{
				TypeName:        "test_resource",
				Config:          plateau,
				IncludeResource: true,
			},
			expectedError:       nil,
			expectedDiagnostics: diag.Diagnostics{},
			expectedResults: []tfprotov6.ListResourceResult{
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

			if diff := cmp.Diff(testCase.expectedResults, slices.Collect(got.Results), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("unexpected results difference: %s", diff)
			}
		})
	}
}
