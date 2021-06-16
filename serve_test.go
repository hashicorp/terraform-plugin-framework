package tfsdk

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
