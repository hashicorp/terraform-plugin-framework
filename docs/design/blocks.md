# Schema Blocks

The framework currently implements nested attribute support in schemas that is only available in the recent Terraform Plugin Protocol version 6. Prior versions of the protocol and SDK implemented support for blocks. Provider developers may have previously existing resources from the older Terraform Plugin SDK that they wish to mux with or bring into the new framework without introducing breaking changes for practitioners.

This design documentation will walk through and recommend options for schema block handling in the framework.

## Background

### Terraform Configuration Language

Terraform's configuration language is built on top of [HashiCorp Configuration Language (HCL)](https://github.com/hashicorp/hcl) and conceptually supports two syntaxes for declaring configuration as described in the [Configuration Syntax documentation](https://www.terraform.io/docs/language/syntax/configuration.html):

- **Attributes** (**Arguments**): Assign a value to a particular name.
- **Blocks**: Container for other content.

With the attributes using an equals sign (`=`), such as:

```terraform
resource "example_thing" "example" {
  some_attr = "abc123" # Example attribute/argument
}
```

While blocks are structural and defined without an equals sign (`=`), such as:

```terraform
resource "example_thing" "example" {
  some_block {
    # ... attributes/arguments ...
  }
}
```

Blocks can be repeated multiple times if the underlying schema type allows multiple declarations, such as:

```terraform
resource "example_thing" "example" {
  some_block {
    # ... attributes/arguments ...
  }

  some_block {
    # ... attributes/arguments ...
  }
}
```

