[![PkgGoDev](https://pkg.go.dev/badge/github.com/hashicorp/terraform-plugin-framework)](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework)

# Terraform Plugin Framework

terraform-plugin-framework is a module for building Terraform providers. It is built on [terraform-plugin-go](https://github.com/hashicorp/terraform-plugin-go). It aims to provide as much of the power, predictability, and versatility of terraform-plugin-go as it can while abstracting away implementation details and repetitive, verbose tasks.

## Status

terraform-plugin-framework is still in **technical preview**. We are committed to moving forward with the module, but cannot guarantee any of its interfaces will not change as long as it is in version 0. We're waiting for more feedback, usage, and maturity before we're comfortable committing to APIs with the same years-long support timelines that [terraform-plugin-sdk](https://github.com/hashicorp/terraform-plugin-sdk) brings. We do not expect practitioner experiences to break or change as a result of these changes, only the abstractions surfaced to provider developers.

terraform-plugin-framework is also not at full feature parity with terraform-plugin-sdk yet. Notably, it doesn't offer support for [using timeouts](https://github.com/hashicorp/terraform-plugin-framework/issues/62). We plan to add these features soon. See [Which SDK Should I Use?](https://terraform.io/docs/plugin/which-sdk.html) on terraform.io for more information.

We believe terraform-plugin-framework is still a suitable and reliable module to build Terraform providers on, and encourage community members that can afford occasional breaking changes to build with it. terraform-plugin-framework will eventually become a new major version of terraform-plugin-sdk, at which point its interfaces will be stable, but we need real-world use and feedback before we can be comfortable making those commitments. When that happens, this repository will be archived.

We recommend only using tagged releases of this module, and examining the CHANGELOG when upgrading to a new release. Breaking changes will only be made in minor versions; patch releases will always maintain backwards compatibility.

We welcome and appreciate issues and PRs discussing both the design and implementation of this module.

## Terraform CLI Compatibility

Providers built with this framework are compatible with Terraform version v0.12 and above.

## Go Compatibility

This project follows the [support policy](https://golang.org/doc/devel/release.html#policy) of Go as its support policy. The two latest major releases of Go are supported by the project.

Currently, that means Go **1.18** or later must be used when including this project as a dependency.

## Getting Started

Documentation for terraform-plugin-framework is still in development. In the meantime, the [GoDoc](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework) is the best source of documentation.

The [`provider.Provider`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/provider#Provider) type is the root of your provider implementation. From there, [`provider.ResourceType`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/provider#ResourceType) and [`provider.DataSourceType`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/provider#DataSourceType) implementations define the [schema](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/tfsdk#Schema) of your resources and data sources, and how to create [`resource.Resource`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/resource#Resource) and [`datasource.DataSource`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/datasource#DataSource) implementations that talk to the API.

## Contributing

See [`.github/CONTRIBUTING.md`](https://github.com/hashicorp/terraform-plugin-framework/blob/main/.github/CONTRIBUTING.md)

## License

[Mozilla Public License v2.0](https://github.com/hashicorp/terraform-plugin-framework/blob/main/LICENSE)
