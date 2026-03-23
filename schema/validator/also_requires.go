// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package validator

import "github.com/hashicorp/terraform-plugin-framework/path"

type AlsoRequiresValidator interface {
	// AlsoRequiresPaths returns attribute paths that this validator applies to.
	AlsoRequiresPaths() path.Expressions
}
