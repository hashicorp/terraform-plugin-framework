// Copyright IBM Corp. 2026
// SPDX-License-Identifier: MPL-2.0

package validator

import "github.com/hashicorp/terraform-plugin-framework/path"

// AlsoRequiresValidator exposes grouped paths where configured values should
// be removed from generated configuration if required peer values are not set.
type AlsoRequiresValidator interface {
	AlsoRequiresPaths() path.Expressions
}
