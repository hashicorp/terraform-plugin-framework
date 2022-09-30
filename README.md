[![PkgGoDev](https://pkg.go.dev/badge/github.com/hashicorp/terraform-plugin-framework)](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework)

# Terraform Plugin Framework

terraform-plugin-framework is a module for building [Terraform providers](https://www.terraform.io/language/providers). It is built on [terraform-plugin-go](https://github.com/hashicorp/terraform-plugin-go). It aims to provide as much of the power, predictability, and versatility of terraform-plugin-go as it can while abstracting away implementation details and repetitive, verbose tasks.

## Getting Started

Developers can get started to build the providers using our new [HashiCorp Learn collections](https://developer.hashicorp.com/terraform/tutorials/providers/plugin-framework-create) or upgrade their existing provider using our [migration guide](https://www.terraform.io/plugin/framework/migrating). 

Learn more about [Terraform Plugin Framework](https://www.terraform.io/plugin/framework).

## Status

terraform-plugin-framework has reached **Public Beta** phase. We are committed to moving forward with the module, but cannot guarantee any of its interfaces will not change as long as it is in version 0. We're waiting for additional feedback, usage, and maturity before we're comfortable committing to APIs with the same years-long support timelines that [terraform-plugin-sdk](https://github.com/hashicorp/terraform-plugin-sdk) brings. We do not expect practitioner experiences to break or change as a result of these changes, only the abstractions surfaced to provider developers.

Refer to [Which SDK Should I Use?](https://terraform.io/docs/plugin/which-sdk.html) for more information.

We believe terraform-plugin-framework is a suitable and reliable module to build Terraform providers on, and encourage community members that can afford occasional breaking changes to build with it. terraform-plugin-framework will eventually become generally available with a new major version release, at which point its interfaces will be stable, but we need real-world use and feedback before we can be comfortable making those commitments. 

We recommend only using tagged releases of this module, and examining the CHANGELOG when upgrading to a new release. Breaking changes will only be made in minor versions; patch releases will always maintain backwards compatibility.

We welcome and appreciate issues and PRs discussing both the design and implementation of this module.

## Terraform CLI Compatibility

Providers built with this framework are compatible with Terraform version v0.12 and above.

## Go Compatibility

This project follows the [support policy](https://golang.org/doc/devel/release.html#policy) of Go as its support policy. The two latest major releases of Go are supported by the project.

Currently, that means Go **1.18** or later must be used when including this project as a dependency.

## Contributing

See [`.github/CONTRIBUTING.md`](https://github.com/hashicorp/terraform-plugin-framework/blob/main/.github/CONTRIBUTING.md)

## License

[Mozilla Public License v2.0](https://github.com/hashicorp/terraform-plugin-framework/blob/main/LICENSE)
