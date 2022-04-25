package tfsdk

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestServeOptsAddress(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		serveOpts ServeOpts
		expected  string
	}{
		"Address": {
			serveOpts: ServeOpts{
				Address: "registry.terraform.io/hashicorp/testing",
			},
			expected: "registry.terraform.io/hashicorp/testing",
		},
		"Address-and-Name-both": {
			serveOpts: ServeOpts{
				Address: "registry.terraform.io/hashicorp/testing",
				Name:    "testing",
			},
			expected: "registry.terraform.io/hashicorp/testing",
		},
		"Name": {
			serveOpts: ServeOpts{
				Name: "testing",
			},
			expected: "testing",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.serveOpts.address(context.Background())

			if got != testCase.expected {
				t.Fatalf("expected %q, got: %s", testCase.expected, got)
			}
		})
	}
}

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
		"Address-and-Name-both": {
			serveOpts: ServeOpts{
				Address: "registry.terraform.io/hashicorp/testing",
				Name:    "testing",
			},
			expectedError: fmt.Errorf("only one of Address or Name should be provided"),
		},
		"Address-and-Name-missing": {
			serveOpts:     ServeOpts{},
			expectedError: fmt.Errorf("either Address or Name must be provided"),
		},
		"Address-invalid-type-only": {
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
		"Name": {
			serveOpts: ServeOpts{
				Name: "testing",
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
		"invalid-type-only": {
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
