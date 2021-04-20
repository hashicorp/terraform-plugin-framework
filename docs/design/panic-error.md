# Panics and Errors

When surfacing exceptions from the Terraform Plugin SDK and its related
projects, we have two tools available to us: panicking, or returning an error.
Logging is an imprecise and unreliable communication method, and most provider
developers will miss it. Static analysis is great and can be used to surface
code that will lead to exceptions, but we can't rely on provider developers
using these tools.

We don't take a hard stance on whether to panic or error, instead asserting
that there are appropriate times to use each method.

This document is meant to explore the trade-offs inherent in returning an error
or panicking, to help navigate those choices when they arise.

## Usability of Function Return Values

A drawback of errors is that they can make using the return values of a
function more difficult if the function returns something besides an error. For
example, [`tftypes.NewValue`][tftypes-newvalue] wants to return a
`tftypes.Value`, but it can encounter exceptions while doing so. There are two
function signatures it can use:

```go
func NewValue(t Type, v interface{}) Value
```

or

```go
func NewValue(t Type, v interface{}) (Value, error)
```

Consider trying to create a Map value. If we use the first method, we can do
the following:

```go
val := tftypes.NewValue(tftypes.Map{ElementType: String}, map[string]tftypes.Value{
	"hello": tftypes.NewValue(tftypes.String, "world"),
	"red": tftypes.NewValue(tftypes.String, "blue"),
})
```

In this example, we're able to create the elements of the Map inline, using the
result of `NewValue` directly, because no error is being returned. Let's look
at the same code, but written with the second function signature, where we
return an error instead of panicking:

```go
helloVal, err := tftypes.NewValue(tftypes.String, "world")
if err != nil {
	// handle error
}
redVal, err := tftypes.NewValue(tftypes.String, "blue")
if err != nil {
	// handle error
}
val, err := tftypes.NewValue(tftypes.Map{ElementType: String}, map[string]tftypes.Value{
	"hello": helloVal,
	"red": redVal,
})
if err != nil {
	// handle error
}
```

We can see in this version that for sufficiently large or complicated values,
understanding what is happening requires much more cognitive overhead when
errors are returned. This doesn't necessarily mean that panicking is right in
this situation; it only means that returning errors comes with a cost that must
be weighed against their benefits and the costs of panicking.

## Reliability

Panics are meant to be rare, exceptional occurrences that force an immediate,
and potentially unexpected halt to all the goroutines running in the process.
Libraries panicking are especially hazardous, because the context the panics
occur in can be more expensive than the library anticipated. A library that
panics too frequently or during regular operation is considered less reliable
than a library that panics rarely and only in exceptional cases. When choosing
between panics and errors for a specific exception, we should keep in mind how
frequently that exception is expected to be encountered.

Panics are so expensive because they are so difficult for the caller to
effectively handle. While callers can use `recover` to handle panics, this is
non-obvious enough and far enough from standard Go practice that it doesn't
really help that much. Errors, on the other hand, have a rich and robust
ecosystem of packages, patterns, and best practices for detecting and handling
them. Using errors allows the caller to automatically fix or work around the
issue: retrying transient errors, creating files that don't exist, or resolving
other exceptions that occurred because of the environment the code was run in.
These automated fixes or workarounds are essential to the resiliency and
reliability of programs, and panics undercut the ability to take advantage of
them.

## Conspicuousness

When returning an error as the only return value of a function, Go will not
complain if the caller ignores the returned error:

```go
func myFunc() error {
	return errors.New("this is why we can't have nice things")
}

func caller() {
	// note we're not checking or handling the error here
	// and the Go compiler doesn't tell us about the unused return value
	myFunc()
}
```

This is most famously seen with the
[helper/schema.ResourceData.Set][resourcedata-set] function in v2 of the
Terraform Plugin SDK. This function sets a takes a path and a value, and then
sets that value in state at that path. It returns an error if the path is
invalid, if the value is an invalid type, or if the type of the value doesn't
match the type expected at that path. This function is ubiquitous in provider
development. Unfortunately, possibly because it's so ubiquitous, provider
developers often forget to check the error value returned, leading to the
exception not being surfaced or addressed, and unexpected behavior from the
provider.

While it's easy to assert that provider developers should just check the error,
the API should be designed such that it's hard to use incorrectly, and
incorrect usage should be loud and yield noticeable feedback.

Panics are more conspicuous than errors, as they cannot be accidentally
ignored. They default to being noticeable.

