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
		Version:             1,
		DeprecationMessage:  "Deprecated in favor of other_resource",
		Description:         "A test resource.",
		MarkdownDescription: "A **test** resource",
		Attributes: map[string]schema.Attribute{
			"required": {
				Type:                types.StringType,
				Required:            true,
				Description:         "A required attribute",
				MarkdownDescription: "A **required** attribute",
			},
			"optional": {
				Type:                types.StringType,
				Optional:            true,
				Description:         "An optional attribute",
				MarkdownDescription: "An _optional_ attribute",
			},
			"computed": {
				Type:                types.StringType,
				Computed:            true,
				Description:         "A read-only attribute",
				MarkdownDescription: "A read-only attribute",
			},
			"optional_computed": {
				Type:                types.StringType,
				Optional:            true,
				Computed:            true,
				Description:         "An optional and computed attribute",
				MarkdownDescription: "An optional and computed attribute",
			},
			"required_sensitive": {
				Type:                types.StringType,
				Required:            true,
				Sensitive:           true,
				Description:         "A required attribute with sensitive values",
				MarkdownDescription: "A _required_ attribute with sensitive values",
			},
			"optional_sensitive": {
				Type:                types.StringType,
				Optional:            true,
				Sensitive:           true,
				Description:         "An optional attribute with sensitive values",
				MarkdownDescription: "An _optional_ attribute with sensitive values",
			},
			"computed_sensitive": {
				Type:                types.StringType,
				Computed:            true,
				Sensitive:           true,
				Description:         "A read-only attribute with sensitive values",
				MarkdownDescription: "A read-only attribute with _sensitive_ values",
			},
			"optional_computed_sensitive": {
				Type:                types.StringType,
				Computed:            true,
				Sensitive:           true,
				Description:         "An optional and computed attribute with sensitive values",
				MarkdownDescription: "An _optional_ and computed attribute with sensitive values",
			},
			"optional_deprecated": {
				Type:                types.StringType,
				Optional:            true,
				DeprecationMessage:  "Deprecated, please use \"optional\" instead",
				Description:         "A deprecated, optional attribute",
				MarkdownDescription: "A **deprecated**, optional attribute",
			},
			"optional_computed_deprecated": {
				Type:                types.StringType,
				Optional:            true,
				Computed:            true,
				DeprecationMessage:  "Deprecated, please use \"optional_computed\" instead",
				Description:         "A deprecated, optional and computed attribute",
				MarkdownDescription: "A **deprecated**, optional and computed attribute",
			},
			"optional_computed_sensitive_deprecated": {
				Type:                types.StringType,
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
				DeprecationMessage:  "Deprecated, please use \"optional_computed_sensitive\" instead",
				Description:         "A deprecated, optional, and computed attribute with sensitive values",
				MarkdownDescription: "A **deprecated**, optional, and computed attribute with sensitive values",
			},
			"string": {
				Type:                types.StringType,
				Optional:            true,
				Description:         "A string attribute",
				MarkdownDescription: "A _string_ attribute",
			},
			"number": {
				Type:                types.NumberType,
				Optional:            true,
				Description:         "A number attribute",
				MarkdownDescription: "A _number_ attribute",
			},
			"bool": {
				Type:                types.BoolType,
				Optional:            true,
				Description:         "A boolean attribute",
				MarkdownDescription: "A _boolean_ attribute",
			},
			"list-string": {
				Type: types.ListType{
					ElemType: types.StringType,
				},
				Optional:            true,
				Description:         "A list of strings",
				MarkdownDescription: "A list of **strings**",
			},
			"list-list-string": {
				Type: types.ListType{
					ElemType: types.ListType{
						ElemType: types.StringType,
					},
				},
				Optional:            true,
				Description:         "A list of lists of strings",
				MarkdownDescription: "A list of lists of _strings_",
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
				Optional:            true,
				Description:         "A list of objects",
				MarkdownDescription: "A list of _objects_",
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
				Optional:            true,
				Description:         "An object attribute",
				MarkdownDescription: "An object _attribute_",
			},
			"empty-object": {
				Type:                types.ObjectType{},
				Optional:            true,
				Description:         "An object attribute with no attributes",
				MarkdownDescription: "An object attribute _with no attributes_",
			},
			// TODO: add maps when we support them
			// TODO: add sets when we support them
			// TODO: add tuples when we support them
			"single-nested-attributes": {
				Attributes: schema.SingleNestedAttributes(map[string]schema.Attribute{
					"foo": {
						Type:                types.StringType,
						Optional:            true,
						Computed:            true,
						Description:         "A nested string attribute",
						MarkdownDescription: "A nested _string_ attribute",
					},
					"bar": {
						Type:                types.NumberType,
						Required:            true,
						Description:         "A nested number attribute",
						MarkdownDescription: "A nested _number_ attribute",
					},
				}),
				Optional:            true,
				Description:         "A single nested attribute",
				MarkdownDescription: "A single _nested_ attribute",
			},
			"list-nested-attributes": {
				Attributes: schema.ListNestedAttributes(map[string]schema.Attribute{
					"foo": {
						Type:                types.StringType,
						Optional:            true,
						Computed:            true,
						Description:         "A nested string attribute",
						MarkdownDescription: "A nested _string_ attribute",
					},
					"bar": {
						Type:                types.NumberType,
						Required:            true,
						Description:         "A nested number attribute",
						MarkdownDescription: "A nested _number_ attribute",
					},
				}, schema.ListNestedAttributesOptions{
					MinItems: 2,
					MaxItems: 10,
				}),
				Optional:            true,
				Description:         "A list nested attribute",
				MarkdownDescription: "A list _nested_ attribute",
			},
		},
	}, nil
}

