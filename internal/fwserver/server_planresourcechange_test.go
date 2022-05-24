package fwserver_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/emptyprovider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestMarkComputedNilsAsUnknown(t *testing.T) {
	t.Parallel()

	s := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
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
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
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
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
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

	got, err := tftypes.Transform(input, fwserver.MarkComputedNilsAsUnknown(context.Background(), input, s))
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

func TestNormaliseRequiresReplace(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []*tftypes.AttributePath
		expected []*tftypes.AttributePath
	}

	tests := map[string]testCase{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"no-duplicates": {
			input: []*tftypes.AttributePath{
				tftypes.NewAttributePath().WithAttributeName("name2"),
				tftypes.NewAttributePath().WithAttributeName("name1"),
				tftypes.NewAttributePath().WithElementKeyInt(1234),
				tftypes.NewAttributePath().WithAttributeName("name1").WithElementKeyString("elementkey"),
			},
			expected: []*tftypes.AttributePath{
				tftypes.NewAttributePath().WithAttributeName("name1"),
				tftypes.NewAttributePath().WithAttributeName("name1").WithElementKeyString("elementkey"),
				tftypes.NewAttributePath().WithAttributeName("name2"),
				tftypes.NewAttributePath().WithElementKeyInt(1234),
			},
		},
		"duplicates": {
			input: []*tftypes.AttributePath{
				tftypes.NewAttributePath().WithAttributeName("name1"),
				tftypes.NewAttributePath().WithAttributeName("name1"),
				tftypes.NewAttributePath().WithElementKeyInt(1234),
				tftypes.NewAttributePath().WithElementKeyInt(1234),
			},
			expected: []*tftypes.AttributePath{
				tftypes.NewAttributePath().WithAttributeName("name1"),
				tftypes.NewAttributePath().WithElementKeyInt(1234),
			},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			actual := fwserver.NormaliseRequiresReplace(context.Background(), tc.input)

			if diff := cmp.Diff(actual, tc.expected, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
				return
			}
		})
	}
}

// TODO: Migrate tfsdk.Provider bits of proto6server.testProviderServer to
// new internal/testing/provider.Provider that allows customization of all
// method implementations via struct fields. Then, create additional test
// cases in this unit test.
//
// For now this testing is covered by proto6server.PlanResourceChange.
//
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/215
func TestServerPlanResourceChange(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		server           *fwserver.Server
		request          *fwserver.PlanResourceChangeRequest
		expectedResponse *fwserver.PlanResourceChangeResponse
	}{
		"empty-provider": {
			server: &fwserver.Server{
				Provider: &emptyprovider.Provider{},
			},
			expectedResponse: &fwserver.PlanResourceChangeResponse{},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwserver.PlanResourceChangeResponse{}
			testCase.server.PlanResourceChange(context.Background(), testCase.request, response)

			if diff := cmp.Diff(response, testCase.expectedResponse); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
