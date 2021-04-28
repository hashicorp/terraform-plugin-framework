# Instantiating New Provider-Controlled Types

Historically, the framework has always tried to use providers, data sources,
and resources that are scoped to the RPC server; you can see this in the SDKv2
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

### Anonymous function types

Wherever we need instances of provider, resource, and data source types from
provider developers, we can ask for them as the results of anonymous function
types.  For example:

```go
type Provider struct {
  Resources map[string]func() framework.Resource
}
```

This is asking for a function that returns a `framework.Resource` when called,
which lets the framework call the function to instantiate a new instance of the
type. This works whether the type is defined by the provider (the function
would return an interface) or by the framework.

### Named function types

Similarly to anonymous function types, we can ask for functions with a named
function type:

```go
type ResourceFactory func() framework.Resource

type Provider struct {
  Resources map[string]ResourceFactory
}
```

This also asks for a function that returns a `framework.Resource`, but gives
the function signature a name. This works whether the resource type is defined
by the provider (the function would return an interface) or by the framework.

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
point to it if the type is defined by the framework, as the framework can just
use manual copying to get the same result.

### Manual copying

If the Resource type is owned by the framework, it will be able to instantiate
its own instances of that type, and would be able to copy over the data used to
populate them:

```go
resource := framework.Resource{
  Schema: map[string]*Schema{},
  Create: createFunc,
  Read: readFunc,
  Update: updateFunc,
  Delete: deleteFunc,
}

newResource := framework.Resource{}
newResource.Schema = resource.Schema
newResource.Create = resource.Create
newResource.Read = resource.Read
newResource.Update = resource.Update
newResource.Delete = resource.Delete
```

we can use this in a helper, to do something like:

```go
type Provider struct {
  Resources map[string]Resource
}
```

and do the instantiation ourselves behind the scenes.

This works only if the type is defined by the framework. Types defined by the
provider would need to use reflection to achieve this outcome.

### Factory types

Instead of using functions, the framework can define an interface for a factory
that can be implemented by a provider-defined type:

```go
type ResourceFactory interface {
  NewResource() Resource
}
```

Which can then be used instead of the type the factory will return when asking
consumers for resources, data sources, and providers:

```go
type Provider struct {
  Resources map[string]ResourceFactory
}
```

This works whether the `Resource` type is defined by the provider (the
`NewResource()` method would return an interface) or by the framework.

### Separating resource types and resource instances

The SDK currently conflates two separate ideas: the type of a resource--like
"random_pet", the resource's name and schema, not the Go type--and a specific
instance of that resource type--like "random_pet.my_resource", a set of
concrete values filled into the resource type, a single state entry.

Rather than conflating resource types and instances, we can separate them out
into two different Go implementations. The resource type can then serve as a
factory that contains information common to all instances, and instances can
surface the implementations that are used to operate only on instances of a
resource type:

```go
type ResourceType interface {
  GetSchema() *tfprotov5.Schema
  NewValue() Resource
}

type Resource interface {
  Create(ctx context.Context, *CreateResourceRequest, *CreateResourceResponse)
  Read(ctx context.Context, *ReadResourceRequest, *ReadResourceResponse)
  Update(ctx context.Context, *UpdateResourceRequest, *UpdateResourceResponse)
  Delete(ctx context.Context, *DeleteResourceRequest, *DeleteResourceResponse)
}
```

We can then use resource types to instantiate new resource instances at
runtime:

```go
type Provider struct {
  Resources map[string]ResourceType
}
```

This works whether the type is defined by the provider (the `NewValue()` method
would return an interface type instead of a struct that the framework defined)
or by the framework (the `NewValue()` method would return a struct type).

## Trade-offs

### Mutability of data

As a general rule of thumb, resources, data sources, and providers should not
change at runtime. It's hard to imagine a scenario where modifying the schema,
validation, or CRUD implementations while the server is running is a good idea,
and we should consider that scenario out of scope for this design.

