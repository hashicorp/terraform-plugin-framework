# 0.3.0 (September 08, 2021)

BREAKING CHANGES:
* Methods on the `tfsdk.Config`, `tfsdk.Plan`, and `tfsdk.State` types now return `[]*tfprotov6.Diagnostic` instead of `error` ([#82](https://github.com/hashicorp/terraform-plugin-framework/issues/82))
* Most uses of `[]*tfprotov6.Diagnostic` have been replaced with a new `diag.Diagnostics` type. Please update your type signatures, and use one of the `diags.New*` helper functions instead of constructing `*tfprotov6.Diagnostic`s by hand. ([#110](https://github.com/hashicorp/terraform-plugin-framework/issues/110))
* The `schema.Attribute` and `schema.Schema` types have been moved to `tfsdk.Attribute` and `tfsdk.Schema`. No changes beyond import names are required. ([#77](https://github.com/hashicorp/terraform-plugin-framework/issues/77))
* With the release of Go 1.17, Go 1.17 is now the lowest supported version of Go to use with terraform-plugin-framework. ([#104](https://github.com/hashicorp/terraform-plugin-framework/issues/104))
* `attr.Value` implementations must now implement a `Type(context.Context)` method that returns the `attr.Type` that created the `attr.Value`. ([#119](https://github.com/hashicorp/terraform-plugin-framework/issues/119))

FEATURES:
* Added support for ModifyPlan functions on Resources. ([#90](https://github.com/hashicorp/terraform-plugin-framework/issues/90))
* Introduced first-class diagnostics (`diag` package). ([#110](https://github.com/hashicorp/terraform-plugin-framework/issues/110))
* Support `attr.Type` validation ([#82](https://github.com/hashicorp/terraform-plugin-framework/issues/82))
* tfsdk: Attributes, Data Sources, Providers, and Resources now support configuration validation ([#75](https://github.com/hashicorp/terraform-plugin-framework/issues/75))

ENHANCEMENTS:
* Added a `tfsdk.ValueAs` helper that allows accessing an `attr.Value` without type assertion, by using the same reflection rules used in the `Config.Get`, `Plan.Get`, and `State.Get` helpers. ([#119](https://github.com/hashicorp/terraform-plugin-framework/issues/119))
* Errors from methods on the `tfsdk.Config`, `tfsdk.Plan`, and `tfsdk.State` types now include rich diagnostic information ([#82](https://github.com/hashicorp/terraform-plugin-framework/issues/82))
* tfsdk: Validate `Attribute` defines at least one of `Required`, `Optional`, or `Computed` ([#111](https://github.com/hashicorp/terraform-plugin-framework/issues/111))

BUG FIXES:
* tfsdk: Diagnostics returned from `(Plan).SetAttribute()` and `(State).SetAttribute()` reflection will now properly include attribute path ([#133](https://github.com/hashicorp/terraform-plugin-framework/issues/133))
* tfsdk: Don't attempt validation on the nested attributes of a null or unknown `SingleNestedAttribute` ([#118](https://github.com/hashicorp/terraform-plugin-framework/issues/118))
* tfsdk: Return warning diagnostic when using `Attribute` or `Schema` type `DeprecationMessage` field ([#93](https://github.com/hashicorp/terraform-plugin-framework/issues/93))

# 0.2.0 (July 22, 2021)

ENHANCEMENTS:
* Added `tfsdk.NewProtocol6Server` to return a `tfprotov6.ProviderServer` implementation for testing and muxing purposes. ([#72](https://github.com/hashicorp/terraform-plugin-framework/issues/72))
* Added support for MapNestedAttributes. ([#79](https://github.com/hashicorp/terraform-plugin-framework/issues/79))
* Responses now default to returning the current state, meaning state will only change when provider developers actively change it. Previously, an empty state value would be returned, which caused problems. ([#74](https://github.com/hashicorp/terraform-plugin-framework/issues/74))

# 0.1.0 (June 24, 2021)

FEATURES:

* Added interfaces extending the attr.Type interface to include attribute and element types. ([#44](https://github.com/hashicorp/terraform-plugin-framework/issues/44))
* Added state, config, and plan types, and support for getting values from them. ([#46](https://github.com/hashicorp/terraform-plugin-framework/issues/46))
* Added support for Object types. ([#38](https://github.com/hashicorp/terraform-plugin-framework/issues/38))
* Added support for bools, numbers, and strings. ([#29](https://github.com/hashicorp/terraform-plugin-framework/issues/29))
* Added support for defining schemas and attributes. ([#27](https://github.com/hashicorp/terraform-plugin-framework/issues/27))
* Added support for lists. ([#36](https://github.com/hashicorp/terraform-plugin-framework/issues/36))
* Added support for maps. ([#37](https://github.com/hashicorp/terraform-plugin-framework/issues/37))
* Added support for provider, resource, and data source types. ([#32](https://github.com/hashicorp/terraform-plugin-framework/issues/32))
* Added the ability to serve providers. ([#45](https://github.com/hashicorp/terraform-plugin-framework/issues/45))
