// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschemadata_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValueSemanticEqualityBool(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request  fwschemadata.ValueSemanticEqualityRequest
		expected *fwschemadata.ValueSemanticEqualityResponse
	}{
		"BoolValue": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path:             path.Root("test"),
				PriorValue:       types.BoolValue(false),
				ProposedNewValue: types.BoolValue(true),
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: types.BoolValue(true),
			},
		},
		"BoolValuableWithSemanticEquals-true": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.BoolValueWithSemanticEquals{
					BoolValue:      types.BoolValue(false),
					SemanticEquals: true,
				},
				ProposedNewValue: testtypes.BoolValueWithSemanticEquals{
					BoolValue:      types.BoolValue(true),
					SemanticEquals: true,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.BoolValueWithSemanticEquals{
					BoolValue:      types.BoolValue(false),
					SemanticEquals: true,
				},
			},
		},
		"BoolValuableWithSemanticEquals-false": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.BoolValueWithSemanticEquals{
					BoolValue:      types.BoolValue(false),
					SemanticEquals: false,
				},
				ProposedNewValue: testtypes.BoolValueWithSemanticEquals{
					BoolValue:      types.BoolValue(true),
					SemanticEquals: false,
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.BoolValueWithSemanticEquals{
					BoolValue:      types.BoolValue(true),
					SemanticEquals: false,
				},
			},
		},
		"BoolValuableWithSemanticEquals-diagnostics": {
			request: fwschemadata.ValueSemanticEqualityRequest{
				Path: path.Root("test"),
				PriorValue: testtypes.BoolValueWithSemanticEquals{
					BoolValue:      types.BoolValue(false),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
				ProposedNewValue: testtypes.BoolValueWithSemanticEquals{
					BoolValue:      types.BoolValue(true),
					SemanticEquals: false,
					SemanticEqualsDiagnostics: diag.Diagnostics{
						diag.NewErrorDiagnostic("test summary 1", "test detail 1"),
						diag.NewErrorDiagnostic("test summary 2", "test detail 2"),
					},
				},
			},
			expected: &fwschemadata.ValueSemanticEqualityResponse{
				NewValue: testtypes.BoolValueWithSemanticEquals{
					BoolValue:      types.BoolValue(true),
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

			fwschemadata.ValueSemanticEqualityBool(context.Background(), testCase.request, got)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