Anonymous and named function types, factory types, and separating resource
types and instances may give the impression that this is possible or supported,
by changing what is returned by the function based on some runtime
considerations.

Reflection and manual copying do not give the impression that this is possible
or supported, as the creation of new values is abstracted from the provider
developer and they may not even know it's happening.

### Type system gymnastics

Anonymous and named function types require the user to specify a function that
returns a resource, with a usually-hardcoded implementation inside the
function. The user may not understand the purpose of the function, and may
consider it extra verbosity. Additionally, depending on how it is used (as a
value in a map, etc.) the provider developer _may_ need to cast their function
implementation to the correct type if we use named function types, which is
verbose and annoying.

Reflection and manual copying require no extra type gymnastics other than the
minimal viable work of defining the resource, which can then be used as a
stamp.

Factory types require the provider developer to define an entire type, likely
with no state of its own, just to implement a method on it, just to return a
static resource definition.

Separating resource types and instances requires the provider developer to
define an entire type, just like factory types, but it perhaps _feels_ less
like type system gymnastics as it also bundles in the schema information with
the factory, providing some separation between the type of resource and an
instance of the resource at runtime; it feels less like we're doing type
gymnastics and more like we're faithfully surfacing a distinction Terraform
makes.

### Reliability

Anonymous and named function types, factory types, and separating resource
types and instances are all straightforward implementations, lean heavily on
the Go compiler, and are relatively reliable as implementation patterns.

Reflection circumvents the Go compiler and has a lot of sharp corner cases to
it, which we may or may not have enough experience to predict, and is
relatively unreliable as an implementation pattern.

Manually copying is a more reliable alternative, compared to reflection, that
yields the same outcome, though the subtlety of things like pointers and slices
in that situation still makes it a less reliable implementation than the other
options, above. It also creates maintenance overhead, as we'll need to remember
to update the copying implementation every time the struct changes.

### Type safety

Anonymous and named function types, factory types, manually copying, and
separating resource types and instances are all type-safe implementations,
working within the Go compiler and its type system.

Reflection circumvents the Go compiler and its type system, and is not a
type-safe implementation.

### Documentation

Named function types, factory types, and separating resource types and
instances all share documentation properties. They can have the purpose of the
function defined explicitly and clearly (positive) but that definition is
likely to be at a distance in the documentation from the types that use it
(negative).

Manually copying and reflection have no special types or outward indication
that the process is happening, meaning there's nowhere to hang documentation
off of except where they're used, which is repetitive (negative); but also
there's not much purpose for that documentation (positive)--assuming the
implementation works correctly.

Anonymous function types likewise have nowhere to hang documentation off of
besides where they're used, which is repetitive (negative).

### Automation

For automation and code-analysis purposes, factory types and separating
resource types and instances are the most friendly, as their intent is explicit
and checked by the compiler. Named function types are the next-most
automatable, as the intent of where the function is used is explicit, but the
definition of the function itself does not have any intent associated with it.
Reflection, manual copying, and anonymous function types can only have their
intent inferred by the name of the property they're set on, which is the
hardest to build automation around.

[sdkv2-provider-func]: https://github.com/hashicorp/terraform-plugin-sdk/blob/893e7238350e1980eb2cce3303689ba59ae47490/plugin/serve.go#L28
[sdkv2-resource-func-call]: https://github.com/hashicorp/terraform-provider-scaffolding/blob/243ba4948171e3902003f678c7c43ec3fafcdc20/internal/provider/provider.go#L33
[sdkv2-resource-schema]: https://github.com/hashicorp/terraform-provider-scaffolding/blob/243ba4948171e3902003f678c7c43ec3fafcdc20/internal/provider/resource_scaffolding.go#L10-L29
[structs-interfaces]: https://github.com/hashicorp/terraform-plugin-framework/blob/main/docs/design/structs-interfaces.md
