package tf

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	// "github.com/hashicorp/terraform-plugin-go/tfprotov6/server"
)

// make a tfprotov6.ProviderServer
// and ResourceServer

type ResourceServer map[string]tfprotov6.ResourceServer

func (s *ResourceServer) ApplyResourceChange() {

}
