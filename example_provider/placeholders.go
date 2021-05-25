package provider

import (
	"context"

	tfsdk "github.com/hashicorp/terraform-plugin-framework"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type Provider interface{}

type Resource interface{}

type CreateResourceRequest struct {
	Plan Plan
}

type Plan struct{}

func (p Plan) Get(ctx context.Context, target interface{}) error {
	return nil
}

func (p Plan) GetAttribute(ctx context.Context, attr *tftypes.AttributePath, target interface{}) error {
	return nil
}

type CreateResourceResponse struct {
	State State
}

func (c CreateResourceResponse) WithError(title string, err error) {
}

type State struct{}

func (s State) Set(ctx context.Context, val tfsdk.AttributeValue) error {
	return nil
}

func (s State) SetAttribute(ctx context.Context, attr *tftypes.AttributePath, value tfsdk.AttributeValue) error {
	return nil
}
