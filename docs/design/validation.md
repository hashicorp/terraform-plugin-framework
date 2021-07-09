# Validation

Practitioners implementing Terraform configurations desire feedback surrounding the syntax, types, and acceptable values. This feedback, typically referred to as validation, is perferably given as early as possible before a configuration is applied. Terraform supports a plugin architecture, which extends the configuration and validation surface area based on the implementation details of those plugins. This framework provides validation hooks for plugins. This design document will outline background information on the problem space, prior framework choices, and proposals for this framework.

## Background

Terraform CLI, this framework, and a Terraform Provider each have differing responsibilities for validation. Depending on the configuration and operation being performed, full information for validation may not yet be visible. Before diving into the intricacies around plugin validation and this framework's design considerations, a general overview of Terraform's configuration and validation mechanisms is provided. Additional information about Terraform concepts not described in detail in this document can be found in the [Terraform Documentation](https://www.terraform.io/docs/).

### Terraform Configuration

The [Terraform configuration language](https://www.terraform.io/docs/language/) is declarative and an implementation of [HashiCorp Configuration Language](https://github.com/hashicorp/hcl) (HCL). HCL provides all the primitives and tokenization required to convert textual configuration files into meaningful concepts and constructs for Terraform. The Terraform CLI is responsible for reading and parsing configurations, performing syntax validation (e.g. feedback around unparseable configurations), and returning user interface output for all validation.

An example of basic configuration syntax validation performed by Terraform CLI:

```console
$ cat main.tf
this is invalid
$ terraform validate
╷
│ Error: Unsupported block type
│ 
│   on main.tf line 1:
│    1: this is invalid
│ 
│ Blocks of type "this" are not expected here.
╵
╷
│ Error: Invalid block definition
│ 
│   on main.tf line 1:
│    1: this is invalid
│ 
│ A block definition must have block content delimited by "{" and
│ "}", starting on the same line as the block header.
```

The [Terraform configuration language defines its own type system](https://www.terraform.io/docs/language/expressions/types.html) which is translated to and from the type system implemented by plugins through the [Terraform Plugin Protocol](#terraform-plugin-protocol) which is described later. This framework is designed to transparently handle those conversions as much as possible, however it is important to note that there are potentially differences in terminology and implementation between the two.

Many values in a Terraform configuration can be referenced in other locations, which can be used to order operations within Terraform:

```terraform
resource "example_foo" "example" {
  some_attribute = "this is a known value"
}

resource "example_bar" "example" {
  known   = example_foo.example.some_attribute  # Known value and expected to be "this is a known value"
  unknown = example_foo.example.other_attribute # Likely unknown value
}
```

In these situations, the value of the `other_attribute` attribute from the `example_foo.example` resource is not present in the configuration, so the value _may_ (see next section) not be known until `example_foo.example` has been applied. These values are typically referred to as unknown values. This distinction is important in Terraform and validation, since this case might need to be explicitly handled by the framework or plugins.

### Terraform Plan

To provide detailed information about actions that Terraform intends to perform, Terraform CLI will generate a plan. For the purposes of validation, the plan is an extension of the available configuration information. Providers have an opportunity to modify the plan before it is finalized, which is where unknown values can potentially be filled in (e.g. with a provider defined default or if the value can be derived from other plan information).

As an example of the human readable output of a plan:

```terraform
$ cat main.tf
resource "random_pet" "example" {
  length = 2
}

$ terraform plan

Terraform used the selected providers to generate the following execution plan. Resource actions are
indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # random_pet.example will be created
  + resource "random_pet" "example" {
      + id        = (known after apply)
      + length    = 2
      + separator = "-"
    }

Plan: 1 to add, 0 to change, 0 to destroy.
```

This plan is surfaced to providers in a machine readable manner through the [Terraform Plugin Protocol](#terraform-plugin-protocol) which is described later.

### Terraform Providers

Terraform Providers are a form of Terraform plugins, which are gRPC (formerly `net/rpc`) server processes that are lifecycle managed by Terraform CLI. Providers implement [managed resources](https://www.terraform.io/docs/language/resources/) and [data sources](https://www.terraform.io/docs/language/data-sources/). Often a managed resource is just called a "resource" although in the underlying implementation details a "resource" may refer to either for legacy reasons as seen later in the previous provider framework.

From a configuration standpoint, Terraform implements the concepts of `provider`, `resource`, and `data` (source) configurations, while providers implement the details inside those configurations in what is called "schema" information. The schema defines attribute naming, types, and behaviors for consumption by the framework and Terraform CLI.

To visualize the difference in a configuration:

```terraform
# Provider configuration
provider "example" { # Defined by the configuration language
  # Defined by the provider schema
  input = "" # a required or optional string attribute named input
}

# Resource configuration
resource "example_thing" "example" { # Defined by the configuration language
  # Defined by the resource schema
  input = 123 # a required or optional number attribute named input
}

# Data source configuration
data "example_thing" "example" { # Defined by the configuration language
  # Defined by the data source schema
  input = true # a required or optional boolean attribute named input
}
```

Terraform supports the following validation for provider implementations:

- Provider configurations
- Resource configurations and plans
- Data Source configurations and plans

Within these, there are two types of validation:

- Single attribute value validation (e.g. string length)
- Multiple attribute validation (e.g. attributes or attribute values that conflict with each other)

The next sections will outline some of the underlying details relevant to implementation proposals in this framework.

### Terraform Plugin Protocol

The specification between Terraform CLI and plugins, such as Terraform Providers, is currently implemented via [Protocol Buffers](https://developers.google.com/protocol-buffers). Below highlights some of the service `rpc` (called by Terraform CLI) and `message` types that are intergral for validation support and applying/destroying a given configuration.

#### `ApplyResourceChange` RPC

Called during the `terraform apply` and `terraform destroy` commands.

```protobuf
service Provider {
    // ...
    rpc ApplyResourceChange(ApplyResourceChange.Request) returns (ApplyResourceChange.Response);
}

message ApplyResourceChange {
    message Request {
        string type_name = 1;
        DynamicValue prior_state = 2;
        DynamicValue planned_state = 3;
        DynamicValue config = 4;
        bytes planned_private = 5; 
        DynamicValue provider_meta = 6;
    }
    message Response {
        DynamicValue new_state = 1;
        bytes private = 2; 
        repeated Diagnostic diagnostics = 3;
    }
}
```

#### `PlanResourceChange` RPC

Called during the `terraform apply`, `terraform destroy`, and `terraform plan` commands.

```protobuf
service Provider {
    // ...
    rpc PlanResourceChange(PlanResourceChange.Request) returns (PlanResourceChange.Response);
}

message PlanResourceChange {
    message Request {
        string type_name = 1;
        DynamicValue prior_state = 2;
        DynamicValue proposed_new_state = 3;
        DynamicValue config = 4;
        bytes prior_private = 5; 
        DynamicValue provider_meta = 6;
    }

    message Response {
        DynamicValue planned_state = 1;
        repeated AttributePath requires_replace = 2;
        bytes planned_private = 3; 
        repeated Diagnostic diagnostics = 4;
    }
}
```

#### `ValidateDataSourceConfig` RPC

Called during the `terraform apply`, `terraform destroy`, `terraform plan`, `terraform refresh`, and `terraform validate` commands if data sources are present.

```protobuf
service Provider {
    // ...
    rpc ValidateDataResourceConfig(ValidateDataResourceConfig.Request) returns (ValidateDataResourceConfig.Response);
}

message ValidateDataResourceConfig {
    message Request {
        string type_name = 1;
        DynamicValue config = 2;
    }
    message Response {
        repeated Diagnostic diagnostics = 1;
    }
}
```

#### `ValidateProviderConfig` RPC

Called during the `terraform apply`, `terraform destroy`, `terraform plan`, `terraform refresh`, and `terraform validate` commands if providers are present.

```protobuf
service Provider {
    // ...
    rpc ValidateProviderConfig(ValidateProviderConfig.Request) returns (ValidateProviderConfig.Response);
}

message ValidateProviderConfig {
    message Request {
        DynamicValue config = 1;
    }
    message Response {
        repeated Diagnostic diagnostics = 2;
    }
}
```

#### `ValidateResourceTypeConfig` RPC

Called during the `terraform apply`, `terraform destroy`, `terraform plan`, `terraform refresh`, and `terraform validate` commands if managed resources are present.

```protobuf
service Provider {
    // ...
    rpc ValidateResourceConfig(ValidateResourceConfig.Request) returns (ValidateResourceConfig.Response);
}

message ValidateResourceConfig {
    message Request {
        string type_name = 1;
        DynamicValue config = 2;
    }
    message Response {
        repeated Diagnostic diagnostics = 1;
    }
}
```

#### `Diagnostics` Message

Diagnostics in the protocol allow providers to return warnings and errors.

```protobuf
message Diagnostic {
    enum Severity {
        INVALID = 0;
        ERROR = 1;
        WARNING = 2;
    }
    Severity severity = 1;
    string summary = 2;
    string detail = 3;
    AttributePath attribute = 4;
}
```

### terraform-plugin-go

The [`terraform-plugin-go` library](https://pkg.go.dev/hashicorp/terraform-plugin-go) is a low-level implementation of the [Terraform Plugin Protocol](#terraform-plugin-protocol) in Go and underpins this framework. This includes packages such as `tfprotov6` and `tftypes`. These are mentioned for completeness as some of these types are not yet abstracted in this framework and may be shown in implementation proposals.

### terraform-plugin-framework

Most of the Go types and functionality from `terraform-plugin-go` will be abstracted by this framework before reaching provider developers. The details represented here are not finalized as this framework is still being designed, however these current details are presented here for additional context in the later proposals.

Generic `tftypes` values are abstracted into an `attr.Value` Go interface type with concrete Go types such as `types.String`:

```go
// Value defines an interface for describing data associated with an attribute.
// Values allow provider developers to specify data in a convenient format, and
// have it transparently be converted to formats Terraform understands.
type Value interface {
    // ToTerraformValue returns the data contained in the Value as
    // a Go type that tftypes.NewValue will accept.
    ToTerraformValue(context.Context) (interface{}, error)

    // Equal must return true if the Value is considered semantically equal
    // to the Value passed as an argument.
    Equal(Value) bool
}

// ... separately ...

var _ attr.Value = String{}

// String represents a UTF-8 string value.
type String struct {
    // Unknown will be true if the value is not yet known.
    Unknown bool

    // Null will be true if the value was not set, or was explicitly set to
    // null.
    Null bool

    // Value contains the set value, as long as Unknown and Null are both
    // false.
    Value string
}
```

Resources (and similarly, but separately, data sources) are currently implmented in their own `Resource` and `ResourceType` Go interface types. Providers are responsible for implementing the concrete Go types.

```go
// A ResourceType is a type of resource. For each type of resource this provider
// supports, it should define a type implementing ResourceType and return an
// instance of it in the map returned by Provider.GeResources.
type ResourceType interface {
    // GetSchema returns the schema for this resource.
    GetSchema(context.Context) (schema.Schema, []*tfprotov6.Diagnostic)

    // NewResource instantiates a new Resource of this ResourceType.
    NewResource(context.Context, Provider) (Resource, []*tfprotov6.Diagnostic)
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
}
```

Similar to the previous framework, schema attributes are currently implemented in their own `Attribute` Go structure type:

```go
// Attribute defines the constraints and behaviors of a single field in a
// schema. Attributes are the fields that show up in Terraform state files and
// can be used in configuration files.
type Attribute struct {
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

Although later designs surrounding the ability to allow providers to define custom schema types may change this particular Go typing detail.

## Prior Implementations

### terraform-plugin-sdk

The previous framework for provider implementations, Terraform Plugin SDK, can be found in the `terraform-plugin-sdk` repository. That framework has existed since the very early days of Terraform, where it was previously contained in a combined CLI and provider codebase, to support the code and testing aspects of provider development.

To implement managed resources and data sources, the previous framework was largely based around Go structure types and declarative definitions of intended behaviors. These were defined in the `helper/schema` package, in particular, the `Schema` and `Resource` types.

#### `helper/schema.Schema`

This type is the main entrypoint for declaring attribute information within a resource or data source. For example,

```go
map[string]*schema.Schema{
    "attribute_name": {
        Type:     schema.TypeString,
        Required: true,
    },
}
```

It supported single attribute value validation via the `ValidateFunc` or `ValidateDiagFunc` fields and multiple attribute validation via a collection of different fields (`AtLeastOneOf`, `ConflictsWith`, `ExactlyOneOf`, `RequiredWith`) which could be combined as necessary. For list, set, and map types, two additional fields (`MaxItems` and `MinItems`) provided validation for the number of elements.

The multiple attribute validation support in the attribute schema is purely existance based, meaning it could not be conditional based on the attribute value. Conditional multiple attribute validation based on values was later added via the resource level `CustomizeDiff`, which will be described later on.

These fields also required a full attribute path in "flatmap" syntax, which had limitations for declaring them against nested attributes. For example:

```go
map[string]*schema.Schema{
    "root_attribute": {
        Type:     schema.TypeString,
        Optional: true,
    },
    "single_block": {
        Type:     schema.TypeList,
        Optional: true,
        MaxItems: 1,
        Elem:     &schema.Resource{
            Schema: map[string]*schema.Schema{
                "list_attribute_one": {
                    Type:          schema.TypeString,
                    Optional:      true,
                    ConflictsWith: []string{"single_block.0.list_attribute_two"}, // only valid due to MaxItems: 1
                },
                "list_attribute_two": {
                    Type:          schema.TypeString,
                    Optional:      true,
                    ConflictsWith: []string{"single_block.0.list_attribute_one"}, // only valid due to MaxItems: 1
                },
            },
        },
    },
    "set_of_blocks": {
        Type:     schema.TypeSet,
        Optional: true,
        Elem:     &schema.Resource{
            Schema: map[string]*schema.Schema{
                "set_attribute_one": {
                    Type:          schema.TypeString,
                    Optional:      true,
                    ConflictsWith: []string{/* No flatmap address syntax for set_attribute_two */}
                },
                "set_attribute_two": {
                    Type:          schema.TypeString,
                    Optional:      true,
                    ConflictsWith: []string{/* No flatmap address syntax for set_attribute_one */}
                },
            },
        },
    },
}
```

##### `AtLeastOneOf`

This field enabled the schema to validate that at least one of the attribute addresses (in "flatmap" syntax) was present in a configuration. For example,

```go
map[string]*schema.Schema{
    "attribute_one": {
        Type:         schema.TypeString,
        Optional:     true,
        AtLeastOneOf: []string{"attribute_one", "attribute_two"},
    },
    "attribute_two": {
        Type:         schema.TypeString,
        Optional:     true,
        AtLeastOneOf: []string{"attribute_one", "attribute_two"},
    },
}
```

Gave the following results:

```terraform
# Failed validation (error returned)
resource "example_thing" "example" {}

# Passed validation
resource "example_thing" "example" {
  attribute_one = "some_value"
}

# Passed validation
resource "example_thing" "example" {
  attribute_two = "some_value"
}

# Passed validation
resource "example_thing" "example" {
  attribute_one = "some_value"
  attribute_two = "some_value"
}
```

##### `ConflictsWith`

This field enabled the schema to validate that multiple of the attribute addresses (in "flatmap" syntax) were present in a configuration. For example,

```go
map[string]*schema.Schema{
    "attribute_one": {
        Type:          schema.TypeString,
        Optional:      true,
        ConflictsWith: []string{"attribute_two"},
    },
    "attribute_two": {
        Type:          schema.TypeString,
        Optional:      true,
        ConflictsWith: []string{"attribute_one"},
    },
}
```

Gave the following results:

```terraform
# Passed validation
resource "example_thing" "example" {}

# Passed validation
resource "example_thing" "example" {
  attribute_one = "some_value"
}

# Passed validation
resource "example_thing" "example" {
  attribute_two = "some_value"
}

# Failed validation (error returned)
resource "example_thing" "example" {
  attribute_one = "some_value"
  attribute_two = "some_value"
}
```

##### `ExactlyOneOf`

This field enabled the schema to validate that one (and only one) of the attribute addresses (in "flatmap" syntax) must be present in a configuration. For example,

```go
map[string]*schema.Schema{
    "attribute_one": {
        Type:         schema.TypeString,
        Optional:     true,
        ExactlyOneOf: []string{"attribute_one", "attribute_two"},
    },
    "attribute_two": {
        Type:         schema.TypeString,
        Optional:     true,
        ExactlyOneOf: []string{"attribute_one", "attribute_two"},
    },
}
```

Gave the following results:

```terraform
# Failed validation (error returned)
resource "example_thing" "example" {}

# Passed validation
resource "example_thing" "example" {
  attribute_one = "some_value"
}

# Passed validation
resource "example_thing" "example" {
  attribute_two = "some_value"
}

# Failed validation (error returned)
resource "example_thing" "example" {
  attribute_one = "some_value"
  attribute_two = "some_value"
}
```

##### `MaxItems`

This field enabled the schema to validate the maximum number of elements in a list, set, or map type. For example,

```go
map[string]*schema.Schema{
    "single_block": {
        Type:     schema.TypeList,
        Optional: true,
        MaxItems: 1,
        Elem:     &schema.Resource{
            Schema: map[string]*schema.Schema{ /* ... nested attributes ... */ },
        },
    },
}
```

Gave the following results:

```terraform
# Passed validation
resource "example_thing" "example" {}

# Passed validation
resource "example_thing" "example" {
  single_block {
    # ... nested attributes ...
  }
}

# Failed validation (error returned)
resource "example_thing" "example" {
  single_block {
    # ... nested attributes ...
  }

  single_block {
    # ... nested attributes ...
  }
}
```

##### `MinItems`

This field enabled the schema to validate the minimum number of elements in a list, set, or map type. For example,

```go
map[string]*schema.Schema{
    "multiple_block": {
        Type:     schema.TypeList,
        Optional: true,
        MinItems: 1,
        Elem:     &schema.Resource{
            Schema: map[string]*schema.Schema{ /* ... nested attributes ... */ },
        },
    },
}
```

Gave the following results:

```terraform
# Passed validation
resource "example_thing" "example" {}

# Failed validation (error returned)
resource "example_thing" "example" {
  multiple_block {
    # ... nested attributes ...
  }
}

# Passed validation
resource "example_thing" "example" {
  multiple_block {
    # ... nested attributes ...
  }

  multiple_block {
    # ... nested attributes ...
  }
}
```

##### `RequiredWith`

This field enabled the schema to validate that any of the attribute addresses (in "flatmap" syntax) were implied as present in a configuration. For example,

```go
map[string]*schema.Schema{
    "attribute_one": {
        Type:     schema.TypeString,
        Optional: true,
    },
    "attribute_two": {
        Type:          schema.TypeString,
        Optional:      true,
        RequiredWith: []string{"attribute_one"},
    },
}
```

Gave the following results:

```terraform
# Passed validation
resource "example_thing" "example" {}

# Failed validation (error returned)
resource "example_thing" "example" {
  attribute_one = "some_value"
}

# Passed validation
resource "example_thing" "example" {
  attribute_two = "some_value"
}

# Passed validation
resource "example_thing" "example" {
  attribute_one = "some_value"
  attribute_two = "some_value"
}
```

##### `ValidateFunc` / `ValidateDiagFunc`

These fields provided single attribute value validation. `ValidateDiagFunc` was a more recent version of `ValidateFunc`, returning `Diagnostics` instead of warning string and error slices.

For example,

```go
//
map[string]*schema.Schema{
    "attribute_name": {
        Type:         schema.TypeString,
        Required:     true,
        ValidateFunc: func(rawValue interface{}, attributePath string) (warnings []string, errors []error) {
            value, ok := rawValue.(string)

            if !ok {
                errors = append(errors, fmt.Errorf("expected type of %s to be string", attributePath))
                return
            }

            if value == "" {
                errors = append(errors, fmt.Errorf("expected %s to not be empty", attributePath))
            }

            return
        },
    },
}
```

Gave the following results:

```terraform
# Failed validation (error returned)
resource "example_thing" "example" {
  attribute_name = ""
}

# Passed validation
resource "example_thing" "example" {
  attribute_name = "some_value"
}
```

These validation functions are expected to perform value type conversion to match the schema and the concepts of null or unknown values are not surfaced due to limitations in the previous framework type system.

Rather than require provider developers to recreate relatively common value validations, a separate `helper/validation` package provides a wide variety of value validation functions and is described below.

###### `helper/validation` Package

This package has common validation functions which can be directly implemented within a `helper/schema.Schema#ValidateFunc`, for example:

```go
map[string]*schema.Schema{
    "attribute_name": {
        Type:         schema.TypeString,
        Required:     true,
        ValidateFunc: validation.StringIsNotEmpty,
    },
}
```

The surface area of this package, as seen in its [Go documentation](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation), is quite large. This documentation will only summarize some of the capabilities of those functions, to highlight the breadth and depth this framework should also support. Whether this framework should also implement these functions or if they should be offered in other packages/modules is a separate decision point.

- Generic String
  - Contains Character/Substring
  - Starts/Ends With
  - Length Between/Maximum/Minimum
  - One of a Collection (Enumeration)
  - Regular Expression
- Generic Float/Integer
  - Between/Maximum/Minimum
  - Multiple Of (Modulo)
  - One of a Collection (Enumeration)
- Encoding
  - base64
- Format
  - JSON
  - YAML
- Networking
  - CIDR
  - IPv4 Address
  - IPv6 Address
  - MAC Address
  - Port Number
- Time/Date
  - Day of week name
  - Month name
  - RFC3339
- URI
  - Scheme

In addition to the above, there are two generic validation helper functions `Any()` and `All()`. These can be used to logically `OR` or `AND` multiple validation functions together:

```go
map[string]*schema.Schema{
    "attribute_name": {
        Type:         schema.TypeString,
        Required:     true,
        ValidateFunc: validation.All(
            validation.StringLenBetween(1, 256),
            validation.StringMatch(regexp.MustCompile(`^[0-9a-zA-Z]+$`), "must contain only alphanumeric characters"),
        ),
    },
}
```

#### `helper/schema.Resource#CustomizeDiff`

As noted above, the multiple attribute validation was limited in the utility it could provide. Terraform CLI and the previous framework were enhanced to support modifying the plan or return an error before it was executed, allowing providers to introduce custom logic around resource recreation and a generic form of validation. This was implemented in the `CustomizeDiff` field of the `Resource` type as a function that had the plan information and provider instance available.

For example:

```go
&schema.Resource{
    // ...
    CustomizeDiff: func(_ context.Context, diff *schema.ResourceDiff, meta interface{}) error {
        if value := diff.Get("attribute_one").(string); value == "special condition" {
            if _, ok := diff.GetOk("attribute_two"); !ok {
                return fmt.Errorf("'attribute_two' must be set when 'attribute_one' is %q", value)
            }
        }

        return nil
    },
}
```

In general, `CustomizeDiff` is not broadly utilized across the ecosystem due to the complexity of properly implementing and testing the functionality.

##### `helper/customdiff` Package

Similar to how the `helper/validation` package of common functionality was created for `ValidateFunc`, a [`helper/customizediff`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff) package was created for common `CustomizeDiff` functionality.

In terms of validation, this provided helpers such as:

```go
&schema.Resource{
    // ...
    CustomizeDiff: customdiff.IfValue(
        "attribute_one",
        func(ctx context.Context, rawValue, meta interface{}) bool {
            value, ok := rawValue.(string)

            // potentially difficult to diagnose type issue
            if !ok {
                return false
            }

            return value == "special condition"
        },
        customdiff.ValidateValue(
            "attribute_two",
            func(ctx context.Context, rawValue, meta interface{}) error {
                value, ok := rawValue.(string)

                if !ok {
                    return fmt.Errorf("incorrect type conversion for attribute_two")
                }

                if value != "" {
                    return fmt.Errorf("cannot provide attribute_two value when attribute_one is \"special condition\"")
                }

                return nil
            },
        ),
    ),
}
```

These likely would have been simplified into further helpers should there have been more `CustomizeDiff` usage.

## Goals

This framework design should strive to accomplish the following with validation support.

Allow provider developers access to all current types of provider validation:

- Provider configurations
- Resource configurations and plans
- Data Source configurations and plans

Including where possible:

- Single attribute value validation (e.g. string length)
- Multiple attribute validation (e.g. attributes or attribute values that conflict with each other)

In terms of implementation, the following core concepts:

- Low level primitives (e.g other portions of the framework, external packages, and provider developers can implement higher level functionality)
- Reusability between single attribute and multiple attribute validation functionality (e.g. attribute value functions)
- Hooks for documentation (e.g. for future tooling such as provider documentation generators to self-document attributes)

Finally, these other considerations:

- Providing the appropriate amount of contextual information for debugging purposes
- Providing the appropriate amount of contextual information for practitioner facing output
- Ease of extending validation (e.g. handling type conversion and/or unknown values in the framework)
- Ease of testing validation (e.g. unit testing)
- Ease and succinctness of common validation scenarios (e.g. verbosity in provider code)
