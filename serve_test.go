package tfsdk

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestServerCancelInFlightContexts(t *testing.T) {
	t.Parallel()

	// let's test and make sure the code we use to Stop will actually
	// cancel in flight contexts how we expect and not, y'know, crash or
	// something

	// first, let's create a bunch of goroutines
	wg := new(sync.WaitGroup)
	s := &server{}
	testCtx := context.Background()
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			ctx = s.registerContext(ctx)
			select {
			case <-time.After(time.Second * 10):
				t.Error("timed out waiting to be canceled")
				return
			case <-ctx.Done():
				return
			}
		}()
	}
	// avoid any race conditions around canceling the contexts before
	// they're all set up
	//
	// we don't need this in prod as, presumably, Terraform would not keep
	// sending us requests after it told us to stop
	time.Sleep(200 * time.Millisecond)

	s.cancelRegisteredContexts(testCtx)

	wg.Wait()
	// if we got here, that means that either all our contexts have been
	// canceled, or we have an error reported
}

func TestMarkComputedNilsAsUnknown(t *testing.T) {
	t.Parallel()

	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			// values should be left alone
			"string-value": {
				Type:     types.StringType,
				Required: true,
			},
			// nil, uncomputed values should be left alone
			"string-nil": {
				Type:     types.StringType,
				Optional: true,
			},
			// nil computed values should be turned into unknown
			"string-nil-computed": {
				Type:     types.StringType,
				Computed: true,
			},
			// nil computed values should be turned into unknown
			"string-nil-optional-computed": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			// non-nil computed values should be left alone
			"string-value-optional-computed": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			// nil objects should be unknown
			"object-nil-optional-computed": {
				Type: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"string-nil": types.StringType,
						"string-set": types.StringType,
					},
				},
				Optional: true,
				Computed: true,
			},
			// non-nil objects should be left alone
			"object-value-optional-computed": {
				Type: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						// nil attributes of objects
						// should be let alone, as they
						// don't have a schema of their
						// own
						"string-nil": types.StringType,
						"string-set": types.StringType,
					},
				},
				Optional: true,
				Computed: true,
			},
			// nil nested attributes should be unknown
			"nested-nil-optional-computed": {
				Attributes: schema.SingleNestedAttributes(map[string]schema.Attribute{
					"string-nil": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
					"string-set": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
				}),
				Optional: true,
				Computed: true,
			},
			// non-nil nested attributes should be left alone on the top level
			"nested-value-optional-computed": {
				Attributes: schema.SingleNestedAttributes(map[string]schema.Attribute{
					// nested computed attributes should be unknown
					"string-nil": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
					// nested non-nil computed attributes should be left alone
					"string-set": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
				}),
				Optional: true,
				Computed: true,
			},
		},
	}
	input := tftypes.NewValue(s.TerraformType(context.Background()), map[string]tftypes.Value{
		"string-value":                   tftypes.NewValue(tftypes.String, "hello, world"),
		"string-nil":                     tftypes.NewValue(tftypes.String, nil),
		"string-nil-computed":            tftypes.NewValue(tftypes.String, nil),
		"string-nil-optional-computed":   tftypes.NewValue(tftypes.String, nil),
		"string-value-optional-computed": tftypes.NewValue(tftypes.String, "hello, world"),
		"object-nil-optional-computed":   tftypes.NewValue(s.Attributes["object-nil-optional-computed"].Type.TerraformType(context.Background()), nil),
		"object-value-optional-computed": tftypes.NewValue(s.Attributes["object-value-optional-computed"].Type.TerraformType(context.Background()), map[string]tftypes.Value{
			"string-nil": tftypes.NewValue(tftypes.String, nil),
			"string-set": tftypes.NewValue(tftypes.String, "foo"),
		}),
		"nested-nil-optional-computed": tftypes.NewValue(s.Attributes["nested-nil-optional-computed"].Attributes.AttributeType().TerraformType(context.Background()), nil),
		"nested-value-optional-computed": tftypes.NewValue(s.Attributes["nested-value-optional-computed"].Attributes.AttributeType().TerraformType(context.Background()), map[string]tftypes.Value{
			"string-nil": tftypes.NewValue(tftypes.String, nil),
			"string-set": tftypes.NewValue(tftypes.String, "bar"),
		}),
	})
	expected := tftypes.NewValue(s.TerraformType(context.Background()), map[string]tftypes.Value{
		"string-value":                   tftypes.NewValue(tftypes.String, "hello, world"),
		"string-nil":                     tftypes.NewValue(tftypes.String, nil),
		"string-nil-computed":            tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"string-nil-optional-computed":   tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		"string-value-optional-computed": tftypes.NewValue(tftypes.String, "hello, world"),
		"object-nil-optional-computed":   tftypes.NewValue(s.Attributes["object-nil-optional-computed"].Type.TerraformType(context.Background()), tftypes.UnknownValue),
		"object-value-optional-computed": tftypes.NewValue(s.Attributes["object-value-optional-computed"].Type.TerraformType(context.Background()), map[string]tftypes.Value{
			"string-nil": tftypes.NewValue(tftypes.String, nil),
			"string-set": tftypes.NewValue(tftypes.String, "foo"),
		}),
		"nested-nil-optional-computed": tftypes.NewValue(s.Attributes["nested-nil-optional-computed"].Attributes.AttributeType().TerraformType(context.Background()), tftypes.UnknownValue),
		"nested-value-optional-computed": tftypes.NewValue(s.Attributes["nested-value-optional-computed"].Attributes.AttributeType().TerraformType(context.Background()), map[string]tftypes.Value{
			"string-nil": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			"string-set": tftypes.NewValue(tftypes.String, "bar"),
		}),
	})

	got, err := tftypes.Transform(input, markComputedNilsAsUnknown(context.Background(), s))
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	diff, err := expected.Diff(got)
	if err != nil {
		t.Errorf("Error diffing values: %s", err)
		return
	}
	if len(diff) > 0 {
		t.Errorf("Unexpected diff (value1 expected, value2 got): %v", diff)
	}
}

