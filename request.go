package tf

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

type ConfigureProviderRequest struct {
	Config *tfprotov6.DynamicValue
}

type CreateResourceRequest struct {
	Config       *tfprotov6.DynamicValue
	PriorState   State
	PlannedState State
}

type ReadResourceRequest struct {
	Config       *tfprotov6.DynamicValue
	CurrentState State
}

type DeleteResourceRequest struct {
	Config *tfprotov6.DynamicValue
}

type UpdateResourceRequest struct {
	Config       *tfprotov6.DynamicValue
	PriorState   State
	PlannedState State
}
