// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fromproto6_test

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromproto6"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestUpgradeResourceIdentityRequest(t *testing.T) {
	t.Parallel()

	testIdentitySchema := identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"test_attribute": schema.StringAttribute{
				Required: true,
			},
		},
	}

	testCases := map[string]struct {
		input               *tfprotov6.UpgradeResourceIdentityRequest
		identitySchema      fwschema.Schema
		resource            resource.Resource
		expected            *fwserver.UpgradeResourceIdentityRequest
		expectedDiagnostics diag.Diagnostics
	}{
		"nil": {
			input:    nil,
			expected: nil,
		},
		"rawIdentity": {
			input: &tfprotov6.UpgradeResourceIdentityRequest{
				RawIdentity: testNewRawState(t, map[string]interface{}{
					"test_attribute": "test-value",
				}),
			},
			identitySchema: testIdentitySchema,
			expected: &fwserver.UpgradeResourceIdentityRequest{
				RawState: testNewRawState(t, map[string]interface{}{
					"test_attribute": "test-value",
				}),
				IdentitySchema: testIdentitySchema,
			},
		},
		"resourceschema": {
			input:          &tfprotov6.UpgradeResourceIdentityRequest{},
			identitySchema: testIdentitySchema,
			expected: &fwserver.UpgradeResourceIdentityRequest{
				IdentitySchema: testIdentitySchema,
			},
		},
		"resourceschema-missing": {
			input:    &tfprotov6.UpgradeResourceIdentityRequest{},
			expected: nil,
			expectedDiagnostics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Unable to Create Empty Identity",
					"An unexpected error was encountered when creating the empty Identity. "+
						"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.\n\n"+
						"Please report this to the provider developer:\n\n"+
						"Missing schema.",
				),
			},
		},
		"version": {
			input: &tfprotov6.UpgradeResourceIdentityRequest{
				Version: 123,
			},
			identitySchema: testIdentitySchema,
			expected: &fwserver.UpgradeResourceIdentityRequest{
				IdentitySchema: testIdentitySchema,
				Version:        123,
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := fromproto6.UpgradeResourceIdentityRequest(context.Background(), testCase.input, testCase.resource, testCase.identitySchema)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