func TestServerGetProviderSchema(t *testing.T) {
	t.Parallel()

	s := new(testServeProvider)
	testServer := &server{
		p: s,
	}
	got, err := testServer.GetProviderSchema(context.Background(), new(tfprotov6.GetProviderSchemaRequest))
	if err != nil {
		t.Errorf("Got unexpected error: %s", err)
		return
	}
	expected := &tfprotov6.GetProviderSchemaResponse{
		Provider: testServeProviderProviderSchema,
		ResourceSchemas: map[string]*tfprotov6.Schema{
			"test_one": testServeResourceTypeOneSchema,
			"test_two": testServeResourceTypeTwoSchema,
		},
		DataSourceSchemas: map[string]*tfprotov6.Schema{
			"test_one": testServeDataSourceOneSchema,
			"test_two": testServeDataSourceTwoSchema,
		},
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
	}
}

func TestServerGetProviderSchemaWithProviderMeta(t *testing.T) {
	t.Parallel()

	s := new(testServeProviderWithMetaSchema)
	testServer := &server{
		p: s,
	}
	got, err := testServer.GetProviderSchema(context.Background(), new(tfprotov6.GetProviderSchemaRequest))
	if err != nil {
		t.Errorf("Got unexpected error: %s", err)
		return
	}
	expected := &tfprotov6.GetProviderSchemaResponse{
		Provider: testServeProviderProviderSchema,
		ResourceSchemas: map[string]*tfprotov6.Schema{
			"test_one": testServeResourceTypeOneSchema,
			"test_two": testServeResourceTypeTwoSchema,
		},
		DataSourceSchemas: map[string]*tfprotov6.Schema{
			"test_one": testServeDataSourceOneSchema,
			"test_two": testServeDataSourceTwoSchema,
		},
		ProviderMeta: &tfprotov6.Schema{
			Version: 2,
			Block: &tfprotov6.SchemaBlock{
				Attributes: []*tfprotov6.SchemaAttribute{
					{
						Name:            "foo",
						Required:        true,
						Type:            tftypes.String,
						Description:     "A **string**",
						DescriptionKind: tfprotov6.StringKindMarkdown,
					},
				},
			},
		},
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
	}
}

