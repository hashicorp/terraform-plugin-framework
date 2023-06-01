// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package totftypes_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/internal/totftypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestAttributePathStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		fw            path.PathStep
		expected      tftypes.AttributePathStep
		expectedError error
	}{
		"nil": {
			fw:            nil,
			expected:      nil,
			expectedError: fmt.Errorf("unknown path.PathStep: <nil>"),
		},
		"PathStepAttributeName": {
			fw:       path.PathStepAttributeName("test"),
			expected: tftypes.AttributeName("test"),
		},
		"PathStepElementKeyInt": {
			fw:       path.PathStepElementKeyInt(1),
			expected: tftypes.ElementKeyInt(1),
		},
		"PathStepElementKeyString": {
			fw:       path.PathStepElementKeyString("test"),
			expected: tftypes.ElementKeyString("test"),
		},
		"PathStepElementKeyValue": {
			fw:       path.PathStepElementKeyValue{Value: types.StringValue("test")},
			expected: tftypes.ElementKeyValue(tftypes.NewValue(tftypes.String, "test")),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := totftypes.AttributePathStep(context.Background(), testCase.fw)

			if err != nil {
				if testCase.expectedError == nil {
					t.Fatalf("expected no error, got: %s", err)
				}

				if !strings.Contains(err.Error(), testCase.expectedError.Error()) {
					t.Fatalf("expected error %q, got: %s", testCase.expectedError, err)
				}
			}

			if err == nil && testCase.expectedError != nil {
				t.Fatalf("got no error, expected: %s", testCase.expectedError)
			}

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
