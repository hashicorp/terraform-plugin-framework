package fwserver

import (
	"context"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// Server implements the framework provider server. Protocol specific
// implementations wrap this handling along with calling all request and
// response type conversions.
type Server struct {
	Provider tfsdk.Provider

	// providerSchema is the cached Provider Schema for RPCs that need to
	// convert configuration data from the protocol. If not found, it will be
	// fetched from the Provider.GetSchema() method.
	providerSchema *tfsdk.Schema

	// providerSchemaDiags is the cached Diagnostics obtained while populating
	// providerSchema. This is to ensure any warnings or errors are also
	// returned appropriately when fetching providerSchema.
	providerSchemaDiags diag.Diagnostics

	// providerSchemaMutex is a mutex to protect concurrent providerSchema
	// access from race conditions.
	providerSchemaMutex sync.Mutex
}

// ProviderSchema returns the Schema associated with the Provider. The Schema
// and Diagnostics are cached on first use.
func (s *Server) ProviderSchema(ctx context.Context) (*tfsdk.Schema, diag.Diagnostics) {
	logging.FrameworkTrace(ctx, "Checking ProviderSchema lock")
	s.providerSchemaMutex.Lock()

	if s.providerSchema != nil {
		return s.providerSchema, nil
	}

	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetSchema")
	providerSchema, diags := s.Provider.GetSchema(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined Provider GetSchema")

	s.providerSchema = &providerSchema
	s.providerSchemaDiags = diags

	s.providerSchemaMutex.Unlock()

	return s.providerSchema, s.providerSchemaDiags
}
