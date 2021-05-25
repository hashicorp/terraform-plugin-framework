package provider

import (
	"context"
	"testing"
)

func TestResourceComputeInstance(t *testing.T) {
	var r resourceComputeInstance

	req := CreateResourceRequest{
		Plan: Plan{},
	}
	resp := CreateResourceResponse{
		State: State{},
	}
	r.Create(context.Background(), req, resp)
}
