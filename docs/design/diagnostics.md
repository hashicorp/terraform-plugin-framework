# Diagnostics

Early in the framework design, [errors were recommended over panics](./panic-error.md). This requires functions throughout the framework to appropriately handle error return values to ensure they are correctly returned back to Terraform for practitioner and developer feedback. Terraform CLI, however, allows two levels of contextualized feedback: warning and error diagnostics. Warnings generally signal an issue, but are not intended to prevent further execution, while errors generally will return early. This type of feedback is generally preferable over warning and error log entries, since Terraform itself is designed around human workflows and direct feedback, rather than a long running service or system.

There are many pieces of framework functionality that warrant enhanced practitioner feedback capabilities:

- Provider (`Configure()`), managed resource (CRUD functions), and data source (`Read()`) logic
- Framework defined provider development issues such as:
  - Invalid schema implementations
  - Invalid getting or setting of configuration, plan, or state values
- [Plan Modification](./plan-modification.md)
- [Validation](./validation.md)

This design documentation will walkthrough and recommend options for abstracted diagnostics handling and extensibility in the framework.

## Background

### Terraform CLI

Warning and error diagnostics are surfaced in the user interface as contextualized elements, typically at the bottom of command output:

```console
$ terraform plan
# ... other plan output ...
╷
│ Error: Summary
│ 
│   on example.tf line #:
│    #: source configuration line
│ 
│ Details
╵
```

This is highly visible feedback to practitioners. Diagnostic severity also influences the exit status of Terraform CLI.

### Terraform Plugin Protocol

The protocol defines the following message types for diagnostics handling:

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

message AttributePath {
    message Step {
        oneof selector {
            // Set "attribute_name" to represent looking up an attribute
            // in the current object value.
            string attribute_name = 1;
            // Set "element_key_*" to represent looking up an element in
            // an indexable collection type.
            string element_key_string = 2;
            int64 element_key_int = 3;
        }
    }
    repeated Step steps = 1;
}
```

It is found in various RPC `Response` messages, such as:

```protobuf
message ValidateProviderConfig {
    message Request {
        DynamicValue config = 1;
    }
    message Response {
        repeated Diagnostic diagnostics = 2;
    }
}
```

All RPCs except `StopProvider`, which never surfaces practitioner errors, include diagnostics support.

### terraform-plugin-go

The `terraform-plugin-go` library, which underpins this framework, provides the following implementation (both `tfprotov5` and `tfprotov6`) of diagnostics:

```go
package tfprotov6

