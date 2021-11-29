# Upgrade Resource State

A resource schema captures the structure and types of the resource state. Any state data that is incorrect from the resource schema will generate errors or not be persisted. Over time, it may be necessary for resources to receive updates to the resource schema. Terraform supports versioning of these resource schemas, which is saved into the Terraform state. When the provider advertises a newer schema version, Terraform will call back to the provider to attempt to upgrade from the saved state version to the one advertised. This operation is performed prior to planning, but with a configured provider.

## Background

The resource state handling between the Terraform CLI and the provider is fairly transparent to practitioners as it is implemented without any particular user interface. Resource state upgrade operations happen as part of the general planning workflow. Practitioners will only have issues if any potential upgrades are incorrectly implemented, such as mismatched types, which will generate errors and likely require further provider action (e.g. a new release with fixed upgrade) or worst case of manual state manipulation.

The next sections will outline some of the underlying details relevant to implementation proposals in this framework.

### Terraform Plugin Protocol

The specification between Terraform CLI and plugins, such as Terraform Providers, is currently implemented via [Protocol Buffers](https://developers.google.com/protocol-buffers). Below highlights some of the service `rpc` (called by the Terraform CLI) and `message` types that are intergral for upgrade resource state support.

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
