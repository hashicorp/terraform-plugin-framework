// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providerserver

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestServeOptsValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		serveOpts     ServeOpts
		expectedError error
	}{
		"Address": {
			serveOpts: ServeOpts{
				Address: "registry.terraform.io/hashicorp/testing",
			},
		},
		"Address-missing": {
			serveOpts:     ServeOpts{},
			expectedError: fmt.Errorf("Address must be provided"),
		},
		"Address-invalid-missing-hostname-and-namespace": {
			serveOpts: ServeOpts{
				Address: "testing",
			},
			expectedError: fmt.Errorf("unable to validate Address: expected hostname/namespace/type format, got: testing"),
		},
		"Address-invalid-missing-hostname": {
			serveOpts: ServeOpts{
				Address: "hashicorp/testing",
			},
			expectedError: fmt.Errorf("unable to validate Address: expected hostname/namespace/type format, got: hashicorp/testing"),
		},
		"ProtocolVersion-invalid": {
			serveOpts: ServeOpts{
				Address:         "registry.terraform.io/hashicorp/testing",
				ProtocolVersion: 999,
			},
			expectedError: fmt.Errorf("ProtocolVersion, if set, must be 5 or 6"),
		},
		"ProtocolVersion-5": {
			serveOpts: ServeOpts{
				Address:         "registry.terraform.io/hashicorp/testing",
				ProtocolVersion: 5,
			},
		},
		"ProtocolVersion-6": {
			serveOpts: ServeOpts{
				Address:         "registry.terraform.io/hashicorp/testing",
				ProtocolVersion: 6,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := testCase.serveOpts.validate(context.Background())

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
		})
	}
}

func TestServeOptsValidateAddress(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		serveOpts     ServeOpts
		expectedError error
	}{
		"valid": {
			serveOpts: ServeOpts{
				Address: "registry.terraform.io/hashicorp/testing",
			},
		},
		"invalid-missing-hostname-and-namepsace": {
			serveOpts: ServeOpts{
				Address: "testing",
			},
			expectedError: fmt.Errorf("expected hostname/namespace/type format, got: testing"),
		},
		"invalid-missing-hostname": {
			serveOpts: ServeOpts{
				Address: "hashicorp/testing",
			},
			expectedError: fmt.Errorf("expected hostname/namespace/type format, got: hashicorp/testing"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := testCase.serveOpts.validateAddress(context.Background())

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
		})
	}
}
