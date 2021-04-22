# Provider, resource, and data source schemas

Providers, resources, and data sources have schemas. This document explores the options for how they relate to each other, and the types provider developers use to define them.

This problem is fairly symmetrical with respect to providers, resources, and data sources, so we choose to focus on resource schemas without loss of generality.

## Relationship of schema to resource

In `helper/schema`, `Resource` is a struct type with a field `Schema`, whose type is `map[string]*Schema`.

In the new framework, `framework.Resource` may be a struct type or interface: see the [Structs and Interfaces](./structs-interfaces.md) design doc. 

Since the schema is defined by the provider code but must be accessible from the framework code, this means that if `framework.Resource` is a struct, the schema can either be a field (`framework.Resource.Schema`), or a method (`framework.Resource.GetSchema()`), i.e. we could have
```go
type Resource struct {
  Schema *tfprotov6.Schema
}
```

or
```go
type Resource struct {
  GetSchema func(context.Context) (*tfprotov6.Schema, []*tfprotov6.Diagnostic)
}
```

If `framework.Resource` is an interface, the only option is a method (`framework.Resource.GetSchema()`), since the actual resource struct will be a provider-defined type:
```go
type Resource interface {
  GetSchema(ctx context.Context) (*tfprotov6.Schema, []*tfprotov6.Diagnostic)
}
```

### Field on struct `framework.Resource.Schema`

This approach is simple and discoverable. The schema is a static piece of data (in conventional providers) known at compile time, so it is idiomatic that it be represented as a field on a struct.

While (depending on the type used to represent the schema - see below) it may not be apparent at compile time if a provider developer has forgotten to fill in a value for `Resource.Schema`, it will be obvious after the most basic of testing if this is the case.

### Method `framework.Resource.GetSchema()`

This approach is also fairly simple and discoverable, and offers the option of returning an error or diag if the schema is not available. This may be a desirable feature for providers written using generated code, where generating the schema could throw an error.

### Both/either

A `Resource` struct could have a `Schema` field, and a `GetSchema` field with func type added later if needed. The framework could use whichever is set, either erroring if both are set or using a documented fallback mechanism to prefer one over the other.

## Types

Whatever the relationship of schema to resource, it is structured data that must be represented by a Go struct or map type. This section details two options for this: the existing `Schema` type in `tfprotov6`, and a proposed `framework.Schema` type.

### `tfprotov6.Schema`

The `tfprotov6` (and `tfprotov5`) package in terraform-plugin-go provides a `Schema` type and related types: https://pkg.go.dev/github.com/hashicorp/terraform-plugin-go/tfprotov6#Schema

