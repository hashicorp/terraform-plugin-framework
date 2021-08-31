[![PkgGoDev](https://pkg.go.dev/badge/github.com/hashicorp/terraform-plugin-framework)](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework)

# Terraform Plugin Framework

terraform-plugin-framework is a module for building Terraform providers. It is built on [terraform-plugin-go](https://github.com/hashicorp/terraform-plugin-go). It aims to provide as much of the power, predictability, and versatility of terraform-plugin-go as it can while abstracting away implementation details and repetitive, verbose tasks.

## Status

![experimental](https://camo.githubusercontent.com/8ad47215ae8b556345c074d2636cdf5e8a7f54068c110d1a1795501b43fab52e/68747470733a2f2f696d672e736869656c64732e696f2f62616467652f7374617475732d6578706572696d656e74616c2d454141413332)

terraform-plugin-framework is still experimental. We are committed to moving forward with the module, but cannot guarantee any of its interfaces will not change as long as it is in version 0. We're waiting for more feedback, usage, and maturity before we're comfortable committing to APIs with the same years-long support timelines that [terraform-plugin-sdk](https://github.com/hashicorp/terraform-plugin-sdk) brings.

terraform-plugin-framework is also not at full feature parity with terraform-plugin-sdk yet. Notably, it doesn't offer support for [validation](https://github.com/hashicorp/terraform-plugin-framework/issues/17), [modifying plans](https://github.com/hashicorp/terraform-plugin-framework/issues/34) (including marking resources as needing to be recreated), [importing resources](https://github.com/hashicorp/terraform-plugin-framework/issues/33), or [upgrading resource state](https://github.com/hashicorp/terraform-plugin-framework/issues/42). We plan to add these features soon.

We believe terraform-plugin-framework is still a suitable and reliable module to build Terraform providers on, and encourage community members that can afford occasional breaking changes to build with it. terraform-plugin-framework will eventually become a new major version of terraform-plugin-sdk, at which point its interfaces will be stable, but we need real-world use and feedback before we can be comfortable making those commitments. When that happens, this repository will be archived.

We recommend only using tagged releases of this module, and examining the CHANGELOG when upgrading to a new release. Breaking changes will only be made in minor versions; patch releases will always maintain backwards compatibility.

We welcome and appreciate issues and PRs discussing both the design and implementation of this module.

## Terraform CLI Compatibility

Plugins built with this framework are only compatible with Terraform versions above v1.0.3.

## Go Compatibility

Prior to its 1.0 release, this module will only support the latest released version of Go, and may use features and functionality introduced in that version of Go.

Currently that means Go **1.17** must be used when building a provider with this framework.

## Getting Started

Documentation for terraform-plugin-framework is still in development. In the meantime, the [GoDoc](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework) is the best source of documentation.

The [`tfsdk.Provider`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/tfsdk#Provider) type is the root of your provider implementation. From there, [`tfsdk.ResourceType`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/tfsdk#ResourceType) and [`tfsdk.DataSourceType`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/tfsdk#DataSourceType) implementations define the [schema](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/schema#Schema) of your resources and data sources, and how to create [`tfsdk.Resource`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/tfsdk#Resource) and [`tfsdk.DataSource`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework/tfsdk#DataSource) implementations that talk to the API.

## Contributing

See [`.github/CONTRIBUTING.md`](https://github.com/hashicorp/terraform-plugin-framework/blob/main/.github/CONTRIBUTING.md)

## License

[Mozilla Public License v2.0](https://github.com/hashicorp/terraform-plugin-framework/blob/main/LICENSE)
