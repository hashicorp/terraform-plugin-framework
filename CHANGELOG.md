## 1.12.0 (September 18, 2024)

NOTES:

* all: This Go module has been updated to Go 1.22 per the [Go support policy](https://go.dev/doc/devel/release#policy). It is recommended to review the [Go 1.22 release notes](https://go.dev/doc/go1.22) before upgrading. Any consumers building on earlier Go versions may experience errors ([#1033](https://github.com/hashicorp/terraform-plugin-framework/issues/1033))

BUG FIXES:

* providerserver: Fixed bug that prevented `moved` operation support between resource types for framework-only providers. ([#1039](https://github.com/hashicorp/terraform-plugin-framework/issues/1039))

## 1.11.0 (August 06, 2024)

NOTES:

* Framework reflection logic (`Config.Get`, `Plan.Get`, etc.) for structs with
`tfsdk` field tags has been updated to support embedded structs that promote exported
fields. For existing structs that embed unexported structs with exported fields, a tfsdk
ignore tag (``tfsdk:"-"``) can be added to ignore all promoted fields.  

For example, the following struct will now return an error diagnostic:
```go
type thingResourceModel struct {
	Attr1 types.String `tfsdk:"attr_1"`
	Attr2 types.Bool   `tfsdk:"attr_2"`

	// Previously, this embedded struct was ignored, will now promote underlying fields
	embeddedModel
}

type embeddedModel struct {
	// No `tfsdk` tag
	ExportedField string
}
```

To preserve the original behavior, a tfsdk ignore tag can be added to ignore the entire embedded struct:
```go
type thingResourceModel struct {
	Attr1 types.String `tfsdk:"attr_1"`
	Attr2 types.Bool   `tfsdk:"attr_2"`

	// This embedded struct will now be ignored
	embeddedModel      `tfsdk:"-"`
}

type embeddedModel struct {
	ExportedField string
}
```
 ([#1021](https://github.com/hashicorp/terraform-plugin-framework/issues/1021))

ENHANCEMENTS:

* all: Added embedded struct support for object to struct conversions with `tfsdk` tags ([#1021](https://github.com/hashicorp/terraform-plugin-framework/issues/1021))

## 1.10.0 (July 09, 2024)

FEATURES:

* types/basetypes: Added `Int32Type` and `Int32Value` implementations for Int32 value handling. ([#1010](https://github.com/hashicorp/terraform-plugin-framework/issues/1010))
* types/basetypes: Added interfaces `basetypes.Int32Typable`, `basetypes.Int32Valuable`, and `basetypes.Int32ValuableWithSemanticEquals` for Int32 custom type and value implementations. ([#1010](https://github.com/hashicorp/terraform-plugin-framework/issues/1010))
* resource/schema: Added `Int32Attribute` implementation for Int32 value handling. ([#1010](https://github.com/hashicorp/terraform-plugin-framework/issues/1010))
* datasource/schema: Added `Int32Attribute` implementation for Int32 value handling. ([#1010](https://github.com/hashicorp/terraform-plugin-framework/issues/1010))
* provider/schema: Added `Int32Attribute` implementation for Int32 value handling. ([#1010](https://github.com/hashicorp/terraform-plugin-framework/issues/1010))
* function: Added `Int32Parameter` and `Int32Return` for Int32 value handling. ([#1010](https://github.com/hashicorp/terraform-plugin-framework/issues/1010))
* resource/schema/int32default: New package with `StaticValue` implementation for Int32 schema-based default values. ([#1010](https://github.com/hashicorp/terraform-plugin-framework/issues/1010))
* resource/schema/int32planmodifier: New package with built-in implementations for Int32 value plan modification. ([#1010](https://github.com/hashicorp/terraform-plugin-framework/issues/1010))
* resource/schema/defaults: New `Int32` interface for Int32 schema-based default implementations. ([#1010](https://github.com/hashicorp/terraform-plugin-framework/issues/1010))
* resource/schema/planmodifier: New `Int32` interface for Int32 value plan modification implementations. ([#1010](https://github.com/hashicorp/terraform-plugin-framework/issues/1010))
* schema/validator: New `Int32` interface for Int32 value schema validation. ([#1010](https://github.com/hashicorp/terraform-plugin-framework/issues/1010))
* types/basetypes: Added `Float32Type` and `Float32Value` implementations for Float32 value handling. ([#1014](https://github.com/hashicorp/terraform-plugin-framework/issues/1014))
* types/basetypes: Added interfaces `basetypes.Float32Typable`, `basetypes.Float32Valuable`, and `basetypes.Float32ValuableWithSemanticEquals` for Float32 custom type and value implementations. ([#1014](https://github.com/hashicorp/terraform-plugin-framework/issues/1014))
* resource/schema: Added `Float32Attribute` implementation for Float32 value handling. ([#1014](https://github.com/hashicorp/terraform-plugin-framework/issues/1014))
* datasource/schema: Added `Float32Attribute` implementation for Float32 value handling. ([#1014](https://github.com/hashicorp/terraform-plugin-framework/issues/1014))
* provider/schema: Added `Float32Attribute` implementation for Float32 value handling. ([#1014](https://github.com/hashicorp/terraform-plugin-framework/issues/1014))
* function: Added `Float32Parameter` and `Float32Return` for Float32 value handling. ([#1014](https://github.com/hashicorp/terraform-plugin-framework/issues/1014))
* resource/schema/float32default: New package with `StaticValue` implementation for Float32 schema-based default values. ([#1014](https://github.com/hashicorp/terraform-plugin-framework/issues/1014))
* resource/schema/float32planmodifier: New package with built-in implementations for Float32 value plan modification. ([#1014](https://github.com/hashicorp/terraform-plugin-framework/issues/1014))
* resource/schema/defaults: New `Float32` interface for Float32 schema-based default implementations. ([#1014](https://github.com/hashicorp/terraform-plugin-framework/issues/1014))
* resource/schema/planmodifier: New `Float32` interface for Float32 value plan modification implementations. ([#1014](https://github.com/hashicorp/terraform-plugin-framework/issues/1014))
* schema/validator: New `Float32` interface for Float32 value schema validation. ([#1014](https://github.com/hashicorp/terraform-plugin-framework/issues/1014))

## 1.9.0 (June 04, 2024)

NOTES:

* resource: If plan modification was dependent on nested attribute plan modification automatically running when the nested object was null/unknown, it may be necessary to add object-level plan modification to convert the nested object to a known object first. ([#995](https://github.com/hashicorp/terraform-plugin-framework/issues/995))
* This release contains support for deferred actions, which is an experimental feature only available in prerelease builds of Terraform 1.9 and later. This functionality is subject to change and is not protected by version compatibility guarantees. ([#999](https://github.com/hashicorp/terraform-plugin-framework/issues/999))

FEATURES:

* resource: Add `Deferred` field to `ReadResponse`, `ModifyPlanResponse`, and `ImportStateResponse` which indicates a resource deferred action to the Terraform client ([#999](https://github.com/hashicorp/terraform-plugin-framework/issues/999))
* datasource: Add `Deferred` field to `ReadResponse` which indicates a data source deferred action to the Terraform client ([#999](https://github.com/hashicorp/terraform-plugin-framework/issues/999))
* resource: Add `ClientCapabilities` field to `ReadRequest`, `ModifyPlanRequest`, and `ImportStateRequest` which specifies optionally supported protocol features for the Terraform client ([#999](https://github.com/hashicorp/terraform-plugin-framework/issues/999))
* datasource: Add `ClientCapabilities` field to `ReadRequest` which specifies optionally supported protocol features for the Terraform client ([#999](https://github.com/hashicorp/terraform-plugin-framework/issues/999))
* provider: Add `Deferred` field to `ConfigureResponse` which indicates a provider deferred action to the Terraform client ([#1002](https://github.com/hashicorp/terraform-plugin-framework/issues/1002))
* provider: Add `ClientCapabilities` field to `ConfigureRequest` which specifies optionally supported protocol features for the Terraform client ([#1002](https://github.com/hashicorp/terraform-plugin-framework/issues/1002))

ENHANCEMENTS:

* function: Introduced implementation errors for collection and object parameters and returns which are missing type information ([#991](https://github.com/hashicorp/terraform-plugin-framework/issues/991))

BUG FIXES:

* resource: Prevented errant collection-based nested object conversion from null/unknown object to known object ([#995](https://github.com/hashicorp/terraform-plugin-framework/issues/995))

## 1.8.0 (April 18, 2024)

BREAKING CHANGES:

* function: Removed `Definition` type `Parameter()` method ([#968](https://github.com/hashicorp/terraform-plugin-framework/issues/968))

NOTES:

* function: Provider-defined function features are now considered generally available and protected by compatibility promises ([#966](https://github.com/hashicorp/terraform-plugin-framework/issues/966))
* attr/xattr: The `TypeWithValidate` interface has been deprecated in preference of the `ValidateableAttribute` interface. A `ValidatableParameter` interface has also been added to the `function` package ([#968](https://github.com/hashicorp/terraform-plugin-framework/issues/968))

FEATURES:

* attr/xattr: Added `ValidateableAttribute` interface for custom value type implementations ([#968](https://github.com/hashicorp/terraform-plugin-framework/issues/968))
* function: Added `ValidateableParameter` interface for custom value type implementations ([#968](https://github.com/hashicorp/terraform-plugin-framework/issues/968))
* `function`: Add `BoolParameterValidator`, `DynamicParameterValidator`, `Float64ParameterValidator`, `Int64ParameterValidator`, `ListParameterValidator`, `MapParameterValidator`, `NumberParameterValidator`, `ObjectParameterValidator`, `SetParameterValidator`, and `StringParameterValidator` interfaces for custom function parameter validation implementations. ([#971](https://github.com/hashicorp/terraform-plugin-framework/issues/971))
* `function`: Add `ParameterWithBoolValidators`, `ParameterWithInt64Validators`, `ParameterWithFloat64Validators`, `ParameterWithDynamicValidators`, `ParameterWithListValidators`, `ParameterWithMapValidators`, `ParameterWithNumberValidators`, `ParameterWithObjectValidators`, `ParameterWithSetValidators`, and `ParameterWithStringValidators` interfaces to enable parameter-based validation support ([#971](https://github.com/hashicorp/terraform-plugin-framework/issues/971))

BUG FIXES:

* types/basetypes: Prevented panic in the `MapValue` types `Equal` method when the receiver has a nil `elementType` ([#961](https://github.com/hashicorp/terraform-plugin-framework/issues/961))
* types/basetypes: Prevented panic in the `ListValue` types `Equal` method when the receiver has a nil `elementType` ([#961](https://github.com/hashicorp/terraform-plugin-framework/issues/961))
* types/basetypes: Prevented panic in the `SetValue` types `Equal` method when the receiver has a nil `elementType` ([#961](https://github.com/hashicorp/terraform-plugin-framework/issues/961))
* resource: Ensured computed-only dynamic attributes will not cause `wrong final value type` errors during planning ([#969](https://github.com/hashicorp/terraform-plugin-framework/issues/969))

## 1.7.0 (March 21, 2024)

BREAKING CHANGES:

* function: All parameters must be explicitly named via the `Name` field ([#964](https://github.com/hashicorp/terraform-plugin-framework/issues/964))
* function: `DefaultParameterNamePrefix` and `DefaultVariadicParameterName` constants have been removed ([#964](https://github.com/hashicorp/terraform-plugin-framework/issues/964))

FEATURES:

* types/basetypes: Added `DynamicType` and `DynamicValue` implementations for dynamic value handling ([#147](https://github.com/hashicorp/terraform-plugin-framework/issues/147))
* types/basetypes: Added interfaces `basetypes.DynamicTypable`, `basetypes.DynamicValuable`, and `basetypes.DynamicValuableWithSemanticEquals` for dynamic custom type and value implementations ([#147](https://github.com/hashicorp/terraform-plugin-framework/issues/147))
* resource/schema: Added `DynamicAttribute` implementation for dynamic value handling ([#147](https://github.com/hashicorp/terraform-plugin-framework/issues/147))
* datasource/schema: Added `DynamicAttribute` implementation for dynamic value handling ([#147](https://github.com/hashicorp/terraform-plugin-framework/issues/147))
* provider/schema: Added `DynamicAttribute` implementation for dynamic value handling ([#147](https://github.com/hashicorp/terraform-plugin-framework/issues/147))
* function: Added `DynamicParameter` and `DynamicReturn` for dynamic value handling` ([#147](https://github.com/hashicorp/terraform-plugin-framework/issues/147))
* resource/schema/dynamicdefault: New package with `StaticValue` implementation for dynamic schema-based default values ([#147](https://github.com/hashicorp/terraform-plugin-framework/issues/147))
* resource/schema/dynamicplanmodifier: New package with built-in implementations for dynamic value plan modification. ([#147](https://github.com/hashicorp/terraform-plugin-framework/issues/147))
* resource/schema/defaults: New `Dynamic` interface for dynamic schema-based default implementations ([#147](https://github.com/hashicorp/terraform-plugin-framework/issues/147))
* resource/schema/planmodifier: New `Dynamic` interface for dynamic value plan modification implementations ([#147](https://github.com/hashicorp/terraform-plugin-framework/issues/147))
* schema/validator: New `Dynamic` interface for dynamic value schema validation ([#147](https://github.com/hashicorp/terraform-plugin-framework/issues/147))

## 1.6.1 (March 05, 2024)

NOTES:

* all: The `v1.6.0` release updated this Go module to Go 1.21 per the [Go support policy](https://go.dev/doc/devel/release#policy). It is recommended to review the [Go 1.21 release notes](https://go.dev/doc/go1.21) before upgrading. Any consumers building on earlier Go versions may experience errors ([#937](https://github.com/hashicorp/terraform-plugin-framework/issues/937))

BUG FIXES:

* resource/schema: Ensured invalid attribute default value errors are raised ([#930](https://github.com/hashicorp/terraform-plugin-framework/issues/930))
* function: Added implementation validation to `function.Definition` to ensure all parameter names (including the variadic parameter) are unique. ([#926](https://github.com/hashicorp/terraform-plugin-framework/issues/926))
* function: Updated the default parameter name to include the position of the parameter (i.e. `param1`, `param2`, etc.). Variadic parameters will default to `varparam`. ([#926](https://github.com/hashicorp/terraform-plugin-framework/issues/926))

## 1.6.0 (February 28, 2024)

BREAKING CHANGES:

* function: Changed the framework type for variadic parameters to `types.TupleType`, where each element is the same element type. Provider-defined functions using a `types.List` for retrieving variadic argument data will need to update their code to use `types.Tuple`. ([#923](https://github.com/hashicorp/terraform-plugin-framework/issues/923))
* function: Altered the `RunResponse` type, replacing `Diagnostics` with `FuncError` ([#925](https://github.com/hashicorp/terraform-plugin-framework/issues/925))
* diag: Removed `DiagnosticWithFunctionArgument` interface. Removed `NewArgumentErrorDiagnostic()`, `NewArgumentWarningDiagnostic()` and `WithFunctionArgument()` functions. Removed `AddArgumentError()` and `AddArgumentWarning()` methods from `Diagnostics`. ([#925](https://github.com/hashicorp/terraform-plugin-framework/issues/925))

FEATURES:

* resource: Added the `ResourceWithMoveState` interface, which enables state moves across resource types with Terraform 1.8 and later ([#917](https://github.com/hashicorp/terraform-plugin-framework/issues/917))

ENHANCEMENTS:

* privatestate: Added support for `SetKey()` method to fully remove key with `nil` or zero-length value ([#910](https://github.com/hashicorp/terraform-plugin-framework/issues/910))
* function: Added `FuncError` type, required for `RunResponse` ([#925](https://github.com/hashicorp/terraform-plugin-framework/issues/925))
* function: Added `NewFuncError()` and `NewArgumentFuncError()` functions, which create a `FuncError` ([#925](https://github.com/hashicorp/terraform-plugin-framework/issues/925))
* function: Added `ConcatFuncErrors()` and `FuncErrorFromDiags()` helper functions for use when working with `FuncError` ([#925](https://github.com/hashicorp/terraform-plugin-framework/issues/925))

## 1.5.0 (January 11, 2024)

NOTES:

* all: Update `google.golang.org/grpc` dependency to address CVE-2023-44487 ([#865](https://github.com/hashicorp/terraform-plugin-framework/issues/865))
* Provider-defined function support is in technical preview and offered without compatibility promises until Terraform 1.8 is generally available. ([#889](https://github.com/hashicorp/terraform-plugin-framework/issues/889))

FEATURES:

* function: New package for implementing provider defined functions ([#889](https://github.com/hashicorp/terraform-plugin-framework/issues/889))

ENHANCEMENTS:

* types/basetypes: Added `TupleType` and `TupleValue` implementations, which are only necessary for dynamic value handling ([#870](https://github.com/hashicorp/terraform-plugin-framework/issues/870))
* diag: Added `NewArgumentErrorDiagnostic()` and `NewArgumentWarningDiagnostic()` functions, which create diagnostics with the function argument position set ([#889](https://github.com/hashicorp/terraform-plugin-framework/issues/889))
* provider: Added `ProviderWithFunctions` interface for implementing provider defined functions ([#889](https://github.com/hashicorp/terraform-plugin-framework/issues/889))
* diag: Added `(Diagnostics).AddArgumentError()` and `(Diagnostics).AddArgumentWarning()` methods for appending function argument diagnostics ([#889](https://github.com/hashicorp/terraform-plugin-framework/issues/889))

## 1.4.2 (October 24, 2023)

BUG FIXES:

* resource: Add `Private` field to `DeleteResource` type, which was missing to allow provider logic to update private state on errors ([#863](https://github.com/hashicorp/terraform-plugin-framework/issues/863))
* resource: Prevented private state data loss if resource destruction returned an error ([#863](https://github.com/hashicorp/terraform-plugin-framework/issues/863))

## 1.4.1 (October 09, 2023)

BUG FIXES:

* providerserver: Prevented `Data Source Type Not Found` and `Resource Type Not Found` errors with Terraform 1.6 and later ([#853](https://github.com/hashicorp/terraform-plugin-framework/issues/853))

## 1.4.0 (September 06, 2023)

NOTES:

* all: This Go module has been updated to Go 1.20 per the [Go support policy](https://go.dev/doc/devel/release#policy). It is recommended to review the [Go 1.20 release notes](https://go.dev/doc/go1.20) before upgrading. Any consumers building on earlier Go versions may experience errors. ([#835](https://github.com/hashicorp/terraform-plugin-framework/issues/835))

FEATURES:

* providerserver: Upgrade to protocol versions 5.4 and 6.4, which can significantly reduce memory usage with Terraform 1.6 and later when a configuration includes multiple instances of the same provider ([#828](https://github.com/hashicorp/terraform-plugin-framework/issues/828))

## 1.3.5 (August 17, 2023)

NOTES:

* internal: Changed provider defined method execution logs from `DEBUG` log level to `TRACE` ([#818](https://github.com/hashicorp/terraform-plugin-framework/issues/818))

BUG FIXES:

* internal/fwserver: Prevented `Invalid Element Type` diagnostics for nested attributes and blocks implementing `CustomType` field ([#823](https://github.com/hashicorp/terraform-plugin-framework/issues/823))

## 1.3.4 (August 03, 2023)

BUG FIXES:

* types/basetypes: Prevented Float64Value Terraform data consistency errors for numbers with high precision floating point rounding errors ([#817](https://github.com/hashicorp/terraform-plugin-framework/issues/817))

## 1.3.3 (July 20, 2023)

BUG FIXES:

* types/basetypes: Minor reduction of memory allocations for `ObjectValue` type `ToTerraformValue()` method, which decreases provider operation durations at scale ([#775](https://github.com/hashicorp/terraform-plugin-framework/issues/775))
* resource: Prevented panic during planning caused by `SetNestedAttribute` with nested attribute `Default` and multiple configured elements ([#783](https://github.com/hashicorp/terraform-plugin-framework/issues/783))
* tfsdk: Prevented `Value Conversion Error` diagnostics when using `Set()` method with base types instead of custom types ([#806](https://github.com/hashicorp/terraform-plugin-framework/issues/806))
* providerserver: Significantly reduced memory usage for framework data handling operations, especially during the `PlanResourceChange` RPC ([#792](https://github.com/hashicorp/terraform-plugin-framework/issues/792))

## 1.3.2 (June 28, 2023)

BUG FIXES:

* resource/schema: Ensured `Default` implementations received request `Path` and have response `Diagnostics` handled ([#778](https://github.com/hashicorp/terraform-plugin-framework/issues/778))
* resource/schema: Prevented panics with `Default` implementations on list, map, and set where no response `Diagnostics` or `PlanValue` was returned ([#778](https://github.com/hashicorp/terraform-plugin-framework/issues/778))
* providerserver: Ensured Terraform CLI interrupts (e.g. Ctrl-c) properly cancel the `context.Context` of inflight requests ([#782](https://github.com/hashicorp/terraform-plugin-framework/issues/782))
* providerserver: Prevented caching of unused data and managed resource schemas ([#784](https://github.com/hashicorp/terraform-plugin-framework/issues/784))

## 1.3.1 (June 14, 2023)

BUG FIXES:

* datasource/schema: Ensure nested attribute and block Equal methods check nested attribute and block definition equality ([#752](https://github.com/hashicorp/terraform-plugin-framework/issues/752))
* provider/metaschema: Ensure nested attribute Equal methods check nested attribute definition equality ([#752](https://github.com/hashicorp/terraform-plugin-framework/issues/752))
* provider/schema: Ensure nested attribute and block Equal methods check nested attribute and block definition equality ([#752](https://github.com/hashicorp/terraform-plugin-framework/issues/752))
* resource/schema: Ensure nested attribute and block Equal methods check nested attribute and block definition equality ([#752](https://github.com/hashicorp/terraform-plugin-framework/issues/752))
* types/basetypes: Prevented panics in `ListType`, `MapType`, and `SetType` methods when `ElemType` field is not set ([#714](https://github.com/hashicorp/terraform-plugin-framework/issues/714))
* resource/schema: Prevented `Value Conversion Error` diagnostics for attributes and blocks implementing both `CustomType` and `PlanModifiers` fields ([#754](https://github.com/hashicorp/terraform-plugin-framework/issues/754))
* types/basetypes: Prevented panic with `ListTypableWithSemanticEquals` and `SetTypableWithSemanticEquals` when proposed new element count was greater than prior element count ([#772](https://github.com/hashicorp/terraform-plugin-framework/issues/772))

## 1.3.0 (June 07, 2023)

NOTES:

* datasource/schema: The `Schema` type `Validate()` method has been deprecated in preference of `ValidateImplementation()` ([#699](https://github.com/hashicorp/terraform-plugin-framework/issues/699))
* provider/metaschema: The `Schema` type `Validate()` method has been deprecated in preference of `ValidateImplementation()` ([#699](https://github.com/hashicorp/terraform-plugin-framework/issues/699))
* provider/schema: The `Schema` type `Validate()` method has been deprecated in preference of `ValidateImplementation()` ([#699](https://github.com/hashicorp/terraform-plugin-framework/issues/699))
* resource/schema: The `Schema` type `Validate()` method has been deprecated in preference of `ValidateImplementation()` ([#699](https://github.com/hashicorp/terraform-plugin-framework/issues/699))

ENHANCEMENTS:

* datasource/schema: Added `Schema` type `ValidateImplementation()` method, which performs framework-defined schema validation and can be used in unit testing ([#699](https://github.com/hashicorp/terraform-plugin-framework/issues/699))
* provider/metaschema: Added `Schema` type `ValidateImplementation()` method, which performs framework-defined schema validation and can be used in unit testing ([#699](https://github.com/hashicorp/terraform-plugin-framework/issues/699))
* provider/schema: Added `Schema` type `ValidateImplementation()` method, which performs framework-defined schema validation and can be used in unit testing ([#699](https://github.com/hashicorp/terraform-plugin-framework/issues/699))
* resource/schema: Added `Schema` type `ValidateImplementation()` method, which performs framework-defined schema validation and can be used in unit testing ([#699](https://github.com/hashicorp/terraform-plugin-framework/issues/699))
* datasource/schema: Raise validation errors if attempting to use top-level `for_each` attribute name, which requires special Terraform configuration syntax to be usable by the data source ([#704](https://github.com/hashicorp/terraform-plugin-framework/issues/704))
* resource/schema: Raise validation errors if attempting to use top-level `for_each` attribute name, which requires special Terraform configuration syntax to be usable by the resource ([#704](https://github.com/hashicorp/terraform-plugin-framework/issues/704))
* datasource/schema: Raise validation errors if attempting to use attribute names with leading numerics (0-9), which are invalid in the Terraform configuration language ([#705](https://github.com/hashicorp/terraform-plugin-framework/issues/705))
* provider/schema: Raise validation errors if attempting to use attribute names with leading numerics (0-9), which are invalid in the Terraform configuration language ([#705](https://github.com/hashicorp/terraform-plugin-framework/issues/705))
* resource/schema: Raise validation errors if attempting to use attribute names with leading numerics (0-9), which are invalid in the Terraform configuration language ([#705](https://github.com/hashicorp/terraform-plugin-framework/issues/705))
* all: Improved SDK logging performance when messages would be skipped due to configured logging level ([#744](https://github.com/hashicorp/terraform-plugin-framework/issues/744))

BUG FIXES:

* datasource/schema: Raise errors with `ListAttribute`, `MapAttribute`, `ObjectAttribute`, and `SetAttribute` implementations instead of panics when missing required `AttributeTypes` or `ElementTypes` fields ([#699](https://github.com/hashicorp/terraform-plugin-framework/issues/699))
* provider/metaschema: Raise errors with `ListAttribute`, `MapAttribute`, `ObjectAttribute`, and `SetAttribute` implementations instead of panics when missing required `AttributeTypes` or `ElementTypes` fields ([#699](https://github.com/hashicorp/terraform-plugin-framework/issues/699))
* provider/schema: Raise errors with `ListAttribute`, `MapAttribute`, `ObjectAttribute`, and `SetAttribute` implementations instead of panics when missing required `AttributeTypes` or `ElementTypes` fields ([#699](https://github.com/hashicorp/terraform-plugin-framework/issues/699))
* resource/schema: Raise errors with `ListAttribute`, `MapAttribute`, `ObjectAttribute`, and `SetAttribute` implementations instead of panics when missing required `AttributeTypes` or `ElementTypes` fields ([#699](https://github.com/hashicorp/terraform-plugin-framework/issues/699))
* tfsdk: Raise framework errors instead of generic upstream errors or panics when encountering unexpected values with `Set()` methods ([#715](https://github.com/hashicorp/terraform-plugin-framework/issues/715))

## 1.2.0 (March 21, 2023)

NOTES:

* New `DEBUG` level `Detected value change between proposed new state and prior state` log messages with the offending attribute path are now emitted when proposed new state value differences would cause the framework to automatically mark all unconfigured `Computed` attributes as unknown during planning. These can be used to troubleshoot potential resource implementation issues, or framework and Terraform plan logic bugs. ([#630](https://github.com/hashicorp/terraform-plugin-framework/issues/630))
* This Go module has been updated to Go 1.19 per the [Go support policy](https://golang.org/doc/devel/release.html#policy). Any consumers building on earlier Go versions may experience errors. ([#682](https://github.com/hashicorp/terraform-plugin-framework/issues/682))

FEATURES:

* resource/schema: Introduce packages, interface types, and built-in static value functionality for schema-based default values ([#674](https://github.com/hashicorp/terraform-plugin-framework/issues/674))

ENHANCEMENTS:

* internal/fwserver: Added `DEBUG` logging to aid troubleshooting unexpected plans with unknown values ([#630](https://github.com/hashicorp/terraform-plugin-framework/issues/630))
* types/basetypes: Add `BoolValue` type `NewBoolPointerValue()` creation function and  `ValueBoolPointer()` method ([#689](https://github.com/hashicorp/terraform-plugin-framework/issues/689))
* types/basetypes: Add `Float64Value` type `NewFloat64PointerValue()` creation function and  `ValueFloat64Pointer()` method ([#689](https://github.com/hashicorp/terraform-plugin-framework/issues/689))
* types/basetypes: Add `Int64Value` type `NewInt64PointerValue()` creation function and  `ValueInt64Pointer()` method ([#689](https://github.com/hashicorp/terraform-plugin-framework/issues/689))
* types/basetypes: Add `StringValue` type `NewStringPointerValue()` creation function and  `ValueStringPointer()` method ([#689](https://github.com/hashicorp/terraform-plugin-framework/issues/689))
* resource/schema: Added `Default` fields to `Attribute` types, which support schema-based default values ([#674](https://github.com/hashicorp/terraform-plugin-framework/issues/674))

BUG FIXES:

* types/basetypes: Fixed `Float64Type` type `ValueFromTerraform` method to handle valid, stringified numbers from Terraform ([#648](https://github.com/hashicorp/terraform-plugin-framework/issues/648))
* resource: Prevented nested attribute and block plan modifications from being undone ([#669](https://github.com/hashicorp/terraform-plugin-framework/issues/669))

# 1.1.1 (January 13, 2022)

BUG FIXES:

* all: Prevented `tftypes.NewValue can't use []tftypes.Value as a tftypes.Object` panics with schemas that included `SingleNestedBlock` ([#624](https://github.com/hashicorp/terraform-plugin-framework/issues/624))

# 1.1.0 (January 13, 2022)

NOTES:

* all: For data handling consistency with attributes, unconfigured list and set blocks will now be represented as a null list or set instead of a known list or set with zero elements. This prevents confusing situations with validation and plan modification, where it was previously required to check block values for the number of elements. Logic that was previously missing null value checks for blocks may require updates. ([#604](https://github.com/hashicorp/terraform-plugin-framework/issues/604))
* tfsdk: The `Config`, `Plan`, and `State` type `PathMatches()` method logic previously returned `Invalid Path Expression for Schema Data` errors based on implementation details of the underlying data, which prevented returning zero matches in cases where the expression is valid for the schema, but there was no actual data at the path. Providers can now determine whether zero matches is consequential for their use case. ([#602](https://github.com/hashicorp/terraform-plugin-framework/issues/602))

ENHANCEMENTS:

* path: Added `Expressions` type `Matches` method for checking if any expression in the collection matches a given path ([#604](https://github.com/hashicorp/terraform-plugin-framework/issues/604))
* tfsdk: Automatically prevented Terraform `nested blocks must be empty to indicate no blocks` errors for responses containing `Plan` and `State` types ([#621](https://github.com/hashicorp/terraform-plugin-framework/issues/621))

BUG FIXES:

* datasource/schema: Prevented `ListNestedBlock` and `SetNestedBlock` type `DeprecationMessage` field from causing `Block Deprecated` warnings with unconfigured blocks ([#604](https://github.com/hashicorp/terraform-plugin-framework/issues/604))
* datasource: Prevented `ConfigValidators` from unexpectedly modifying or removing prior validator diagnostics ([#619](https://github.com/hashicorp/terraform-plugin-framework/issues/619))
* provider/schema: Prevented `ListNestedBlock` and `SetNestedBlock` type `DeprecationMessage` field from causing `Block Deprecated` warnings with unconfigured blocks ([#604](https://github.com/hashicorp/terraform-plugin-framework/issues/604))
* provider: Prevented `ConfigValidators` from unexpectedly modifying or removing prior validator diagnostics ([#619](https://github.com/hashicorp/terraform-plugin-framework/issues/619))
* resource/schema: Prevented `ListNestedBlock` and `SetNestedBlock` type `DeprecationMessage` field from causing `Block Deprecated` warnings with unconfigured blocks ([#604](https://github.com/hashicorp/terraform-plugin-framework/issues/604))
* resource: Prevented `ConfigValidators` from unexpectedly modifying or removing prior validator diagnostics ([#619](https://github.com/hashicorp/terraform-plugin-framework/issues/619))
* tfsdk: Fixed false positive `Invalid Path Expression for Schema Data` error to be schema-determined instead of data-determined ([#602](https://github.com/hashicorp/terraform-plugin-framework/issues/602))
* types/basetypes: Fixed `ObjectType` type `ApplyTerraform5AttributePathStep` method to return an error instead of `nil` for invalid attribute name steps ([#602](https://github.com/hashicorp/terraform-plugin-framework/issues/602))

# 1.0.1 (December 19, 2022)

BUG FIXES:

* resource/schema/planmodifier: Prevented `assignment to entry in nil map` panic for `Object` type plan modifiers ([#591](https://github.com/hashicorp/terraform-plugin-framework/issues/591))
* types/basetypes: Prevented type mutation via the `ObjectType` type `AttributeTypes()` method return ([#591](https://github.com/hashicorp/terraform-plugin-framework/issues/591))
* types/basetypes: Prevented value mutation via the `ListValue`, `MapValue`, and `SetValue` type `Elements()` method return ([#591](https://github.com/hashicorp/terraform-plugin-framework/issues/591))
* types/basetypes: Prevented value mutation via the `ObjectValue` type `AttributeTypes()` and `Attributes()` method returns ([#591](https://github.com/hashicorp/terraform-plugin-framework/issues/591))

# 1.0.0 (December 13, 2022)

NOTES:

* The Terraform Plugin Framework is now generally available with semantic versioning compatibility promises. ([#578](https://github.com/hashicorp/terraform-plugin-framework/issues/578))
* types: Framework type implementations have been moved into the underlying `basetypes` package. Value creation functions and type aliases have been created in the `types` package that should prevent any breaking changes. ([#567](https://github.com/hashicorp/terraform-plugin-framework/issues/567))

BREAKING CHANGES:

* provider: The `Provider` interface now requires the `Metadata` method. It can be left empty or set the `MetadataResponse` type `TypeName` field to populate `datasource.MetadataRequest` and `resource.MetadataRequest` type `ProviderTypeName` fields. ([#580](https://github.com/hashicorp/terraform-plugin-framework/issues/580))
* resource: The `RequiresReplace()` plan modifier has been removed. Use a type-specific plan modifier instead, such as `resource/schema/stringplanmodifier.RequiresReplace()` or `resource/schema/stringplanmodifier.RequiresReplaceIfConfigured()` ([#576](https://github.com/hashicorp/terraform-plugin-framework/issues/576))
* resource: The `RequiresReplaceIf()` plan modifier has been removed. Use a type-specific plan modifier instead, such as `resource/schema/stringplanmodifier.RequiresReplaceIf()` ([#576](https://github.com/hashicorp/terraform-plugin-framework/issues/576))
* resource: The `Resource` type `GetSchema` method has been removed. Use the `Schema` method instead. ([#576](https://github.com/hashicorp/terraform-plugin-framework/issues/576))
* resource: The `StateUpgrader` type `PriorSchema` field type has been migrated from `tfsdk.Schema` to `resource/schema.Schema`, similar to other resource schema handling ([#573](https://github.com/hashicorp/terraform-plugin-framework/issues/573))
* resource: The `UseStateForUnknown()` plan modifier has been removed. Use a type-specific plan modifier instead, such as `resource/schema/stringplanmodifier.UseStateForUnknown()` ([#576](https://github.com/hashicorp/terraform-plugin-framework/issues/576))
* tfsdk: The `AttributePlanModifier` interface has been removed. Use the type-specific plan modifier interfaces in the `resource/schema/planmodifier` package instead. ([#576](https://github.com/hashicorp/terraform-plugin-framework/issues/576))
* tfsdk: The `AttributeValidator` interface has been removed. Use the type-specific validator interfaces in the `schema/validator` package instead. ([#576](https://github.com/hashicorp/terraform-plugin-framework/issues/576))
* tfsdk: The `Attribute`, `Block`, and `Schema` types have been removed. Use the similarly named types in the `datasource/schema`, `provider/schema`, and `resource/schema` packages instead. ([#576](https://github.com/hashicorp/terraform-plugin-framework/issues/576))
* tfsdk: The `ListNestedAttributes`, `MapNestedAttributes`, `SetNestedAttributes`, and `SingleNestedAttributes` functions have been removed. Use the similarly named types in the `datasource/schema`, `provider/schema`, and `resource/schema` packages instead. ([#576](https://github.com/hashicorp/terraform-plugin-framework/issues/576))
* types: The type-specific `Typable` and `Valuable` interfaces have been moved into the underlying `basetypes` package. ([#567](https://github.com/hashicorp/terraform-plugin-framework/issues/567))

FEATURES:

* types/basetypes: New package which contains embeddable types for custom types ([#567](https://github.com/hashicorp/terraform-plugin-framework/issues/567))

BUG FIXES:

* datasource: Add `Validate` function to `Schema` to prevent usage of reserved and invalid names for attributes and blocks ([#548](https://github.com/hashicorp/terraform-plugin-framework/issues/548))
* provider: Add `Validate` function to `MetaSchema` to prevent usage of reserved and invalid names for attributes and blocks ([#548](https://github.com/hashicorp/terraform-plugin-framework/issues/548))
* provider: Add `Validate` function to `Schema` to prevent usage of reserved and invalid names for attributes and blocks ([#548](https://github.com/hashicorp/terraform-plugin-framework/issues/548))
* resource: Add `Validate` function to `Schema` to prevent usage of reserved and invalid names for attributes and blocks ([#548](https://github.com/hashicorp/terraform-plugin-framework/issues/548))

# 0.17.0 (November 30, 2022)

NOTES:

* datasource: The `DataSource` type `GetSchema` method has been deprecated. Use the `Schema` method instead. ([#546](https://github.com/hashicorp/terraform-plugin-framework/issues/546))
* provider: The `Provider` type `GetSchema` method has been deprecated. Use the `Schema` method instead. ([#553](https://github.com/hashicorp/terraform-plugin-framework/issues/553))
* resource: The `RequiresReplace()` plan modifier has been deprecated. Use a type-specific plan modifier instead, such as `resource/schema/stringplanmodifier.RequiresReplace()` or `resource/schema/stringplanmodifier.RequiresReplaceIfConfigured()` ([#565](https://github.com/hashicorp/terraform-plugin-framework/issues/565))
* resource: The `RequiresReplaceIf()` plan modifier has been deprecated. Use a type-specific plan modifier instead, such as `resource/schema/stringplanmodifier.RequiresReplaceIf()` ([#565](https://github.com/hashicorp/terraform-plugin-framework/issues/565))
* resource: The `Resource` type `GetSchema` method has been deprecated. Use the `Schema` method instead. ([#558](https://github.com/hashicorp/terraform-plugin-framework/issues/558))
* resource: The `UseStateForUnknown()` plan modifier has been deprecated. Use a type-specific plan modifier instead, such as `resource/schema/stringplanmodifier.UseStateForUnknown()` ([#565](https://github.com/hashicorp/terraform-plugin-framework/issues/565))
* tfsdk: The `Attribute`, `Block`, and `Schema` types have been deprecated. Use the similarly named types in the `datasource/schema`, `provider/schema`, and `resource/schema` packages instead. ([#563](https://github.com/hashicorp/terraform-plugin-framework/issues/563))
* tfsdk: The `ListNestedAttributes`, `MapNestedAttributes`, `SetNestedAttributes`, and `SingleNestedAttributes` functions have been deprecated. Use the similarly named types in the `datasource/schema`, `provider/schema`, and `resource/schema` packages instead. ([#563](https://github.com/hashicorp/terraform-plugin-framework/issues/563))

BREAKING CHANGES:

* provider: The `ProviderWithMetaSchema` type `GetMetaSchema` method has been replaced with the `MetaSchema` method ([#562](https://github.com/hashicorp/terraform-plugin-framework/issues/562))
* tfsdk: The `Attribute` type `FrameworkType()` method has been removed. Use the `GetType()` method instead which returns the same information. ([#543](https://github.com/hashicorp/terraform-plugin-framework/issues/543))
* tfsdk: The `Attribute` type `GetType()` method now returns type information whether the attribute implements the `Type` field or `Attributes` field. ([#543](https://github.com/hashicorp/terraform-plugin-framework/issues/543))
* tfsdk: The `Config`, `Plan`, and `State` type `Schema` field type has been updated from `tfsdk.Schema` to the generic `fwschema.Schema` interface to enable additional schema implementations ([#544](https://github.com/hashicorp/terraform-plugin-framework/issues/544))

FEATURES:

* datasource/schema: New package which contains schema interfaces and types relevant to data sources ([#546](https://github.com/hashicorp/terraform-plugin-framework/issues/546))
* provider/schema: New package which contains schema interfaces and types relevant to providers ([#553](https://github.com/hashicorp/terraform-plugin-framework/issues/553))
* resource/schema/planmodifier: New package which contains type-specific schema plan modifier interfaces ([#557](https://github.com/hashicorp/terraform-plugin-framework/issues/557))
* resource/schema: New package which contains schema interfaces and types relevant to resources ([#558](https://github.com/hashicorp/terraform-plugin-framework/issues/558))
* resource/schema: New packages, such as `stringplanmodifier` which contain type-specific schema plan modifier implementations ([#565](https://github.com/hashicorp/terraform-plugin-framework/issues/565))
* schema/validator: New package which contains type-specific schema validator interfaces ([#542](https://github.com/hashicorp/terraform-plugin-framework/issues/542))

BUG FIXES:

* diag: Allow diagnostic messages with incorrect UTF-8 encoding to pass through with the invalid sequences replaced with the Unicode Replacement Character. This avoids returning the unhelpful message "string field contains invalid UTF-8" in that case. ([#549](https://github.com/hashicorp/terraform-plugin-framework/issues/549))
* internal/fwserver: Ensured blocks are ignored when marking computed nils as unknown during resource change planning ([#552](https://github.com/hashicorp/terraform-plugin-framework/issues/552))

# 0.16.0 (November 15, 2022)

BREAKING CHANGES:

* types: The `Bool` type `Null`, `Unknown`, and `Value` fields have been removed. Use the `BoolNull()`, `BoolUnknown()`, and `BoolValue()` creation functions and `IsNull()`, `IsUnknown()`, and `ValueBool()` methods instead. ([#523](https://github.com/hashicorp/terraform-plugin-framework/issues/523))
* types: The `Float64` type `Null`, `Unknown`, and `Value` fields have been removed. Use the `Float64Null()`, `Float64Unknown()`, and `Float64Value()` creation functions and `IsNull()`, `IsUnknown()`, and `ValueFloat64()` methods instead. ([#523](https://github.com/hashicorp/terraform-plugin-framework/issues/523))
* types: The `Int64` type `Null`, `Unknown`, and `Value` fields have been removed. Use the `Int64Null()`, `Int64Unknown()`, and `Int64Value()` creation functions and `IsNull()`, `IsUnknown()`, and `ValueInt64()` methods instead. ([#523](https://github.com/hashicorp/terraform-plugin-framework/issues/523))
* types: The `List` type `Elems`, `ElemType`, `Null`, and `Unknown` fields have been removed. Use the `ListNull()`, `ListUnknown()`, `ListValue()`, and `ListValueMust()` creation functions and `Elements()`, `ElementsAs()`, `ElementType()`, `IsNull()`, and `IsUnknown()` methods instead. ([#523](https://github.com/hashicorp/terraform-plugin-framework/issues/523))
* types: The `Map` type `Elems`, `ElemType`, `Null`, and `Unknown` fields have been removed. Use the `MapNull()`, `MapUnknown()`, `MapValue()`, and `MapValueMust()` creation functions and `Elements()`, `ElementsAs()`, `ElementType()`, `IsNull()`, and `IsUnknown()` methods instead. ([#523](https://github.com/hashicorp/terraform-plugin-framework/issues/523))
* types: The `Number` type `Null`, `Unknown`, and `Value` fields have been removed. Use the `NumberNull()`, `NumberUnknown()`, and `NumberValue()` creation functions and `IsNull()`, `IsUnknown()`, and `ValueBigFloat()` methods instead. ([#523](https://github.com/hashicorp/terraform-plugin-framework/issues/523))
* types: The `Object` type `Attrs`, `AttrTypes`, `Null`, and `Unknown` fields have been removed. Use the `ObjectNull()`, `ObjectUnknown()`, `ObjectValue()`, and `ObjectValueMust()` creation functions and `As()`, `Attributes()`, `AttributeTypes()`, `IsNull()`, and `IsUnknown()` methods instead. ([#523](https://github.com/hashicorp/terraform-plugin-framework/issues/523))
* types: The `Set` type `Elems`, `ElemType`, `Null`, and `Unknown` fields have been removed. Use the `SetNull()`, `SetUnknown()`, `SetValue()`, and `SetValueMust()` creation functions and `Elements()`, `ElementsAs()`, `ElementType()`, `IsNull()`, and `IsUnknown()` methods instead. ([#523](https://github.com/hashicorp/terraform-plugin-framework/issues/523))
* types: The `String` type `Null`, `Unknown`, and `Value` fields have been removed. Use the `StringNull()`, `StringUnknown()`, and `StringValue()` creation functions and `IsNull()`, `IsUnknown()`, and `ValueString()` methods instead. ([#523](https://github.com/hashicorp/terraform-plugin-framework/issues/523))

ENHANCEMENTS:

* attr: Added `ValueState` type, which custom types can use to consistently represent the three possible value states (known, null, and unknown) ([#523](https://github.com/hashicorp/terraform-plugin-framework/issues/523))
* types: Added `BoolTypable` and `BoolValuable` interface types, which enable embedding existing boolean types for custom types ([#536](https://github.com/hashicorp/terraform-plugin-framework/issues/536))
* types: Added `Float64Typable` and `Float64Valuable` interface types, which enable embedding existing float64 types for custom types ([#536](https://github.com/hashicorp/terraform-plugin-framework/issues/536))
* types: Added `Int64Typable` and `Int64Valuable` interface types, which enable embedding existing int64 types for custom types ([#536](https://github.com/hashicorp/terraform-plugin-framework/issues/536))
* types: Added `ListTypable` and `ListValuable` interface types, which enable embedding existing list types for custom types ([#536](https://github.com/hashicorp/terraform-plugin-framework/issues/536))
* types: Added `MapTypable` and `MapValuable` interface types, which enable embedding existing map types for custom types ([#536](https://github.com/hashicorp/terraform-plugin-framework/issues/536))
* types: Added `NumberTypable` and `NumberValuable` interface types, which enable embedding existing number types for custom types ([#536](https://github.com/hashicorp/terraform-plugin-framework/issues/536))
* types: Added `ObjectTypable` and `ObjectValuable` interface types, which enable embedding existing object types for custom types ([#536](https://github.com/hashicorp/terraform-plugin-framework/issues/536))
* types: Added `SetTypable` and `SetValuable` interface types, which enable embedding existing set types for custom types ([#536](https://github.com/hashicorp/terraform-plugin-framework/issues/536))
* types: Added `StringTypable` and `StringValuable` interface types, which enable embedding existing string types for custom types ([#536](https://github.com/hashicorp/terraform-plugin-framework/issues/536))

BUG FIXES:

* types: Prevented Terraform errors where the zero-value for any `attr.Value` types such as `String` would be a known value instead of null ([#523](https://github.com/hashicorp/terraform-plugin-framework/issues/523))
* types: Prevented indeterminate behavior for any `attr.Value` types where they could be any combination of null, unknown, and/or known ([#523](https://github.com/hashicorp/terraform-plugin-framework/issues/523))

# 0.15.0 (October 26, 2022)

NOTES:

* types: The `Bool` type `Null`, `Unknown`, and `Value` fields have been deprecated in preference of the `BoolNull()`, `BoolUnknown()`, and `BoolValue()` creation functions and `IsNull()`, `IsUnknown()`, and `ValueBool()` methods. The fields will be removed in a future release. ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: The `Float64` type `Null`, `Unknown`, and `Value` fields have been deprecated in preference of the `Float64Null()`, `Float64Unknown()`, and `Float64Value()` creation functions and `IsNull()`, `IsUnknown()`, and `ValueFloat64()` methods. The fields will be removed in a future release. ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: The `Int64` type `Null`, `Unknown`, and `Value` fields have been deprecated in preference of the `Int64Null()`, `Int64Unknown()`, and `Int64Value()` creation functions and `IsNull()`, `IsUnknown()`, and `ValueInt64()` methods. The fields will be removed in a future release. ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: The `List` type `Elems`, `ElemType`, `Null`, and `Unknown` fields have been deprecated in preference of the `ListNull()`, `ListUnknown()`, `ListValue()`, and `ListValueMust()` creation functions and `Elements()`, `ElementsAs()`, `ElementType()`, `IsNull()`, and `IsUnknown()` methods. The fields will be removed in a future release. ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: The `Map` type `Elems`, `ElemType`, `Null`, and `Unknown` fields have been deprecated in preference of the `MapNull()`, `MapUnknown()`, `MapValue()`, and `MapValueMust()` creation functions and `Elements()`, `ElementsAs()`, `ElementType()`, `IsNull()`, and `IsUnknown()` methods. The fields will be removed in a future release. ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: The `Number` type `Null`, `Unknown`, and `Value` fields have been deprecated in preference of the `NumberNull()`, `NumberUnknown()`, and `NumberValue()` creation functions and `IsNull()`, `IsUnknown()`, and `ValueBigFloat()` methods. The fields will be removed in a future release. ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: The `Object` type `Attrs`, `AttrTypes`, `Null`, and `Unknown` fields have been deprecated in preference of the `ObjectNull()`, `ObjectUnknown()`, `ObjectValue()`, and `ObjectValueMust()` creation functions and `As()`, `Attributes()`, `AttributeTypes()`, `IsNull()`, and `IsUnknown()` methods. The fields will be removed in a future release. ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: The `Set` type `Elems`, `ElemType`, `Null`, and `Unknown` fields have been deprecated in preference of the `SetNull()`, `SetUnknown()`, `SetValue()`, and `SetValueMust()` creation functions and `Elements()`, `ElementsAs()`, `ElementType()`, `IsNull()`, and `IsUnknown()` methods. The fields will be removed in a future release. ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: The `String` type `Null`, `Unknown`, and `Value` fields have been deprecated in preference of the `StringNull()`, `StringUnknown()`, and `StringValue()` creation functions and `IsNull()`, `IsUnknown()`, and `ValueString()` methods. The fields will be removed in a future release. ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))

ENHANCEMENTS:

* types: Added `BoolNull()`, `BoolUnknown()`, and `BoolValue()` functions, which create immutable `Bool` values ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: Added `Bool` type `ValueBool()` method, which returns the `bool` of the known value or `false` if null or unknown ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: Added `Float64Null()`, `Float64Unknown()`, and `Float64Value()` functions, which create immutable `Float64` values ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: Added `Float64` type `ValueFloat64()` method, which returns the `float64` of the known value or `0.0` if null or unknown ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: Added `Int64Null()`, `Int64Unknown()`, and `Int64Value()` functions, which create immutable `Int64` values ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: Added `Int64` type `ValueInt64()` method, which returns the `int64` of the known value or `0` if null or unknown ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: Added `ListNull()`, `ListUnknown()`, `ListValue()`, and `ListValueMust()` functions, which create immutable `List` values ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: Added `ListValueFrom()`, `MapValueFrom()`, `ObjectValueFrom()`, and `SetValueFrom()` functions, which can create value types from standard Go types using reflection similar to `tfsdk.ValueFrom()` ([#522](https://github.com/hashicorp/terraform-plugin-framework/issues/522))
* types: Added `List` type `Elements()` method, which returns the `[]attr.Value` of the known values or `nil` if null or unknown ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: Added `MapNull()`, `MapUnknown()`, `MapValue()`, and `MapValueMust()` functions, which create immutable `Map` values ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: Added `Map` type `Elements()` method, which returns the `map[string]attr.Value` of the known values or `nil` if null or unknown ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: Added `NumberNull()`, `NumberUnknown()`, and `NumberValue()` functions, which create immutable `Number` values ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: Added `Number` type `ValueBigFloat()` method, which returns the `*big.Float` of the known value or `nil` if null or unknown ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: Added `SetNull()`, `SetUnknown()`, `SetValue()`, and `SetValueMust()` functions, which create immutable `Set` values ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: Added `Set` type `Elements()` method, which returns the `[]attr.Value` of the known values or `nil` if null or unknown ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: Added `StringNull()`, `StringUnknown()`, and `StringValue()` functions, which create immutable `String` values ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))
* types: Added `String` type `ValueString()` method, which returns the `string` of the known value or `""` if null or unknown ([#502](https://github.com/hashicorp/terraform-plugin-framework/issues/502))

# 0.14.0 (October 4, 2022)

NOTES:

* The Terraform Plugin Framework is now in beta. Feedback towards a general availability release in the future with compatibility promises is appreciated. ([#500](https://github.com/hashicorp/terraform-plugin-framework/issues/500))

BREAKING CHANGES:

* attr: The `Type` interface now requires the `ValueType` method, which is used for enhancing error diagnostics from the framework ([#497](https://github.com/hashicorp/terraform-plugin-framework/issues/497))

ENHANCEMENTS:

* internal/reflect: Added `attr.Value` type suggestions to error diagnostics ([#497](https://github.com/hashicorp/terraform-plugin-framework/issues/497))

# 0.13.0 (September 15, 2022)

NOTES:

* tfsdk: Schema definitions may now introduce single nested mode blocks, however this support is only intended for migrating terraform-plugin-sdk timeouts blocks. New implementations should prefer single nested attributes instead. ([#477](https://github.com/hashicorp/terraform-plugin-framework/issues/477))

BREAKING CHANGES:

* datasource: The `DataSource` interface now requires the `GetSchema` and `Metadata` methods. ([#478](https://github.com/hashicorp/terraform-plugin-framework/issues/478))
* provider: The `DataSourceType` and `ResourceType` types have been removed. Use the `GetSchema`, `Metadata`, and optionally the `Configure` methods on `datasource.DataSource` and `resource.Resource` implementations instead. ([#478](https://github.com/hashicorp/terraform-plugin-framework/issues/478))
* provider: The `Provider` interface `GetDataSources` and `GetResources` methods have been removed. Use the `DataSources` and `Resources` methods instead. ([#478](https://github.com/hashicorp/terraform-plugin-framework/issues/478))
* resource: The `Resource` interface now requires the `GetSchema` and `Metadata` methods. ([#478](https://github.com/hashicorp/terraform-plugin-framework/issues/478))

ENHANCEMENTS:

* tfsdk: Added single nested mode block support ([#477](https://github.com/hashicorp/terraform-plugin-framework/issues/477))

BUG FIXES:

* internal/fwserver: Ensured nested block plan modifiers correctly set their request `AttributeConfig`, `AttributePlan`, and `AttributeState` values ([#479](https://github.com/hashicorp/terraform-plugin-framework/issues/479))
* types: Ensured `List`, `Map`, and `Set` types with `xattr.TypeWithValidate` elements run validation on those elements ([#481](https://github.com/hashicorp/terraform-plugin-framework/issues/481))

# 0.12.0 (September 12, 2022)

NOTES:

* datasource: The `DataSource` type `GetSchema` and `Metadata` methods will be required in the next version. ([#472](https://github.com/hashicorp/terraform-plugin-framework/issues/472))
* provider: The `DataSourceType` type has been deprecated in preference of moving the `GetSchema` method to the `datasource.DataSource` type  and optionally implementing the `NewResource` method logic to a new `Configure` method. The `DataSourceType` type will be removed in the next version. ([#472](https://github.com/hashicorp/terraform-plugin-framework/issues/472))
* provider: The `Provider` type `GetDataSources` method has been deprecated in preference of the `DataSources` method. All `datasource.DataSource` types must implement the `Metadata` method after migrating. Support for the `GetDataSources` method will be removed in the next version. ([#472](https://github.com/hashicorp/terraform-plugin-framework/issues/472))
* provider: The `Provider` type `GetResources` method has been deprecated in preference of the `Resources` method. All `resource.Resource` types must implement the `Metadata` method after migrating. Support for the `GetResources` method will be removed in the next version. ([#472](https://github.com/hashicorp/terraform-plugin-framework/issues/472))
* provider: The `ResourceType` type has been deprecated in preference of moving the `GetSchema` method to the `resource.Resource` type and optionally implementing the `NewResource` method logic to a new `Configure` method.  The `ResourceType` type will be removed in the next version. ([#472](https://github.com/hashicorp/terraform-plugin-framework/issues/472))
* resource: The `Resource` type `GetSchema` and `Metadata` methods will be required in the next version. ([#472](https://github.com/hashicorp/terraform-plugin-framework/issues/472))

BREAKING CHANGES:

* tfsdk: The `Schema` type `AttributeAtPath()` method signature has be updated with a `path.Path` parameter and `diag.Diagnostics` return. Use the `AttributeAtTerraformPath()` method instead if `*tftypes.AttributePath` or specific `error` handling is still necessary. ([#450](https://github.com/hashicorp/terraform-plugin-framework/issues/450))
* tfsdk: The previously deprecated `Schema` type `AttributeType()` method has been removed. Use the `Type()` method instead. ([#450](https://github.com/hashicorp/terraform-plugin-framework/issues/450))
* tfsdk: The previously deprecated `Schema` type `AttributeTypeAtPath()` method has been removed. Use the `TypeAtPath()` or `TypeAtTerraformPath()` method instead. ([#450](https://github.com/hashicorp/terraform-plugin-framework/issues/450))
* tfsdk: The previously deprecated `Schema` type `TerraformType()` method has been removed. Use `Type().TerraformType()` instead. ([#450](https://github.com/hashicorp/terraform-plugin-framework/issues/450))

ENHANCEMENTS:

* datasource: Added `DataSource` type `Configure`, `GetSchema`, and `Metadata` method support ([#472](https://github.com/hashicorp/terraform-plugin-framework/issues/472))
* provider: Added `ConfigureResponse` type `DataSourceData` field, which will set the `datasource.ConfigureRequest.ProviderData` field ([#472](https://github.com/hashicorp/terraform-plugin-framework/issues/472))
* provider: Added `ConfigureResponse` type `ResourceData` field, which will set the `resource.ConfigureRequest.ProviderData` field ([#472](https://github.com/hashicorp/terraform-plugin-framework/issues/472))
* provider: Added `Provider` type `Metadata` method support, which the `MetadataResponse.TypeName` field will set the `datasource.MetadataRequest.ProviderTypeName` and `resource.MetadataRequest.ProviderTypeName` fields ([#472](https://github.com/hashicorp/terraform-plugin-framework/issues/472))
* resource: Added `Resource` type `Configure`, `GetSchema`, and `Metadata` method support ([#472](https://github.com/hashicorp/terraform-plugin-framework/issues/472))

BUG FIXES:

* internal/fwserver: Delayed deprecated attribute/block warnings for unknown values, which may be null ([#465](https://github.com/hashicorp/terraform-plugin-framework/issues/465))
* internal/fwserver: Fixed alignment of set type plan modification ([#468](https://github.com/hashicorp/terraform-plugin-framework/issues/468))

# 0.11.1 (August 15, 2022)

BUG FIXES:
* resource: Prevented `Error Decoding Private State` errors on resources previously managed by terraform-plugin-sdk ([#452](https://github.com/hashicorp/terraform-plugin-framework/issues/452))

# 0.11.0 (August 11, 2022)

NOTES:

* This Go module has been updated to Go 1.18 per the [Go support policy](https://golang.org/doc/devel/release.html#policy). Any consumers building on earlier Go versions may experience errors. ([#445](https://github.com/hashicorp/terraform-plugin-framework/issues/445))
* tfsdk: The `Schema` type `AttributeAtPath()` method signature will be updated from a `*tftypes.AttributePath` parameter to `path.Path` in the next release. Switch to the `AttributeAtTerraformPath()` method if `*tftypes.AttributePath` handling is still necessary. ([#440](https://github.com/hashicorp/terraform-plugin-framework/issues/440))
* tfsdk: The `Schema` type `AttributeType()` method has been deprecated in preference of the `Type()` method. ([#440](https://github.com/hashicorp/terraform-plugin-framework/issues/440))
* tfsdk: The `Schema` type `AttributeTypeAtPath()` method has been deprecated for the `TypeAtPath()` and `TypeAtTerraformPath()` methods. ([#440](https://github.com/hashicorp/terraform-plugin-framework/issues/440))
* tfsdk: The `Schema` type `TerraformType()` method has been deprecated in preference of calling `Type().TerraformType()`. ([#440](https://github.com/hashicorp/terraform-plugin-framework/issues/440))

BREAKING CHANGES:

* tfsdk: Go types relating to data source handling have been migrated to the new `datasource` package. Consult the pull request description for a full listing of find-and-replace information. ([#432](https://github.com/hashicorp/terraform-plugin-framework/issues/432))
* tfsdk: Go types relating to provider handling have been migrated to the new `provider` package. Consult the pull request description for a full listing of find-and-replace information. ([#432](https://github.com/hashicorp/terraform-plugin-framework/issues/432))
* tfsdk: Go types relating to resource handling have been migrated to the new `resource` package. Consult the pull request description for a full listing of find-and-replace information. ([#432](https://github.com/hashicorp/terraform-plugin-framework/issues/432))
* tfsdk: The `RequiresReplace()`, `RequiresReplaceIf()`, and `UseStateForUnknown()` plan modifier functions, which only apply to managed resources, have been moved to the `resource` package. ([#434](https://github.com/hashicorp/terraform-plugin-framework/issues/434))
* tfsdk: The `ResourceImportStatePassthroughID()` function has been moved to `resource.ImportStatePassthroughID()`. ([#432](https://github.com/hashicorp/terraform-plugin-framework/issues/432))
* tfsdk: The `Schema` type `AttributeAtPath` method now returns a `fwschema.Attribute` interface instead of a `tfsdk.Attribute` type. Consumers will need to update from direct field usage to similarly named interface method calls. ([#438](https://github.com/hashicorp/terraform-plugin-framework/issues/438))

FEATURES:

* datasource: New package, which colocates all data source implementation types from the `tfsdk` package ([#432](https://github.com/hashicorp/terraform-plugin-framework/issues/432))
* provider: New package, which colocates all provider implementation types from the `tfsdk` package ([#432](https://github.com/hashicorp/terraform-plugin-framework/issues/432))
* resource: Enabled provider developers to read/write private state data. ([#433](https://github.com/hashicorp/terraform-plugin-framework/issues/433))
* resource: New package, which colocates all resource implementation types from the `tfsdk` package ([#432](https://github.com/hashicorp/terraform-plugin-framework/issues/432))

ENHANCEMENTS:

* tfsdk: Added `Block` type `MaxItems` and `MinItems` field validation for Terraform 0.12 through 0.15.1 ([#422](https://github.com/hashicorp/terraform-plugin-framework/issues/422))

BUG FIXES:

* internal/fwserver: Ensured `UpgradeResourceState` calls from Terraform 0.12 properly ignored attributes not defined in the schema ([#426](https://github.com/hashicorp/terraform-plugin-framework/issues/426))
* path: Ensured `Expression` type `Copy()` method appropriately copied root expressions and `Equal()` checked for root versus relative expressions ([#420](https://github.com/hashicorp/terraform-plugin-framework/issues/420))

# 0.10.0 (July 18, 2022)

BREAKING CHANGES:

* attr: The `TypeWithValidate` interface has been moved under the `attr/xattr` package and the `*tftypes.AttributePath` parameter is replaced with `path.Path` ([#390](https://github.com/hashicorp/terraform-plugin-framework/issues/390))
* diag: The `DiagnosticWithPath` interface `Path` method `*tftypes.AttributePath` return is replaced with `path.Path` ([#390](https://github.com/hashicorp/terraform-plugin-framework/issues/390))
* diag: The `Diagnostics` type `AddAttributeError` and `AddAttributeWarning` method `*tftypes.AttributePath` parameters are replaced with `path.Path` ([#390](https://github.com/hashicorp/terraform-plugin-framework/issues/390))
* diag: The `NewAttributeErrorDiagnostic` and `NewAttributeWarningDiagnostic` function `*tftypes.AttributePath` parameters are replaced with `path.Path` ([#390](https://github.com/hashicorp/terraform-plugin-framework/issues/390))
* tfsdk: The `Config`, `Plan`, and `State` types `GetAttribute` and `SetAttribute` methods `*tftypes.AttributePath` parameters are replaced with `path.Path` ([#390](https://github.com/hashicorp/terraform-plugin-framework/issues/390))
* tfsdk: The `DataSourceConfigValidator` interface `Validate` method is now `ValidateDataSource` to support generic validators that satisfy `DataSourceConfigValidator`, `ProviderConfigValidator`, and `ResourceConfigValidator` ([#405](https://github.com/hashicorp/terraform-plugin-framework/issues/405))
* tfsdk: The `ModifyAttributePlanRequest`, `ModifyResourcePlanResponse`, and `ValidateAttributeRequest` type `AttributePath *tftypes.AttributePath` fields are replaced with `AttributePath path.Path` ([#390](https://github.com/hashicorp/terraform-plugin-framework/issues/390))
* tfsdk: The `PlanResourceChange` RPC on destroy is now enabled. To prevent unexpected Terraform errors, the framework attempts to catch errant provider logic in plan modifiers when destroying. Resource level plan modifiers may require updates to handle a completely null proposed new state (plan) and ensure it remains completely null on resource destruction. ([#409](https://github.com/hashicorp/terraform-plugin-framework/issues/409))
* tfsdk: The `ProviderConfigValidator` interface `Validate` method is now `ValidateProvider` to support generic validators that satisfy `DataSourceConfigValidator`, `ProviderConfigValidator`, and `ResourceConfigValidator` ([#405](https://github.com/hashicorp/terraform-plugin-framework/issues/405))
* tfsdk: The `RequiresReplaceIf` and `ResourceImportStatePassthroughID` function `*tftypes.AttributePath` parameters are replaced with `path.Path` ([#390](https://github.com/hashicorp/terraform-plugin-framework/issues/390))
* tfsdk: The `ResourceConfigValidator` interface `Validate` method is now `ValidateResource` to support generic validators that satisfy `DataSourceConfigValidator`, `ProviderConfigValidator`, and `ResourceConfigValidator` ([#405](https://github.com/hashicorp/terraform-plugin-framework/issues/405))

FEATURES:

* Support plan modifiers returning warning and error diagnostics on resource destruction with Terraform 1.3 and later ([#409](https://github.com/hashicorp/terraform-plugin-framework/issues/409))
* path: Introduced attribute path expressions ([#396](https://github.com/hashicorp/terraform-plugin-framework/issues/396))
* path: Introduced framework abstraction for attribute path handling ([#390](https://github.com/hashicorp/terraform-plugin-framework/issues/390))

ENHANCEMENTS:

* diag: Added `Diagnostics` type `Equal()` method ([#402](https://github.com/hashicorp/terraform-plugin-framework/issues/402))
* diag: `ErrorsCount`, `WarningsCount`, `Errors` and `Warnings` functions have been added to `diag.Diagnostics` ([#392](https://github.com/hashicorp/terraform-plugin-framework/issues/392))
* providerserver: Added sdk.proto logger request duration and response diagnostics logging ([#398](https://github.com/hashicorp/terraform-plugin-framework/issues/398))
* tfsdk: Added `AttributePathExpression` field to `ModifyAttributePlanRequest` and `ValidateAttributeRequest` types ([#396](https://github.com/hashicorp/terraform-plugin-framework/issues/396))
* tfsdk: Added `PathMatches` method to `Config`, `Plan`, and `State` types ([#396](https://github.com/hashicorp/terraform-plugin-framework/issues/396))
* tfsdk: Added framework-specific error diagnostics when `Resource` implementations errantly return no errors and empty state after `Create` and `Update` methods ([#406](https://github.com/hashicorp/terraform-plugin-framework/issues/406))
* types: Method `IsNull()` for `Number` type will now return true if the struct is zero-value initialized. ([#384](https://github.com/hashicorp/terraform-plugin-framework/issues/384))

# 0.9.0 (June 15, 2022)

BREAKING CHANGES:

* attr: The `Value` interface now includes the `IsNull()` and `IsUnknown()` methods ([#335](https://github.com/hashicorp/terraform-plugin-framework/issues/335))
* attr: The `Value` interface now includes the `String()` method ([#376](https://github.com/hashicorp/terraform-plugin-framework/issues/376))
* tfsdk: `ListNestedAttributes`, `SetNestedAttributes` and `MapNestedAttributes` functions lost the second argument `opts`, as it was unused. ([#349](https://github.com/hashicorp/terraform-plugin-framework/issues/349))

FEATURES:

* providerserver: Implemented native protocol version 5 support ([#368](https://github.com/hashicorp/terraform-plugin-framework/issues/368))

ENHANCEMENTS:

* providerserver: Added `NewProtocol5()` and `NewProtocol5WithError()` functions, which return a protocol version 5 compatible provider server ([#368](https://github.com/hashicorp/terraform-plugin-framework/issues/368))
* providerserver: Added `ServeOpts` type `ProtocolVersion` field, which can be set to `5` or `6` and defaults to `6` ([#368](https://github.com/hashicorp/terraform-plugin-framework/issues/368))
* tfsdk: New function `ValueFrom` that takes a Go value and populates a compatible `attr.Value`, given a descriptive `attr.Type`. ([#350](https://github.com/hashicorp/terraform-plugin-framework/issues/350))
* tfsdk: Removed `ListNestedAttributesOptions`, `SetNestedAttributesOptions` and `MapNestedAttributesOptions` types, as they were empty (no fields) and unused. ([#349](https://github.com/hashicorp/terraform-plugin-framework/issues/349))
* types: Added `IsNull()` and `IsUnknown()` methods to all types ([#335](https://github.com/hashicorp/terraform-plugin-framework/issues/335))
* types: Added `String()` method to all types ([#376](https://github.com/hashicorp/terraform-plugin-framework/issues/376))

BUG FIXES:

* tfsdk: Prevented configuration handling error when `Schema` contained `Blocks` ([#371](https://github.com/hashicorp/terraform-plugin-framework/issues/371))
* types: Prevented panic being thrown when `.ToTerraformValue` is called on an `attr.Value` type where `ElemType / AttrsType` were not set. ([#354](https://github.com/hashicorp/terraform-plugin-framework/issues/354))
* types: Prevented potential loss of number precision with `Int64` between 54 and 64 bits ([#325](https://github.com/hashicorp/terraform-plugin-framework/issues/325))

# 0.8.0 (May 6, 2022)

BREAKING CHANGES:

* diag: Removed `Diagnostics` type `ToTfprotov6Diagnostics()` method. This was not intended for usage by provider developers. ([#313](https://github.com/hashicorp/terraform-plugin-framework/issues/313))
* tfsdk: The `ModifySchemaPlanRequest`, `ModifySchemaPlanResponse`, `ValidateSchemaRequest`, and `ValidateSchemaResponse` types have been removed. These were not intended for provider developer usage. ([#310](https://github.com/hashicorp/terraform-plugin-framework/issues/310))
* tfsdk: The `NewProtocol6Server()` function, `Serve()` function, and `ServeOpts` type have been removed. Use the `providerserver` package instead. ([#310](https://github.com/hashicorp/terraform-plugin-framework/issues/310))
* tfsdk: The `ResourceImportStateNotImplemented()` function has been removed. Remove the `Resource` type `ImportState` method instead for resources that should not support import. ([#312](https://github.com/hashicorp/terraform-plugin-framework/issues/312))

ENHANCEMENTS:

* tfsdk: Propagated `tf_data_source_type`, `tf_req_id`, `tf_resource_type`, and `tf_rpc` fields in log entries ([#315](https://github.com/hashicorp/terraform-plugin-framework/issues/315))

BUG FIXES:

* all: Prevented `This log was generated by an SDK subsystem logger that wasn't created before being used.` warning messages in logging ([#314](https://github.com/hashicorp/terraform-plugin-framework/issues/314))
* tfsdk: Prevented `Unable to create logging subsystem with AdditionalLocationOffset due to missing root logger options` warning logs during acceptance testing ([#315](https://github.com/hashicorp/terraform-plugin-framework/issues/315))

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