## Practitioner Experience

Panic output tends to be an overwhelming stacktrace spanning multiple different
packages, modules, and abstraction layers. It surfaces a lot of information,
with the most relevant message at the top.

Practitioners are likely to encounter these panics in logs or in their
terminals, and finding the most relevant message takes some practice and some
familiarity with what you're looking at, which practitioners should not be
expected to have.

Practitioners being exposed to panics is therefore a confusing and alarming
situation, and one that should be avoided.

## Recontextualising

One of the benefits of errors is that they can have context added by the caller
before being surfaced. For example, in the test framework shipped with v2 of
the Terraform Plugin SDK, errors are "decorated" with the test step that they
occurred during, surfacing more information to the developer about where the
error came from. The test framework can do this because errors are just values.
Panics, on the other hand, are much harder to decorate in this fashion, and
tend to only convey the information available at the level of abstraction that
prompted them. This suggests that panics lower in the abstraction should be
considered more expensive, as they convey less context to the developer.

## Nature of the Exception

There seem to be two broad categories of exceptions that a provider can run
into: logic exceptions and environmental exceptions. Either the exception
occurred because the code was faulty, or the exception occurred because the
environment the code was deployed to did not meet expectations.

Logic exceptions are errors the programmer has made while writing the provider.
Type assertions on an interface, accessing indexes that don't exist on a slice,
or otherwise performing an operation that cannot be performed on the data in
question but is not prohibited by Go's type system are all examples of this.
The standard library panics in all these cases, as there's no meaningful error
handling to be done. The programmer made a mistake, and the code needs to be
corrected.

Environmental exceptions occur when code that has sound logic is deployed into
an environment that does not meet its needs. Files not existing, networking
errors, and other various error conditions are all common exceptions that can
occur in a specific environment. What groups these exceptions together is that
the solution isn't to fix the _code_, but to fix the environment the code is
running in; create the file, resolve the networking issues, etc. The standard
library tends to use errors in these cases, as there is meaningful error
handling that can be done, often in an automated fashion.

When determining whether to use a panic or error, interrogate whether the
exception being surfaced requires the code to be fixed, in which case a panic
may make sense, or if it can be surfaced by environmental factors like the data
the practitioner passes in, the state the API is in, the computer the code is
run on, or other environmental factors.

## Recommendations

Our recommended rule of thumb is to use an error unless an exception:

* is likely to be surfaced during even cursory testing. This usually means
  type-based exceptions, as exceptions that can occur only in the presence of
  certain values require testing those specific values.
* benefits substantially from a more concise API.
* is unequivocally a logic error, with no reasonable handling strategy
  available to callers.
* requires programmer attention and justifies the conspicuous nature of panics

When an function or method uses a panic, it is highly recommended that a
version that returns an error is also created. For example,
[`tftypes.NewValue`][tftypes-newvalue] has a twin
[`tftypes.ValidateValue`][tftypes-validatevalue]:

```go
// returns the Value, panics in the face of exceptions
func NewValue(t Type, v interface{}) Value 

// returns an error if NewValue would panic
func ValidateValue(t Type, v interface{}) error
```

`NewValue` decided to use panic given the API benefits it conveyed, based on
the knowledge that panics _only_ happen when the type of `v` is incompatible
with the `Type` specified. In the vast majority of circumstances, `v` will be
of a type known at compile time, and therefore panics will show up if the code
is executed even once in testing. But there are edge cases in which an API may
return a value with an interface type, in which case, the type of `v` may not
be known at compile time. In those edge cases, providers should call
`tftypes.ValidateValue` before calling `tftypes.NewValue`, secure in the
knowledge that if no error was returned from `tftypes.ValidateValue` they are
safe from panics.

## Prior Art

* [Don't Panic! Handling Errors and Bugs in Go](https://apparently.me.uk/go-api-panic-or-error/) by [@apparentlymart][apparentlymart]
* [Effective Go](https://golang.org/doc/effective_go) ([panic](https://golang.org/doc/effective_go#panic) and [error](https://golang.org/doc/effective_go#errors))

[tftypes-newvalue]: https://pkg.go.dev/github.com/hashicorp/terraform-plugin-go/tftypes#NewValue
[tftypes-validatevalue]: https://pkg.go.dev/github.com/hashicorp/terraform-plugin-go/tftypes#ValidateValue
[resourcedata-set]: https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema#ResourceData.Set
[apparentlymart]: https://github.com/apparentlymart