func TestServerConfigureProvider(t *testing.T) {
	t.Parallel()

	type testCase struct {
		tfVersion     string
		config        tftypes.Value
		expectedDiags []*tfprotov6.Diagnostic
	}

	tests := map[string]testCase{
		"basic": {
			tfVersion: "1.0.0",
			config: tftypes.NewValue(testServeProviderProviderType, map[string]tftypes.Value{
				"required":          tftypes.NewValue(tftypes.String, "this is a required value"),
				"optional":          tftypes.NewValue(tftypes.String, nil),
				"computed":          tftypes.NewValue(tftypes.String, nil),
				"optional_computed": tftypes.NewValue(tftypes.String, "they filled this one out"),
				"sensitive":         tftypes.NewValue(tftypes.String, "hunter42"),
				"deprecated":        tftypes.NewValue(tftypes.String, "oops"),
				"string":            tftypes.NewValue(tftypes.String, "a new string value"),
				"number":            tftypes.NewValue(tftypes.Number, 1234),
				"bool":              tftypes.NewValue(tftypes.Bool, true),
				"list-string": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, "world"),
				}),
				"list-list-string": tftypes.NewValue(tftypes.List{ElementType: tftypes.List{ElementType: tftypes.String}}, []tftypes.Value{
					tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "blue"),
						tftypes.NewValue(tftypes.String, "green"),
					}),
					tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "rojo"),
						tftypes.NewValue(tftypes.String, "azul"),
						tftypes.NewValue(tftypes.String, "verde"),
					}),
				}),
				"list-object": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Bool,
					"baz": tftypes.Number,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"foo": tftypes.String,
						"bar": tftypes.Bool,
						"baz": tftypes.Number,
					}}, map[string]tftypes.Value{
						"foo": tftypes.NewValue(tftypes.String, "hello, world"),
						"bar": tftypes.NewValue(tftypes.Bool, true),
						"baz": tftypes.NewValue(tftypes.Number, 4567),
					}),
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"foo": tftypes.String,
						"bar": tftypes.Bool,
						"baz": tftypes.Number,
					}}, map[string]tftypes.Value{
						"foo": tftypes.NewValue(tftypes.String, "goodnight, moon"),
						"bar": tftypes.NewValue(tftypes.Bool, false),
						"baz": tftypes.NewValue(tftypes.Number, 8675309),
					}),
				}),
				"object": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo":  tftypes.String,
					"bar":  tftypes.Bool,
					"baz":  tftypes.Number,
					"quux": tftypes.List{ElementType: tftypes.String},
				}}, map[string]tftypes.Value{
					"foo": tftypes.NewValue(tftypes.String, "testing123"),
					"bar": tftypes.NewValue(tftypes.Bool, true),
					"baz": tftypes.NewValue(tftypes.Number, 123),
					"quux": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "blue"),
						tftypes.NewValue(tftypes.String, "green"),
					}),
				}),
				"empty-object": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{}),
				"single-nested-attributes": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Number,
				}}, map[string]tftypes.Value{
					"foo": tftypes.NewValue(tftypes.String, "almost done"),
					"bar": tftypes.NewValue(tftypes.Number, 12),
				}),
				"list-nested-attributes": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Number,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"foo": tftypes.String,
						"bar": tftypes.Number,
					}}, map[string]tftypes.Value{
						"foo": tftypes.NewValue(tftypes.String, "let's do the math"),
						"bar": tftypes.NewValue(tftypes.Number, 18973),
					}),
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"foo": tftypes.String,
						"bar": tftypes.Number,
					}}, map[string]tftypes.Value{
						"foo": tftypes.NewValue(tftypes.String, "this is why we can't have nice things"),
						"bar": tftypes.NewValue(tftypes.Number, 14554216),
					}),
				}),
			}),
		},
		"config-unknown-value": {
			tfVersion: "1.0.0",
			config: tftypes.NewValue(testServeProviderProviderType, map[string]tftypes.Value{
				"required":          tftypes.NewValue(tftypes.String, "this is a required value"),
				"optional":          tftypes.NewValue(tftypes.String, nil),
				"computed":          tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"optional_computed": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"sensitive":         tftypes.NewValue(tftypes.String, "hunter42"),
				"deprecated":        tftypes.NewValue(tftypes.String, "oops"),
				"string":            tftypes.NewValue(tftypes.String, "a new string value"),
				"number":            tftypes.NewValue(tftypes.Number, 1234),
				"bool":              tftypes.NewValue(tftypes.Bool, true),
				"list-string": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "hello"),
					tftypes.NewValue(tftypes.String, "world"),
				}),
				"list-list-string": tftypes.NewValue(tftypes.List{ElementType: tftypes.List{ElementType: tftypes.String}}, tftypes.UnknownValue),
				"list-object": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Bool,
					"baz": tftypes.Number,
				}}}, tftypes.UnknownValue),
				"object": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo":  tftypes.String,
					"bar":  tftypes.Bool,
					"baz":  tftypes.Number,
					"quux": tftypes.List{ElementType: tftypes.String},
				}}, map[string]tftypes.Value{
					"foo":  tftypes.NewValue(tftypes.String, "testing123"),
					"bar":  tftypes.NewValue(tftypes.Bool, true),
					"baz":  tftypes.NewValue(tftypes.Number, 123),
					"quux": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, tftypes.UnknownValue),
				}),
				"empty-object": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{}}, map[string]tftypes.Value{}),
				"single-nested-attributes": tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Number,
				}}, map[string]tftypes.Value{
					"foo": tftypes.NewValue(tftypes.String, "almost done"),
					"bar": tftypes.NewValue(tftypes.Number, 12),
				}),
				"list-nested-attributes": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"foo": tftypes.String,
					"bar": tftypes.Number,
				}}}, tftypes.UnknownValue),
			}),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := new(testServeProvider)
			testServer := &server{
				p: s,
			}
			dv, err := tfprotov6.NewDynamicValue(testServeProviderProviderType, tc.config)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}

			providerSchema, diags := s.GetSchema(context.Background())
			if len(diags) > 0 {
				t.Errorf("Unexpected diags: %+v", diags)
				return
			}
			got, err := testServer.ConfigureProvider(context.Background(), &tfprotov6.ConfigureProviderRequest{
				TerraformVersion: tc.tfVersion,
				Config:           &dv,
			})
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if s.configuredTFVersion != tc.tfVersion {
				t.Errorf("Expected Terraform version to be %q, got %q", tc.tfVersion, s.configuredTFVersion)
			}
			if diff := cmp.Diff(got.Diagnostics, tc.expectedDiags); diff != "" {
				t.Errorf("Unexpected diff in diagnostics (+wanted, -got): %s", diff)
			}
			if diff := cmp.Diff(s.configuredVal, tc.config); diff != "" {
				t.Errorf("Unexpected diff in config (+wanted, -got): %s", diff)
				return
			}
			if diff := cmp.Diff(s.configuredSchema, providerSchema); diff != "" {
				t.Errorf("Unexpected diff in schema (+wanted, -got): %s", diff)
				return
			}
		})
	}
}

