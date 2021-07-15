# Validation

Practitioners implementing Terraform configurations desire feedback surrounding the syntax, types, and acceptable values. This feedback, typically referred to as validation, is preferably given as early as possible before a configuration is applied. Terraform supports a plugin architecture, which extends the configuration and validation surface area based on the implementation details of those plugins. This framework provides validation hooks for plugins. This design document will outline background information on the problem space, prior framework choices, and proposals for this framework.

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

#### `ValidateResourceConfig` RPC

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

Providers are currently implemented in the `Provider` Go interface type. Provider implementations are responsible for implementing the concrete Go type.

```go
// Provider is the core interface that all Terraform providers must implement.
type Provider interface {
    // GetSchema returns the schema for this provider's configuration. If
    // this provider has no configuration, return nil.
    GetSchema(context.Context) (schema.Schema, []*tfprotov6.Diagnostic)

    // Configure is called at the beginning of the provider lifecycle, when
    // Terraform sends to the provider the values the user specified in the
    // provider configuration block. These are supplied in the
    // ConfigureProviderRequest argument.
    // Values from provider configuration are often used to initialise an
    // API client, which should be stored on the struct implementing the
    // Provider interface.
    Configure(context.Context, ConfigureProviderRequest, *ConfigureProviderResponse)

    // GetResources returns a map of the resource types this provider
    // supports.
    GetResources(context.Context) (map[string]ResourceType, []*tfprotov6.Diagnostic)

    // GetDataSources returns a map of the data source types this provider
    // supports.
    GetDataSources(context.Context) (map[string]DataSourceType, []*tfprotov6.Diagnostic)
}
```

Data Sources are currently implemented in their own `DataSource` and `DataSourceType` Go interface types. Providers are responsible for implementing the concrete Go types.

```go
// A DataSourceType is a type of data source. For each type of data source this
// provider supports, it should define a type implementing DataSourceType and
// return an instance of it in the map returned by Provider.GetDataSources.
type DataSourceType interface {
    // GetSchema returns the schema for this data source.
    GetSchema(context.Context) (schema.Schema, []*tfprotov6.Diagnostic)

    // NewDataSource instantiates a new DataSource of this DataSourceType.
    NewDataSource(context.Context, Provider) (DataSource, []*tfprotov6.Diagnostic)
}

// DataSource implements a data source instance.
type DataSource interface {
    // Read is called when the provider must read data source values in
    // order to update state. Config values should be read from the
    // ReadDataSourceRequest and new state values set on the
    // ReadDataSourceResponse.
    Read(context.Context, ReadDataSourceRequest, *ReadDataSourceResponse)
}
```

Managed resources are currently implemented in their own `Resource` and `ResourceType` Go interface types. Providers are responsible for implementing the concrete Go types.

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

Values of `Attribute` in this framework are abstracted from the generic `tftypes` values into an `attr.Value` Go interface type:

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
```

This framework then implements concrete Go types such as `types.String`:

```go
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

The multiple attribute validation support in the attribute schema is purely existence based, meaning it could not be conditional based on the attribute value. Conditional multiple attribute validation based on values was later added via the resource level `CustomizeDiff`, which will be described later on.

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

- Providing the appropriate amount of contextual information for debugging purposes (e.g. logging)
- Providing the appropriate amount of contextual information for practitioner facing output (e.g. paths and values involved with validation decisions)
- Ease of extending validation (e.g. handling type conversion and/or unknown values in the framework)
- Ease of testing validation (e.g. unit testing)
- Ease and succinctness of common validation scenarios (e.g. verbosity in provider code)
- Allowing potential future enhancements of validation behavioral decisions based on configuration (e.g. converting validation errors to warnings or logs)

## Proposals

### Single Attribute Value Validation

This validation would be applicable to the `schema.Attribute` types declared within the `GetSchema()` of `DataSourceType`, `Provider`, and `ResourceType` implementations. For most of these proposals, the framework would walk through all attribute paths during the `ValidateDataSourceConfig`, `ValidateProviderConfig`, and `ValidateResourceConfig` calls, executing the declared validation in each attribute if present.

#### Declaring Value Validation for Attributes

This section includes examples and details with `schema.Attribute` implemented as a Go structure type as it exists today. Future design considerations around creating specialized or custom attribute types may warrant switching this to an interface type with separate concrete types.

##### `ValueValidator` Field on `schema.Attribute`

Similar to the previous framework, a new field can be added to the `schema.Attribute` type. For example:

```go
schema.Attribute{
    // ...
    ValueValidator: T,
}
```

Implementators would be responsible for ensuring that single function covered all necessary validation. The framework could provide wrapper functions similar to the previous `All()` and `Any()` to allow simpler validations built from multiple functions. For example:

```go
schema.Attribute{
    // ...
    ValueValidator: All(
        T,
        T,
    ),
}
```

As seen with the previous framework in practice however, it was very common to implement the `All()` wrapper function. New provider developers would be responsible for understanding that multiple validations are possible in the single function field and knowing that custom validation functions may not be necessary to write if using the wrapper functions.

This proposal colocates the value validation behaviors in the schema definition, meaning it is easier for provider developers to discover this type of validation and correlate the validation logic to the name and type information.

##### `ValueValidators` Field on `schema.Attribute`

A new field that accepts a list of functions can be added to the `schema.Attribute` type. For example:

