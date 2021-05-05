# Provider-level runtime state

This document explores the several ways that provider-level runtime data could be injected into resource functions.

## Data

The following are examples of provider-level runtime data:

### `TerraformVersion`

The version of Terraform currently running. 

Providers often want to include the running Terraform version in User-Agent headers in HTTP requests. The framework will include helpers for this, either in constructing the User-Agent string only (as in SDKv2's [`schema.Provider.UserAgent`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema#Provider.UserAgent)), or in providing an HTTP client or RoundTripper helper which providers can inject into their own API clients.

`TerraformVersion` is sent from Terraform Core to the provider in the `ConfigureProvider` RPC ([protocol](https://github.com/hashicorp/terraform/blob/d15f7394a19f8f4d604b632df54c1d0cc0c9cc85/docs/plugin-protocol/tfplugin6.0.proto#L237), [tfprotov6 wrapper](https://github.com/hashicorp/terraform-plugin-go/blob/e0e351efc90e60fa583afa6a2a2e3d8231f25e24/tfprotov6/provider.go#L140)). It is not included in any resource lifecycle RPCs such as `ApplyResourceChange`. 

`TerraformVersion` is *framework-owned*, in that the framework, not the provider, is responsible for setting its value. However, the provider should also have access to this value: for example, for constructing a custom User-Agent header.

### Provider configuration (`ConfigureProviderData`)

Providers typically configure an API client, which is used to make API requests in resource CRUD functions, using API tokens and other parameters specified in the provider configuration block or read from environment variables.

Values from the provider configuration block are supplied to the provider in the `ConfigureProvider` RPC. In this document we refer to such data as `ConfigureProviderData` since it is only available after the `ConfigureProvider` RPC, in the understanding that providers typically use this to store a configured API client.

### Provider Metadata

Provider metadata, not to be confused with the `meta` parameter used in `helper/schema` (see History below), is an experimental Terraform feature used for module-specific configuration. 

`ProviderMeta` is included in resource lifecycle RPCs such as `ReadResource` ([protocol](https://github.com/hashicorp/terraform/blob/d15f7394a19f8f4d604b632df54c1d0cc0c9cc85/docs/plugin-protocol/tfplugin6.0.proto#L250), [tfprotov6 wrapper](https://github.com/hashicorp/terraform-plugin-go/blob/e0e351efc90e60fa583afa6a2a2e3d8231f25e24/tfprotov6/resource.go#L142)).

### Other runtime data

While not yet implemented in the Terraform protocol, in future we may want to surface more data to the provider via the `ConfigureProvider` RPC, such as provider address and version.

## History

In `helper/schema`, resource CRUD functions have the following signature:

```go
type CreateContextFunc func(context.Context, *ResourceData, interface{}) diag.Diagnostics
```

The parameters with type `*ResourceData` and `interface{}` contain the runtime state accessible inside the CRUD functions.

### ResourceData

[`ResourceData`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema#ResourceData) is a struct with no exported fields, on which providers call `Get()` or `Set()` to access config and state values.

`ResourceData` is a notoriously unclear and overloaded abstraction. Are the values obtained from `Get()` the values from config, values from state, planned values, or something else? (This problem is outside the scope of the present document, and its solution will most likely involve dedicated `Request` and `Response` objects in the signature of each CRUD function, with unambiguous ways to get state and config values.)

This struct also stores `providerMeta`, which is accessible through a getter. 

### `meta`

The `interface{}` parameter in CRUD function types such as `CreateContextFunc` is conventionally known as `meta`, corresponding to `ConfigureProviderData`, and is  typically used to store a configured API client. Consider the following code from terraform-provider-github ([resource_github_membership.go](https://github.com/integrations/terraform-provider-github/blob/f7f029822d637f08f8460935a2e56f26f9d3eda1/github/resource_github_membership.go#L76...L84)), which implements a resource `Read` function:

```go
func resourceGithubMembershipRead(d *schema.ResourceData, meta interface{}) error {
	err := checkOrganization(meta)
	if err != nil {
		return err
	}

	client := meta.(*Owner).v3client

	orgName := meta.(*Owner).name
```

Note that a type assertion (`meta.(*Owner)`) is necessary before `meta` can be used.


## Design option 1: the `framework.RuntimeData` struct

Suppose the runtime state be represented by a struct defined in the framework:

```go
type RuntimeData struct {
  terraformVersion string
  ConfigureProviderData interface{}
}
```

Here, `ConfigureProviderData` represents the provider-owned data supplied by the `ConfigureProvider` request RPC, which maps to the `meta` parameter in `helper/schema` (see Data and History above).

In future, we may want to add additional framework-owned data to `RuntimeData`, such as provider address and version.

The problem becomes how to make this data available to both the framework code and the provider code.

### Option 1a: CRUD function parameters

Like SDKv2, the framework could include a parameter in CRUD functions:

```go
type CreateFunc func(context.Context, RuntimeData, ResourceCreateRequest, ResourceCreateResponse)
```

The provider could then access runtime data either via an exported field on the `RuntimeData` struct, or a getter. See the design document on [Structs and Interfaces](./structs-interfaces.md) for more information on how this implementation would be impacted by the decision to define `framework.Resource` as a struct or interface.

The framework can introduce new runtime data in a backwards-compatible manner by adding fields or methods to `RuntimeData`.

#### Disadvantages

As in SDKv2, the provider-owned data `ConfigureProviderData` has type `interface{}`, which means its type must be asserted in the provider code every time it is used. This is verbose.

An extra `RuntimeData` parameter in resource CRUD functions may make them harder to unit test.

### Option 1b: Resource or Provider type

`RuntimeData` could be included in the `Resource` or `Provider` type, so that it could be accessed from within provider `Resource` methods (assuming the resource instance contains a reference to the provider instance if data is stored there). 

As detailed in the [Structs and Interfaces](./structs-interfaces.md) design document, this approach would require additional tradeoffs if resources/providers are defined as interfaces in the framework: provider developers may need to implement getters on their resource/provider struct types, or at the very least embed `framework.RuntimeData` in their structs, with the following provider code:

```go
type provider struct {
  framework.RuntimeData
}
```

## Design option 2: `provider.Client`

There are two types of runtime data: provider-owned, and framework-owned. In Design Option 1 above, both are stored in the `RuntimeData` struct, which means that the provider-owned data must have type `interface{}`, since there is no way for the provider to supply a type for that data. Instead, it must assert the type of `RuntimeData.ConfigureProviderData` in a similar way to `meta` in `helper/schema` (see History above).

If we want the provider to be able to define the type of provider-owned runtime data, the provider must be able to define its own struct type `Provider` (see the [Structs and Interfaces](./structs-interfaces.md) design document for why this is not possible if the provider is an instance of a `framework.Provider` struct), on which, for example, the API client could be stored, so the provider code could look like the following:

```go
type provider struct {
  Client *github.Client
}
```

The provider code would need to set this field in the configure func:

```go
func (p *Provider) Configure(ctx context.Context, req *tfprotov5.ConfigureProviderRequest, res *tfprotov5.ConfigureProviderResponse) {
  // use config values from req to configure a client "c" of type github.Client
  
  p.Client = c
  
  return nil
}
```

Then, assuming that the provider instance was available to resource instances, being injected in the resource factory for example, resources could use the configured client during CRUD functions:

```go
func (r myResource) Create(ctx context.Context, req framework.ResourceCreateRequest, resp framework.ResourceCreateResponse) {
  // determine correct parameters "params" for API request from config and state
  
  r.Provider.Client.CreateResource(params)
}
```

The signatures of `Configure` and `Create` functions here are illustrative examples.

### Tradeoffs

This approach has the clear advantage, over Option 1, that the provider API client is now strongly typed. No type assertions are necessary, and it is clear at compile time prior to any resource implementations whether `provider.Client` is of the correct type.

On the other hand, depending on how `Provider` is surfaced to resources, this approach may require providers or the framework to use mutexes to write concurrency-safe code. For example, in order to prevent concurrent access of `provider.Client`, the framework could create an `RWMutex`, locking it before `ConfigureProvider` and unlocking it afterwards, and read-locking/unlocking it before/after every CRUD call. Handling these mutexes in the framework is no great disadvantage, but requiring provider developers to do so would be a significant increase in complexity.

As shown in Option 1, framework-owned runtime data such as `TerraformVersion` could be stored on a `framework.RuntimeData` struct, which is available to the provider code either as a CRUD function parameter or a field on `provider.Provider`, which embeds `framework.RuntimeData`. Corresponding tradeoffs apply. 

## Recommendations

After investigating this issue, we decided that the framework-owned data is,
usually, used for User-Agent generation, which tends to be configured as part
of the client instantiation, which usually happens in ConfigureProvider.
Because of this, it's somewhat rare that this information is even needed in the
CRUD functions at all. This means that the need to pipe this state through
outside the ConfigureProvider RPC is relatively rare, and probably doesn't need
a first-class solution in the framework. To that end, we're going with option
2, without a `framework.RuntimeData` type. Essentially, we're going to ignore
`TerraformVersion` outside the ConfigureProvider RPC, and provider developers
that want to hoist it into their CRUD functions can do so by explicitly
including it in their provider-defined type, with their API client and other
information. This minimizes complexity for most developers, while allowing
developers that need that information in their CRUD functions a path forward.
