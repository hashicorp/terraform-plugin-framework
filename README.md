[![PkgGoDev](https://pkg.go.dev/badge/github.com/hashicorp/terraform-plugin-framework)](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework)

# Terraform Plugin Framework

terraform-plugin-framework is a module for building [Terraform providers](https://www.terraform.io/language/providers). It is built on [terraform-plugin-go](https://github.com/hashicorp/terraform-plugin-go). It aims to provide as much of the power, predictability, and versatility of terraform-plugin-go as it can while abstracting away implementation details and repetitive, verbose tasks.

## Getting Started

* Try the [Terraform Plugin Framework collection](https://learn.hashicorp.com/collections/terraform/providers-plugin-framework) on HashiCorp Learn.
* Clone the [terraform-provider-scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework) template repository on GitHub for new providers.
* Follow the [Terraform Plugin Framework migration guide](https://www.terraform.io/plugin/framework/migrating) for converting existing terraform-plugin-sdk providers.
* Read the [Terraform Plugin Framework website](https://www.terraform.io/plugin/framework) for full documentation.
* Browse the [terraform-plugin-framework module](http://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework) on the Go package documentation website.
* Ask questions in the [Terraform Plugin Development section](https://discuss.hashicorp.com/c/terraform-providers/tf-plugin-sdk/43) on HashiCorp Discuss.


## Status

terraform-plugin-framework has reached **General Availability** phase and follows [semantic versioning](https://semver.org/) for Go and Terraform compatibility promises. We recommend only using tagged releases of this Go module and examining the CHANGELOG when upgrading to a new release. Major version releases contain breaking changes to existing provider code. Minor version releases introduce new functionality. Patch version releases contain bug fixes or documentation updates.

Refer to [Plugin Framework Benefits](https://developer.hashicorp.com/terraform/plugin/framework-benefits) for more information about benefits over [terraform-plugin-sdk](https://github.com/hashicorp/terraform-plugin-sdk).

## Terraform CLI Compatibility

Providers built with this framework are compatible with Terraform version v0.12 and above.

## Go Compatibility

This project follows the [support policy](https://golang.org/doc/devel/release.html#policy) of Go as its support policy. The two latest major releases of Go are supported by the project.

Currently, that means Go **1.19** or later must be used when including this project as a dependency.

## Contributing

See [`.github/CONTRIBUTING.md`](https://github.com/hashicorp/terraform-plugin-framework/blob/main/.github/CONTRIBUTING.md)

## License

[Mozilla Public License v2.0](https://github.com/hashicorp/terraform-plugin-framework/blob/main/LICENSE)
