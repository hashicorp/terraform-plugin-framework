package action

import (
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// TODO: create an interface for LinkedResource in internal package

type LinkedResources map[string]LinkedResource

type LinkedResource struct {
	ResourceTypeName string

	AttributePath path.Path
}

type LinkedAttributes struct {
	AttributePath path.Path
}