Resource schemas defined using these types can be quite verbose. For example, consider the following implementation of a resource with one attribute and one nested block, adapted from the `sql_migrate` resource from [terraform-provider-sql](https://github.com/paultyng/terraform-provider-sql/blob/main/internal/provider/data_query.go):

```go
var schema = &tfprotov5.Schema{
	Block: &tfprotov5.SchemaBlock{
		BlockTypes: []*tfprotov5.SchemaNestedBlock{
			{
				TypeName: "migration",
				Nesting:  tfprotov5.SchemaNestedBlockNestingModeList,
				Block: &tfprotov5.SchemaBlock{
					Attributes: []*tfprotov5.SchemaAttribute{
						{
							Name:            "id",
							Required:        true,
							Description:     "Identifier can be any string to help identifying the migration in the source.",
							DescriptionKind: tfprotov5.StringKindMarkdown,
							Type:            tftypes.String,
						},
						{
							Name:            "up",
							Required:        true,
							Description:     "The query to run when applying this migration.",
							DescriptionKind: tfprotov5.StringKindMarkdown,
							Type:            tftypes.String,
						},
						{
							Name:            "down",
							Required:        true,
							Description:     "The query to run when undoing this migration.",
							DescriptionKind: tfprotov5.StringKindMarkdown,
							Type:            tftypes.String,
						},
					},
				},
			},
		},
		Attributes: []*tfprotov5.SchemaAttribute{
			&tfprotov5.SchemaAttribute{
				Name:     "complete_migrations",
				Computed: true,
				Description: "The completed migrations that have been run against your database. This list is used as " +
					"storage to migrate down or as a trigger for downstream dependencies.",
				DescriptionKind: tfprotov5.StringKindMarkdown,
				Type: tftypes.List{
					ElementType: tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"id":   tftypes.String,
							"up":   tftypes.String,
							"down": tftypes.String,
						},
					},
				},
			},
		},
	},
}
```

### `framework.Schema`

Like `helper/schema`, the framework could wrap the `tfprotov6.Schema` types with helpers to reduce verbosity. A minimally verbose option could look something like:

```go
var schema = map[string]*framework.Schema{
	"migration": {
		Type:    framework.NestedBlockType,
		Nesting: framework.SchemaNestedBlockNestingModeList,
		Attributes: {
			"id": {
				Required:    true,
				Description: "Identifier can be any string to help identifying the migration in the source.",
				Type:        framework.StringType,
			},
			"up": {
				Required:    true,
				Description: "The query to run when applying this migration.",
				Type:        framework.StringType,
			},
			"down": {
				Required:    true,
				Description: "The query to run when undoing this migration.",
				Type:        framework.StringType,
			},
		},
	},
	"complete_migrations": {
		Computed: true,
		Description: "The completed migrations that have been run against your database. This list is used as " +
			"storage to migrate down or as a trigger for downstream dependencies.",
		Type: framework.TypeList{
			ElementType: framework.TypeObject{
				AttributeTypes: map[string]framework.AttributeType{
					"id":   framework.TypeString,
					"up":   framework.TypeString,
					"down": framework.TypeString,
				},
			},
		},
	},
}
```

Types such as `Schema` and `StringType` would likely be exported by a package (like `helper/schema`), not the top-level `framework` package, whose name is used here for simplicity.

The work being done here by the `framework.Schema` type, from the provider developer's perspective, is:
 - Remove the need for a `Block` field in the schema struct for the root block, since this is unambiguous;
 - Use the attribute or nested block `Name` as a map key;
 - Handle defaults for fields such as `DescriptionKind` (`Nesting` could also be set appropriately);
 - Allow nested blocks to be defined in the same way as attributes with a special `NestedBlockType`.

#### Benefits: verbosity, discoverability

Apart from reducing verbosity, this approach has a number of benefits. Users will not have to import `tftypes` or `tfprotov6` packages, removing the burden of having to understand why the terraform-plugin-go module exists separately from the framework. The framework schema types are more discoverable, being documented inside the framework code for framework users, not in plugin-go.

#### Tradeoffs: compatibility

The API of the framework schema package for defining resource schemas (i.e. the functionality shown in the examples above) should not have to change unless there is a change in the Terraform protocol concerning schemas.

Compatibility can be evaluated by considering the following two situations.

Firstly, consider what happens if a backwards-compatible change is made in the Terraform protocol, such as the addition of `SchemaObject` in tfprotov6. In this case, the framework schema package should also be able to make the change in a backwards-compatible way, unless it concerns one of the assumptions made above (for example, that attributes/blocks will always have a unique `Name` that can be used as a map key). This seems unlikely, but it is possible: for example, a new `BlockType` is added which has no equivalent to `Name` or `TypeName`.

Secondly, consider what happens if a backwards-incompatible change is made in the Terraform protocol. In this case, a new Terraform protocol major version would be needed anyway, so the compatibility status of `framework.Schema` is no worse than `tfprotov5.schema`. 

Note that in the case of Terraform protocol v5 and v6, despite a major version increase, the functional change made is backwards-compatible with respect to the schema, so `tfprotov6.Schema` is a backwards-compatible extension of `tfprotov5.Schema`. Wrapping these types in a `framework.Schema` type hides this complexity from users and saves them the work of updating their use of `tfprotovN.Schema` if something similar happens in the future. 
