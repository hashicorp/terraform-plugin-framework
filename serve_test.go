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

type testServeProvider struct{}

func (t *testServeProvider) GetSchema(_ context.Context) (schema.Schema, []*tfprotov6.Diagnostic) {
	return schema.Schema{
		Version:            1,
		DeprecationMessage: "Deprecated in favor of other_resource",
		Attributes: map[string]schema.Attribute{
			"required": {
				Type:     types.StringType,
				Required: true,
			},
			"optional": {
				Type:     types.StringType,
				Optional: true,
			},
			"computed": {
				Type:     types.StringType,
				Computed: true,
			},
			"optional_computed": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"sensitive": {
				Type:      types.StringType,
				Optional:  true,
				Sensitive: true,
			},
			"deprecated": {
				Type:               types.StringType,
				Optional:           true,
				DeprecationMessage: "Deprecated, please use \"optional\" instead",
			},
			"string": {
				Type:     types.StringType,
				Optional: true,
			},
			"number": {
				Type:     types.NumberType,
				Optional: true,
			},
			"bool": {
				Type:     types.BoolType,
				Optional: true,
			},
			"list-string": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional: true,
			},
			"list-list-string": {
				Type: types.ListType{
					ElemType: types.ListType{
						ElemType: types.StringType,
					},
				},
				Optional: true,
			},
			"list-object": {
				Type: types.ListType{
					ElemType: types.ObjectType{
						AttrTypes: map[string]attr.Type{
							"foo": types.StringType,
							"bar": types.BoolType,
							"baz": types.NumberType,
						},
					},
				},
				Optional: true,
			},
			"object": {
				Type: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"foo": types.StringType,
						"bar": types.BoolType,
						"baz": types.NumberType,
						"quux": types.ListType{
							ElemType: types.StringType,
						},
					},
				},
				Optional: true,
			},
			"empty-object": {
				Type:     types.ObjectType{},
				Optional: true,
			},
			// TODO: add maps when we support them
			// TODO: add sets when we support them
			// TODO: add tuples when we support them
			"single-nested-attributes": {
				Attributes: schema.SingleNestedAttributes(map[string]schema.Attribute{
					"foo": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
					"bar": {
						Type:     types.NumberType,
						Required: true,
					},
				}),
				Optional: true,
			},
			"list-nested-attributes": {
				Attributes: schema.ListNestedAttributes(map[string]schema.Attribute{
					"foo": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
					},
					"bar": {
						Type:     types.NumberType,
						Required: true,
					},
				}, schema.ListNestedAttributesOptions{}),
				Optional: true,
			},
		},
	}, nil
}