func TestServerReadResource(t *testing.T) {
	t.Parallel()

	type testCase struct {
		// request input
		currentState tftypes.Value
		providerMeta tftypes.Value
		private      []byte
		resource     string
		resourceType tftypes.Type

		impl func(context.Context, ReadResourceRequest, *ReadResourceResponse)

		// response expectations
		expectedNewState tftypes.Value
		expectedDiags    []*tfprotov6.Diagnostic
		expectedPrivate  []byte
	}

	tests := map[string]testCase{
		"one_basic": {
			currentState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name":              tftypes.NewValue(tftypes.String, "foo"),
				"favorite_colors":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				"created_timestamp": tftypes.NewValue(tftypes.String, "a minute ago, but like, as a timestamp"),
			}),
			resource:     "test_one",
			resourceType: testServeResourceTypeOneType,

			impl: func(_ context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "foo"),
					"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "orange"),
						tftypes.NewValue(tftypes.String, "yellow"),
					}),
					"created_timestamp": tftypes.NewValue(tftypes.String, "now"),
				})
			},

			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "foo"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "now"),
			}),
		},
		"one_provider_meta": {
			currentState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "my name"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "a long, long time ago"),
			}),
			resource:     "test_one",
			resourceType: testServeResourceTypeOneType,

			providerMeta: tftypes.NewValue(testServeProviderMetaType, map[string]tftypes.Value{
				"foo": tftypes.NewValue(tftypes.String, "my provider_meta value"),
			}),

			impl: func(_ context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "my name"),
					"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
						tftypes.NewValue(tftypes.String, "red"),
						tftypes.NewValue(tftypes.String, "blue"),
					}),
					"created_timestamp": tftypes.NewValue(tftypes.String, "a long, long time ago"),
				})
			},

			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "my name"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "blue"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "a long, long time ago"),
			}),
		},
		"one_remove": {
			currentState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "my name"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "a long, long time ago"),
			}),
			resource:     "test_one",
			resourceType: testServeResourceTypeOneType,

			impl: func(_ context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeOneType, nil)
			},

			expectedNewState: tftypes.NewValue(testServeResourceTypeOneType, nil),
		},
		"two_basic": {
			currentState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123foo"),
				"disks": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					},
				}, tftypes.UnknownValue),
			}),
			resource:     "test_two",
			resourceType: testServeResourceTypeTwoType,

			impl: func(_ context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
					"id": tftypes.NewValue(tftypes.String, "123foo"),
					"disks": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"name":    tftypes.String,
								"size_gb": tftypes.Number,
								"boot":    tftypes.Bool,
							},
						},
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"name":    tftypes.String,
								"size_gb": tftypes.Number,
								"boot":    tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"name":    tftypes.NewValue(tftypes.String, "my-disk"),
							"size_gb": tftypes.NewValue(tftypes.Number, 100),
							"boot":    tftypes.NewValue(tftypes.Bool, true),
						}),
					}),
				})
			},

			expectedNewState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123foo"),
				"disks": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 100),
						"boot":    tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
			}),
		},
		"two_diags": {
			currentState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123foo"),
				"disks": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					},
				}, tftypes.UnknownValue),
			}),
			resource:     "test_two",
			resourceType: testServeResourceTypeTwoType,

			impl: func(_ context.Context, req ReadResourceRequest, resp *ReadResourceResponse) {
				resp.State.Raw = tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
					"id": tftypes.NewValue(tftypes.String, "123foo"),
					"disks": tftypes.NewValue(tftypes.List{
						ElementType: tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"name":    tftypes.String,
								"size_gb": tftypes.Number,
								"boot":    tftypes.Bool,
							},
						},
					}, []tftypes.Value{
						tftypes.NewValue(tftypes.Object{
							AttributeTypes: map[string]tftypes.Type{
								"name":    tftypes.String,
								"size_gb": tftypes.Number,
								"boot":    tftypes.Bool,
							},
						}, map[string]tftypes.Value{
							"name":    tftypes.NewValue(tftypes.String, "my-disk"),
							"size_gb": tftypes.NewValue(tftypes.Number, 100),
							"boot":    tftypes.NewValue(tftypes.Bool, true),
						}),
					}),
				})
				resp.Diagnostics = []*tfprotov6.Diagnostic{
					{
						Summary:   "This is a warning",
						Severity:  tfprotov6.DiagnosticSeverityWarning,
						Detail:    "This is your final warning",
						Attribute: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(0),
					},
					{
						Summary:  "This is an error",
						Severity: tfprotov6.DiagnosticSeverityError,
						Detail:   "Oops.",
					},
				}
			},

			expectedNewState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123foo"),
				"disks": tftypes.NewValue(tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					},
				}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"name":    tftypes.String,
							"size_gb": tftypes.Number,
							"boot":    tftypes.Bool,
						},
					}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 100),
						"boot":    tftypes.NewValue(tftypes.Bool, true),
					}),
				}),
			}),

			expectedDiags: []*tfprotov6.Diagnostic{
				{
					Summary:   "This is a warning",
					Severity:  tfprotov6.DiagnosticSeverityWarning,
					Detail:    "This is your final warning",
					Attribute: tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(0),
				},
				{
					Summary:  "This is an error",
					Severity: tfprotov6.DiagnosticSeverityError,
					Detail:   "Oops.",
				},
			},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := &testServeProvider{
				readResourceImpl: tc.impl,
			}
			testServer := &server{
				p: s,
			}
			var pmSchema schema.Schema
			if tc.providerMeta.Type() != nil {
				sWithMeta := &testServeProviderWithMetaSchema{s}
				testServer.p = sWithMeta
				schema, diags := sWithMeta.GetMetaSchema(context.Background())
				if len(diags) > 0 {
					t.Errorf("Unexpected diags: %+v", diags)
					return
				}
				pmSchema = schema
			}

			rt, diags := testServer.getResourceType(context.Background(), tc.resource)
			if len(diags) > 0 {
				t.Errorf("Unexpected diags: %+v", diags)
				return
			}
			schema, diags := rt.GetSchema(context.Background())
			if len(diags) > 0 {
				t.Errorf("Unexpected diags: %+v", diags)
				return
			}

			dv, err := tfprotov6.NewDynamicValue(tc.resourceType, tc.currentState)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			req := &tfprotov6.ReadResourceRequest{
				TypeName:     tc.resource,
				Private:      tc.private,
				CurrentState: &dv,
			}
			if tc.providerMeta.Type() != nil {
				providerMeta, err := tfprotov6.NewDynamicValue(testServeProviderMetaType, tc.providerMeta)
				if err != nil {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				req.ProviderMeta = &providerMeta
			}
			got, err := testServer.ReadResource(context.Background(), req)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if s.readResourceCalledResourceType != tc.resource {
				t.Errorf("Called wrong resource. Expected to call %q, actually called %q", tc.resource, s.readResourceCalledResourceType)
				return
			}
			if diff := cmp.Diff(got.Diagnostics, tc.expectedDiags); diff != "" {
				t.Errorf("Unexpected diff in diagnostics (+wanted, -got): %s", diff)
			}
			if diff := cmp.Diff(s.readResourceCurrentStateValue, tc.currentState); diff != "" {
				t.Errorf("Unexpected diff in current state (+wanted, -got): %s", diff)
				return
			}
			if diff := cmp.Diff(s.readResourceCurrentStateSchema, schema); diff != "" {
				t.Errorf("Unexpected diff in state schema (+wanted, -got): %s", diff)
				return
			}
			if tc.providerMeta.Type() != nil {
				if diff := cmp.Diff(s.readResourceProviderMetaValue, tc.providerMeta); diff != "" {
					t.Errorf("Unexpected diff in provider meta (+wanted, -got): %s", diff)
					return
				}
				if diff := cmp.Diff(s.readResourceProviderMetaSchema, pmSchema); diff != "" {
					t.Errorf("Unexpected diff in provider meta schema (+wanted, -got): %s", diff)
					return
				}
			}
			gotNewState, err := got.NewState.Unmarshal(tc.resourceType)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(gotNewState, tc.expectedNewState); diff != "" {
				t.Errorf("Unexpected diff in new state (+wanted, -got): %s", diff)
				return
			}
			if string(got.Private) != string(tc.expectedPrivate) {
				t.Errorf("Expected private to be %q, got %q", tc.expectedPrivate, got.Private)
				return
			}
		})
	}
}

