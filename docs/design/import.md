# Import

Practitioners may have previously existing resources that they wish to bring under Terraform's management. Conceptually, this is called importing as Terraform must create the state of the resource. Providers are responsible for receiving a request for resource import and respond with the relevant resource state(s).

This design documentation will walk through and recommend options for import handling in the framework.

## Background

### Terraform CLI

Terraform CLI supports two forms of interacting with resources not in its statefile:

- [`terraform add`](https://www.terraform.io/docs/cli/commands/add.html) (*Experimental*, Terraform CLI version 1.1.0 and later): Given a resource address and using the resource schema provided by the provider, generate a configuration template.
- [`terraform import`](https://www.terraform.io/docs/cli/commands/import.html) (Terraform CLI version 0.7.0 and later): Given a resource address and an identifier, forwards the request to the relevant provider and expects relevant resource state(s) with enough information to perform a refresh, then performs that resource refresh to ensure the resource state is fully populated.

Since `terraform add` uses already available schema information and does not require additional provider interaction, this design focuses on the latter, which is a currently unimplemented integration point for providers in the framework.

Practitioners currently interact with `terraform import` by providing a resource address and import identifier, e.g.

```shell
terraform import aws_security_group.example sg-12345678
```

This is surfaced in the protocol as "type name" (parsed from the resource address) and "id" (passed through) as shown in the next section. Future enhancements may allow Terraform CLI to surface the entire resource configuration across the protocol as well, but there is no timeline for that design or implementation.

Terraform CLI also supports the ability to import multiple resource states during a single import. For example, resource import of an `aws_s3_bucket` could automatically import an `aws_s3_bucket_policy` into the state, if it exists. When encountering the resources beyond the first, Terraform CLI will save them into the state using the same label, e.g.

```console
$ terraform import aws_s3_bucket.example example-bucket
...
aws_s3_bucket_policy.example: Refreshing state... (ID: example-bucket)
aws_s3_bucket.example: Refreshing state... (ID: example-bucket)
```

In a more advanced example, the EC2 Security Group resource (`aws_security_group`) previously imported all EC2 Security Group Rules as resources (`aws_security_group_rule`). These would be saved into the state as if they were resources defined via `count`, e.g. `aws_security_group_rule.example[0]`, `aws_security_group_rule.example[1]`, etc.

### Terraform Plugin Protocol

The protocol defines the following implementation details for import handling:

```protobuf
service Provider {
    // ... other RPCs ...
    rpc ImportResourceState(ImportResourceState.Request) returns (ImportResourceState.Response);
}

message ImportResourceState {
    message Request {
        string type_name = 1;
        string id = 2;
    }

    message ImportedResource {
        string type_name = 1;
        DynamicValue state = 2;
        bytes private = 3;
    }

    message Response {
        repeated ImportedResource imported_resources = 1;
        repeated Diagnostic diagnostics = 2;
    }
}
```

### terraform-plugin-go

The `terraform-plugin-go` library, which underpins this framework, provides the following implementation (both `tfprotov5` and `tfprotov6`) of import types:

```go
// ImportResourceStateRequest is the request Terraform sends when it wants a
// provider to import one or more resources specified by an ID.
type ImportResourceStateRequest struct {
    // TypeName is the type of resource Terraform wants to import.
    TypeName string

    // ID is the user-supplied identifying information about the resource
    // or resources. Providers decide and communicate to users the format
    // for the ID, and use it to determine what resource or resources to
    // import.
    ID string
}

// ImportResourceStateResponse is the response from the provider about the
// imported resources.
type ImportResourceStateResponse struct {
    // ImportedResources are the resources the provider found and was able
    // to import.
    ImportedResources []*ImportedResource

    // Diagnostics report errors or warnings related to importing the
    // requested resource or resources. Returning an empty slice indicates
    // a successful validation with no warnings or errors generated.
    Diagnostics []*Diagnostic
}

// ImportedResource represents a single resource that a provider has
// successfully imported into state.
type ImportedResource struct {
    // TypeName is the type of resource that was imported.
    TypeName string

    // State is the provider's understanding of the imported resource's
    // state, represented as a `DynamicValue`. See the documentation for
    // `DynamicValue` for information about safely creating the
    // `DynamicValue`.
    //
    // The state should be represented as a tftypes.Object, with each
    // attribute and nested block getting its own key and value.
    State *DynamicValue

    // Private should be set to any state that the provider would like sent
    // with requests for this resource. This state will be associated with
    // the resource, but will not be considered when calculating diffs.
    Private []byte
}
```

## Prior Implementations

### terraform-plugin-sdk

The previous framework provided an [`Importer` field in the `helper/schema.Resource` type](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema#Resource.Importer):

```go
type Resource struct {
    // ... other fields ...

    // Importer is the ResourceImporter implementation for this resource.
    // If this is nil, then this resource does not support importing. If
    // this is non-nil, then it supports importing and ResourceImporter
    // must be validated. The validity of ResourceImporter is verified
    // by InternalValidate on Resource.
    Importer *ResourceImporter
}

// ResourceImporter defines how a resource is imported in Terraform. This
// can be set onto a Resource struct to make it Importable. Not all resources
// have to be importable; if a Resource doesn't have a ResourceImporter then
// it won't be importable.
//
// "Importing" in Terraform is the process of taking an already-created
// resource and bringing it under Terraform management. This can include
// updating Terraform state, generating Terraform configuration, etc.
type ResourceImporter struct {
    // State is called to convert an ID to one or more InstanceState to
    // insert into the Terraform state.
    //
    // Deprecated: State is deprecated in favor of StateContext.
    // Only one of the two functions can bet set.
    State StateFunc

    // StateContext is called to convert an ID to one or more InstanceState to
    // insert into the Terraform state. If this isn't specified, then
    // the ID is passed straight through. This function receives a context
    // that will cancel if Terraform sends a cancellation signal.
    StateContext StateContextFunc
}

// StateFunc is the function called to import a resource into the Terraform state.
//
// Deprecated: Please use the context aware equivalent StateContextFunc.
type StateFunc func(*ResourceData, interface{}) ([]*ResourceData, error)

// StateContextFunc is the function called to import a resource into the
// Terraform state. It is given a ResourceData with only ID set. This
// ID is going to be an arbitrary value given by the user and may not map
// directly to the ID format that the resource expects, so that should
// be validated.
//
// This should return a slice of ResourceData that turn into the state
// that was imported. This might be as simple as returning only the argument
// that was given to the function. In other cases (such as AWS security groups),
// an import may fan out to multiple resources and this will have to return
// multiple.
//
// To create the ResourceData structures for other resource types (if
// you have to), instantiate your resource and call the Data function.
type StateContextFunc func(context.Context, *ResourceData, interface{}) ([]*ResourceData, error)
```

When defined, the resource could respond to the `ImportResource` RPC, otherwise an error was returned that the resource does not support import.

If the resource state could be entirely fetched using an import identifier that matched the resource identifier, the `ImportStatePassthrough` and `ImportStatePassthroughContext` helpers simplified provider implementations, e.g.

```go
Importer: &schema.ResourceImporter{
    State: schema.ImportStatePassthrough,
},
```

Otherwise, custom provider implementations were required. The following example shows multiple resource import that was previously implemented in `aws_s3_bucket` resource to optionally also import an associated `aws_s3_bucket_policy` resource:

```go
// Importer: &schema.ResourceImporter{
//     State: resourceAwsS3BucketImportState,
// },

func resourceAwsS3BucketImportState(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
    results := make([]*schema.ResourceData, 1)
    results[0] = d

    conn := meta.(*AWSClient).s3conn
    pol, err := conn.GetBucketPolicy(&s3.GetBucketPolicyInput{
        Bucket: aws.String(d.Id()),
    })
    if err != nil {
        if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NoSuchBucketPolicy" {
            // Bucket without policy
            return results, nil
        }
        return nil, fmt.Errorf("Error importing AWS S3 bucket policy: %s", err)
    }

    policy := resourceAwsS3BucketPolicy()
    pData := policy.Data(nil)
    pData.SetId(d.Id())
    pData.SetType("aws_s3_bucket_policy")
    pData.Set("bucket", d.Id())
    pData.Set("policy", pol.Policy)
    results = append(results, pData)

    return results, nil
}
```

The [previous implementation of `aws_security_group` import]((https://github.com/hashicorp/terraform-provider-aws/blob/v2.70.0/aws/import_aws_security_group.go)), shows the more advanced case of importing an unknown number of resources.

## Caveats

Almost all resources in the Terraform ecosystem with import support implement a 1:1 import instead of utilizing multiple resource import.

In practice, the following pain points generally led provider developers to avoid multiple resource import:

- Risk of resource destruction for incorrect implementations ([example issue reference](https://github.com/hashicorp/terraform-provider-aws/issues/6036)).
- Pracitioners not having the required permissions or not wishing to manage the secondary imported resources ([example issue reference](https://github.com/hashicorp/terraform-provider-aws/issues/9508)).
- The implicit resource addresses for secondary imported resources requiring additional practitioner action such as `state mv` ([example issue reference](https://github.com/hashicorp/terraform-provider-aws/issues/9001)).

There were also some previous pain points associated with multiple resource import that have either been partially or wholly resolved:

- Risk of resource destruction for missing configurations ([example issue reference](https://github.com/hashicorp/terraform-provider-aws/issues/9001)).
- Terraform CLI unable to refresh secondary resource states using the correct provider instance ([example issue reference](https://github.com/hashicorp/terraform-provider-aws/issues/394)).

While the framework design should not necessarily prohibit the inclusion of this functionality, enabling it with the current Terraform CLI handling could further contribute to provider and practitioner confusion.

## Goals

Import support in this framework should be:

- Able to support diagnostics and cancellation contexts.
- Available as exported functionality for provider developers.
- Abstracted from terraform-plugin-go and convertable into those types to separate implementation concerns.
- Ergonomic to implement Go code (e.g. have helper methods for common use cases).

Additional consideration should be given to:

- Whether implementations must return a partial or full state, since Terraform CLI currently implements a full resource refresh.
- Accessing the provider instance (e.g. imports that may require additional remote information).
- Returning multiple resource states.
- Potential future `terraform import` capabilities (e.g. configuration-based import).

## Proposals

These proposals are split into sections based on separate design considerations.

### Defining Resource Import Support

#### ResourceType Interface

The framework can provide an extension interface type on the existing `ResourceType`:

```go
// ResourceTypeWithImportState represents a resource type with import support.
type ResourceTypeWithImportState interface {
    ResourceType

    ImportState(context.Context, ImportResourceStateRequest, *ImportResourceStateResponse)
}
```

When present, the resource can respond to the `ImportResourceState` RPC for the resource, otherwise it will return an error.

This satisfies the main goals of this design, however it is notably missing the ability to access the provider instance.

#### Resource Interface

The framework can provide an extension interface type on the existing `Resource`:

```go
// ResourceWithImportState represents a resource type with import support.
type ResourceWithImportState interface {
    Resource

    ImportState(context.Context, ImportResourceStateRequest, *ImportResourceStateResponse)
}
```

When present, the resource can respond to the `ImportResourceState` RPC for the resource, otherwise it will return an error.

This satisfies the main goals of this design and includes the ability to access the provider instance. There is a slight drawback that provider developers may look for this support on `ResourceType` first, so it may require some additional documentation.

#### ResourceType and Resource Interface

It is feasible to include import support using both methods described above, however it may introduce provider developer confusion about when to choose which method. The framework would need to decide what to do if both are defined. If an error is desired, a methodology for quickly unit testing resource implementations would be desirable, otherwise that type of error may only be surfaced during acceptance testing (if it performs an import) or to practitioners after a provider is built/released.

#### Requiring Implementation

It is feasible to include import support on either of the existing `Resource` or `ResourceType` interface types as a new required method. e.g.

```go
type Resource interface {
    // ... existing Create, Read, etc. ...

    ImportState(context.Context, ImportResourceStateRequest, *ImportResourceStateResponse)
}
```

Doing so would be a breaking change for providers already using early versions this framework, but longer term this would force provider developers to always consider import functionality. This may be desireable as it is a popular request in the ecosystem. To support cases where import support is not easy to implement or a desired provider design choice, the framework could then provide helper(s) to explicitly return an "import is not supported" diagnostic as the import implementation.

### Response Handling

#### Single State Only

The framework can support a 1:1 mapping of import request to expected resource state, e.g.

```go
type ImportResourceStateResponse struct {
    Diagnostics diags.Diagnostics

    // State is the imported state for the resource address in the request.
    State State
}
```

Upfront, this optimizes for the most common use case today. Additional upsides include the framework not needing carefully document the [caveats](#caveats) above or expose additional functionality around managing multiple imported resources and their states. It would, however, prevent provider developers from implementing multiple resource import support as is supported by Terraform CLI and the protocol.

#### Multiple States Only

The framework can match the underlying implementation details by requiring multiple states in the response, e.g.

```go
type ImportResourceStateResponse struct {
    Diagnostics diags.Diagnostics

    // ImportedResources is the imported state for all resources.
    ImportedResources []ImportedResource
}

type ImportedResource struct {
    // State is the imported state for this resource.
    State State

    // TypeName is the resource address associated with this resource.
    TypeName string
}
```

This leaves a lot of the implementation details up to provider developers and introduces all the [caveats](#caveats) mentioned above. For example, the implementation must include passing through the request resource address (`TypeName`) and a new `State` must be constructed from the resource type.

#### Single and Multple States

The framework could offer a hybrid of both approaches, e.g.

```go
type ImportResourceStateResponse struct {
    Diagnostics diags.Diagnostics

    // State is the imported state for the resource address in the request.
    //
    // Depending on the implementation, this would either compliment or
    // conflict with ImportedResources.
    State State

    // ImportedResources is the imported state for either all resources
    // (superceding State) or only resources outside the resource address in
    // the request.
    ImportedResources []ImportedResource
}
```

This would allow provider developers to choose between simpler or more advanced implementations.

## Recommendations

It is recommended that import support be implemented as a new required method on the existing `Resource` interface type. This will allow the framework to own the abstraction, allow access to the provider instance, and encourage implementation. Import methods should use a request and response model similar to all other framework implementations.

The `ImportResourceStateResponse` type should initially implement single state support so provider developers do not need to manually construct a new state from the resource, pass through the correct resource address, or worry about the complexity and caveats associated with multiple resource import. If there is a strong desire or Terraform CLI has an improved story around multiple resource import, the framework can extend the type to support a separate field for multiple resource states.
