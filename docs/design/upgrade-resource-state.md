# Upgrade Resource State

A resource schema captures the structure and types of the resource state. Any state data that does not conform to the resource schema will generate errors or not be persisted. Over time, it may be necessary for resources to make updates to their schemas. Terraform supports versioning of these resource schemas and the current version is saved into the Terraform state. When the provider advertises a newer schema version, Terraform will call back to the provider to attempt to upgrade from the saved schema version to the one advertised. This operation is performed prior to planning, but with a configured provider.

## Background

The resource state handling between the Terraform CLI and the provider is fairly transparent to practitioners as it is implemented without any particular user interface. Resource state upgrade operations happen as part of the general planning workflow. Practitioners will only have issues if any potential upgrades are incorrectly implemented, such as mismatched types, which will generate errors and likely require further provider action (e.g. a new release with fixed upgrade) or worst case of manual state manipulation.

The next sections will outline some of the underlying details relevant to implementation proposals in this framework.

### Terraform Plugin Protocol

The specification between Terraform CLI and plugins, such as Terraform Providers, is currently implemented via [Protocol Buffers](https://developers.google.com/protocol-buffers). Highlighted below are some of the service `rpc` (called by the Terraform CLI) and `message` types that are integral for upgrade resource state support.

#### `UpgradeResourceState` RPC

```protobuf
service Provider {
    // ...
    rpc UpgradeResourceState(UpgradeResourceState.Request) returns (UpgradeResourceState.Response);
}

message UpgradeResourceState {
    message Request {
        string type_name = 1;

        // version is the schema_version number recorded in the state file
        int64 version = 2;

        // raw_state is the raw states as stored for the resource.  Core does
        // not have access to the schema of prior_version, so it's the
        // provider's responsibility to interpret this value using the
        // appropriate older schema. The raw_state will be the json encoded
        // state, or a legacy flat-mapped format.
        RawState raw_state = 3;
    }
    message Response {
        // new_state is a msgpack-encoded data structure that, when interpreted with
        // the _current_ schema for this resource type, is functionally equivalent to
        // that which was given in prior_state_raw.
        DynamicValue upgraded_state = 1;

        // diagnostics describes any errors encountered during migration that could not
        // be safely resolved, and warnings about any possibly-risky assumptions made
        // in the upgrade process.
        repeated Diagnostic diagnostics = 2;
    }
}
```

### terraform-plugin-go

The [`terraform-plugin-go` library](https://pkg.go.dev/hashicorp/terraform-plugin-go) is a low-level implementation of the [Terraform Plugin Protocol](#terraform-plugin-protocol) in Go and underpins this framework. This includes packages such as `tfprotov6` and `tftypes`. These are mentioned for completeness as some of these types are not yet abstracted in this framework and may be shown in implementation proposals.

### terraform-plugin-framework

Most of the Go types and functionality from `terraform-plugin-go` will be abstracted by this framework before reaching provider developers. The details represented here are not finalized as this framework is still being designed, however these current details are presented here for additional context in the later proposals.

Managed resources are currently implemented in the `ResourceType` and `Resource` Go interface types. Provider implementations are responsible for implementing these as concrete Go types.

```go
// A ResourceType is a type of resource. For each type of resource this provider
// supports, it should define a type implementing ResourceType and return an
// instance of it in the map returned by Provider.GetResources.
type ResourceType interface {
    // GetSchema returns the schema for this resource.
    GetSchema(context.Context) (Schema, diag.Diagnostics)

    // NewResource instantiates a new Resource of this ResourceType.
    NewResource(context.Context, Provider) (Resource, diag.Diagnostics)
}

// Resource represents a resource instance. This is the core interface that all
// resources must implement.
type Resource interface {
    // Create is called when the provider must create a new resource. Config
    // and planned state values should be read from the
    // CreateResourceRequest and new state values set on the
    // CreateResourceResponse.
    Create(context.Context, CreateResourceRequest, *CreateResourceResponse)

    // Read is called when the provider must read resource values in order
    // to update state. Planned state values should be read from the
    // ReadResourceRequest and new state values set on the
    // ReadResourceResponse.
    Read(context.Context, ReadResourceRequest, *ReadResourceResponse)

    // Update is called to update the state of the resource. Config, planned
    // state, and prior state values should be read from the
    // UpdateResourceRequest and new state values set on the
    // UpdateResourceResponse.
    Update(context.Context, UpdateResourceRequest, *UpdateResourceResponse)

    // Delete is called when the provider must delete the resource. Config
    // values may be read from the DeleteResourceRequest.
    Delete(context.Context, DeleteResourceRequest, *DeleteResourceResponse)

    // ImportState is called when the provider must import the resource.
    //
    // If import is not supported, it is recommended to use the
    // ResourceImportStateNotImplemented() call in this method.
    //
    // If setting an attribute with the import identifier, it is recommended
    // to use the ResourceImportStatePassthroughID() call in this method.
    ImportState(context.Context, ImportResourceStateRequest, *ImportResourceStateResponse)
}
```

The existing `Schema` type also has a placeholder `Version` field, which will update the saved state:

```go
// Schema is used to define the shape of practitioner-provider information,
// like resources, data sources, and providers. Think of it as a type
// definition, but for Terraform.
type Schema struct {
    // ... other fields ...

    // Version indicates the current version of the schema. Schemas are
    // versioned to help with automatic upgrade process. This is not
    // typically required unless there is a change in the schema, such as
    // changing an attribute type, that needs manual upgrade handling.
    // Versions should only be incremented by one each release.
    Version int64
}
```

Since the underlying [`tfprotov6.ResourceServer`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-go/tfprotov6#ResourceServer) interface requires an `UpgradeResourceState` implmentation, the framework currently implements a stub implementation that ensures:

- The resource type exists in this provider.
- That given state in the request is passed back in the response (e.g. as a passthrough).

The implementation proposals are to handle schema versioning and allow provider developers to adjust the state between each version.

## Prior Implementations

### terraform-plugin-sdk

The previous framework for provider implementations, Terraform Plugin SDK, can be found in the `terraform-plugin-sdk` repository. That framework has existed since the very early days of Terraform, where it was previously contained in a combined CLI and provider codebase, to support the code and testing aspects of provider development.

To implement managed resources and data sources, the previous framework was largely based around Go structure types and declarative definitions of intended behaviors. These were defined in the `helper/schema` package, in particular, the `Resource` type. The relevant fields are shown below.

```go
type Resource struct {
    // ...

    // SchemaVersion is the version number for this resource's Schema
    // definition. The current SchemaVersion stored in the state for each
    // resource. Provider authors can increment this version number
    // when Schema semantics change. If the State's SchemaVersion is less than
    // the current SchemaVersion, the InstanceState is yielded to the
    // MigrateState callback, where the provider can make whatever changes it
    // needs to update the state to be compatible to the latest version of the
    // Schema.
    //
    // When unset, SchemaVersion defaults to 0, so provider authors can start
    // their Versioning at any integer >= 1
    SchemaVersion int

    // MigrateState is responsible for updating an InstanceState with an old
    // version to the format expected by the current version of the Schema.
    //
    // It is called during Refresh if the State's stored SchemaVersion is less
    // than the current SchemaVersion of the Resource.
    //
    // The function is yielded the state's stored SchemaVersion and a pointer to
    // the InstanceState that needs updating, as well as the configured
    // provider's configured meta interface{}, in case the migration process
    // needs to make any remote API calls.
    //
    // Deprecated: MigrateState is deprecated and any new changes to a resource's schema
    // should be handled by StateUpgraders. Existing MigrateState implementations
    // should remain for compatibility with existing state. MigrateState will
    // still be called if the stored SchemaVersion is less than the
    // first version of the StateUpgraders.
    MigrateState StateMigrateFunc

    // StateUpgraders contains the functions responsible for upgrading an
    // existing state with an old schema version to a newer schema. It is
    // called specifically by Terraform when the stored schema version is less
    // than the current SchemaVersion of the Resource.
    //
    // StateUpgraders map specific schema versions to a StateUpgrader
    // function. The registered versions are expected to be ordered,
    // consecutive values. The initial value may be greater than 0 to account
    // for legacy schemas that weren't recorded and can be handled by
    // MigrateState.
    StateUpgraders []StateUpgrader
}

type StateMigrateFunc func(int, *terraform.InstanceState, interface{}) (*terraform.InstanceState, error)

type StateUpgrader struct {
    // Version is the version schema that this Upgrader will handle, converting
    // it to Version+1.
    Version int

    // Type describes the schema that this function can upgrade. Type is
    // required to decode the schema if the state was stored in a legacy
    // flatmap format.
    Type cty.Type

    // Upgrade takes the JSON encoded state and the provider meta value, and
    // upgrades the state one single schema version. The provided state is
    // deocded into the default json types using a map[string]interface{}. It
    // is up to the StateUpgradeFunc to ensure that the returned value can be
    // encoded using the new schema.
    Upgrade StateUpgradeFunc
}

// See StateUpgrader
type StateUpgradeFunc func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error)
```

An example provider implementation:

```go
func exampleResource() *schema.Resource {
    return &schema.Resource{
        // ... current schema ...

        SchemaVersion: 1,
        StateUpgraders: []schema.StateUpgrader{
            {
                Type:    exampleResourceSchemaV0().CoreConfigSchema().ImpliedType(),
                Upgrade: exampleResourceUpgradeV0,
                Version: 0,
            },
        },
    }
}

func exampleResourceSchemaV0() *schema.Resource {
    return &schema.Resource{
        Schema: map[string]*schema.Schema{
            // ... previous schema ...
        },
    }
}

func exampleResourceUpgradeV0(_ context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
    rawState["example_attribute"] = false

    return rawState, nil
}
```

## Goals

Upgrade resource support in this framework should be:

- Able to support diagnostics, cancellation context, and the provider instance.
- Available as exported functionality for provider developers.
- Abstracted from terraform-plugin-go and convertable into those types to separate implementation concerns.
- Ergonomic to implement Go code (e.g. have helper methods for common use cases).

Additional consideration should be given to:

- Whether provider implementations can use partial schemas and state information.
- Whether schema versioning should be implemented within the existing resource design, outside of it, or wrap it.
- Whether the framework can raise helpful errors when there is missing resource information.

## Proposals

### Additional Attribute and Block Fields

The framework could introduce additional fields to the `Attribute` and `Block` types, which signal what to do with schema upgrades. This would allow a singular `Schema` to be a single source of truth for state data over time.

For example in the framework:

```go
// Existing type
type Attribute struct {
    // ... existing fields ...

    // New field
    // If prior version matches during UpgradeResourceState, do something.
    // The request and response types would need to allow get/set on whole states
    StateUpgrades map[int64]func(context.Context, AttributeStateUpgradeRequest, *AttributeStateUpgradeResponse)
}
```

However, there is an immediate drawback that there would be no access to the provider client during upgrade state operations. This is considered a non-starter due to the goals of this functionality. There are also quite a few additional drawbacks to this type of implementation including an ever growing `Schema` size over time, when (if ever) it might be safe to remove state upgrades, and no ability to control any potential ordering considerations across multiple attributes.

### New VersionedResourceType Interface

The framework could wrap the existing `ResourceType` type with a versioning interface, then requiring that provider developers fully define every version of a `ResourceType` and `Resource` when implementing a managed resource. The `Version` field of the `Schema` type would be removed to prevent conflicting code.

For example:

```go
// New type
type VersionedResourceType interface {
    GetResourceTypeVersions(context.Context) map[int64]ResourceType
}

// Existing type
type Provider interface {
    // Required method update from map[string]ResourceType
    GetResources(context.Context) (map[string]VersionedResourceType, diag.Diagnostics)
}

// Existing type
type Resource interface {
    // New required method
    // If there is no previous version, a blank implementation would be required.
    UpgradeState(context.Context, /*...*/) /*...*/
}
```

This would ensure framework resources:

- Contain all schema and logic necessary over time to perform any resource state upgrades, which simplifies the framework implementation.
- Do not require additional provider developer discovery for resource versioning features.
- Have access to the provider client during upgrade state operations.

However, there are quite a few drawbacks to this type of implementation:

- Additional and generally unnecessary complexity for most provider developers.
- Lots of code build up over time unless provider developers find ways to reduce it.
- Existing framework provider implementations would require updates to `GetResources` and each `Resource`.
- Confusing blank `(Resource).UpgradeState()` implementation for single version resources. It could be made as an optional method, but then provider developers need to discover it and know when to use it, removing a main benefit of this type of approach.

### New ResourceTypeWithUpgradeState Interface

The framework could allow provider developers to optionally extend their `ResourceType` with a new `UpgradeState` method.

For example in the framework:

```go
// New type
type ResourceTypeWithUpgradeState interface {
    UpgradeState(context.Context, /*...*/) /*...*/
}
```

Which would result in the following optional provider implementation:

```go
func (rt ExampleResourceType) UpgradeState(ctx context.Context, /*...*/) /*...*/ {
    /* ... */
}
```

However, there is an immediate drawback that there would be no access to the provider client during upgrade state operations. This is considered a non-starter due to the goals of this functionality.

### New ResourceWithUpgradeState Interface

The framework could allow provider developers to optionally extend their `Resource` with a new `UpgradeState` method.

For example in the framework:

```go
// New type
type ResourceWithUpgradeState interface {
    UpgradeState(context.Context, /*...*/) /*...*/
}
```

Which would result in the following optional provider implementation:

```go
func (r ExampleResource) UpgradeState(ctx context.Context, /*...*/) /*...*/ {
    /* ... */
}
```

There are considerations for the possible method parameters and returns, which is why they are omitted above and they will be discussed below.

This would ensure framework resources:

- Have access to the provider client during upgrade state operations.
- Are simpler in implementation until the additional complexity is necessary.

However, there are some drawbacks:

- If `Schema.Version` is greater than `0`, there are no compile time errors if the `UpgradeState` method is not defined. Framework-defined unit testing may be possible to ensure that if the `Schema.Version` returned from `(ResourceType).GetSchema()` is greater than `0`, that `(ResourceType).NewResource()` returns a type that supports the `ResourceWithUpgradeState` interface, however it would require provider developers to ensure they implement the additional unit testing.

#### Direct Request and Response Parameters

The framework could require provider developers to implement the `UpgradeState` method signature similar to other `Resource` functionality, such as `Create`, `Read`, etc.

For example in the framework:

```go
// New type
type ResourceWithUpgradeState interface {
    UpgradeState(context.Context, UpgradeResourceStateRequest, *UpgradeResourceStateResponse)
}

// New type
type UpgradeResourceStateRequest struct {
    // Current state version
    // May be worth renaming to CurrentStateVersion in the actual implementation
    Version int64

    // JSON encoded or flatmap state
    // The type would likely be abstracted in the framework, but showing for brevity.
    RawState tfprotov6.RawState
}

// New type
type UpgradeResourceStateResponse struct {
    // Upgraded state, which must be sent as msgpack
    // The underlying *tftypes.DynamicValue type is very unfriendly for
    // provider developers at this layer so a next option is to rely on
    // the existing State type.
    State State

    Diagnostics diag.Diagnostics
}
```

Which would result in the following optional provider implementation:

```go
func (r ExampleResource) UpgradeState(context.Context, req tfsdk.UpgradeResourceStateRequest, *tfsdk.UpgradeResourceStateResponse) {
    // potentially branching logic based on req.Version
    // complex logic to convert req.RawState into the resp.State
}
```

In this setup, the framework could only implement a thin abstraction over the request data since it has no previous schema information. One framework workaround to remove this hurdle for provider developers might be to introduce additional method on `ResourceType` or `Resource` to fetch previous schema.

```go
// New type
type UpgradeResourceStateRequest struct {
    // Current state version
    // May be worth renaming to CurrentStateVersion in the actual implementation
    Version int64

    // JSON encoded or flatmap state
    // The type would likely be abstracted in the framework, but showing for brevity.
    RawState tfprotov6.RawState

    // New field that is populated if GetPreviousSchemas returns a Schema
    State State
}

// Option 1: Only on Resource, requiring previous schemas method
// Pro: Colocates all state upgrade information
// Con: Extra provider developer effort, even if not wanting the framework State
// Con: Schema information is split between ResourceType and Resource

// New type
type ResourceWithUpgradeState interface {
    // If completely nil, UpgradeResourceStateRequest.State is never populated
    // For version matches, UpgradeResourceStateRequest.State is populated
    GetPreviousSchemas(context.Context) map[int64]Schema
    UpgradeState(context.Context, UpgradeResourceStateRequest, *UpgradeResourceStateResponse)
}

// Option 2: Optional on Resource
// Pro: Provider developer effort only when interested in framework State
// Either: State upgrade information is semi-colocated on Resource
// Con: Schema information is split between both ResourceType and Resource

// New type
type ResourceWithUpgradeState interface {
    UpgradeState(context.Context, UpgradeResourceStateRequest, *UpgradeResourceStateResponse)
}

// New type
// If not implemented, UpgradeResourceStateRequest.State is never populated
type ResourceWithGetPreviousSchemas interface {
    // For version matches, UpgradeResourceStateRequest.State is populated
    GetPreviousSchemas(context.Context) map[int64]Schema
}

// Option 3: Optional on ResourceType
// Pro: Schema information is all colocated on ResourceType
// Con: State upgrade information is split between Resource and ResourceType

// New type
// If not implemented, UpgradeResourceStateRequest.State is never populated
type ResourceTypeWithGetPreviousSchemas interface {
    // For version matches, UpgradeResourceStateRequest.State is populated
    GetPreviousSchemas(context.Context) map[int64]Schema
}

// New type
type ResourceWithUpgradeState interface {
    UpgradeState(context.Context, UpgradeResourceStateRequest, *UpgradeResourceStateResponse)
}
```

There are some overall benefits to this approach:

- Follows request/response model of many other parts of the framework.
- Ultimate provider implementation flexibility due to relatively thin abstractions.

However, there are some drawbacks:

- Framework behavior around `State` availability happens at a distance.
- Provider developers must handle version selection logic.
- Varying levels of unit testing difficulty, generally leaning harder.

#### Return Map of StateUpgraders

The framework could require provider developers to implement the `UpgradeState` method signature which colocates all state upgrade information.

For example in the framework:

```go
// New type
type ResourceWithUpgradeState interface {
    // Version to state upgrader implementation
    // Prefer a map over a slice as ordering is irrelevant or
    // trying to force an implementation based on slice indexing
    // would be confusing. The framework can return a helpful error
    // if it receives a version not implemented.
    UpgradeState(context.Context) map[int64]ResourceStateUpgrader
}

// New type
// Potentially could be an interface as well, but a concrete type may be
// a better fit to reduce required provider implementation details.
type ResourceStateUpgrader struct {
    // Optionally populate UpgradeResourceStateRequest.State
    Schema *Schema

    StateUpgrader func(context.Context, UpgradeResourceStateRequest, *UpgradeResourceStateResponse)
}

// New type
type UpgradeResourceStateRequest struct {
    // JSON encoded or flatmap state
    // The type would likely be abstracted in the framework, but showing for brevity.
    RawState tfprotov6.RawState

    // Populated if ResourceStateUpgrader.Schema is present
    State State
}

// New type
type UpgradeResourceStateResponse struct {
    State State

    Diagnostics diag.Diagnostics
}
```

Which would result in the following optional provider implementation:

```go
func (r ExampleResource) UpgradeState(ctx context.Context) map[int64]tfsdk.ResourceStateUpgrader {
    return map[int64]tfsdk.ResourceStateUpgrader{
        0: {
            StateUpgrader: func(ctx context.Context, req tfsdk.UpgradeResourceStateRequest, resp tfsdk.UpgradeResourceStateResponse) {
                // logic to handle req.RawState
                // resp.State = ...
            },
        },
        1: {
            Schema: tfsdk.Schema{ /* ... */ },
            StateUpgrader: func(ctx context.Context, req tfsdk.UpgradeResourceStateRequest, resp tfsdk.UpgradeResourceStateResponse) {
                // logic to handle req.State
                // resp.State = ...
            },
        }
    }
}
```

Whether to implement separate methods for state upgrade logic is up to the provider developer.

This would ensure framework resources:

- Succinctly define state upgrade information in one place (it is a provider developer decision whether to add non-framework methods or other coding techniques to fully expand the details).
- Have a standard methodology for handling version selection, with duplicate detection at compile time. Missing versions can have helpful errors raised.
- Have access to the provider instance, if desired.
- Can be unit tested in a relatively straightforward manner.

However, there are some drawbacks:

- With great provider developer flexibility comes the potential lack of conventional coding practices. Choosing a default documentation style would likely lean towards inline definitions of the state upgrade information, unless there is a very strong convention that forms.

## Recommendations

It is recommended to implement this by creating a new `ResourceWithUpgradeState` interface type with a `UpgradeState` method that returns a map of versions to `ResourceStateUpgrader`. The framework will need to additionally implement the ability to create a new `State`, a capability previously not needed. The framework should also provide unit testing functionality to ensure that if the `Schema.Version` returned from `(ResourceType).GetSchema()` is greater than `0`, that `(ResourceType).NewResource()` returns a type that supports the `ResourceWithUpgradeState` interface.

The ability to handle flatmap `RawState` would require some special consideration which can be handled after the more common JSON encoding use case, however the design does preclude implementing this detail later.
