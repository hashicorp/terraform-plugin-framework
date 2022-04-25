package tfsdk

import (
	"context"
	"fmt"
	"strings"
)

// ServeOpts are options for serving the provider.
type ServeOpts struct {
	// Address is the full address of the provider. Full address form has three
	// parts separated by forward slashes (/): Hostname, namespace, and
	// provider type ("name").
	//
	// For example: registry.terraform.io/hashicorp/random.
	Address string

	// Name is the name of the provider, in full address form. For example:
	// registry.terraform.io/hashicorp/random.
	//
	// Deprecated: Use Address field instead.
	Name string

	// Debug runs the provider in a mode acceptable for debugging and testing
	// processes, such as delve, by managing the process lifecycle. Information
	// needed for Terraform CLI to connect to the provider is output to stdout.
	// os.Interrupt (Ctrl-c) can be used to stop the provider.
	Debug bool
}

// Get provider address, based on whether Address or Name is specified.
//
// Deprecated: Will be removed in preference of just using the Address field.
func (opts ServeOpts) address(_ context.Context) string {
	if opts.Address != "" {
		return opts.Address
	}

	return opts.Name
}

// Validate a given provider address. This is only used for the Address field
// to preserve backwards compatibility for the Name field.
//
// This logic is manually implemented over importing
// github.com/hashicorp/terraform-registry-address as its functionality such as
// ParseAndInferProviderSourceString and ParseRawProviderSourceString allow
// shorter address formats, which would then require post-validation anyways.
func (opts ServeOpts) validateAddress(_ context.Context) error {
	addressParts := strings.Split(opts.Address, "/")
	formatErr := fmt.Errorf("expected hostname/namespace/type format, got: %s", opts.Address)

	if len(addressParts) != 3 {
		return formatErr
	}

	if addressParts[0] == "" || addressParts[1] == "" || addressParts[2] == "" {
		return formatErr
	}

	return nil
}

// Validation checks for provider defined ServeOpts.
//
// Current checks which return errors:
//
//    - If both Address and Name are set
//    - If neither Address nor Name is set
//    - If Address is set, it is a valid full provider address
func (opts ServeOpts) validate(ctx context.Context) error {
	if opts.Address == "" && opts.Name == "" {
		return fmt.Errorf("either Address or Name must be provided")
	}

	if opts.Address != "" && opts.Name != "" {
		return fmt.Errorf("only one of Address or Name should be provided")
	}

	if opts.Address != "" {
		err := opts.validateAddress(ctx)

		if err != nil {
			return fmt.Errorf("unable to validate Address: %w", err)
		}
	}

	return nil
}
