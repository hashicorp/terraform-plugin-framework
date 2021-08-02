package tfsdk

// ValidateDataSourceConfigRequest represents a request to validate the
// configuration of a data source. An instance of this request struct is
// supplied as an argument to the DataSource ValidateConfig receiver method
// or automatically passed through to each ConfigValidator.
type ValidateDataSourceConfigRequest struct {
	// Config is the configuration the user supplied for the data source.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	Config Config
}

// ValidateProviderConfigRequest represents a request to validate the
// configuration of a provider. An instance of this request struct is
// supplied as an argument to the Provider ValidateConfig receiver method
// or automatically passed through to each ConfigValidator.
type ValidateProviderConfigRequest struct {
	// Config is the configuration the user supplied for the provider.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	Config Config
}

// ValidateResourceConfigRequest represents a request to validate the
// configuration of a resource. An instance of this request struct is
// supplied as an argument to the Resource ValidateConfig receiver method
// or automatically passed through to each ConfigValidator.
type ValidateResourceConfigRequest struct {
	// Config is the configuration the user supplied for the resource.
	//
	// This configuration may contain unknown values if a user uses
	// interpolation or other functionality that would prevent Terraform
	// from knowing the value at request time.
	Config Config
}