To handle situations where the block configuration is not static, the configuration language added support for [`dynamic` block expressions](https://www.terraform.io/docs/language/expressions/dynamic-blocks.html). In contrast, working with attributes is relatively simpler as there are many expressions that can operate directly on any collection value, such as [`for` expressions](https://www.terraform.io/docs/language/expressions/for.html).

Blocks also have other limitations in Terraform, such as not supporting sensitivity. Instead, each nested attribute must implement sensitivity as appropriate.

### Terraform Plugin Protocol

In terms of the protocol, all configuration information from the plugin is referred to as the schema. Prior to the recent version 6 of the protocol, implementing complex schema types could only be done using blocks. Version 6 introduced the new `nested_type` field within `Attribute` so attributes can encode nested object schema information.

Version 6 of the protocol defines the following implementation details for schema:

```protobuf
message Schema {
    message Block {
        int64 version = 1;
        repeated Attribute attributes = 2;
        repeated NestedBlock block_types = 3;
        string description = 4;
        StringKind description_kind = 5;
        bool deprecated = 6;
    }

    message Attribute {
        string name = 1;
        bytes type = 2;
        Object nested_type = 10;
        string description = 3;
        bool required = 4;
        bool optional = 5;
        bool computed = 6;
        bool sensitive = 7;
        StringKind description_kind = 8;
        bool deprecated = 9;
    }

    message NestedBlock {
        enum NestingMode {
            INVALID = 0;
            SINGLE = 1;
            LIST = 2;
            SET = 3;
            MAP = 4;
            GROUP = 5;
        }

        string type_name = 1;
        Block block = 2;
        NestingMode nesting = 3;
        int64 min_items = 4;
        int64 max_items = 5;
    }

    message Object {
        enum NestingMode {
            INVALID = 0;
            SINGLE = 1;
            LIST = 2;
            SET = 3;
            MAP = 4;
        }

        repeated Attribute attributes = 1;
        NestingMode nesting = 3;
        int64 min_items = 4;
        int64 max_items = 5;
    }

    // The version of the schema.
    // Schemas are versioned, so that providers can upgrade a saved resource
    // state when the schema is changed. 
    int64 version = 1;

    // Block is the top level configuration block for this schema.
    Block block = 2;
}
```

### terraform-plugin-go

The `terraform-plugin-go` library, which underpins this framework, provides the following implementation of schema types:

```go
// Schema is how Terraform defines the shape of data. It can be thought of as
// the type information for resources, data sources, provider configuration,
// and all the other data that Terraform sends to providers. It is how
// providers express their requirements for that data.
type Schema struct {
	// Version indicates which version of the schema this is. Versions
	// should be monotonically incrementing numbers. When Terraform
	// encounters a resource stored in state with a schema version lower
	// that the schema version the provider advertises for that resource,
	// Terraform requests the provider upgrade the resource's state.
	Version int64

	// Block is the root level of the schema, the collection of attributes
	// and blocks that make up a resource, data source, provider, or other
	// configuration block.
	Block *SchemaBlock
}

// SchemaBlock represents a block in a schema. Blocks are how Terraform creates
// groupings of attributes. In configurations, they don't use the equals sign
// and use dynamic instead of list comprehensions.
//
// Blocks will show up in state and config Values as a tftypes.Object, with the
// attributes and nested blocks defining the tftypes.Object's AttributeTypes.
type SchemaBlock struct {
	// TODO: why do we have version in the block, too?
	Version int64

	// Attributes are the attributes defined within the block. These are
	// the fields that users can set using the equals sign or reference in
	// interpolations.
	Attributes []*SchemaAttribute

	// BlockTypes are the nested blocks within the block. These are used to
	// have blocks within blocks.
	BlockTypes []*SchemaNestedBlock

	// Description offers an end-user friendly description of what the
	// block is for. This will be surfaced to users through editor
	// integrations, documentation generation, and other settings.
	Description string

	// DescriptionKind indicates the formatting and encoding that the
	// Description field is using.
	DescriptionKind StringKind

	// Deprecated, when set to true, indicates that a block should no
	// longer be used and users should migrate away from it. At the moment
	// it is unused and will have no impact, but it will be used in future
	// tooling that is powered by provider schemas to enable richer user
	// experiences. Providers should set it when deprecating blocks in
	// preparation for these tools.
	Deprecated bool
}

// SchemaAttribute represents a single attribute within a schema block.
// Attributes are the fields users can set in configuration using the equals
// sign, can assign to variables, can interpolate, and can use list
// comprehensions on.
type SchemaAttribute struct {
	// Name is the name of the attribute. This is what the user will put
	// before the equals sign to assign a value to this attribute.
	Name string

	// Type indicates the type of data the attribute expects. See the
	// documentation for the tftypes package for information on what types
	// are supported and their behaviors.
	Type tftypes.Type

	// NestedType indicates that this is a NestedBlock-style object masquerading
	// as an attribute. This field conflicts with Type.
	NestedType *SchemaObject

	// Description offers an end-user friendly description of what the
	// attribute is for. This will be surfaced to users through editor
	// integrations, documentation generation, and other settings.
	Description string

	// Required, when set to true, indicates that this attribute must have
	// a value assigned to it by the user or Terraform will throw an error.
	Required bool

	// Optional, when set to true, indicates that the user does not need to
	// supply a value for this attribute, but may.
	Optional bool

	// Computed, when set to true, indicates the the provider will supply a
	// value for this field. If Optional and Required are false and
	// Computed is true, the user will not be able to specify a value for
	// this field without Terraform throwing an error. If Optional is true
	// and Computed is true, the user can specify a value for this field,
	// but the provider may supply a value if the user does not. It is
	// always a violation of Terraform's protocol to substitute a value for
	// what the user entered, even if Computed is true.
	Computed bool

	// Sensitive, when set to true, indicates that the contents of this
	// attribute should be considered sensitive and not included in output.
	// This does not encrypt or otherwise protect these values in state, it
	// only offers protection from them showing up in plans or other
	// output.
	Sensitive bool

	// DescriptionKind indicates the formatting and encoding that the
	// Description field is using.
	DescriptionKind StringKind

	// Deprecated, when set to true, indicates that a attribute should no
	// longer be used and users should migrate away from it. At the moment
	// it is unused and will have no impact, but it will be used in future
	// tooling that is powered by provider schemas to enable richer user
	// experiences. Providers should set it when deprecating attributes in
	// preparation for these tools.
	Deprecated bool
}

// SchemaNestedBlock is a nested block within another block. See SchemaBlock
// for more information on blocks.
type SchemaNestedBlock struct {
	// TypeName is the name of the block. It is what the user will specify
	// when using the block in configuration.
	TypeName string

	// Block is the block being nested inside another block. See the
	// SchemaBlock documentation for more information on blocks.
	Block *SchemaBlock

	// Nesting is the kind of nesting the block is using. Different nesting
	// modes have different behaviors and imply different kinds of data.
	Nesting SchemaNestedBlockNestingMode

	// MinItems is the minimum number of instances of this block that a
	// user must specify or Terraform will return an error.
	//
	// MinItems can only be set for SchemaNestedBlockNestingModeList and
	// SchemaNestedBlockNestingModeSet. SchemaNestedBlockNestingModeSingle
	// can also set MinItems and MaxItems both to 1 to indicate that the
	// block is required to be set. All other SchemaNestedBlockNestingModes
	// must leave MinItems set to 0.
	MinItems int64

	// MaxItems is the maximum number of instances of this block that a
	// user may specify before Terraform returns an error.
	//
	// MaxItems can only be set for SchemaNestedBlockNestingModeList and
	// SchemaNestedBlockNestingModeSet. SchemaNestedBlockNestingModeSingle
	// can also set MinItems and MaxItems both to 1 to indicate that the
	// block is required to be set. All other SchemaNestedBlockNestingModes
	// must leave MaxItems set to 0.
	MaxItems int64
}
```

## Prior Implementations

### terraform-plugin-sdk

The previous framework provided the [`helper/schema.Resource` type](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema#Resource):

```go
// Resource represents a thing in Terraform that has a set of configurable
// attributes and a lifecycle (create, read, update, delete).
//
// The Resource schema is an abstraction that allows provider writers to
// worry only about CRUD operations while off-loading validation, diff
// generation, etc. to this higher level library.
//
// In spite of the name, this struct is not used only for terraform resources,
// but also for data sources. In the case of data sources, the Create,
// Update and Delete functions must not be provided.
type Resource struct {
	// ... other fields ...

	// Schema is the schema for the configuration of this resource.
	//
	// The keys of this map are the configuration keys, and the values
	// describe the schema of the configuration value.
	//
	// The schema is used to represent both configurable data as well
	// as data that might be computed in the process of creating this
	// resource.
	Schema map[string]*Schema
}

// Schema is used to describe the structure of a value.
//
// Read the documentation of the struct elements for important details.
type Schema struct {
	// ... other fields ...

	// Type is the type of the value and must be one of the ValueType values.
	//
	// This type not only determines what type is expected/valid in configuring
	// this value, but also what type is returned when ResourceData.Get is
	// called. The types returned by Get are:
	//
	//   TypeBool - bool
	//   TypeInt - int
	//   TypeFloat - float64
	//   TypeString - string
	//   TypeList - []interface{}
	//   TypeMap - map[string]interface{}
	//   TypeSet - *schema.Set
	//
	Type ValueType

	// If one of these is set, then this item can come from the configuration.
	// Both cannot be set. If Optional is set, the value is optional. If
	// Required is set, the value is required.
	//
	// One of these must be set if the value is not computed. That is:
	// value either comes from the config, is computed, or is both.
	Optional bool
	Required bool

	// Description is used as the description for docs, the language server and
	// other user facing usage. It can be plain-text or markdown depending on the
	// global DescriptionKind setting.
	Description string

	// The fields below relate to diffs.
	//
	// If Computed is true, then the result of this value is computed
	// (unless specified by config) on creation.
	//
	// If ForceNew is true, then a change in this resource necessitates
	// the creation of a new resource.
	//
	// StateFunc is a function called to change the value of this before
	// storing it in the state (and likewise before comparing for diffs).
	// The use for this is for example with large strings, you may want
	// to simply store the hash of it.
	Computed  bool
	ForceNew  bool
	StateFunc SchemaStateFunc

	// The following fields are only set for a TypeList, TypeSet, or TypeMap.
	//
	// Elem represents the element type. For a TypeMap, it must be a *Schema
	// with a Type that is one of the primitives: TypeString, TypeBool,
	// TypeInt, or TypeFloat. Otherwise it may be either a *Schema or a
	// *Resource. If it is *Schema, the element type is just a simple value.
	// If it is *Resource, the element type is a complex structure,
	// potentially managed via its own CRUD actions on the API.
	Elem interface{}

	// The following fields are only set for a TypeList or TypeSet.
	//
	// MaxItems defines a maximum amount of items that can exist within a
	// TypeSet or TypeList. Specific use cases would be if a TypeSet is being
	// used to wrap a complex structure, however more than one instance would
	// cause instability.
	//
	// MinItems defines a minimum amount of items that can exist within a
	// TypeSet or TypeList. Specific use cases would be if a TypeSet is being
	// used to wrap a complex structure, however less than one instance would
	// cause instability.
	//
	// If the field Optional is set to true then MinItems is ignored and thus
	// effectively zero.
	MaxItems int
	MinItems int

	// When Deprecated is set, this attribute is deprecated.
	//
	// A deprecated field still works, but will probably stop working in near
	// future. This string is the message shown to the user with instructions on
	// how to address the deprecation.
	Deprecated string

	// Sensitive ensures that the attribute's value does not get displayed in
	// logs or regular output. It should be used for passwords or other
	// secret fields. Future versions of Terraform may encrypt these
	// values.
	Sensitive bool
}
```

An example schema would be:

```go
func resourceExampleThing() *schema.Resource {
	return &schema.Resource{
		// ... other fields ...

		Schema: map[string]*schema.Schema{
			"example_block": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"example_attribute": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}
```

Which allowed the following configuration:

```go
resource "example_thing" "example" {
    example_block {
        example_attribute = "inside the block"
    }
}
```

### terraform-plugin-framework

This framework already provides attribute support in the following manner:

```go
// Attribute defines the constraints and behaviors of a single field in a
// schema. Attributes are the fields that show up in Terraform state files and
// can be used in configuration files.
type Attribute struct {
	// ... other fields ...

	// Type indicates what kind of attribute this is. You'll most likely
	// want to use one of the types in the types package.
	//
	// If Type is set, Attributes cannot be.
	Type attr.Type

	// Attributes can have their own, nested attributes. This nested map of
	// attributes behaves exactly like the map of attributes on the Schema
	// type.
	//
	// If Attributes is set, Type cannot be.
	Attributes NestedAttributes

	// Description is used in various tooling, like the language server, to
	// give practitioners more information about what this attribute is,
	// what it's for, and how it should be used. It should be written as
	// plain text, with no special formatting.
	Description string

	// MarkdownDescription is used in various tooling, like the
	// documentation generator, to give practitioners more information
	// about what this attribute is, what it's for, and how it should be
	// used. It should be formatted using Markdown.
	MarkdownDescription string

	// Required indicates whether the practitioner must enter a value for
	// this attribute or not. Required and Optional cannot both be true,
	// and Required and Computed cannot both be true.
	Required bool

	// Optional indicates whether the practitioner can choose not to enter
	// a value for this attribute or not. Optional and Required cannot both
	// be true.
	Optional bool

	// Computed indicates whether the provider may return its own value for
	// this attribute or not. Required and Computed cannot both be true. If
	// Required and Optional are both false, Computed must be true, and the
	// attribute will be considered "read only" for the practitioner, with
	// only the provider able to set its value.
	Computed bool

	// Sensitive indicates whether the value of this attribute should be
	// considered sensitive data. Setting it to true will obscure the value
	// in CLI output. Sensitive does not impact how values are stored, and
	// practitioners are encouraged to store their state as if the entire
	// file is sensitive.
	Sensitive bool

	// DeprecationMessage defines a message to display to practitioners
	// using this attribute, warning them that it is deprecated and
	// instructing them on what upgrade steps to take.
	DeprecationMessage string
}
```

An example schema similar to blocks (but implemented with nested attributes) would be:

```go
func (t exampleThingResourceType) GetSchema(_ context.Context) (Schema, diag.Diagnostics) {
	return Schema{
		Attributes: map[string]Attribute{
			"example_attribute": {
				Optional: true,
				Computed: true,
				Attributes: ListNestedAttributes(map[string]Attribute{
					"example_nested_attribute": {
						Required: true,
						Type:     types.StringType,
					},
				}, ListNestedAttributesOptions{}),
			},
		},
	}, nil
}
```

Which allowed the following configuration:

```terraform
resource "example_thing" "example" {
    example_attribute = [
        {
            example_nested_attribute = "inside the attribute"
        }
    ]
}
```

A consideration for block support should be whether the schema definition and value handling can be similar to this existing support to reduce provider developer burden.

## Goals

Block support in this framework should be:

- Available as exported functionality for provider developers.
- Abstracted from terraform-plugin-go and convertable into those types to separate implementation concerns.
- Ergonomic to implement Go code (e.g. have helper methods for common use cases).

Additional consideration should be given to:

- Allowing similar value retrieval and writing as attributes, which can enable an easier transition from blocks to nested attributes.
- Allowing similar schema definitions as attributes, which can enable an easier transition from blocks to nested attributes.
- Validating any constraints imposed by Terraform, such as lack of sensitivity.

## Proposals

### Blocks as New Nested Attribute Types

Blocks can be implemented within the existing `Attributes` handling inside `Attribute`.

For example:

```go
"example_block": {
	Attributes: ListNestedBlocks(map[string]Attribute{
		"example_attribute": {
			Type:     types.StringType,
			Required: true,
		},
	}, ListNestedBlocksOptions{}),
},
```

While this prevents the addition of a new field on `Attribute`, it can potentially conflate the two concepts and their supported functionality. Future enhancements that only apply to nested attributes may bleed into the overloaded schema definition. Any functionality within the framework would also need to differentiate the `Attributes` types by querying the nesting mode returned by the helper functions, rather than field existance.

### Blocks as New Attribute Field

Blocks can be implemented as a new field inside `Attribute`.

For example:

```go
"example_block": {
	Blocks: ListNestedBlocks(map[string]Attribute{
		"example_attribute": {
			Type:     types.StringType,
			Required: true,
		},
	}, ListNestedBlocksOptions{}),
},
```

This allows the framework to differentiate between block and nested attributes by fields.

### Blocks as a Separate Type

Blocks can be implemented as a new field inside `Schema` and have a separate type from `Attribute`.

For example:

```go
type Block struct {
    // ... other fields, no Sensitive field ...

    Attributes  map[string]Attribute
    Blocks      map[string]Block
    Description string
    NestingMode BlockNestingMode
}

type Schema struct {
    // ... other fields ...

    Attributes map[string]Attribute
    Blocks     map[string]Block
}
```

This would more closely represent the underlying protocol differences, however it could be a more complex implementation and likely a more confusing provider development experience. Provider developers wanting to switch from blocks to attributes in the future would then need to perform more difficult code changes.

## Recommendations

It is recommended that list and set block support be implemented as a new `Attribute` field, named `Blocks`, that is defined similarly to `Attributes` and conflicts with the `Type` and `Attributes` fields. Other block nesting modes such as map, single, and group should be avoided as they are practically untested in practice since the older Terraform Plugin SDK never supported them. `Attribute` validation can impose any constraints for blocks, such as lack of sensitivity. The schema will do all the necessary conversion into the underlying `SchemaNestedBlock` and `SchemaBlock` types, while data access will be no different than existing `List` or `Set` of `Object` handling (similar to the nested attribute counterparts).

`Attributes` will always be recommended for new schema definitions over `Blocks`. Whether the new `Blocks` field is explicitly marked as `Deprecated` in Go documentation, is a potential implementation detail for consideration.

An example schema may look like:

```go
"example_block": {
	Blocks: ListNestedBlocks(map[string]Attribute{
		"example_attribute": {
			Type:     types.StringType,
			Required: true,
		},
	}, ListNestedBlocksOptions{}),
},
```

This has a few benefits for provider developers as `Blocks` will feel as familiar as `Attributes`. It will also allow for easier conversion to `Attributes` in the future.
