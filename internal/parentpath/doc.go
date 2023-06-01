// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package parentpath contains path functionality intended for previous steps
// of a given Path or Expression.
//
// This functionality is not included in the exported path package because
// its external utility is unknown and being unexported means the functionality
// can be modified without violating compatibility promises. If provider
// developers are interested in any of this functionality, relevant parts can
// be migrated to the path package or by creating a path/parentpath package.
package parentpath
