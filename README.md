[![PkgGoDev](https://pkg.go.dev/badge/github.com/hashicorp/terraform-plugin-framework)](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-framework)

# Terraform Plugin Framework

terraform-plugin-framework is an experimental module with no backwards compatibility guarantees, as indicated by its version 0 SemVer releases. We are mindful of our consumers and will strive to either maintain backwards compatibility or offer low-effort upgrade paths to provider developers, but our compatibility guarantees on this module are weaker than the [terraform-plugin-sdk](https://github.com/hashicorp/terraform-plugin-sdk) module. The purpose of this repository is to get real world experience with the design and implementation of this code before it is bound by the SDK's compatibility guarantees. This module is intentionally limited in its lifespan, and will become the next major version of the terraform-plugin-sdk module once were confident in its design and implementation, and this repository will be archived. When this happens, we will provide an upgrade path from users of the terraform-plugin-framework repository to the terraform-plugin-sdk release.

We recommend only using tagged releases of this module, and examining the CHANGELOG when upgrading to a new release. Breaking changes will only be made in minor versions; patch releases will always maintain backwards compatibility.

We welcome and appreciate issues and PRs discussing both the design and implementation of this module.

## Terraform CLI Compatibility

Plugins built with this framework are only compatible with Terraform versions above 0.12.0. Terraform 0.12.26 or higher is needed to run acceptance tests for providers built with this framework.

## Go Compatibility

Prior to its 1.0 release, this module will only support the latest released version of Go, and may use features and functionality introduced in that version of Go.

Currently that means Go **1.16** must be used when building a provider with this framework.

## Contributing

See [`.github/CONTRIBUTING.md`](https://github.com/hashicorp/terraform-plugin-framework/blob/main/.github/CONTRIBUTING.md)

## License

[Mozilla Public License v2.0](https://github.com/hashicorp/terraform-plugin-framework/blob/main/LICENSE)