var testServeProviderProviderSchema = &tfprotov6.Schema{
	Version: 1,
	Block: &tfprotov6.SchemaBlock{
		Deprecated:      true,
		Description:     "A **test** resource",
		DescriptionKind: tfprotov6.StringKindMarkdown,
		Attributes: []*tfprotov6.SchemaAttribute{
			{
				Name:            "bool",
				Type:            tftypes.Bool,
				Optional:        true,
				Description:     "A _boolean_ attribute",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
			{
				Name:            "computed",
				Type:            tftypes.String,
				Computed:        true,
				Description:     "A read-only attribute",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
			{
				Name:            "computed_sensitive",
				Type:            tftypes.String,
				Computed:        true,
				Sensitive:       true,
				Description:     "A read-only attribute with _sensitive_ values",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
			{
				Name: "empty-object",
				Type: tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{},
				},
				Optional:        true,
				Description:     "An object attribute _with no attributes_",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
			{
				Name: "list-list-string",
				Type: tftypes.List{
					ElementType: tftypes.List{
						ElementType: tftypes.String,
					},
				},
				Optional:        true,
				Description:     "A list of lists of _strings_",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
			{
				Name: "list-nested-attributes",
				NestedType: &tfprotov6.SchemaObject{
					Nesting:  tfprotov6.SchemaObjectNestingModeList,
					MaxItems: 10,
					MinItems: 2,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:            "bar",
							Type:            tftypes.Number,
							Required:        true,
							Description:     "A nested _number_ attribute",
							DescriptionKind: tfprotov6.StringKindMarkdown,
						},
						{
							Name:            "foo",
							Type:            tftypes.String,
							Optional:        true,
							Computed:        true,
							Description:     "A nested _string_ attribute",
							DescriptionKind: tfprotov6.StringKindMarkdown,
						},
					},
				},
				Optional:        true,
				Description:     "A list _nested_ attribute",
				DescriptionKind: tfprotov6.StringKindMarkdown,
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
				Optional:        true,
				Description:     "A list of _objects_",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
			{
				Name: "list-string",
				Type: tftypes.List{
					ElementType: tftypes.String,
				},
				Optional:        true,
				Description:     "A list of **strings**",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
			{
				Name:            "number",
				Type:            tftypes.Number,
				Optional:        true,
				Description:     "A _number_ attribute",
				DescriptionKind: tfprotov6.StringKindMarkdown,
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
				Optional:        true,
				Description:     "An object _attribute_",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
			{
				Name:            "optional",
				Type:            tftypes.String,
				Optional:        true,
				Description:     "An _optional_ attribute",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
			{
				Name:            "optional_computed",
				Type:            tftypes.String,
				Optional:        true,
				Computed:        true,
				Description:     "An optional and computed attribute",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
			{
				Name:            "optional_computed_deprecated",
				Type:            tftypes.String,
				Optional:        true,
				Computed:        true,
				Deprecated:      true,
				Description:     "A **deprecated**, optional and computed attribute",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
			{
				Name:            "optional_computed_sensitive",
				Type:            tftypes.String,
				Computed:        true,
				Sensitive:       true,
				Description:     "An _optional_ and computed attribute with sensitive values",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
			{
				Name:            "optional_computed_sensitive_deprecated",
				Type:            tftypes.String,
				Optional:        true,
				Computed:        true,
				Sensitive:       true,
				Deprecated:      true,
				Description:     "A **deprecated**, optional, and computed attribute with sensitive values",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
			{
				Name:            "optional_deprecated",
				Type:            tftypes.String,
				Optional:        true,
				Deprecated:      true,
				Description:     "A **deprecated**, optional attribute",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
			{
				Name:            "optional_sensitive",
				Type:            tftypes.String,
				Optional:        true,
				Sensitive:       true,
				Description:     "An _optional_ attribute with sensitive values",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
			{
				Name:            "required",
				Type:            tftypes.String,
				Required:        true,
				Description:     "A **required** attribute",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
			{
				Name:            "required_sensitive",
				Type:            tftypes.String,
				Required:        true,
				Sensitive:       true,
				Description:     "A _required_ attribute with sensitive values",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
			{
				Name: "single-nested-attributes",
				NestedType: &tfprotov6.SchemaObject{
					Nesting: tfprotov6.SchemaObjectNestingModeSingle,
					Attributes: []*tfprotov6.SchemaAttribute{
						{
							Name:            "bar",
							Type:            tftypes.Number,
							Required:        true,
							Description:     "A nested _number_ attribute",
							DescriptionKind: tfprotov6.StringKindMarkdown,
						},
						{
							Name:            "foo",
							Type:            tftypes.String,
							Optional:        true,
							Computed:        true,
							Description:     "A nested _string_ attribute",
							DescriptionKind: tfprotov6.StringKindMarkdown,
						},
					},
				},
				Optional:        true,
				Description:     "A single _nested_ attribute",
				DescriptionKind: tfprotov6.StringKindMarkdown,
			},
			{
				Name:            "string",
				Type:            tftypes.String,
				Optional:        true,
				Description:     "A _string_ attribute",
				DescriptionKind: tfprotov6.StringKindMarkdown,
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

func (t *testServeProvider) GetResources(_ context.Context) (map[string]ResourceType, []*tfprotov6.Diagnostic) {
	return nil, nil
}

func (t *testServeProvider) GetDataSources(_ context.Context) (map[string]DataSourceType, []*tfprotov6.Diagnostic) {
	return nil, nil
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
		Provider:        testServeProviderProviderSchema,
		ResourceSchemas: map[string]*tfprotov6.Schema{
			// TODO: include resource schemas
		},
		DataSourceSchemas: map[string]*tfprotov6.Schema{
			// TODO: include data source schemas
		},
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (-wanted, +got): %s", diff)
	}
}
