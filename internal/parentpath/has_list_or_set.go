// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package parentpath

import "github.com/hashicorp/terraform-plugin-framework/path"

// HasListOrSet returns true if any step of given path is a list or set. This
// cannot detect if the last step is a list or set.
//
// This functionality could also theoretically be a method on the path.Path
// type, e.g. ParentHasListOrSet(), rather than a separate parentpath function.
func HasListOrSet(p path.Path) bool {
	for _, pathStep := range p.Steps() {
		switch pathStep.(type) {
		case path.PathStepElementKeyInt, path.PathStepElementKeyValue:
			// This type of step is after a list or set attribute
			return true
		}
	}

	return false
}
