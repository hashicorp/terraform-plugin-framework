package fwserver

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// Server implements the framework provider server. Protocol specific
// implementations wrap this handling along with calling all request and
// response type conversions.
type Server struct {
	Provider tfsdk.Provider
}
