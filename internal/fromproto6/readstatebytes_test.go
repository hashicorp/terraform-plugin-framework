// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto6_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testschema"
	"github.com/hashicorp/terraform-plugin-framework/statestore"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestReadStateBytesRequest(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input               *tfprotov6.ReadStateBytesRequest
		stateStore          statestore.StateStore
		statestoreSchema    fwschema.Schema
		expected            *fwserver.ReadStateBytesRequest
		expectedDiagnostics diag.Diagnostics
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"empty": {
			input:            &tfprotov6.ReadStateBytesRequest{},
			statestoreSchema: nil,
			expected:         nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Missing StateBytes Schema",
					"An unexpected error was encountered when handling the request. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"id-missing": {
			input: &tfprotov6.ReadStateBytesRequest{
				StateId: "",
			},
			statestoreSchema: testschema.Schema{},
			expected:         nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Missing State ID",
					"An unexpected error was encountered when handling the request. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing State ID.",
				),
			},
		},
		"state-id": {
			input: &tfprotov6.ReadStateBytesRequest{
				StateId: "test-value",
			},
			statestoreSchema: testschema.Schema{},
			expected: &fwserver.ReadStateBytesRequest{
				StateID: "test-value",
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto6.ReadStateBytesRequest(context.Background(), testCase.input, testCase.stateStore, testCase.statestoreSchema)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
