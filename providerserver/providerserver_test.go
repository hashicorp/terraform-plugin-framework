// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package providerserver

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testprovider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestNewProtocol5(t *testing.T) {
	t.Parallel()

	provider := &testprovider.Provider{}

	providerServerFunc := NewProtocol5(provider)
	providerServer := providerServerFunc()

	// Simple verification
	_, err := providerServer.GetProviderSchema(context.Background(), &tfprotov5.GetProviderSchemaRequest{})

	if err != nil {
		t.Fatalf("unexpected error calling ProviderServer: %s", err)
	}
}

func TestNewProtocol5WithError(t *testing.T) {
	t.Parallel()

	provider := &testprovider.Provider{}

	providerServer, err := NewProtocol5WithError(provider)()

	if err != nil {
		t.Fatalf("unexpected error creating ProviderServer: %s", err)
	}

	// Simple verification
	_, err = providerServer.GetProviderSchema(context.Background(), &tfprotov5.GetProviderSchemaRequest{})

	if err != nil {
		t.Fatalf("unexpected error calling ProviderServer: %s", err)
	}
}

func TestNewProtocol6(t *testing.T) {
	t.Parallel()

	provider := &testprovider.Provider{}

	providerServerFunc := NewProtocol6(provider)
	providerServer := providerServerFunc()

	// Simple verification
	_, err := providerServer.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})

	if err != nil {
		t.Fatalf("unexpected error calling ProviderServer: %s", err)
	}
}

func TestNewProtocol6WithError(t *testing.T) {
	t.Parallel()

	provider := &testprovider.Provider{}

	providerServer, err := NewProtocol6WithError(provider)()

	if err != nil {
		t.Fatalf("unexpected error creating ProviderServer: %s", err)
	}

	// Simple verification
	_, err = providerServer.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})

	if err != nil {
		t.Fatalf("unexpected error calling ProviderServer: %s", err)
	}
}
