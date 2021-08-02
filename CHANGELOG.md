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