type Diagnostic struct {
    Severity  DiagnosticSeverity
    Summary   string
    Detail    string
    Attribute *tftypes.AttributePath
}
```

These types include no helper methods or slice type alias.

## Prior Implementations

### HCL and Terraform

This context is transcribed from [this comment](https://github.com/hashicorp/terraform-plugin-framework/pull/108#discussion_r691409060).

HCL's `hcl.Diagnostics` implements the `Error` interface, but it's at the level of the group of diagnostics rather than at the level of a single diagnostic so that a group of diagnostics can return together as a single `error` value.

While this worked okay for the internals of HCL where everything was generally in agreement about how diagnostics work, it ended up causing friction once we started using diagnostics in Terraform where there's more of a blend of native diagnostics handling and traditional `error` handling. In particular, we got caught out a few times by the fact that a `hcl.Diagnostics` which contains only `hcl.DiagWarning` diagnostics appears as a non-`nil` `error`, even though it's not actually describing an error condition.

In response to those problems, we created [Terraform's own `tfdiags` package](https://pkg.go.dev/github.com/hashicorp/terraform/internal/tfdiags), which has [`tfdiags.Diagnostics`](https://pkg.go.dev/github.com/hashicorp/terraform/internal/tfdiags#Diagnostics) as a more general analog to `hcl.Diagnostics`. Part of that design was to intentionally make `tfdiags.Diagnostics` _not_ implement `error`, and instead it has a method [`Err`](https://pkg.go.dev/github.com/hashicorp/terraform/internal/tfdiags#Diagnostics.Err) which returns an `error` which is `nil` unless `diags.HasErrors()`, thus preserving the usual meaning of a non-`nil` error at the expense of then losing track of warnings that aren't accompanied by an error (because there's nowhere to put them in an `error` value).

For some particularly gnarly cases we also have [`ErrWithWarnings`](https://pkg.go.dev/github.com/hashicorp/terraform/internal/tfdiags#Diagnostics.ErrWithWarnings) which essentially recovers the HCL approach of potentially returning a weird `error` that might actually only be reporting warnings. In the few cases where we use that the caller needs to be careful to check whether the `error` value has the type [`NonFatalError`](https://pkg.go.dev/github.com/hashicorp/terraform/internal/tfdiags#NonFatalError) and treat it as warnings only in that case.

Tangentially, we also made [the `Append` method of `Diagnostics`](https://pkg.go.dev/github.com/hashicorp/terraform/internal/tfdiags#Diagnostics) able to accept naked `error` values and turn them into (often lower-quality) diagnostic errors. It also recognizes `error` values previously returned from `Diagnostics.Err` or `Diagnostics.NonFatalErr` and recovers the original diagnostics from them, which allows losslessly(-ish) sending diagnostics through a function that returns `error` because e.g. it needs to interact with other typical Go error-handling patterns.

### terraform-plugin-sdk

The previous framework provided a thin abstraction layer in the [`diag` package](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/diag):

```go
type Diagnostic struct {
    Severity      Severity
    Summary       string
    Detail        string
    AttributePath cty.Path
}

func (d Diagnostic) Validate() error

type Diagnostics []Diagnostic

func (diags Diagnostics) HasError() bool

func Errorf(format string, a ...interface{}) Diagnostics

func FromErr(err error) Diagnostics
```

## Goals

Diagnostics in this framework should be:

- Able to natively support warning and error diagnostics.
- Implementable by provider developers.
- Abstracted from terraform-plugin-go and convertable into those types to separate implementation concerns.
- Ergonomic to create, read, and delete in Go code (e.g. have helper methods for common use cases).

Additional consideration should be given to:

- Extensibility and additional data storage (e.g. distinguishing between types of diagnostics or what generated an error).
- The ability to deduplicate or manipulate diagnostics.

## Proposals

### Thin Abstraction

The most trivial implementation would be creating new Go types in this framework that provide a thin abstraction layer, e.g.

```go
package diag

type Diagnostic struct {
    Severity  DiagnosticSeverity
    Summary   string
    Detail    string
    Attribute *tftypes.AttributePath // See also: https://github.com/hashicorp/terraform-plugin-framework/issues/81
}

type Diagnostics []Diagnostic

func (d Diagnostics) Append(in ...Diagnostic)
func (d Diagnostics) HasError() bool
func (d Diagnostics) toTfprotov6Diagnostics() []*tfprotov6.Diagnostic
```

This solves many requirements, allowing the framework to own the abstraction and provide ergonomic methods. However, this implementation is tightly coupled to the underlying implementation and does not introduce protections against backwards compatibility issues. Since the data fields are strongly associated with the type, implementors inside and outside the framework are bound to the initial implementation details. If those details are not exported and instead hidden behind a constructor function, e.g.

```go
type Diagnostic struct {
    severity  DiagnosticSeverity
    summary   string
    detail    string
    attribute *tftypes.AttributePath // See also: https://github.com/hashicorp/terraform-plugin-framework/issues/81
}

