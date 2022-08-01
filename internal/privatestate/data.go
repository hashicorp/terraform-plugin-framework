package privatestate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Data contains private state data for the framework and providers.
type Data struct {
	// Potential future usage:
	// Framework contains private state data for framework usage.
	Framework map[string][]byte

	// Provider contains private state data for provider usage.
	Provider ProviderData
}

// Bytes returns a JSON encoded slice of bytes containing the merged
// framework and provider private state data.
func (d Data) Bytes(_ context.Context) ([]byte, diag.Diagnostics) {
	var diags diag.Diagnostics

	bytes, err := json.Marshal(d)
	if err != nil {
		diags.AddError(
			"Error Encoding Private State",
			fmt.Sprintf("An error was encountered when encoding private state: %s.\n\n"+
				"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.", err),
		)

		return nil, diags
	}

	return bytes, diags
}

// NewData creates a new Data based on the given slice of bytes.
// It must be a JSON encoded slice of bytes, that is map[string][]byte.
func NewData(ctx context.Context, data []byte) (Data, diag.Diagnostics) {
	var (
		u     map[string][]byte
		diags diag.Diagnostics
	)

	err := json.Unmarshal(data, &u)
	if err != nil {
		diags.AddError(
			"Error Decoding Private State",
			fmt.Sprintf("An error was encountered when decoding private state: %s.\n\n"+
				"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.", err),
		)

		return Data{}, diags
	}

	output := Data{
		Framework: make(map[string][]byte),
		Provider:  make(map[string][]byte),
	}

	for k, v := range u {
		if isInvalidProviderDataKey(ctx, k) {
			output.Framework[k] = v
			continue
		}

		output.Provider[k] = v
	}

	return output, diags
}

// ProviderData contains private state data for provider usage.
type ProviderData map[string][]byte

// GetKey returns the private state data associated with the given key.
//
// If the key is reserved for framework usage, an error diagnostic
// is returned. If the key is valid, but private state data is not found,
// nil is returned.
//
// The naming of keys only matters in context of a single resource,
// however care should be taken that any historical keys are not reused
// without accounting for older resource instances that may still have
// older data at the key.
func (d ProviderData) GetKey(ctx context.Context, key string) ([]byte, diag.Diagnostics) {
	diags := ValidateProviderDataKey(ctx, key)

	if diags.HasError() {
		return nil, diags
	}

	value, ok := d[key]

	if !ok {
		return nil, nil
	}

	return value, nil
}

// SetKey sets the private state data at the given key.
//
// If the key is reserved for framework usage, an error diagnostic
// is returned. The data must be valid JSON and UTF-8 safe or an error
// diagnostic is returned.
//
// The naming of keys only matters in context of a single resource,
// however care should be taken that any historical keys are not reused
// without accounting for older resource instances that may still have
// older data at the key.
func (d ProviderData) SetKey(ctx context.Context, key string, value []byte) diag.Diagnostics {
	diags := ValidateProviderDataKey(ctx, key)

	if diags.HasError() {
		return diags
	}

	if !utf8.Valid(value) {
		diags.AddError("UTF-8 Invalid",
			"Values stored in private state must be valid UTF-8")

		return diags
	}

	if !json.Valid(value) {
		diags.AddError("JSON Invalid",
			"Values stored in private state must be valid JSON")

		return diags
	}

	d[key] = value

	return nil
}

// ValidateProviderDataKey determines whether the key supplied is allowed on the basis of any
// restrictions that are in place, such as key prefixes that are reserved for use with
// framework private state data.
func ValidateProviderDataKey(ctx context.Context, key string) diag.Diagnostics {
	if isInvalidProviderDataKey(ctx, key) {
		return diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Restricted Namespace",
				"Using a period ('.') as a prefix for a key used in private state is not allowed",
			),
		}
	}

	return nil
}

// isInvalidProviderDataKey determines whether the supplied key has a prefix that is reserved for
// keys in Data.Framework
func isInvalidProviderDataKey(_ context.Context, key string) bool {
	return strings.HasPrefix(key, ".")
}
