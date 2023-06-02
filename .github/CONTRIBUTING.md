# Contributing to terraform-plugin-framework

**First:** if you're unsure or afraid of _anything_, just ask
or submit the issue describing the problem you're aiming to solve.

Any bug fix and feature has to be considered in the context
of many thousands providers and the wider Terraform ecosystem.
This is great as your contribution can have a big positive impact,
but we have to assess potential negative impact too (e.g. breaking
existing providers which may not use a new feature).

## Table of Contents

- [I have a question](#i-have-a-question)
- [I want to report a vulnerability](#i-want-to-report-a-vulnerability)
- [New Issue](#new-issue)
- [New Pull Request](#new-pull-request)

## I have a question

> **Note:** We use GitHub for tracking bugs and feature requests related to this project.

For questions, please use [HashiCorp Discuss](https://discuss.hashicorp.com/c/terraform-providers/tf-plugin-sdk/43).

## I want to report a vulnerability

Please disclose security vulnerabilities responsibly by following the procedure
described at https://www.hashicorp.com/security#vulnerability-reporting

## New Issue

We welcome issues of all kinds including feature requests, bug reports, or documentation contributions. Below are guidelines for well-formed issues of each type.

### Bug Reports

- **Test against latest release**: Make sure you test against the latest avaiable version of both Terraform and terraform-plugin-framework. It is possible we already fixed the bug you're experiencing.
- **Search for duplicates**: It is helpful to keep bug reports consolidated to one thread, so do a quick search on existing bug reports to check if anybody else has reported the same issue. You can scope searches by the label `bug` to help narrow things down.
- **Include steps to reproduce**: Provide steps to reproduce the issue, along with code examples (both HCL and Go, where applicable) and/or real code, so we can try to reproduce it. Without this, it makes it much harder (sometimes impossible) to fix the issue.
- **Consider adding an integration test case**: A demo provider, [terraform-provider-corner](https://github.com/hashicorp/terraform-provider-corner), is available as an easy test bed for reproducing bugs in provider code. Consider opening a PR to terraform-provider-corner demonstrating the bug. Please see [Integration tests](#integration-tests) below for how to do this.

### Feature Requests

- **Search for possible duplicate requests**: It is helpful to keep requests consolidated to one thread, so do a quick search on existing requests to check if anybody else has reported the same issue. You can scope searches by the label `enhancement` to help narrow things down.
- **Include a use case description**: In addition to describing the behavior of the feature you'd like to see added, it's helpful to also lay out the reason why the feature would be important and how it would benefit the wider Terraform ecosystem. Use case in context of 1 provider is good, wider context of more providers is better.

### Documentation Contributions

- **Search for possible duplicate suggestions**: It is helpful to keep suggestions consolidated to one thread, so do a quick search on existing issues to check if anybody else has suggested the same change. You can scope searches by the label `documentation` to help narrow things down.
- **Describe the questions you're hoping the documentation will answer**: It is very helpful when writing documentation to have specific questions like "how do I implement a default value?" in mind. This helps us ensure the documentation is targeted, specific, and framed in a useful way.
- **Contribute**: This repository contains the markdown files that generate versioned documentation for [developer.hashicorp.com/terraform/plugin/framework](https://developer.hashicorp.com/terraform/plugin/framework). Please open a pull request with documentation changes. Refer to the [website README](../website/README.md) for more information.

## New Pull Request

Thank you for contributing!

- **Early validation of idea and implementation plan**: Most code changes in this project, unless trivial typo fixes or cosmetic, should be discussed and approved by maintainers in an issue before an implementation is created. This project is complicated enough that there are often several ways to implement something, each of which has different implications and tradeoffs. Working through an implementation plan with the maintainers before you dive into implementation will help ensure that your efforts may be approved and merged.
- **Unit and Integration Tests**: It may go without saying, but every new patch should be covered by tests wherever possible (see [Testing](#testing) below).
- **Go Modules**: We use [Go Modules](https://github.com/golang/go/wiki/Modules) to manage and version all our dependencies. Please make sure that you reflect dependency changes in your pull requests appropriately (e.g. `go get`, `go mod tidy` or other commands). Refer to the [dependency updates](#dependency-updates) section for more information about how this project maintains existing dependencies.
- **Changelog**: Refer to the [changelog](#changelog) section for more information about how to create changelog entries.
- **License Headers**: All source code requires a license header at the top of the file, refer to [License Headers](#license-headers) for information on how to autogenerate these headers.

### Dependency Updates

Dependency management is performed by [dependabot](https://docs.github.com/en/code-security/supply-chain-security/keeping-your-dependencies-updated-automatically/about-dependabot-version-updates). Where possible, dependency updates should occur through that system to ensure all Go module files are appropriately updated and to prevent duplicated effort of concurrent update submissions. Once available, updates are expected to be verified and merged to prevent latent technical debt.

### Changelog

HashiCorp’s open-source projects have always maintained user-friendly, readable `CHANGELOG`s that allow practitioners and developers to tell at a glance whether a release should have any effect on them, and to gauge the risk of an upgrade. This provider uses the [Changie](https://changie.dev/) automation tool for changelog automation.

#### Creating Changelog Entries

Creating a new entry for the `CHANGELOG`:

- [Install Changie](https://changie.dev/guide/installation/), if not already done
- Run `changie new` from the root directory of this project
- Choose the appropriate `kind` of change
- When prompted, type the associated GitHub issue or pull request number
- Fill out the entry body using a `package: details` format
- Repeat this process for any additional entries

The `.yaml` files created in the `.changes/unreleased` folder should be pushed the repository along with any code changes.

#### Pull Request Types to CHANGELOG

The CHANGELOG is intended to show developer-impacting changes to the codebase for a particular version. If every change or commit to the code resulted in an entry, the CHANGELOG would become less useful for developers. The lists below are general guidelines and examples for when a decision needs to be made to decide whether a change should have an entry.

##### Changes that should not have a CHANGELOG entry

- Documentation updates
- Testing updates
- Code refactoring

##### Changes that may have a CHANGELOG entry

- Dependency updates: If the update contains relevant bug fixes or enhancements that affect developers, those should be called out.

##### Changes that should have a CHANGELOG entry

- Major features
- Enhancements
- Bug fixes
- Deprecation notes
- Breaking changes

### License Headers

All source code files (excluding autogenerated files like `go.mod`, prose, and files excluded in [.copywrite.hcl](../.copywrite.hcl)) must have a license header at the top.

This can be autogenerated by running `make generate` or running `go generate ./...` in the [/tools](../tools) directory.

## Linting

GitHub Actions workflow bug and style checking is performed via [`actionlint`](https://github.com/rhysd/actionlint).

To run the GitHub Actions linters locally, install the `actionlint` tool, and run:

```shell
actionlint
```

Go code bug and style checking is performed via [`golangci-lint`](https://golangci-lint.run/).

To run the Go linters locally, install the `golangci-lint` tool, and run:

```shell
golangci-lint run ./...
```

## Testing

Code contributions should be supported by both unit and integration tests wherever possible. 

### GitHub Actions Tests

GitHub Actions workflow testing is performed via [`act`](https://github.com/nektos/act).

To run the GitHub Actions testing locally (setting appropriate event):

```shell
act --artifact-server-path /tmp --env ACTIONS_RUNTIME_TOKEN=test -P ubuntu-latest=ghcr.io/catthehacker/ubuntu:act-latest pull_request
```

The command options can be added to a `~/.actrc` file:

```text
--artifact-server-path /tmp
--env ACTIONS_RUNTIME_TOKEN=test
-P ubuntu-latest=ghcr.io/catthehacker/ubuntu:act-latest
```

So they do not need to be specified every invocation:

```shell
act pull_request
```

To test the `ci-go/terraform-provider-corner` job, a valid GitHub Personal Access Token (PAT) with public read permissions is required. It can be passed in via the `-s GITHUB_TOKEN=...` command option.

### Go Unit Tests

Go code unit testing is perfomed via Go's built-in testing functionality.

To run the Go unit testing locally:

```shell
go test ./...
```

This codebase follows Go conventions for unit testing. Some guidelines include:

- **File Naming**: Test files should be named `*_test.go` and usually reside in the same package as the code being tested.
- **Test Naming**: Test functions must include the `Test` prefix and should be named after the function or method under test. An `Example()` function test should be named `TestExample`. A `Data` type `Example()` method test should be named `TestDataExample`.
- **Concurrency**: Where possible, unit tests should be able to run concurrently and include a call to [`t.Parallel()`](https://pkg.go.dev/testing#T.Parallel). Usage of mutable shared data, such as environment variables or global variables that are used with reads and writes, is strongly discouraged.
- **Table Driven**: Where possible, unit tests should be written using the [table driven testing](https://github.com/golang/go/wiki/TableDrivenTests) style.
- **go-cmp**: Where possible, comparison testing should be done via [`go-cmp`](https://pkg.go.dev/github.com/google/go-cmp). In particular, the [`cmp.Diff()`](https://pkg.go.dev/github.com/google/go-cmp/cmp#Diff) and [`cmp.Equal()`](https://pkg.go.dev/github.com/google/go-cmp/cmp#Equal) functions are helpful.

A common template for implementing unit tests is:

```go
func TestExample(t *testing.T) {
    t.Parallel()

    testCases := map[string]struct{
        // fields to store inputs and expectations
    }{
        "test-description": {
            // fields from above
        },
    }

    for name, testCase := range testCases {
        // Do not omit this next line
        name, testCase := name, testCase

        t.Run(name, func(t *testing.T) {
            t.Parallel()

            // Implement test referencing testCase fields
        })
    }
}
```

### Integration tests

We use a special "corner case" Terraform provider for integration testing of terraform-plugin-framework, called [terraform-provider-corner](https://github.com/hashicorp/terraform-provider-corner).

Integration testing for terraform-plugin-framework involves compiling this provider against the version of the framework to be tested, and running the provider's acceptance tests. The `"provider-corner integration test"` CI job does this automatically for each PR commit and each commit to `main`. This ensures that changes to terraform-plugin-framework do not cause regressions.

#### Creating a test case in terraform-provider-corner

The terraform-provider-corner repo contains several provider servers (which are combined in order to test [terraform-plugin-mux](https://github.com/hashicorp/terraform-plugin-mux)) to test different versions of the Terraform Plugin SDK and Framework.

To add a test case for terraform-plugin-framework, add or modify resource code as appropriate in the [`frameworkprovider`](https://github.com/hashicorp/terraform-provider-corner/tree/main/internal/frameworkprovider). Then, create an acceptance test for the desired behaviour.

Creating a test case in terraform-provider-corner is a very helpful way to illustrate your bug report or feature request with easily reproducible provider code. We welcome PRs to terraform-provider-corner that demonstrate bugs and edge cases.

#### Adding integration tests to your terraform-plugin-framework PR

When fixing a bug or adding a new feature to the framework, it is helpful to create a test case in real provider code. Since the test will fail until your change is included in a terraform-plugin-framework release used by terraform-provider-corner, we recommend doing the following:

0. Fork and clone the terraform-plugin-framework and terraform-provider-corner repositories to your local machine. Identify the bug you want to fix or the feature you want to add to terraform-plugin-framework. 
1. On your local fork of terraform-provider-corner, create a failing acceptance test demonstrating this behaviour. The test should be named `TestAccFramework*`.
2. Add a `replace` directive to the `go.mod` file in your local terraform-provider-corner, pointing to your local fork of terraform-plugin-framework.
3. Make the desired code change on your local fork of terraform-plugin-framework. Don't forget unit tests as well!
4. Verify that the acceptance test now passes.
5. Make a PR to terraform-plugin-framework proposing the code change.
6. Make a PR to terraform-provider-corner adding the new acceptance test and noting that it depends on the PR to terraform-plugin-framework.

Maintainers will ensure that the acceptance test is merged into terraform-provider-corner once the terraform-plugin-framework change is merged and released.

## Maintainers Guide

This section is dedicated to the maintainers of this project.

### Releases

Run the [`release` GitHub Actions workflow](https://github.com/hashicorp/terraform-plugin-framework/actions/workflows/release.yml).