func NewDiagnostic(severity DiagnosticSeverity, summary string, detail string, attribute *tftypes.AttributePath) Diagnostic
func NewErrorDiagnostic(summary string, detail string, attribute *tftypes.AttributePath) Diagnostic
func NewWarningDiagnostic(summary string, detail string, attribute *tftypes.AttributePath) Diagnostic
```

This type of implementation will still be difficult to extend or change.

As another extensibility example, as `Diagnostics` pass through various functions across subsystems, determining the "source" or "cause" of a diagnostic must be done by `Summary` and/or `Detail` value similar to checking Go `error` types by value using `(error).Is()`. Extending the `Diagnostic` type is possible to add other identifying fields or creating new types, e.g.

```go
type ValidationDiagnostic struct {
  Diagnostic
}
```

However, the implementation for diagnostics then differs from other design choices in the framework where interfaces are preferred. For example, it becomes difficult to implement equality and future enhancements for diagnostics since the interface is not strongly defined for extensions.

### Basic Interface

Diagnostics could be modeled around interfaces:

```go
package diag

type Diagnostic interface {
  Severity() Severity
  Summary()  string
  Detail()   string

  Equal(other Diagnostic) bool
}

type DiagnosticWithLogger interface {
  Diagnostic

  Log(ctx context.Context)
}

// This could be folded into Diagnostic
type DiagnosticWithPath interface {
  Diagnostic

  Path() *tftypes.AttributePath // See also: https://github.com/hashicorp/terraform-plugin-framework/issues/81
}

type Diagnostics []Diagnostic

func (d Diagnostics) Append(in ...Diagnostic)
func (d Diagnostics) HasError() bool
func (d Diagnostics) toTfprotov6Diagnostics() []*tfprotov6.Diagnostic
```

With an example implementation:

```go
// ErrorDiagnostic is a generic diagnostic with error severity.
type ErrorDiagnostic struct {
  Summary string
  Detail  string
}

func (d ErrorDiagnostic) Severity() Severity { return SeverityError }
func (d ErrorDiagnostic) Summary() string { return d.Summary }
func (d ErrorDiagnostic) Detail() string { return d.Detail }
func (d ErrorDiagnostic) Equal(other Diagnostic) bool {
  other, ok := other.(ErrorDiagnostic)

  if !ok {
    return false
  }

  return other.Summary() == d.Summary() && other.Detail() == d.Detail()
}

// ValidationErrorDiagnostic represents an error during validation.
type ValidationErrorDiagnostic struct {
  ErrorDiagnostic
}

func (d ValidationErrorDiagnostic) Equal(other Diagnostic) bool {
  other, ok := other.(ValidationErrorDiagnostic)

  if !ok {
    return false
  }

  return other.Summary() == d.Summary() && other.Detail() == d.Detail()
}
```

It then becomes possible to distinguish between diagnostic types (in this case the "source" of a diagnostic) while still supporting additional functionality defined by the interface. These extended diagnostics can also implement other framework-defined capabilities, such as implementing the example `Log()` method in this situation.

### Error Interface

On top of modeling as an interface, diagnostics could be further marked as an extension of the `error` interface:

```go
package diag

type Diagnostic interface {
  error

  DiagnosticSeverity() Severity
  DiagnosticSummary()  string
  DiagnosticDetail()   string
}

// This could be folded into Diagnostic
type DiagnosticWithPath interface {
  Diagnostic

  DiagnosticPath() *tftypes.AttributePath // See also: https://github.com/hashicorp/terraform-plugin-framework/issues/81
}
```

This would allow a single diagnostic to be usable in place of where an `error` would be returned and the type would pick up associated `errors` package capabilities. All implementations would be required for including an `Error()` method and optionally other `Unwrap()` etc. methods to satisfy other `errors` capabilities, which could introduce verbosity concerns. It also might be unclear what the semantics of a warning diagnostic are in terms of the `error` interface -- this implementation would treat all warning diagnostics as errors, which seems less than ideal.

## Recommendations

It is recommended that diagnostics be implemented as its own interface types without extending the `error` interface. This will allow the framework to own the abstraction wholly without potentially introducing semantic differences between warnings and errors in certain situations. The framework can initially provide generic error and warning diagnostic implementations as concrete types as well as a slice type alias with associated methods to ease common use case operations.
