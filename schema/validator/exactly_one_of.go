// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package validator

import "github.com/hashicorp/terraform-plugin-framework/path"

// ExactlyOneOfValidator exposes grouped paths where generated configuration
// should preserve exactly one value, optionally applying a default when all are
// null.
type ExactlyOneOfValidator interface {
	ExactlyOneOfPaths() path.Expressions
}