func TestServerPlanResourceChange(t *testing.T) {
	t.Parallel()

	type testCase struct {
		// request input
		priorState       tftypes.Value
		proposedNewState tftypes.Value
		config           tftypes.Value
		priorPrivate     []byte
		providerMeta     tftypes.Value
		resource         string
		resourceType     tftypes.Type

		// response expectations
		expectedPlannedState    tftypes.Value
		expectedRequiresReplace []*tftypes.AttributePath
		expectedPlannedPrivate  []byte
		expectedDiags           []*tfprotov6.Diagnostic
	}

	tests := map[string]testCase{
		"one_basic": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "when the earth was young"),
			}),
			proposedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "when the earth was young"),
			}),
			config: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
					tftypes.NewValue(tftypes.String, "yellow"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_one",
			resourceType: testServeResourceTypeOneType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
					tftypes.NewValue(tftypes.String, "red"),
					tftypes.NewValue(tftypes.String, "orange"),
				}),
				"created_timestamp": tftypes.NewValue(tftypes.String, "when the earth was young"),
			}),
		},
		"two_delete": {
			priorState: tftypes.NewValue(testServeResourceTypeTwoType, map[string]tftypes.Value{
				"id": tftypes.NewValue(tftypes.String, "123456"),
				"disks": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
					"name":    tftypes.String,
					"size_gb": tftypes.Number,
					"boot":    tftypes.Bool,
				}}}, []tftypes.Value{
					tftypes.NewValue(tftypes.Object{AttributeTypes: map[string]tftypes.Type{
						"name":    tftypes.String,
						"size_gb": tftypes.Number,
						"boot":    tftypes.Bool,
					}}, map[string]tftypes.Value{
						"name":    tftypes.NewValue(tftypes.String, "my-disk"),
						"size_gb": tftypes.NewValue(tftypes.Number, 10),
						"boot":    tftypes.NewValue(tftypes.Bool, false),
					}),
				}),
			}),
			proposedNewState:     tftypes.NewValue(testServeResourceTypeTwoType, nil),
			config:               tftypes.NewValue(testServeResourceTypeTwoType, nil),
			resource:             "test_two",
			resourceType:         testServeResourceTypeTwoType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeTwoType, nil),
		},
		"one_add": {
			priorState: tftypes.NewValue(testServeResourceTypeOneType, nil),
			proposedNewState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name":              tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				"created_timestamp": tftypes.NewValue(tftypes.String, nil),
			}),
			config: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name":              tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				"created_timestamp": tftypes.NewValue(tftypes.String, nil),
			}),
			resource:     "test_one",
			resourceType: testServeResourceTypeOneType,
			expectedPlannedState: tftypes.NewValue(testServeResourceTypeOneType, map[string]tftypes.Value{
				"name":              tftypes.NewValue(tftypes.String, "hello, world"),
				"favorite_colors":   tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				"created_timestamp": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			}),
		},
	}

	for name, tc := range tests {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			s := &testServeProvider{}
			testServer := &server{
				p: s,
			}

			priorStateDV, err := tfprotov6.NewDynamicValue(tc.resourceType, tc.priorState)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			proposedStateDV, err := tfprotov6.NewDynamicValue(tc.resourceType, tc.proposedNewState)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			configDV, err := tfprotov6.NewDynamicValue(tc.resourceType, tc.config)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			req := &tfprotov6.PlanResourceChangeRequest{
				TypeName:         tc.resource,
				PriorPrivate:     tc.priorPrivate,
				PriorState:       &priorStateDV,
				ProposedNewState: &proposedStateDV,
				Config:           &configDV,
			}
			if tc.providerMeta.Type() != nil {
				providerMeta, err := tfprotov6.NewDynamicValue(testServeProviderMetaType, tc.providerMeta)
				if err != nil {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				req.ProviderMeta = &providerMeta
			}
			got, err := testServer.PlanResourceChange(context.Background(), req)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(got.Diagnostics, tc.expectedDiags); diff != "" {
				t.Errorf("Unexpected diff in diagnostics (+wanted, -got): %s", diff)
			}
			gotPlannedState, err := got.PlannedState.Unmarshal(tc.resourceType)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if diff := cmp.Diff(gotPlannedState, tc.expectedPlannedState); diff != "" {
				t.Errorf("Unexpected diff in planned state (+wanted, -got): %s", diff)
				return
			}
			if string(got.PlannedPrivate) != string(tc.expectedPlannedPrivate) {
				t.Errorf("Expected planned private to be %q, got %q", tc.expectedPlannedPrivate, got.PlannedPrivate)
				return
			}
			if diff := cmp.Diff(got.RequiresReplace, tc.expectedRequiresReplace); diff != "" {
				t.Errorf("Unexpected diff in requires replace (+wanted, -got): %s", diff)
				return
			}
		})
	}
}

func TestServerApplyResourceChange(t *testing.T) {
	t.Parallel()

	// TODO: test applying
}

func TestServerReadDataSource(t *testing.T) {
	t.Parallel()

	// TODO: test reading data source
}
