package providerserver

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/internal/testing/emptyprovider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestNewProtocol6(t *testing.T) {
	provider := &emptyprovider.Provider{}

	providerServerFunc := NewProtocol6(provider)
	providerServer := providerServerFunc()

	// Simple verification
	_, err := providerServer.GetProviderSchema(context.Background(), &tfprotov6.GetProviderSchemaRequest{})

	if err != nil {
		t.Fatalf("unexpected error calling ProviderServer: %s", err)
	}
}

func TestNewProtocol6WithError(t *testing.T) {
	provider := &emptyprovider.Provider{}

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
