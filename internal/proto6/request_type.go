package proto6

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// IsCreate returns true if the request is creating a resource.
func IsCreate(ctx context.Context, req *tfprotov6.ApplyResourceChangeRequest, typ tftypes.Type) (bool, error) {
	priorState, err := req.PriorState.Unmarshal(typ)
	if err != nil {
		return false, err
	}
	// if our prior state isn't null, the state already exists, this can't
	// be a create request
	if !priorState.IsNull() {
		return false, nil
	}

	// otherwise, it's a create request
	return true, nil
}

// IsUpdate returns true if the request is updating a resource.
func IsUpdate(ctx context.Context, req *tfprotov6.ApplyResourceChangeRequest, typ tftypes.Type) (bool, error) {
	priorState, err := req.PriorState.Unmarshal(typ)
	if err != nil {
		return false, err
	}
	// if our prior state is null, the state doesn't exist, so this can't be
	// an update request
	if priorState.IsNull() {
		return false, nil
	}

	plannedState, err := req.PlannedState.Unmarshal(typ)
	if err != nil {
		return false, err
	}
	// if our planned state is null, this is a delete request, and it can't be
	// an update too
	if plannedState.IsNull() {
		return false, nil
	}

	// otherwise, this is an update
	return true, nil
}

// IsDestroy returns true if the request is deleting a resource.
func IsDestroy(ctx context.Context, req *tfprotov6.ApplyResourceChangeRequest, typ tftypes.Type) (bool, error) {
	plannedState, err := req.PlannedState.Unmarshal(typ)
	if err != nil {
		return false, err
	}
	// if our planned state isn't null, this can't be a delete request
	if !plannedState.IsNull() {
		return false, nil
	}

	// otherwise, this is a delete request
	return true, nil
}
