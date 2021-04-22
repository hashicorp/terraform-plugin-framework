# Instantiating New Provider-Controlled Types

Historically, the provider has always tried to use providers, data sources, and
resources that are scoped to the RPC server; you can see this in the SDKv2
[`ProviderFunc`][sdkv2-provider-func] type, which is used to instantiate a
provider when the server starts up. But you can also see it in the practice of
[using a function][sdkv2-resource-func-call] to [register resource
schemas][sdkv2-resource-schema], even though there's no enforced requirement
that providers do this. This keeps the variables representing providers, data
sources, and resources scoped to the gRPC server, which is very helpful when
running multiple servers at the same time, as in the acceptance test drivers.

There are two main goals when instantiating new instances of a type: to provide
isolation from other instances of that type that may be running (otherwise, a
global can just be used) and to (sometimes) allow the provider to populate some
values on that type (e.g., filling in a schema).

There are a few ways these values can be instantiated, and some of the details
are dependent on whether the types in question are [framework-controlled or
provider-controlled][structs-interfaces].

This document is meant to catalogue our available options and explore their
benefits and trade-offs.

## Options

### Anonymous functions

Wherever we need these types from provider developers, we can ask for them as
the results of anonymous functions. For example:

```go
type Provider struct {
  Resources map[string]func() framework.Resource
}
```

This is asking for a function that returns a `framework.Resource` when called,
which lets the framework call the function to instantiate a new instance of the
type. This works whether the type is defined by the provider (the function
would return an interface) or by the framework.

### Named functions

Similarly to anonymous functions, we can ask for named functions:

```go
type ResourceFactory func() framework.Resource

type Provider struct {
  Resources map[string]ResourceFactory
}
```

This also asks for a function that returns a `framework.Resource`, but gives
the function signature a name. This works whether the type is defined by the
provider (the function would return an interface) or by the framework.

### Reflection

In theory, given a single instance of the type, we can use the `reflect`
package to create new instances of the type:


```go
resource := computeInstanceResource{}

typ := reflect.TypeOf(resource)
newResource := reflect.New(typ)

// newResource is now a newly instantiated variable of the same type as
// resource
```

We can use this in a helper, to do something like:

```go
type Provider struct {
  Resources map[string]Resource
}
```

and do the instantiation ourselves behind the scenes.

This works if the type is defined by the provider, but there's not really any
point to it if the type is defined by the framework, as the framework can then
just use non-reflection instantiation to get to the same result. A separate
initialization callback or other hook would need to be used to allow providers
to define values for this new instance, to set things like the schema and CRUD
functions on it.

### Factory types

Instead of using functions, we can have an interface for a factory type that
providers define:

```go
type ResourceFactory interface {
  NewFactory() Resource
}
```

Which can then be used instead of the type the factory will return when asking
users for resources, data sources, and providers:

```go
type Provider struct {
  Resources map[string]ResourceFactory
}
```

This works whether the type is defined by the provider (the `NewFactory()`
method would return an interface) or by the framework.

### Separating types and values

Rather than conflating types and values, we can separate them out into two
different Go implementations. A type is a factory that contains information
common to values, values are specific instances of a type:

```go
type ResourceType interface {
  GetSchema() *tfprotov5.Schema
  NewValue() ResourceValue
}

type ResourceValue interface {
  Create(ctx context.Context, *CreateResourceRequest, *CreateResourceResponse)
  Read(ctx context.Context, *ReadResourceRequest, *ReadResourceResponse)
  Update(ctx context.Context, *UpdateResourceRequest, *UpdateResourceResponse)
  Delete(ctx context.Context, *DeleteResourceRequest, *DeleteResourceResponse)
}
```

We can then use types to instantiate new values at runtime:

```go
type Provider struct {
  Resources map[string]ResourceType
}
```

This works whether the type is defined by the provider (the `NewValue()` method
would return an interface) or by the framework.

## Trade-offs

### Mutability of data

### Type system gymnastics

### Reliability

### Type safety

### Documentation

[sdkv2-provider-func][https://github.com/hashicorp/terraform-plugin-sdk/blob/893e7238350e1980eb2cce3303689ba59ae47490/plugin/serve.go#L28]
[sdkv2-resource-func-call][https://github.com/hashicorp/terraform-provider-scaffolding/blob/243ba4948171e3902003f678c7c43ec3fafcdc20/internal/provider/provider.go#L33]
[sdkv2-resource-schema][https://github.com/hashicorp/terraform-provider-scaffolding/blob/243ba4948171e3902003f678c7c43ec3fafcdc20/internal/provider/resource_scaffolding.go#L10-L29]
[structs-interfaces][https://github.com/hashicorp/terraform-plugin-framework/blob/main/docs/design/structs-interfaces.md]