var testServeProviderProviderSchema = &tfprotov6.Schema{
	Version: 1,
	Block: &tfprotov6.SchemaBlock{
		Deprecated: true,
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "bool",
				Type:     tftypes.Bool,
				Optional: true,
			},
			{
				Name:     "computed",
				Type:     tftypes.String,
				Computed: true,
			},
			{
				Name:       "deprecated",
				Type:       tftypes.String,
				Optional:   true,
				Deprecated: true,
			},
			{
				Name: "empty-object",
				Type: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{},
				},
				Optional: true,
			},
			{
				Name: "list-list-string",
				Type: tftypes.List{
					ElementType: tftypes.List{
						ElementType: tftypes.String,
					},
				},
				Optional: true,
			},
			{
				Name: "list-nested-attributes",
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeList,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "bar",
							Type:     tftypes.Number,
							Required: true,
						},
						{
							Name:     "foo",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
					},
				},
				Optional: true,
			},
			{
				Name: "list-object",
				Type: tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"foo": tftypes.String,
							"bar": tftypes.Bool,
							"baz": tftypes.Number,
						},
					},
				},
				Optional: true,
			},
			{
				Name: "list-string",
				Type: tftypes.List{
					ElementType: tftypes.String,
				},
				Optional: true,
			},
			{
				Name:     "number",
				Type:     tftypes.Number,
				Optional: true,
			},
			{
				Name: "object",
				Type: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"foo": tftypes.String,
						"bar": tftypes.Bool,
						"baz": tftypes.Number,
						"quux": tftypes.List{
							ElementType: tftypes.String,
						},
					},
				},
				Optional: true,
			},
			{
				Name:     "optional",
				Type:     tftypes.String,
				Optional: true,
			},
			{
				Name:     "optional_computed",
				Type:     tftypes.String,
				Optional: true,
				Computed: true,
			},
			{
				Name:     "required",
				Type:     tftypes.String,
				Required: true,
			},
			{
				Name:      "sensitive",
				Type:      tftypes.String,
				Optional:  true,
				Sensitive: true,
			},
			{
				Name: "single-nested-attributes",
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeSingle,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "bar",
							Type:     tftypes.Number,
							Required: true,
						},
						{
							Name:     "foo",
							Type:     tftypes.String,
							Optional: true,
							Computed: true,
						},
					},
				},
				Optional: true,
			},
			{
				Name:     "string",
				Type:     tftypes.String,
				Optional: true,
			},
			// TODO: add maps when we support them
			// TODO: add sets when we support them
			// TODO: add tuples when we support them
		},
	},
}

type testServeProviderWithMetaSchema struct {
	*testServeProvider
}

func (t *testServeProviderWithMetaSchema) GetMetaSchema(context.Context) (schema.Schema, []*tfprotov6.Diagnostic) {
	return schema.Schema{
		Version: 2,
		Attributes: map[string]schema.Attribute{
			"foo": {
				Type:                types.StringType,
				Required:            true,
				Description:         "A string",
				MarkdownDescription: "A **string**",
			},
		},
	}, nil
}

type testServeResourceOne struct{}

func (rt testServeResourceOne) GetSchema(_ context.Context) (schema.Schema, []*tfprotov6.Diagnostic) {
	return schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"name": schema.Attribute{
				Required: true,
				Type:     types.StringType,
			},
			"favorite_colors": schema.Attribute{
				Optional: true,
				Type:     types.ListType{ElemType: types.StringType},
			},
			"created_timestamp": schema.Attribute{
				Computed: true,
				Type:     types.StringType,
			},
		},
	}, nil
}

func (rt testServeResourceOne) NewResource(_ Provider) (Resource, []*tfprotov6.Diagnostic) {
	panic("not implemented") // TODO: Implement
}

var testServeResourceOneSchema = &tfprotov6.Schema{
	Version: 1,
	Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "created_timestamp",
				Computed: true,
				Type:     tftypes.String,
			},
			{
				Name:     "favorite_colors",
				Optional: true,
				Type:     tftypes.List{ElementType: tftypes.String},
			},
			{
				Name:     "name",
				Required: true,
				Type:     tftypes.String,
			},
		},
	},
}

type testServeResourceTwo struct{}

func (rt testServeResourceTwo) GetSchema(_ context.Context) (schema.Schema, []*tfprotov6.Diagnostic) {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Attribute{
				Optional: true,
				Computed: true,
				Type:     types.StringType,
			},
			"disks": schema.Attribute{
				Optional: true,
				Computed: true,
				Attributes: schema.ListNestedAttributes(map[string]schema.Attribute{
					"name": {
						Required: true,
						Type:     types.StringType,
					},
					"size_gb": {
						Required: true,
						Type:     types.NumberType,
					},
					"boot": {
						Required: true,
						Type:     types.BoolType,
					},
				}, schema.ListNestedAttributesOptions{}),
			},
		},
	}, nil
}

func (rt testServeResourceTwo) NewResource(_ Provider) (Resource, []*tfprotov6.Diagnostic) {
	panic("not implemented") // TODO: Implement
}

