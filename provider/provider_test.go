package provider_test

import (
	"context"
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/metadata"
	"testing"
)

// TODO: write test for a more complicated type like list or map or set
// TODO: write test for an object type

func TestName(t *testing.T) {
	// TDT - table driven tests, define a bunch of test cases and then run them all

	// Arrange, Act, Assert
	// Arrange = table + test cases, mock/spy setup
	// Act = the actual logic that takes a test case and runs it
	// Assert = thing that says we got what we expected, defined in the test cases
	t.Parallel()

	testCases := map[string]struct {
		// DONE: define data input + output types for test cases here
		input    provider.Provider
		expected metadata.ProviderMetadata
	}{ // We want to write 1 resource schema with 2 attributes: string and a boolean
		"bool-string-test": {
			input: &testprovider.Provider{
				ResourcesMethod: func(_ context.Context) []func() resource.Resource {
					return []func() resource.Resource{
						func() resource.Resource {
							return &testprovider.Resource{
								SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
									resp.Schema = resourceschema.Schema{
										Attributes: map[string]resourceschema.Attribute{
											"string_attribute": resourceschema.StringAttribute{
												Description: "This is a string test string",
												Required:    true,
											},
											"bool_attribute": resourceschema.BoolAttribute{
												Description: "This is a bool test string",
												Required:    true,
											},
										},
									}
								},
								MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
									resp.TypeName = "bool_string_attribute"
								},
							}
						},
					}
				},
			},
			expected: metadata.ProviderMetadata{
				DataSourceSchemas: nil,
				Functions:         nil,
				Provider:          nil,
				ProviderServer:    nil,
				ResourceSchemas: map[string]metadata.SchemaBlock{
					"bool_string_attribute": metadata.SchemaBlock{
						Block: &metadata.Block{
							Attributes: map[string]metadata.Attribute{
								"string_attribute": {
									Computed:        metadata.BoolPointer(false),
									Deprecated:      metadata.BoolPointer(false),
									Description:     metadata.StringPointer("This is a string test string"),
									DescriptionKind: metadata.DescriptionPointer(metadata.Plain),
									Optional:        metadata.BoolPointer(false),
									Required:        metadata.BoolPointer(true),
									Sensitive:       metadata.BoolPointer(false),
									Type:            metadata.AnyPointer(json.RawMessage([]byte(`"string"`))),
								},
								"bool_attribute": {
									Computed:        metadata.BoolPointer(false),
									Deprecated:      metadata.BoolPointer(false),
									Description:     metadata.StringPointer("This is a bool test string"),
									DescriptionKind: metadata.DescriptionPointer(metadata.Plain),
									Optional:        metadata.BoolPointer(false),
									Required:        metadata.BoolPointer(true),
									Sensitive:       metadata.BoolPointer(false),
									Type:            metadata.AnyPointer(json.RawMessage([]byte(`"bool"`))),
								},
							},
						},
					},
				},
				Version: "",
			},
		},
		"list-test": {
			input: &testprovider.Provider{
				ResourcesMethod: func(_ context.Context) []func() resource.Resource {
					return []func() resource.Resource{
						func() resource.Resource {
							return &testprovider.Resource{
								SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
									resp.Schema = resourceschema.Schema{
										Attributes: map[string]resourceschema.Attribute{
											"list_attribute": resourceschema.ListAttribute{
												Description: "This is a list test string",
												ElementType: types.StringType,
												Required:    true,
											},
										},
									}
								},
								MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
									resp.TypeName = "list-type-attribute"
								},
							}
						},
					}
				},
			},
			expected: metadata.ProviderMetadata{
				DataSourceSchemas: nil,
				Functions:         nil,
				Provider:          nil,
				ProviderServer:    nil,
				ResourceSchemas: map[string]metadata.SchemaBlock{
					"list-type-attribute": metadata.SchemaBlock{
						Block: &metadata.Block{
							Attributes: map[string]metadata.Attribute{
								"list_attribute": {
									Computed:        metadata.BoolPointer(false),
									Deprecated:      metadata.BoolPointer(false),
									Description:     metadata.StringPointer("This is a list test string"),
									DescriptionKind: metadata.DescriptionPointer(metadata.Plain),
									Optional:        metadata.BoolPointer(false),
									Required:        metadata.BoolPointer(true),
									Sensitive:       metadata.BoolPointer(false),
									Type:            metadata.AnyPointer(json.RawMessage([]byte(`["list","string"]`))),
								},
							},
						},
					},
				},
				Version: "",
			},
		},
		"nested-single-test": {
			input: &testprovider.Provider{
				ResourcesMethod: func(_ context.Context) []func() resource.Resource {
					return []func() resource.Resource{
						func() resource.Resource {
							return &testprovider.Resource{
								SchemaMethod: func(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
									resp.Schema = resourceschema.Schema{
										Attributes: map[string]resourceschema.Attribute{
											"nested_attribute": resourceschema.SingleNestedAttribute{
												Required:    true,
												Description: "This is a nested test string",
												Attributes: map[string]resourceschema.Attribute{
													"nested_nested_attribute": resourceschema.BoolAttribute{
														Description: "This is a bool test string",
														Required:    true,
													},
												},
											},
										},
									}
								},
								MetadataMethod: func(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
									resp.TypeName = "nested-type-attribute"
								},
							}
						},
					}
				},
			},
			expected: metadata.ProviderMetadata{
				DataSourceSchemas: nil,
				Functions:         nil,
				Provider:          nil,
				ProviderServer:    nil,
				ResourceSchemas: map[string]metadata.SchemaBlock{
					"nested-type-attribute": metadata.SchemaBlock{
						Block: &metadata.Block{
							Attributes: map[string]metadata.Attribute{
								"nested_attribute": {
									Computed:        metadata.BoolPointer(false),
									Deprecated:      metadata.BoolPointer(false),
									Description:     metadata.StringPointer("This is a nested test string"),
									DescriptionKind: metadata.DescriptionPointer(metadata.Plain),
									NestedType: metadata.AnyPointer(metadata.NestedAttributeType{
										Attributes: map[string]metadata.Attribute{
											"nested_nested_attribute": metadata.Attribute{
												Computed:        metadata.BoolPointer(false),
												Deprecated:      metadata.BoolPointer(false),
												Description:     metadata.StringPointer("This is a bool test string"),
												DescriptionKind: metadata.DescriptionPointer(metadata.Plain),
												Optional:        metadata.BoolPointer(false),
												Required:        metadata.BoolPointer(true),
												Sensitive:       metadata.BoolPointer(false),
												Type:            metadata.AnyPointer(json.RawMessage([]byte(`"bool"`))),
											},
										},
										NestingMode: metadata.AnyPointer(metadata.PurpleSingle),
									}),
									Optional:  metadata.BoolPointer(false),
									Required:  metadata.BoolPointer(true),
									Sensitive: metadata.BoolPointer(false),
									Type:      metadata.AnyPointer(json.RawMessage([]byte(`["object",{"nested_nested_attribute":"bool"}]`))),
								},
							},
						},
					},
				},
				Version: "",
			},
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual := provider.BuildMetadata(context.Background(), testCase.input)

			if diff := cmp.Diff(actual, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
