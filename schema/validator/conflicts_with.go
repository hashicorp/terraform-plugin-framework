// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package validator

import "github.com/hashicorp/terraform-plugin-framework/path"

// ConflictsWithValidator exposes grouped paths where only one configured value
// should be preserved during generated configuration handling.
type ConflictsWithValidator interface {
	ConflictsWithPaths() path.Expressions
}