var testServeResourceTwoSchema = &tfprotov6.Schema{
	Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "disks",
				Optional: true,
				Computed: true,
				NestedType: &tfprotov6.SchemaObject{
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:     "boot",
							Required: true,
							Type:     tftypes.Bool,
						},
						{
							Name:     "name",
							Required: true,
							Type:     tftypes.String,
						},
						{
							Name:     "size_gb",
							Required: true,
							Type:     tftypes.Number,
						},
					},
					Nesting: tfprotov6.SchemaObjectNestingModeList,
				},
			},
			{
				Name:     "id",
				Optional: true,
				Computed: true,
				Type:     tftypes.String,
			},
		},
	},
}

type testServeDataSourceOne struct{}

func (dt testServeDataSourceOne) GetSchema(_ context.Context) (schema.Schema, []*tfprotov6.Diagnostic) {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"current_time": {
				Type:     types.StringType,
				Computed: true,
			},
			"current_date": {
				Type:     types.StringType,
				Computed: true,
			},
			"is_dst": {
				Type:     types.BoolType,
				Computed: true,
			},
		},
	}, nil
}

func (dt testServeDataSourceOne) NewDataSource(_ Provider) (DataSource, []*tfprotov6.Diagnostic) {
	panic("not implemented") // TODO: Implement
}

var testServeDataSourceOneSchema = &tfprotov6.Schema{
	Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "current_date",
				Computed: true,
				Type:     tftypes.String,
			},
			{
				Name:     "current_time",
				Computed: true,
				Type:     tftypes.String,
			},
			{
				Name:     "is_dst",
				Computed: true,
				Type:     tftypes.Bool,
			},
		},
	},
}

type testServeDataSourceTwo struct{}

func (dt testServeDataSourceTwo) GetSchema(_ context.Context) (schema.Schema, []*tfprotov6.Diagnostic) {
	return schema.Schema{
		Attributes: map[string]schema.Attribute{
			"family": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
		},
	}, nil
}

func (dt testServeDataSourceTwo) NewDataSource(_ Provider) (DataSource, []*tfprotov6.Diagnostic) {
	panic("not implemented") // TODO: Implement
}

var testServeDataSourceTwoSchema = &tfprotov6.Schema{
	Block: &tfprotov6.SchemaBlock{
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:     "family",
				Optional: true,
				Computed: true,
				Type:     tftypes.String,
			},
			{
				Name:     "id",
				Computed: true,
				Type:     tftypes.String,
			},
			{
				Name:     "name",
				Optional: true,
				Computed: true,
				Type:     tftypes.String,
			},
		},
	},
}

func (t *testServeProvider) GetResources(_ context.Context) (map[string]ResourceType, []*tfprotov6.Diagnostic) {
	return map[string]ResourceType{
		"test_one": testServeResourceOne{},
		"test_two": testServeResourceTwo{},
	}, nil
}

func (t *testServeProvider) GetDataSources(_ context.Context) (map[string]DataSourceType, []*tfprotov6.Diagnostic) {
	return map[string]DataSourceType{
		"test_one": testServeDataSourceOne{},
		"test_two": testServeDataSourceTwo{},
	}, nil
}

func (t *testServeProvider) Configure(_ context.Context, _ ConfigureProviderRequest, _ *ConfigureProviderResponse) {
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
			"test_one": testServeResourceOneSchema,
			"test_two": testServeResourceTwoSchema,
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
			"test_one": testServeResourceOneSchema,
			"test_two": testServeResourceTwoSchema,
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

	// TODO: test configuring the provider
}

func TestServerReadResource(t *testing.T) {
	t.Parallel()

	// TODO: test reading resource
}

func TestServerPlanResourceChange(t *testing.T) {
	t.Parallel()

	// TODO: test planning
}

func TestServerApplyResourceChange(t *testing.T) {
	t.Parallel()

	// TODO: test applying
}

func TestServerReadDataSource(t *testing.T) {
	t.Parallel()

	// TODO: test reading data source
}
