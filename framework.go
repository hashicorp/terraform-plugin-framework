package framework

import "github.com/hashicorp/terraform-plugin-go/tfprotov5/tftypes"

// this is just a placeholder file to make CircleCI happy with an empty
// repository.
//
// We created a package so Go vet will be happy.
//
// We're using tftypes so we have an import, so our cache behavior will stop
// breaking on the lack of a go.sum file.
var _ = tftypes.NewValue(tftypes.String, "delete me")
