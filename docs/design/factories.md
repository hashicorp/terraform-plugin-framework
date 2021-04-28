# Instantiating New Provider-Controlled Types

Historically, the framework has always tried to use providers, data sources,
and resources that are scoped to the RPC server; you can see this in the SDKv2
[`ProviderFunc`][sdkv2-provider-func] type, which is used to instantiate a
provider when the server starts up. But you can also see it in the practice of
[using a function][sdkv2-resource-func-call] to [register resource
schemas][sdkv2-resource-schema-usage], even though there's no enforced
requirement that providers do this. This keeps the variables representing
providers, data sources, and resources scoped to the gRPC server, which is very
helpful when running multiple servers at the same time, as in the acceptance
test drivers.

There are two main goals when instantiating new instances of a type: to provide
isolation from other instances of that type that may be running (otherwise, a
global can just be used) and to (sometimes) allow the provider to populate some
values on that type (e.g., filling in a schema).

There are a few ways these values can be instantiated, and some of the details
are dependent on whether the types in question are [framework-controlled or
provider-controlled][structs-interfaces].

This document is meant to catalogue our available options and explore their
benefits and trade-offs. It, like all these design documents, is meant to
explore only its specific scope; orthogonal issues are handled in separate
design documents.

**Note**: there are many code samples below. Code samples are meant to be
illustrative of patterns; the specific types of arguments and return values,
the names of things, etc. aren't meant to be part of the proposal. The purpose
of the code is just to illustrate the pattern, not to suggest a final
implementation that will be used. Code will be PRed, and these and other
concerns that don't impact the pattern can be resolved at that time.

## Options

### Anonymous function types

Wherever we need instances of provider, resource, and data source types from
provider developers, we can ask for them as the results of anonymous function
types.  For example:

```go
// in the framework
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
// in the framework
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
// in the provider
resource := computeInstanceResource{}

typ := reflect.TypeOf(resource)
newResource := reflect.New(typ)

// newResource is now a newly instantiated variable of the same type as
// resource
```

We can use this in a helper, to do something like:

```go
// in the framework
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
// in the provider
resource := framework.Resource{
  Schema: map[string]*Schema{},
  Create: createFunc,
  Read: readFunc,
  Update: updateFunc,
  Delete: deleteFunc,
}
```

```go
// in the framework
newResource := framework.Resource{}
newResource.Schema = resource.Schema
newResource.Create = resource.Create
newResource.Read = resource.Read
newResource.Update = resource.Update
newResource.Delete = resource.Delete
```

we can use this in a helper, to do something like:

```go
// in the framework
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
// in the framework
type ResourceFactory interface {
  NewResource() Resource
}
```

Which can then be used instead of the type the factory will return when asking
consumers for resources, data sources, and providers:

```go
// in the framework
type Provider struct {
  Resources map[string]ResourceFactory
}
```

This works whether the `Resource` type is defined by the provider (the
`NewResource()` method would return an interface) or by the framework.

### Separating resource types and resource instances

Resources are currently treated as a single logical concept: they have a
schema, and they have CRUD functions, and the same type is used to implement
both of these.

There are, however, two underlying concepts that are surfaced as "resources"
right now: resource types and resource instances.

Resource types are the resource in abstract form. `random_pet` is a resource
type. It has a schema, but no config, state, or plan. It doesn't show up in a
practitioner's configuration files at all. It has no lifecycle.

Resource instances are the resource in concrete form. `random_pet.my_pet` is a
resource instance. It has a schema, but also has a config, state, and plan. It
shows up in the pracitioner's configuration files. It has a lifecycle.

At the moment, both of these concepts are surfaced as a single `Resource` type.
This leads to two problems:

First, `helper/schema` uses [a single instance][sdkv2-resource-registration] of
the `Resource` type for all RPC calls. _If_ we use a provider-defined type for
resources, this may lead providers to try and store information generated
during RPC calls in their `framework.Resource` implementation:

```go
// in the provider
type myResource struct {
  readResult tftypes.Value
}

func (m *myResource) Read(ctx context.Context, req framework.ReadResourceRequest, resp framework.ReadResourceResponse) {
  // fetch state from the API here
  // this next line assumes the state from the API is "hello, world"
  // this is unlikely, but sufficient to illustrate the point
  m.readResult = tftypes.NewValue(tftypes.String, "hello, world")
}

func (m *myResource) Create(ctx context.Context, req framework.CreateResourceRequest, resp framework.CreateResourceResponse) {
  var readResult string
  err := m.readResult.As(&readResult)
  if err != nil {
    panic(err)
  }
  // make an API call here using readResult
}
```

