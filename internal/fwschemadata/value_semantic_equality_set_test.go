package fwschemadata_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValueSemanticEqualitySet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  fwschemadata.ValueSemanticEqualityRequest
		expected *fwschemadata.ValueSemanticEqualityResponse
	}{
		// Type and ElementType without semantic equality
		"SetValue": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("prior"),
					},
				),
				ProposedNewValue: types.SetValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("new"),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.SetValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("new"),
					},
				),
			},
		},
		// ElementType with semantic equality
		"SetValue-StringValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: true,
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("prior"),
							SemanticEquals: true,
						},
					},
				),
				ProposedNewValue: types.SetValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: true,
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("new"),
							SemanticEquals: true,
						},
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.SetValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: true,
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("prior"),
							SemanticEquals: true,
						},
					},
				),
			},
		},
		"SetValue-StringValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: false,
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("prior"),
							SemanticEquals: false,
						},
					},
				),
				ProposedNewValue: types.SetValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: false,
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("new"),
							SemanticEquals: false,
						},
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.SetValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: false,
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("new"),
							SemanticEquals: false,
						},
					},
				),
			},
		},
		"SetValue-StringValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: true,
						SemanticEqualsDiagnostics: diag.Diagnostics{
							diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
							diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
						},
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("prior"),
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
				),
				ProposedNewValue: types.SetValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: true,
						SemanticEqualsDiagnostics: diag.Diagnostics{
							diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
							diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
						},
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("new"),
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.SetValueMust(
					testtypes.StringTypeWithSemanticEquals{
						SemanticEquals: true,
						SemanticEqualsDiagnostics: diag.Diagnostics{
							diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
							diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
						},
					},
					[]attr.Value{
						testtypes.StringValueWithSemanticEquals{
							StringValue:    types.StringValue("new"),
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
				),
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
		// Nested ElementType with semantic equality
		"SetValue-SetValue-StringValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					[]attr.Value{
						types.SetValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("prior"),
									SemanticEquals: true,
								},
							},
						),
					},
				),
				ProposedNewValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					[]attr.Value{
						types.SetValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new"),
									SemanticEquals: true,
								},
							},
						),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
						},
					},
					[]attr.Value{
						types.SetValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("prior"),
									SemanticEquals: true,
								},
							},
						),
					},
				),
			},
		},
		"SetValue-SetValue-StringValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: false,
						},
					},
					[]attr.Value{
						types.SetValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: false,
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("prior"),
									SemanticEquals: false,
								},
							},
						),
					},
				),
				ProposedNewValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: false,
						},
					},
					[]attr.Value{
						types.SetValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: false,
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new"),
									SemanticEquals: false,
								},
							},
						),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: false,
						},
					},
					[]attr.Value{
						types.SetValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: false,
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new"),
									SemanticEquals: false,
								},
							},
						),
					},
				),
			},
		},
		"SetValue-SetValue-StringValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
					[]attr.Value{
						types.SetValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
								SemanticEqualsDiagnostics: diag.Diagnostics{
									diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
									diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
								},
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("prior"),
									SemanticEquals: true,
									SemanticEqualsDiagnostics: diag.Diagnostics{
										diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
										diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
									},
								},
							},
						),
					},
				),
				ProposedNewValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
					[]attr.Value{
						types.SetValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
								SemanticEqualsDiagnostics: diag.Diagnostics{
									diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
									diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
								},
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new"),
									SemanticEquals: true,
									SemanticEqualsDiagnostics: diag.Diagnostics{
										diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
										diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
									},
								},
							},
						),
					},
				),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.SetValueMust(
					types.SetType{
						ElemType: testtypes.StringTypeWithSemanticEquals{
							SemanticEquals: true,
							SemanticEqualsDiagnostics: diag.Diagnostics{
								diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
								diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
							},
						},
					},
					[]attr.Value{
						types.SetValueMust(
							testtypes.StringTypeWithSemanticEquals{
								SemanticEquals: true,
								SemanticEqualsDiagnostics: diag.Diagnostics{
									diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
									diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
								},
							},
							[]attr.Value{
								testtypes.StringValueWithSemanticEquals{
									StringValue:    types.StringValue("new"),
									SemanticEquals: true,
									SemanticEqualsDiagnostics: diag.Diagnostics{
										diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
										diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
									},
								},
							},
						),
					},
				),
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
		// Type with semantic equality
		"SetValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: true,
				},
				ProposedNewValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: true,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: true,
				},
			},
		},
		"SetValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: false,
				},
				ProposedNewValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: false,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: false,
				},
			},
		},
		"SetValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("prior"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				ProposedNewValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.SetValueWithSemanticEquals{
					SetValue: types.SetValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("new"),
						},
					),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				Diagnostics: diag.Diagnostics{
					diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
					diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
				},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testCase.request.ProposedNewValue,
			}

			fwschemadata.ValueSemanticEqualitySet(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
