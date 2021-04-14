# Contributing to terraform-plugin-framework

**First:** if you're unsure or afraid of _anything_, just ask
or submit the issue describing the problem you're aiming to solve.

Any bug fix and feature has to be considered in the context
of many (100+) providers and the wider Terraform ecosystem.
This is great as your contribution can have a big positive impact,
but we have to assess potential negative impact too (e.g. breaking
existing providers which may not use a new feature).

## Table of Contents

 - [I just have a question](#i-just-have-a-question)
 - [I want to report a vulnerability](#i-want-to-report-a-vulnerability)
 - [New Issue](#new-issue)
 - [New Pull Request](#new-pull-request)

## I just have a question

> **Note:** We use GitHub for tracking bugs and feature requests related to Plugin SDK.

For questions, please see relevant channels at https://www.terraform.io/community.html

## I want to report a vulnerability

Please disclose security vulnerabilities responsibly by following the procedure
described at https://www.hashicorp.com/security#vulnerability-reporting

## New Issue

We welcome issues of all kinds including feature requests, bug reports or documentation suggestions. Below are guidelines for well-formed issues of each type.

### Bug Reports

 - [ ] **Test against latest release**: Make sure you test against the latest avaiable version of both Terraform and terraform-plugin-framework. It is possible we already fixed the bug you're experiencing.

 - [ ] **Search for duplicates**: It's helpful to keep bug reports consolidated to one thread, so do a quick search on existing bug reports to check if anybody else has reported the same thing. You can scope searches by the label `bug` to help narrow things down.

 - [ ] **Include steps to reproduce**: Provide steps to reproduce the issue, along with code examples (both HCL and Go, where applicable) and/or real code, so we can try to reproduce it. Without this, it makes it much harder (sometimes impossible) to fix the issue.

### Feature Requests

 - [ ] **Search for possible duplicate requests**: It's helpful to keep requests consolidated to one thread, so do a quick search on existing requests to check if anybody else has reported the same thing. You can scope searches by the label `enhancement` to help narrow things down.

 - [ ] **Include a use case description**: In addition to describing the behavior of the feature you'd like to see added, it's helpful to also lay out the reason why the feature would be important and how it would benefit the wider Terraform ecosystem. Use case in context of 1 provider is good, wider context of more providers is better.

## New Pull Request

Thank you for contributing!

We are happy to review pull requests without associated issues,
but we highly recommend starting by describing and discussing
your problem or feature and attaching use cases to an issue first
before raising a pull request.

- [ ] **Early validation of idea and implementation plan**: terraform-plugin-framework is complicated enough that there are often several ways to implement something, each of which has different implications and tradeoffs. Working through a plan of attack with the team before you dive into implementation will help ensure that you're working in the right direction.

- [ ] **Unit Tests**: It may go without saying, but every new patch should be covered by tests wherever possible.

- [ ] **Go Modules**: We use [Go Modules](https://github.com/golang/go/wiki/Modules) to manage and version all our dependencies. Please make sure that you reflect dependency changes in your pull requests appropriately (e.g. `go get`, `go mod tidy` or other commands). Where possible it is better to raise a separate pull request with just dependency changes as it's easier to review such PR(s).

### Cosmetic changes, code formatting, and typos

In general we do not accept PRs containing only the following changes:

 - Correcting spelling or typos
 - Code formatting, including whitespace
 - Other cosmetic changes that do not affect functionality
 
While we appreciate the effort that goes into preparing PRs, there is always a tradeoff between benefit and cost. The costs involved in accepting such contributions include the time taken for thorough review, the noise created in the git history, and the increased number of GitHub notifications that maintainers must attend to.

#### Exceptions

We belive that one should "leave the campsite cleaner than you found it", so you are welcome to clean up cosmetic issues in the neighbourhood when submitting a patch that makes functional changes or fixes.