```go
schema.Attribute{
    // ...
    ValueValidators: []T{
        T,
        T,
    },
}
```

In this case, the framework would perform the validation similar to the previous framework `All()` wrapper function. The logical `AND` type of value validation is overwhelmingly more common in practice, which will simplify provider implementations. This still allows for an `Any()` based wrapper (logical `OR`) to be inserted if necessary.

Colocating the value validation behaviors in the schema definition, means it is easier for provider developers to discover this type of validation and correlate the validation logic to the name and type information. This proposal will feel familiar to existing provider developers. New provider developers will immediately know that multiple validations are supported.

##### New Attribute With Value Validation Type(s)

The `schema.Attribute` type could be converted to a Go interface type and split into capabilities, similar to other interface types in the framework. For example:

```go
type Attribute interface {
    Type(context.Context) attr.Type
    // ...
}

type AttributeWithValueValidators struct {
    Attribute
    ValueValidators []T
}

// or more interfaces

type AttributeWithValueValidators interface {
    Attribute
    ValueValidators(/* ... */) []T
}
```

This type of proposal, in isolation, feels extraneous given the current attribute implementation. The framework does not appear to benefit from this splitting and it seems desirable that all attributes should be able to optionally enable value validation. Future considerations to allow declaring custom attribute types, outside of validation handling, are more likely to drive this type of potential change.

##### Resource Level Attribute Value Validation Handling

