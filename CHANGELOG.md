# 0.7.0 (April 28, 2022)

NOTES:

* tfsdk: Providers may now optionally remove `RemoveResource()` calls from `Resource` type `Delete` methods ([#301](https://github.com/hashicorp/terraform-plugin-framework/issues/301))
* tfsdk: The `NewProtocol6Server()` function has been deprecated in preference of `providerserver.NewProtocol6()` and `providerserver.NewProtocol6WithError()` functions, which will simplify muxing and testing implementations. The `tfsdk.NewProtocol6Server()` function will be removed in the next minor version. ([#308](https://github.com/hashicorp/terraform-plugin-framework/issues/308))
* tfsdk: The `ResourceImportStateNotImplemented()` function has been deprecated. Instead, the `ImportState` method can be removed from the `Resource` and the framework will automatically return an error diagnostic if import is attempted. ([#297](https://github.com/hashicorp/terraform-plugin-framework/issues/297))
* tfsdk: The `Resource` interface no longer requires the `ImportState` method. A separate `ResourceWithImportState` interface now defines the same `ImportState` method. ([#297](https://github.com/hashicorp/terraform-plugin-framework/issues/297))
* tfsdk: The `Serve()` function has been deprecated in preference of the `providerserver.Serve()` function. The `tfsdk.Serve()` function will be removed in the next minor version. ([#308](https://github.com/hashicorp/terraform-plugin-framework/issues/308))
* tfsdk: The `ServeOpts` type has been deprecated in preference of the `providerserver.ServeOpts` type. When migrating, the `Name` field has been replaced with `Address`. The `tfsdk.ServeOpts` type will be removed in the next minor version. ([#308](https://github.com/hashicorp/terraform-plugin-framework/issues/308))
* tfsdk: The previously unexported `server` type has been temporarily exported to aid in the migration to the new `providerserver` package. It is not intended for provider developer usage and will be moved into an internal package in the next minor version. ([#308](https://github.com/hashicorp/terraform-plugin-framework/issues/308))

FEATURES:

* Introduced `providerserver` package, which contains all functions and types necessary for serving a provider in production or acceptance testing. ([#308](https://github.com/hashicorp/terraform-plugin-framework/issues/308))
* tfsdk: Added optional `ResourceWithUpgradeState` interface, which allows for provider defined logic when the `UpgradeResourceState` RPC is called ([#292](https://github.com/hashicorp/terraform-plugin-framework/issues/292))

ENHANCEMENTS:

* tfsdk: Added `DEBUG` level logging for all framework handoffs to provider defined logic ([#300](https://github.com/hashicorp/terraform-plugin-framework/issues/300))
* tfsdk: Added `ResourceWithImportState` interface, which allows `Resource` implementations to optionally define the `ImportState` method. ([#297](https://github.com/hashicorp/terraform-plugin-framework/issues/297))
* tfsdk: Added automatic `(DeleteResourceResponse.State).RemoveResource()` call after `Resource` type `Delete` method execution if there are no errors ([#301](https://github.com/hashicorp/terraform-plugin-framework/issues/301))

# 0.6.1 (March 29, 2022)

BUG FIXES:

* types: Prevented panics with missing type information during `Float64`, `Int64`, and `Set` validation logic ([#259](https://github.com/hashicorp/terraform-plugin-framework/issues/259))

# 0.6.0 (March 10, 2022)

NOTES:

* The underlying `terraform-plugin-log` dependency has been updated to v0.3.0, which includes a breaking change in the optional additional fields parameter of logging function calls to ensure correctness and catch coding errors during compilation. Any early adopter provider logging which calls those functions may require updates. ([#268](https://github.com/hashicorp/terraform-plugin-framework/issues/268))

BREAKING CHANGES:

* The `ToTerraformValue` method of the `attr.Value` interface now returns a `tftypes.Value`, instead of an `interface{}`. Existing types need to be updated to call `tftypes.ValidateValue` and `tftypes.NewValue`, passing the value they were returning before, instead of returning the value directly. ([#231](https://github.com/hashicorp/terraform-plugin-framework/issues/231))
* tfsdk: The `ListNestedAttributesOptions`, `MapNestedAttributeOptions`, and `SetNestedAttributeOptions` type `MaxItems` and `MinItems` fields have been removed since the protocol and framework never supported this type of nested attribute validation. Use attribute validators instead. ([#249](https://github.com/hashicorp/terraform-plugin-framework/issues/249))

ENHANCEMENTS:

* Added the ability to get an attribute as a generic `attr.Value` when using `GetAttribute`. ([#232](https://github.com/hashicorp/terraform-plugin-framework/issues/232))
* Logging can now be used by calling `tflog.Trace`, `tflog.Debug`, `tflog.Info`, `tflog.Warn`, or `tflog.Error`. See [the tflog docs](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-log/tflog) for more information. ([#234](https://github.com/hashicorp/terraform-plugin-framework/issues/234))
* tfsdk: Added `Debug` field to `ServeOpts` for running providers via debugger and testing processes ([#243](https://github.com/hashicorp/terraform-plugin-framework/issues/243))

BUG FIXES:

* tfsdk: Removed `Schema` restriction that it must contain at least one attribute or block ([#252](https://github.com/hashicorp/terraform-plugin-framework/issues/252))
* tfsdk: Support protocol version 5 and verify valid resource type in `UpgradeResourceState` RPC ([#263](https://github.com/hashicorp/terraform-plugin-framework/issues/263))

# 0.5.0 (November 30, 2021)

BREAKING CHANGES:

* Fixed RequiresReplace and RequiresReplaceIf to be more judicious about when they require a resource to be destroyed and recreated. They will no longer require resources to be recreated when _any_ attribute changes, instead limiting it only to the attribute they're declared on. They will also not require resources to be recreated when they're being created or deleted. Finally, they won't require a resource to be recreated if the user has no value in the config for the attribute and the attribute is computed; this is to prevent the resource from being destroyed and recreated when the provider changes the value without any user prompting. Providers that wish to destroy and recreate the resource when an optional and computed attribute is removed from the user's config should do so in their own plan modifier. ([#213](https://github.com/hashicorp/terraform-plugin-framework/issues/213))
* RequiresReplaceIf no longer overrides previous plan modifiers' value for RequiresReplace if the function returns false. ([#213](https://github.com/hashicorp/terraform-plugin-framework/issues/213))
* diag: The `AttributeErrorDiagnostic` and `AttributeWarningDiagnostic` types have been removed. Any usage can be replaced with `DiagnosticWithPath`. ([#219](https://github.com/hashicorp/terraform-plugin-framework/issues/219))
* tfsdk: The `AddAttributeError`, `AddAttributeWarning`, `AddError`, and `AddWarning` methods on the `ConfigureProviderResponse`, `CreateResourceResponse`, `DeleteResourceResponse`, `ModifyAttributePlanResponse`, `ModifyResourcePlanResponse`, `ReadDataSourceResponse`, `ReadResourceResponse`, and `UpdateResourceResponse` types have been removed in preference of the same methods on the `Diagnostics` field of these types. For example, code such as `resp.AddError("...", "...")` can be updated to `resp.Diagnostics.AddError("...", "...")`. ([#198](https://github.com/hashicorp/terraform-plugin-framework/issues/198))
* tfsdk: The `Config`, `Plan`, and `State` type `GetAttribute` methods now return diagnostics only and require the target as the last parameter, similar to the `Get` method. ([#167](https://github.com/hashicorp/terraform-plugin-framework/issues/167))

FEATURES:

* Added `tfsdk.UseStateForUnknown()` as a built-in plan modifier, which will automatically replace an unknown value in the plan with the value from the state. This mimics the behavior of computed and optional+computed values in Terraform Plugin SDK versions 1 and 2. Provider developers will likely want to use it for "write-once" attributes that never change once they're set in state. ([#204](https://github.com/hashicorp/terraform-plugin-framework/issues/204))
* tfsdk: Support list and set blocks in schema definitions ([#188](https://github.com/hashicorp/terraform-plugin-framework/issues/188))

ENHANCEMENTS:

* diag: Added `WithPath()` function to wrap or overwrite diagnostic path information. ([#219](https://github.com/hashicorp/terraform-plugin-framework/issues/219))
* tfsdk: The `Config`, `Plan`, and `State` type `GetAttribute` methods can now be used to fetch values directly into `attr.Value` implementations and Go types. ([#167](https://github.com/hashicorp/terraform-plugin-framework/issues/167))

BUG FIXES:

* tfsdk: Fetch null values from valid missing `Config`, `Plan`, and `State` paths in `GetAttribute()` method ([#185](https://github.com/hashicorp/terraform-plugin-framework/issues/185))
* types: Ensure `Float64` `Type()` method returns `Float64Type` ([#202](https://github.com/hashicorp/terraform-plugin-framework/issues/202))
* types: Prevent panic with uninitialized `Number` `Value` ([#200](https://github.com/hashicorp/terraform-plugin-framework/issues/200))
* types: Prevent panics when `ValueFromTerraform` received `nil` values ([#208](https://github.com/hashicorp/terraform-plugin-framework/issues/208))

# 0.4.2 (September 29, 2021)

BUG FIXES:
* Fix bug in which updating `Computed`-only attributes would lead to a "Provider produced inconsistent result after apply" error ([#176](https://github.com/hashicorp/terraform-plugin-framework/issues/176)/[#184](https://github.com/hashicorp/terraform-plugin-framework/issues/184))

# 0.4.1 (September 27, 2021)

NOTES:
* Upgraded to terraform-plugin-go v0.4.0 which contains its own breaking changes. Please see https://github.com/hashicorp/terraform-plugin-go/blob/main/CHANGELOG.md#040-september-24-2021 for more details. ([#179](https://github.com/hashicorp/terraform-plugin-framework/issues/179))

# 0.4.0 (September 24, 2021)

BREAKING CHANGES:
* `attr.Type` implementations must now have a `String()` method that returns a human-friendly name for the type. ([#120](https://github.com/hashicorp/terraform-plugin-framework/issues/120))
* tfsdk: `Resource` implementations must now include the `ImportState(context.Context, ImportResourceStateRequest, *ImportResourceStateResponse)` method. If import is not supported, call the `ResourceImportStateNotImplemented()` function or return an error. ([#149](https://github.com/hashicorp/terraform-plugin-framework/issues/149))

FEATURES:
* tfsdk: Support resource import ([#149](https://github.com/hashicorp/terraform-plugin-framework/issues/149))
* types: Support `Set` and `SetType` ([#126](https://github.com/hashicorp/terraform-plugin-framework/issues/126))
* types: Support for `Float64`, `Float64Type`, `Int64`, and `Int64Type` ([#166](https://github.com/hashicorp/terraform-plugin-framework/issues/166))

ENHANCEMENTS:
* Added a `tfsdk.ConvertValue` helper that will convert any `attr.Value` into any compatible `attr.Type`. Compatibility happens at the terraform-plugin-go level; the type that the `attr.Value`'s `ToTerraformValue` method produces must be compatible with the `attr.Type`'s `TerraformType()`. Generally, this means that the `attr.Type` of the `attr.Value` and the `attr.Type` being converted to must both produce the same `tftypes.Type` when their `TerraformType()` method is called. ([#120](https://github.com/hashicorp/terraform-plugin-framework/issues/120))

BUG FIXES:
* attr: Ensure `List` types implementing `attr.TypeWithValidate` call `ElementType` validation only if that type implements `attr.TypeWithValidate` ([#126](https://github.com/hashicorp/terraform-plugin-framework/issues/126))
* tfsdk: `(Plan).SetAttribute()` and `(State).SetAttribute()` will now create missing attribute paths instead of silently failing to update. ([#165](https://github.com/hashicorp/terraform-plugin-framework/issues/165))

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