This code could _sometimes_ work if we're not careful about always generating a
new `myResource` for each RPC call we handle. But if the inner workings of the
SDK or Terraform's graph change in any way, it's likely to break this code,
which may not be obvious to provider developers. We can mitigate this by
requiring provider developers to register a function with the provider, not a
value, and always calling the function to get a fresh value at the beginning of
every RPC call, though provider developers may not understand that we're doing
that.

Second, there exists a certain kind of state that it's very reasonable for
providers to want to have available to all RPC calls for every instance of
their resources. This state usually is _used_ by RPC calls, not _created_ by
it. An example of this we see a lot in the wild is a mutex that constrains the
number of requests that can be made in parallel, to not provoke API rate
limiting. Currently, the only way to keep this state is to register it as
global mutable state. This is problematic in testing scenarios, as all provider
servers will need to share that same state; it's not just one server's resource
instances, it's all the resource instances for all the servers created by any
concurrently-running tests.

We have the option of surfacing this distinction explicitly, allowing providers
to store this state that should be shared among all instances of a resource
type. We could define the resource type and the resource instance as separate
Go types:

```go
// in the framework
type ResourceType interface {
  GetSchema() *tfprotov5.Schema
  NewResource(p framework.Provider) Resource
}

type Resource interface {
  Create(ctx context.Context, *CreateResourceRequest, *CreateResourceResponse)
  Read(ctx context.Context, *ReadResourceRequest, *ReadResourceResponse)
  Update(ctx context.Context, *UpdateResourceRequest, *UpdateResourceResponse)
  Delete(ctx context.Context, *DeleteResourceRequest, *DeleteResourceResponse)
}
```

This allows provider developers to define their resource types and thread
through the state shared by all instances of the resource:

```go
// in the provider
type myResourceType struct {
  reqMutex sync.Mutex
}

func (m *myResourceType) NewResource(p framework.Provider) framework.Resource{
  return &myResource{
    reqMutex: &m.reqMutex,
    client: p.(*Client),
  }
}

func (m *myResourceType) GetSchema() *tfprotov5.Schema {
  return &tfprotov5.Schema{
    // hard-code schema here
  }
}

type myResource struct {
  reqMutex sync.Mutex
  client *Client
}

func (m *myResource) Create(ctx, req, resp) {
  // this ensures that only one Create call can happen at a time
  m.reqMutex.Lock()
  defer m.reqMutex.Unlock()

  // make API call
}
```

We can then use resource types to instantiate new resource instances at
runtime:

```go
// in the framework
type Provider struct {
  Resources map[string]ResourceType
}
```

And provider developers can instantiate the state they need the resource
instances to share:

```go
// in the provider
func NewProvider() *framework.Provider{
  return &framework.Provider{
    Resources: map[string]framework.ResourceType{
      "my_resource": myResourceType{
        reqMutex: sync.Mutex{},
      },
    },
  }
}
```

This largely requires the `Resource` type to be defined by the provider, as
there's no place to thread the resource-global state through on a
framework-defined `Resource` type. In theory, you could still separate the two
concepts with a framework-defined `Resource` type, but most if not all of the
benefit is lost.

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

### Schema Access

Anonymous functions, named functions, factory types, reflection, and manually
copying all let the provider developer access the resource's schema inside the
CRUD functions of that resource, by just calling the method.

If resource types and resource instances are separated into two Go types,
however, this avenue will no longer be available, and we'll either need to
surface the resource type or its schema in the CRUD functions somehow, or
decide that provider developers don't need access to it.

[sdkv2-provider-func]: https://github.com/hashicorp/terraform-plugin-sdk/blob/893e7238350e1980eb2cce3303689ba59ae47490/plugin/serve.go#L28
[sdkv2-resource-func-call]: https://github.com/hashicorp/terraform-provider-scaffolding/blob/243ba4948171e3902003f678c7c43ec3fafcdc20/internal/provider/provider.go#L33
[sdkv2-resource-schema-usage]: https://github.com/hashicorp/terraform-provider-scaffolding/blob/243ba4948171e3902003f678c7c43ec3fafcdc20/internal/provider/resource_scaffolding.go#L10-L29
[sdkv2-resource-registration]: https://github.com/hashicorp/terraform-plugin-sdk/blob/e512e3737c6c64e51a1bca47aab84f9a90042cc8/helper/schema/provider.go#L63
[structs-interfaces]: https://github.com/hashicorp/terraform-plugin-framework/blob/main/docs/design/structs-interfaces.md