This proposal would introduce no changes to `schema.Attribute`. Instead, this would require value validation declarations at the `DataSource` and `Resource` level similar to other proposed attribute validations in the [Declaring Multiple Attribute Validation for Resources](#declaring-multiple-attribute-validation-for-resources) section.

This proposal makes value validation behaviors occur at a distance, meaning it is harder for provider developers to correlate the validation logic to the name/path and type information. It would also be very verbose for even moderately sized schemas with thorough value validation. The only real potential benefit to this type of value validation is that the framework implementation is very straightforward, to just go through this single list of validations instead of walking all attributes.

It could be possible to implement another proposal in this space, while also supporting this one, however this could introduce unnecessary complexity into the implementation.

#### Defining Attribute Value Validation Functions

This section includes examples with incoming types as `tftypes.AttributePath` and the `attr.Value` interface type with an output type of `error`. These implementation details are discussed later and only shown for simpler illustrative purposes here.

##### `AttributeValueValidatorFunc` Type

A new Go type could be created that defines the signature of a value validation function, similar to the previous framework `SchemaValidateFunc`. For example:

```go
type AttributeValueValidatorFunc func(context.Context, path *tftypes.AttributePath, value attr.Value) error
```

To support passing through the provider instance to the function, the parameters would also need to include a `tfsdk.Provider` interface type:

```go
type AttributeValueValidatorFunc func(context.Context, provider tfsdk.Provider, path *tftypes.AttributePath, value attr.Value) error
```

While the simplest implementation, this proposal does not allow for documentation hooks.

##### `attr.ValueValidator` Interface

A new Go interface type could be created that defines an extensible value validation function type. For example:

```go
type ValueValidator interface {
    Description(context.Context) string
    MarkdownDescription(context.Context) string
    Validate(context.Context, path *tftypes.AttributePath, value attr.Value) error
}
```

With an example implementation:

```go
type stringLengthBetweenValidator struct {
    ValueValidator

    maximum int
    minimum int
}

func (v stringLengthBetweenValidator) Description(_ context.Context) string {
    return fmt.Sprintf("length must be between %d and %d", v.minimum, v.maximum)
}

func (v stringLengthBetweenValidator) MarkdownDescription(_ context.Context) string {
    return fmt.Sprintf("length must be between `%d` and `%d`", v.minimum, v.maximum)
}

func (v stringLengthBetweenValidator) Validate(ctx context.Context, path *tftypes.AttributePath, rawValue attr.Value) error {
    value, ok := rawValue.(types.String)
    
    if !ok {
        return fmt.Errorf("%s with incorrect type: %T", path, rawValue)
    }

    if value.Unknown {
        return fmt.Errorf("%s with unknown value", path)
    }

    if len(value.Value) < v.minimum || len(value.Value) > v.maximum {
        return fmt.Errorf("%s with value %q %s", path, value.Value, v.Description(ctx))
    }

    return nil
}

func StringLengthBetween(minimum int, maximum int) stringLengthBetweenValidator {
    return stringLengthBetweenValidator{
        maximum: maximum,
        minimum: minimum,
    }
}
```

This helps solve the documentation issue with the following example slice type alias and receiver method:

```go
// ValueValidators implements iteration functions across ValueValidator
type ValueValidators []ValueValidator

// Descriptions returns all ValueValidator Description
func (vs ValueValidators) Descriptions(ctx context.Context) []string {
    result := make([]string, 0, len(vs))

    for _, v := range vs {
        result = append(result, v.Description(ctx))
    }
    return result
}
```

To support passing through the provider instance, a separate interface type could be introduced that includes a function call with the `tfsdk.Provider` interface type:

```go
type ValueValidatorWithProvider interface {
    ValueValidator
    ValidateWithProvider(context.Context, provider tfsdk.Provider, path *tftypes.AttributePath, value attr.Value) error
}
```

#### Attribute Value Validation Function Value Parameter

Regardless the choice of concrete or interface types for the value validation functions, the parameters and returns for the implementations will play a crucial role on the extensibility and development experience.

##### `attr.Value` Type

The simplest implementation in the framework that could occur in all function types or interfaces is directly supplying an `attr.Value` and requiring implementations to handle all type conversion:

```go
func (v someValidator) Validate(ctx context.Context, path *tftypes.AttributePath, rawValue attr.Value) error {
    value, ok := rawValue.(types.String)
    
    if !ok {
        return fmt.Errorf("%s with incorrect type: %T", path, rawValue)
    }

    // ... rest of logic ...
```

This proposal would strongly encourage all implementations to handle value type conversions since using a stronger type in the function signature would risk panics that the framework cannot prevent and the compiler cannot check. Any error handling here could become inconsistent across implementations. This type conversion logic feels like an unnecessary burden on implementors and could reduce the developer experience as this logic would always need to be repeated with little to no actual utility.

##### `types.T` Type

If using an `attr.ValueValidator` interface approach, multiple new Go interface types could be created that define extensible value validation functions with strong typing. For example:

```go
// ValueValidator describes common validation functionality
type ValueValidator interface {
    Description(context.Context) string
    MarkdownDescription(context.Context) string
}

// StringValueValidator describes String value validation
type StringValueValidator interface {
    ValueValidator
    Validate(context.Context, *tftypes.AttributePath, types.String) error
}
```

Then, this framework can handle the appropriate type conversions and error handling:

```go
// Validate performs all validation functions.
//
// Each type performs conversion or returns a conversion error
// prior to executing the typed validation function.
func (vs ValueValidators) Validate(ctx context.Context, path *tftypes.AttributePath, rawValue attr.Value) error {
    for _, validator := range vs {
        switch typedValidator := validator.(type) {
        case StringValueValidator:
            value, ok := rawValue.(types.String)

            if !ok {
                return fmt.Errorf("%s with incorrect type: %T", path, rawValue)
            }

            if err := typedValidator.Validate(ctx, path, value); err != nil {
                return err
            }
        default:
            return fmt.Errorf("unknown validator type: %T", validator)
        }
    }

    return nil
}
```

Leaving the implementations to only be concerned with the typed value:

```go
func (v stringLengthBetweenValidator) Validate(ctx context.Context, path *tftypes.AttributePath, value types.String) error {
    if value.Unknown {
        return fmt.Errorf("%s with unknown value", path)
    }

    if len(value.Value) < v.minimum || len(value.Value) > v.maximum {
        return fmt.Errorf("%s with value %q %s", path, value.Value, v.Description(ctx))
    }

    return nil
}
```

This proposal allows each validation function to be succinctly defined with the expected value type. It may be possible to get the validation function implementations even closer to the true value logic if unknown values are also handled automatically by this framework, however that decision can be made further along in the design process.

Even with this type of implementation, it is theoretically possible to create a "generic" type handler for escaping the strongly typed logic if necessary:

```go
// GenericValueValidator describes value validation without a strong type.
//
// While it is generally preferred to use the typed validation interfaces,
// such as StringValueValidator, this interface allows custom implementations
// where the others may not be suitable. The Validate function is responsible
// for protecting against attr.Value type assertion panics.
type GenericValueValidator interface {
    ValueValidator
    Validate(context.Context, *tftypes.AttributePath, attr.Value) error
}
```

Offering the largest amount of flexibility for implementors to choose the level of desired abstraction, while not hindering more advanced implementations.

To support passing through the provider instance, separate interface types could be introduced that include a function call with the `tfsdk.Provider` interface type:

```go
type StringValueValidatorWithProvider interface {
    ValueValidator
    ValidateWithProvider(context.Context, provider tfsdk.Provider, path *tftypes.AttributePath, value types.String) error
}
```

#### Attribute Value Validation Function Path Parameter

Another consideration with attribute value validation functions is whether the implementation should be responsible for adding context around the attribute path under validation and how that information (if provided) is surfaced to the function body.

##### No Attribute Path Parameter

Validation function implementations could potentially not have access to the attribute path under validation, instead relying on surrounding logic to handle wrapping errors or logging to include the path. For example:

```go
tflog.Debug(ctx, "validating attribute path (%s) attribute value (%s): %s", attributePath.String(), value, validator.Description())

err := validator.Validate(ctx, value)

if err != nil {
    return fmt.Errorf("%s: %w", attributePath.String(), err)
}
```

This could be a double edged sword for extensibility. Implementators do not need to worry about handling the attribute path in error messages that are returned to practitioners or manually adding logging around it. This does however prevent the ability to provide that additional context to the validation logic, if for example the logic warrants making decisions based on the given path or additional logging that includes the full path. In practice with validation functions in the previous framework, path based decisions are rare at best, and this framework could be opinionated against that particular pattern.

##### Adding Attribute Path to Context

This framework could inject additional validation information into the `context.Context` being passed through to the validation functions. For example:

```go
const ValidationAttributePathKey = "validation_attribute_path"

validationCtx := context.WithValue(ctx, ValidationAttributePathKey, attributePath)
validator.Validate(ctx, value)
```

With implementations referencing this data:

```go
func (v someValidator) Validate(ctx context.Context, rawValue attr.Value) error {
    // ...
    rawAttributePath := ctx.Value(ValidationAttributePathKey)

    attributePath, ok := rawAttributePath.(*tftypes.AttributePath)

    if !ok {
        return fmt.Errorf("unexpected %s context value type: %T", ValidationAttributePathKey, rawAttributePath)
    }
    // ...
```

This experience seems subpar for developers though as they must know about the special context value(s) available and how to reference them appropriately, especially to avoid a type assertion panic. In this case, it seems more appropriately to pass the parameter directly, if necessary.

##### `string` Type

The attribute path could be passed to validation functions as its string representation. For example:

```go
validator.Validate(ctx, attributePath.String(), value)
```

This would allow implementors to ignore the details of what the attribute path is or how to represent it appropriately. However, this seems unnecessarily limiting should the path information need to be used in the logic. In this case, calling a Go conventional `String()` receiver method on the actual attribute path type does not feel like a development burden for implementors as necessary.

##### `*tftypes.AttributePath` Type

The attribute path could be passed to validation functions directly using `*tftypes.AttributePath` or its abstraction in this framework. For example:

```go
validator.Validate(ctx, attributePath, value)
```

This provides the ultimate flexibility for implementors, making the path information fully available in logic, logging, etc. This framework's design could also borrow ideas from the [No Attribute Path Parameter](#no-attribute-path-parameter) section and automatically handle logging and wrapping where appropriate, leaving it completely optional for implementators to handle the path information.

#### Attribute Value Validation Function Returns

Depending on the validation function design, there could be important details about the validation process that need to be surfaced to callers. This section walks through different proposals on how information can be returned to callers.

##### Attribute Value Validation Function `bool` Return

Validation functions could implement return information via a `bool` type. For example:

```go
func (v stringLengthBetweenValidator) Validate(ctx context.Context, path *tftypes.AttributePath, rawValue attr.Value) bool {
    value, ok := rawValue.(types.String)
    
    if !ok {
        return false
    }

    if value.Unknown {
        return false
    }

    return len(value.Value) > v.minimum && len(value.Value) < v.maximum
}
```

This proposal encodes no information in the response from these functions beyond a simple boolean "validation passed" versus "validation failed" value. Information such as whether validation failed due to type conversion problems or validation could not be performed due to an unknown value is hidden. Giving the ability for functions to surface details about unsuccessful validation back to callers is likely required broader utility in this framework and extensions to it.

In this scenario, it is this framework's responsibility to generate the appropriate diagnostic back. Implementors will not be able to influence the level, summary, or details associated with that diagnostic.

##### Attribute Value Validation Function `error` Return

Validation functions could implement return information via an untyped `error`. For example:

```go
func (v stringLengthBetweenValidator) Validate(ctx context.Context, path *tftypes.AttributePath, rawValue attr.Value) error {
    value, ok := rawValue.(types.String)
    
    if !ok {
        return fmt.Errorf("%s with incorrect type: %T", path, rawValue)
    }

    if value.Unknown {
        return fmt.Errorf("%s with unknown value", path)
    }

    if len(value.Value) < v.minimum || len(value.Value) > v.maximum {
        return fmt.Errorf("%s with value %q %s", path, value.Value, v.Description(ctx))
    }

    return nil
}
```

In this scenario, callers will know that validation did not pass, but not necessarily why. This proposal is only marginally better than the `bool` return value, as some manual error message context can be provided about the problem that caused the failure. However short of perfectly consistent error messaging which is not feasible to enforce in all implementors, callers will still not reasonably be able to perform actions based on the differing reasons for errors.

In this scenario, it is this framework's responsibility to generate the appropriate diagnostic back. Implementors will not be able to influence the level or summary associated with that diagnostic. The details would likely include the error messaging.

##### Attribute Value Validation Function Typed Error Return

This framework could provide typed errors for validation functions. For example:

```go
type ValueValidatorInvalidTypeError struct {
    Path *tftypes.AttributePath
    Value attr.Value
}

// Error implements the error interface
func (e ValueValidatorInvalidTypeError) Error() string {
    // ...
}

type ValueValidatorInvalidValueError struct {
    Description string
    Path *tftypes.AttributePath
    Value attr.Value
}

// Error implements the error interface
func (e ValueValidatorInvalidValueError) Error() string {
    // ...
}

type ValueValidatorUnknownValueError struct {
    Path *tftypes.AttributePath
}

// Error implements the error interface
func (e ValueValidatorUnknownValueError) Error() string {
    // ...
}
```

With implementators able to return these such as:

```go
func (v stringLengthBetweenValidator) Validate(ctx context.Context, path *tftypes.AttributePath, rawValue attr.Value) error {
    value, ok := rawValue.(types.String)

    if !ok {
        return ValueValidatorInvalidTypeError{
            Path: path,
            Value: rawValue,
        }
    }

    if value.Unknown {
        return ValueValidatorUnknownValueError{
            Path: path,
        }
    }

    if len(value.Value) < v.minimum || len(value.Value) > v.maximum {
        return ValueValidatorInvalidValueError{
            Description: v.Description(ctx),
            Path: path,
            Value: value,
        }
    }

    return nil
}
```

This framework could also go further and require using one of these error types:

```go
type ValueValidatorError interface {}

// ...

type ValueValidatorInvalidTypeError struct {
    ValueValidatorError

    Path *tftypes.AttributePath
    Value attr.Value
}

// ...

type ValueValidator interface {
    // ...
    Validate(context.Context, *tftypes.AttributePath, attr.Value) ValueValidatorError
}
```

Meaning that extensibility is guaranteed to follow certain compile time rules.

In either the `error` or `ValueValidatorError` interface type scenarios, this allows callers to react to the responses by checking for underlying error types. For example, it is possible to implement a generic `Not()` (logical `NOT`) validation function that catches invalid values but passes through other errors:

```go
func (v notValidator) Validate(ctx context.Context, path *tftypes.AttributePath, rawValue attr.Value) error {
    var invalidValueError ValueValidatorInvalidValueError

    err := v.validator.Validate(ctx, path, rawValue)

    if err == nil {
        return ValueValidatorInvalidValueError{
            Description: v.Description(ctx),
            Path: path,
            Value: rawValue,
        }
    }

    if errors.As(err, &invalidValueError) {
        return nil
    }

    return err
}
```

In this scenario, it is this framework's responsibility to generate the appropriate diagnostic back. Implementors will not be able to influence the level or summary associated with that diagnostic. The details would likely include the error messaging based on the error type implementations, although if it was warranted for extensibility, there could also be a "generic" `ValueValidatorError` type (or when there is an unrecognized `error` type) that this framework would pass over except transferring the messaging through to the diagnostic. Additional warning-only types could also be provided to allow further diagnostic customization.

##### Attribute Value Validation Function Diagnostic Return

Validation functions could directly return a `*tfprotov6.Diagnostic` or abstracted type from this framework. For example:

```go
func (v stringLengthBetweenValidator) Validate(ctx context.Context, path *tftypes.AttributePath, rawValue attr.Value) *tfprotov6.Diagnostic {
    value, ok := rawValue.(types.String)

    if !ok {
        return &tfprotov6.Diagnostic{
            Severity: tfprotov6.DiagnosticSeverityError,
            Summary: "Incorrect validation type",
            Details: fmt.Sprintf("%s with incorrect type: %T", path, rawValue),
        }
    }

    if value.Unknown {
        return &tfprotov6.Diagnostic{
            Severity: tfprotov6.DiagnosticSeverityError,
            Summary: "Unknown validation value",
            Details: fmt.Sprintf("received unknown value at path: %s", path),
        }
    }

    if len(value.Value) < v.minimum || len(value.Value) > v.maximum {
        return &tfprotov6.Diagnostic{
            Severity: tfprotov6.DiagnosticSeverityError,
            Summary: "Value validation failed",
            Details: fmt.Sprintf("%s with value %q %s", path, value.Value, v.Description(ctx))
        }
    }

    return nil
}
```

In this scenario, it the implementor's responsibility to generate the appropriate diagnostic back, but they have full control of the output. It could be difficult for the framework to enforce implementation rules around these responses or potentially allow configuration overrides for them without creating more abstractions on top of this type or additional helper functions. Differing diagnostic implementations could introduce confusion for practitioners.

In general, this proposal feels very similar to either the generic `error` type or typed error proposals above (depending on the implmentation details) with minimal utility over them beyond complete output customization.

### Multiple Attribute Validation

This framework should also provide the ability to handle validation situations across multiple attributes as noted in the goals. Some of the proposals from the [Single Attribute Value Validation](#single-attribute-value-validation) section are applicable for these proposals as well, so they are largely omitted here for brevity. Examples showing `attr.Value`, `*tftypes.AttributePath`, and bare `error` types are for illustrative purposes, whose final forms would be determined by those proposals.

#### Declaring Multiple Attribute Validation for Attributes

The previous framework implemented behaviors, such as `ConflictsWith`, as an individual field per behavior within each attribute. This section of proposals targets this specific functionality. One major caveat to these proposals is that they should not be considered exclusive to attribute value validations as it may be desirable to provide some consistency between the two implementations to improve developer experience.

This section includes examples and details with `schema.Attribute` implemented as a Go structure type as it exists today. Future design considerations around creating specialized or custom attribute types may warrant switching this to an interface type with separate concrete types.

##### Individual Behavior Fields on `schema.Attribute`

Similar to the previous framework, individual fields for each attribute validation could be added to the `schema.Attribute` type which accepts multiple attribute paths. For example:

```go
schema.Attribute{
    // ...
    ConflictsWith: []tftypes.AttributePath,
}
```

A potential downside is that these behaviors cannot support the notion of conditional logic without changes to the implementations, since they can only be existence based if passed an attribute path. Allowing value validations in the declarations (on either side), could allieviate this issue. For example:

```go
schema.Attribute{
    // ...
    ConflictsWith: []func(AttributeValueValidator, tftypes.AttributePath, AttributeValueValidator),
}
```

Regardless of the potential value handling, this proposal would feel familiar for existing provider developers and be relatively trivial for them to implement. One noticable downside to this approach however is that there can be any number of related, but disjointed attribute behaviors. The previous framework supported four of these already and there is logical room for addtional behaviors, making updates to the `schema.Attribute` type a limiting factor in this validation space. This proposal also differs from value validation proposals, which are focused around a single field.

##### `PathValidator` Field on `schema.Attribute`

A new field for attribute validation can be added to the `schema.Attribute` type. For example:

```go
schema.Attribute{
    // ...
    PathValidator: T,
}
```

Implementators would be responsible for ensuring that single function covered all necessary validation. The framework could provide wrapper functions similar to the previous `All()` and `Any()` of `ValidateFunc` to allow simpler validations built from multiple functions. For example:

```go
schema.Attribute{
    // ...
    PathValidator: All(
        T,
        T,
    ),
}
```

As seen with the previous framework in practice however, it was very common to implement the `All()` wrapper function. New provider developers would be responsible for understanding that multiple validations are possible in the single function field and knowing that custom validation functions may not be necessary to write if using the wrapper functions.

This proposal colocates the attribute validation behaviors in the schema definition, meaning it is easier for provider developers to discover this type of validation and correlate the validation logic to the name and type information.

##### `PathValidators` Field on `schema.Attribute`

A new field that accepts a list of functions can be added to the `schema.Attribute` type. For example:

```go
schema.Attribute{
    // ...
    PathValidators: []T{
        T,
        T,
    },
}
```

In this case, the framework would perform the validation similar to the previous framework `All()` wrapper function for `ValidateFunc`. The logical `AND` type of value validation is overwhelmingly more common in practice, which will simplify provider implementations. This still allows for an `Any()` based wrapper (logical `OR`) to be inserted if necessary.

Colocating the attribute validation behaviors in the schema definition, means it is easier for provider developers to discover this type of validation and correlate the validation logic to the name and type information. This proposal will feel familiar to existing provider developers. New provider developers will immediately know that multiple validations are supported.

##### Combined `Validators` Field on `schema.Attribute`

A new field that accepts the union of [`ValueValidators` field on `schema.Attribute`](#valuevalidators-field-on-schemaattribute) and [`PathValidators` field on `schema.Attribute`](#pathvalidators-field-on-schemaattribute) can be added to the `schema.Attribute` type. For example:

```go
schema.Attribute(
    // ...
    Validators: []I(
        T1,
        T2,
    )
)
```

Since value validation functions would inherently be implemented different than path validation functions and they are conceptually similar but different in certain ways, this could be complex to implement or understand correctly. When trying to handle documentation output for example, this framework or callers would need to distinguish between the two validation types to ensure the intended validation meanings are correct.

##### Resource Level Attribute Path Validation Handling

Rather than adjusting the `schema.Attribute` type for this type of validation, it could be forced to the resource (or data source) level. The [Declaring Multiple Attribute Validation for Resources](#declaring-multiple-attribute-validation-for-resources) proposals presented later are revelant for this section. To prevent proposal duplication, please see that section for more details and associated tradeoffs.

#### Declaring Multiple Attribute Validation for Resources

In the previous framework, the `CustomizeDiff` functionality enabled resource (or data source) level validation as a logical catch-all. These proposals cover the next iteration of that type of functionality.

##### `PlanModifications` for Resources

The [Plan Modifications design documentation](./plan-modifications.md) outlines proposals which broadly replace the previous framework's `CustomizeDiff` functionality. See that documentation for considerations and recommendations there. In this proposal for validation, new functions for validation would be provided within that framework, rather than introducing separate handling.

Implementing against that design could prove complex for the framework as they are intended to serve differing purposes. It could also be confusing for provider developers in the same way that `CustomizeDiff` was confusing where differing logical rules applied to differing attribute value and operation scenarios.

##### `AttributeValidators` for Resources

This introduces a new extension interface type for `ResourceType` and `DataSourceType`. For example:

```go
type DataSourceTypeWithAttributeValidators interface {
    DataSourceType
    AttributeValidators(context.Context) AttributeValidators
}

type ResourceTypeWithAttributeValidators interface {
    ResourceType
    AttributeValidators(context.Context) AttributeValidators
}
```

Where `AttributeValidators` is a slice of types to be discussed later.

As an example sketch, provider developers could introduce a function that fulfills the new interface with example helpers such as:

```go
func (t *customResourceType) AttributeValidators(ctx context.Context) AttributeValidators {
    return AttributeValidators{
      ConflictingAttributes(*tftypes.AttributePath, *tftypes.Attribute),
      ConflictingAttributesWithValues(*tftypes.AttributePath, ValueValidator, *tftypes.AttributePath, ValueValidator),
      PrerequisiteAttribute(*tftypes.AttributePath, *tftypes.AttributePath),
      PrerequisiteAttributeWithValue(*tftypes.AttributePath, ValueValidator, *tftypes.AttributePath),
    }
}
```

This setup would allow for the framework to provide flexible resource level validation with a low amount of friction for provider developers. Helper functions would be extensible and make the behaviors clear.

#### Defining Attribute Validation Functions

This section includes examples with parameter types as `tftypes.AttributePath` and the `attr.Value` interface type with an return type of `error`. These implementation details are shown for simpler illustrative purposes here, but will likely depend on the outcome from the [Single Attribute Value Validation](#single-attribute-value-validation) proposals.

##### `AttributeValidatorsFunc` Type

A new Go type could be created that defines the signature of a value validation function, similar to the previous framework `SchemaValidateFunc`. For example:

```go
type AttributeValidatorsFunc func(context.Context, path1 *tftypes.AttributePath, value1 attr.Value, path2 *tftypes.AttributePath, value2 attr.Value) error
```

To support passing through the provider instance to the function, the parameters would also need to include a `tfsdk.Provider` interface type:

```go
type AttributeValidatorsFunc func(context.Context, provider tfsdk.Provider, path1 *tftypes.AttributePath, value1 attr.Value, path2 *tftypes.AttributePath, value2 attr.Value) error
```

This proposal does not allow for documentation hooks. It could be confusing for implementors as they could be responsible for more complex validation logic or provider developers if many iterations of validation are implemented across many different functions since each would be unique. It might be possible to reduce this burden by passing in a `ValueValidator` as well.

##### `AttributeValidator` Interface

A new Go interface type could be created that defines an extensible attribute validation function type. For example:

```go
type AttributeValidator interface {
    Description(context.Context) string
    MarkdownDescription(context.Context) string
    Validate(context.Context, path1 *tftypes.AttributePath, value1 attr.Value, path2 *tftypes.AttributePath, value2 attr.Value) error
}
```

With an example implementation:

```go
type conflictingAttributesValidator struct {
    AttributeValidator

    path1 *tftypes.AttributePath
    path2 *tftypes.AttributePath
}

func (v conflictingAttributesValidator) Description(_ context.Context) string {
    return fmt.Sprintf("%s and %s cannot both be configured", v.path1.String(), v.path2.String())
}

func (v conflictingAttributesValidator) MarkdownDescription(_ context.Context) string {
    return fmt.Sprintf("`%s` and `%s` cannot both be configured", v.path1.String(), v.path2.String())
}

func (v conflictingAttributesValidator) Validate(ctx context.Context, _ *tftypes.AttributePath, _ attr.Value, _ *tftypes.AttributePath, _ attr.Value) error {
    if /* v.path1 configured */ && /* v.path2 configured */ {
        return fmt.Errorf("%s", v.Description(ctx))
    }

    return nil
}

func ConflictingAttributes(path1 *tftypes.AttributePath, path2 *tftypes.AttributePath) conflictingAttributesValidator {
    return conflictingAttributesValidator{
        path1: path1,
        path2: path1,
    }
}
```

This helps solve the documentation issue with the following example slice type alias and receiver method:

```go
// AttributeValidators implements iteration functions across AttributeValidator
type AttributeValidators []AttributeValidator

// Descriptions returns all AttributeValidator Description
func (vs AttributeValidators) Descriptions(ctx context.Context) []string {
    result := make([]string, 0, len(vs))

    for _, v := range vs {
        result = append(result, v.Description(ctx))
    }
    return result
}
```

To support passing through the provider instance, a separate interface type could be introduced that includes a function call with the `tfsdk.Provider` interface type:

```go
type AttributeValidatorWithProvider interface {
    AttributeValidator
    ValidateWithProvider(context.Context, provider tfsdk.Provider, path1 *tftypes.AttributePath, value1 attr.Value, path2 *tftypes.AttributePath, value2 attr.Value) error
}
```

## Recommendations

This section will summarize the proposals into specific recommendations for each topic. Code examples are provided in following sections to illustrate the concepts. The final section provides some future considerations for the framework and terraform-plugin-go.

### Overview

Defining all validation functionality via interface types will offer the framework the most flexibility for future enhancements while ensuring consistent implementations. Furthermore for attribute value validation functions, providing strongly typed interfaces for common value types will reduce implementor burden and ensure consistent invalid type error messaging, rather than potential panic scenarios.

Attribute value validations should be implemented as a slice of the interface type on `schema.Attribute`. Multiple attribute validation, such as declaring conflicting attributes on attributes themselves, should be implemented as a separate slice of that differing interface type on `schema.Attribute`. This would be in addition to supporting that functionality with resource level multiple attribute validation.

Attribute value validations should be required to accept the attribute path in its native type as a parameter. This will allow a flexible implementation for provider developers that may desire advanced logic based on the path.

Validation functions should be required to return framework defined error types. These errors will either result in the framework returning consistent error diagnostics or callers (such as wrapper validation functions) otherwise handling these results in a predictable manner. Error types that equate to consistent warning diagnostics can also be provided, if desired.

Resource level multiple attribute validation functions should be implemented separately from plan modifications to separate concerns. For example:

### Attribute Level Example Implementation

Example framework code:

```go
// Well defined error types
type ValueValidatorError interface {}

type ValueValidatorInvalidTypeError interface {
    ValueValidatorError
}

type ValueValidatorUnknownValueError interface {
    ValueValidatorError
}

type ValueValidatorUnsuccessfulValidationError interface {
    ValueValidatorError
}

// ValueValidator is an interface type for implementing common validation functionality.
type ValueValidator interface {
    Description(context.Context) string
    MarkdownDescription(context.Context) string
}

// ValueValidators is a type alias for a slice of ValueValidator.
type ValueValidators []ValueValidator

// Descriptions returns all ValueValidator Description
func (vs ValueValidators) Descriptions(ctx context.Context) []string {
    // ...
}

// MarkdownDescriptions returns all ValueValidator MarkdownDescription
func (vs ValueValidators) MarkdownDescriptions(ctx context.Context) []string {
    // ...
}

// Validates performs all ValueValidator Validate or ValidateWithProvider
func (vs ValueValidators) Validates(ctx context.Context) diag.Diagnostics {
    // ...
}

// GenericValueValidator describes value validation without a strong type.
//
// While it is generally preferred to use the typed validation interfaces,
// such as StringValueValidator, this interface allows custom implementations
// where the others may not be suitable. The Validate function is responsible
// for protecting against attr.Value type assertion panics.
type GenericValueValidator interface {
    ValueValidator
    Validate(context.Context, *tftypes.AttributePath, attr.Value) error
}

// StringValueValidator is an interface type for implementing String value validation.
type StringValueValidator interface {
    ValueValidator
    Validate(context.Context, *tftypes.AttributePath, types.String) ValueValidatorError
}

// StringValueValidatorWithProvider is an interface type for implementing String value validation with a provider instance.
type StringValueValidatorWithProvider interface {
    StringValueValidator
    ValidateWithProvider(context.Context, tfsdk.Provider, *tftypes.AttributePath, types.String) ValueValidatorError
}

type Attribute struct {
    // ...
    PathValidators  AttributeValidators // described below
    ValueValidators ValueValidators
}
```

Example validation function code:

```go
type stringLengthBetweenValidator struct {
    StringValueValidator

    maximum int
    minimum int
}

func (v stringLengthBetweenValidator) Description(_ context.Context) string {
    return fmt.Sprintf("length must be between %d and %d", v.minimum, v.maximum)
}

func (v stringLengthBetweenValidator) MarkdownDescription(_ context.Context) string {
    return fmt.Sprintf("length must be between `%d` and `%d`", v.minimum, v.maximum)
}

func (v stringLengthBetweenValidator) Validate(ctx context.Context, path *tftypes.AttributePath, value types.String) ValueValidatorError {
    if value.Unknown {
        return ValueValidatorUnknownValueError{
            Path: path,
        }
    }

    if len(value.Value) < v.minimum || len(value.Value) > v.maximum {
        return ValueValidatorUnsuccessfulValidationError{
            Description: v.Description(ctx),
            Path: path,
            Value: value,
        }
    }

    return nil
}

func StringLengthBetween(minimum int, maximum int) stringLengthBetweenValidator {
    return stringLengthBetweenValidator{
        maximum: maximum,
        minimum: minimum,
    }
}
```

Example provider code:

```go
schema.Attribute{
    Type:             types.StringType,
    Required:         true,
    PathValidators:  AttributeValidators{
        ConflictsWithAttribute(tftypes.NewAttributePath().AttributeName("other_attribute")),
    },
    ValueValidators: ValueValidators{
        StringLengthBetween(1, 256),
    },
}
```

### Resource Level Example Implementation

Example framework code:

```go
// Well defined error types
type AttributeValidatorError interface {}

type AttributeValidatorInvalidTypeError interface {
    AttributeValidatorError
}

type AttributeValidatorUnknownValueError interface {
    AttributeValidatorError
}

type AttributeValidatorUnsuccessfulValidationError interface {
    AttributeValidatorError
}

// AttributeValidator is an interface type for declaring multiple attribute validations.
type AttributeValidator interface {
    Description(context.Context) string
    MarkdownDescription(context.Context) string
    Validate(context.Context, path1 *tftypes.AttributePath, value1 attr.Value, path2 *tftypes.AttributePath, value2 attr.Value) AttributeValidatorError
}

// AttributeValidators is a type alias for a slice of AttributeValidator.
type AttributeValidators []AttributeValidator

// Descriptions returns all AttributeValidator Description
func (vs AttributeValidators) Descriptions(ctx context.Context) []string {
    // ...
}

// MarkdownDescriptions returns all AttributeValidator MarkdownDescription
func (vs AttributeValidators) MarkdownDescriptions(ctx context.Context) []string {
    // ...
}

// Validates performs all AttributeValidator Validate or ValidateWithProvider
func (vs AttributeValidators) Validates(ctx context.Context) diag.Diagnostics {
    // ...
}

// AttributeValidatorWithProvider is an interface type for declaring multiple attribute validation that requires a provider instance.
type AttributeValidatorWithProvider interface {
    AttributeValidator
    ValidateWithProvider(context.Context, provider tfsdk.Provider, path1 *tftypes.AttributePath, value1 attr.Value, path2 *tftypes.AttributePath, value2 attr.Value) AttributeValidatorError
}

// DataSourceTypeWithAttributeValidators is an interface type that extends DataSourceType to include attribute validations.
type DataSourceTypeWithAttributeValidators interface {
    DataSourceType
    AttributeValidators(context.Context) AttributeValidators
}

// ResourceTypeWithAttributeValidators is an interface type that extends ResourceType to include attribute validations.
type ResourceTypeWithAttributeValidators interface {
    ResourceType
    AttributeValidators(context.Context) AttributeValidators
}
```

Example provider code:

```go
func (t *customResourceType) AttributeValidators(ctx context.Context) AttributeValidators {
    return AttributeValidators{
        ConflictingAttributes(
            tftypes.NewAttributePath().AttributeName("first_attribute"),
            tftypes.NewAttributePath().AttributeName("second_attribute"),
        ),
    }
}
```

### Future Considerations

It is recommended that the framework provide an abstracted `*tftypes.AttributePath` rather than depend on that type directly, but this can converted after an initial implementation. This is purely for decoupling the two projects, similar to other abstracted types already created in the framework.

To better support provider-based validation functionality in the future, it is also recommended that the `Provider` interface type also add a new `Configured(context.Context) bool` function or another methodology for easily checking the configuration state of a provider instance. Adding a setter function could also allow the framework to manage the provider configuration state automatically. This would simplify validations that require provider instances since it will likely be required that implementations need to check on this status as part of the validation logic.

It is recommended that the framework or the upstream terraform-plugin-go module provide functionality to declare relative attribute paths, such as "this" and "parent" methods to better enable nested attribute declarations. This will enable provider developers to create attribute paths such as:

```go
NewAttributePath(CurrentPath().Parent().AttributeName("other_attr"))
```
