---
name: 🐛 Bug report
about: Let us know about an unexpected error, a crash, or an incorrect behavior.
labels: bug
---

### Module version
<!---
Inspect your go.mod as below to find the version, and paste the result between the ``` marks below.

go list -m github.com/hashicorp/terraform-plugin-framework/...

If you are not running the latest version of the framework, please try upgrading
because your bug may have already been fixed.
-->

```
...
```

### Relevant provider source code

<!--
Paste any Go code that you believe to be relevant to the bug
e.g. schema or implementation of CRUD for a given resource or data source
-->
```go
...
```

### Terraform Configuration Files
<!--
Paste the relevant parts of your Terraform configuration between the ``` marks below.

For large Terraform configs, please use a service like Dropbox and share a link to the ZIP file. For security, you can also encrypt the files using our GPG public key.
-->

```hcl
...
```

### Debug Output
<!--
Full debug output can be obtained by running Terraform with the environment variable `TF_LOG=trace`. Please create a GitHub Gist containing the debug output. Please do _not_ paste the debug output in the issue, since debug output is long.

Debug output may contain sensitive information. Please review it before posting publicly, and if you are concerned feel free to encrypt the files using the HashiCorp security public key.
-->


### Expected Behavior
<!--
What should have happened?
-->

### Actual Behavior
<!--
What actually happened?
-->

### Steps to Reproduce
<!--
Please list the full steps required to reproduce the issue, for example:
1. `terraform init`
2. `terraform apply`
-->

### References
<!--
Are there any other GitHub issues (open or closed) or Pull Requests that should be linked here? For example:

- #6017
-->
