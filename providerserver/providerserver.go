package providerserver

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwserver"
	"github.com/hashicorp/terraform-plugin-framework/internal/proto6server"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
)

// NewProtocol6 returns a protocol version 6 ProviderServer implementation
// based on the given Provider and suitable for usage with the
// github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server.Serve()
// function and various terraform-plugin-mux functions.
func NewProtocol6(p tfsdk.Provider) func() tfprotov6.ProviderServer {
	return func() tfprotov6.ProviderServer {
		return &proto6server.Server{
			FrameworkServer: fwserver.Server{
				Provider: p,
			},
			Provider: p,
		}
	}
}

// NewProtocol6WithError returns a protocol version 6 ProviderServer
// implementation based on the given Provider and suitable for usage with
// github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource.TestCase.ProtoV6ProviderFactories.
//
// The error return is not currently used, but it may be in the future.
func NewProtocol6WithError(p tfsdk.Provider) func() (tfprotov6.ProviderServer, error) {
	return func() (tfprotov6.ProviderServer, error) {
		return &proto6server.Server{
			FrameworkServer: fwserver.Server{
				Provider: p,
			},
			Provider: p,
		}, nil
	}
}

// Serve serves a provider, blocking until the context is canceled.
func Serve(ctx context.Context, providerFunc func() tfsdk.Provider, opts ServeOpts) error {
	err := opts.validate(ctx)

	if err != nil {
		return fmt.Errorf("unable to validate ServeOpts: %w", err)
	}

	var tf6serverOpts []tf6server.ServeOpt

	if opts.Debug {
		tf6serverOpts = append(tf6serverOpts, tf6server.WithManagedDebug())
	}

	return tf6server.Serve(
		opts.Address,
		func() tfprotov6.ProviderServer {
			provider := providerFunc()

			return &proto6server.Server{
				FrameworkServer: fwserver.Server{
					Provider: provider,
				},
				Provider: provider,
			}
		},
		tf6serverOpts...,
	)
}
