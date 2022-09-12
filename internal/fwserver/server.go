package fwserver

import (
	"context"
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Server implements the framework provider server. Protocol specific
// implementations wrap this handling along with calling all request and
// response type conversions.
type Server struct {
	Provider provider.Provider

	// DataSourceConfigureData is the
	// [provider.ConfigureResponse.DataSourceData] field value which is passed
	// to [datasource.ConfigureRequest.ProviderData].
	DataSourceConfigureData any

	// ResourceConfigureData is the
	// [provider.ConfigureResponse.ResourceData] field value which is passed
	// to [resource.ConfigureRequest.ProviderData].
	ResourceConfigureData any

	// dataSourceSchemas is the cached DataSource Schemas for RPCs that need to
	// convert configuration data from the protocol. If not found, it will be
	// fetched from the DataSourceType.GetSchema() method.
	dataSourceSchemas map[string]fwschema.Schema

	// dataSourceSchemasDiags is the cached Diagnostics obtained while populating
	// dataSourceSchemas. This is to ensure any warnings or errors are also
	// returned appropriately when fetching dataSourceSchemas.
	dataSourceSchemasDiags diag.Diagnostics

	// dataSourceSchemasMutex is a mutex to protect concurrent dataSourceSchemas
	// access from race conditions.
	dataSourceSchemasMutex sync.Mutex

	// dataSourceFuncs is the cached DataSource functions for RPCs that need to
	// access data sources. If not found, it will be fetched from the
	// Provider.DataSources() method.
	dataSourceFuncs map[string]func() datasource.DataSource

	// dataSourceTypes is the cached DataSourceTypes for RPCs that need to
	// access data sources. If not found, it will be fetched from the
	// Provider.GetDataSources() method.
	//nolint:staticcheck // Internal implementation
	dataSourceTypes map[string]provider.DataSourceType

	// dataSourceTypesDiags is the cached Diagnostics obtained while populating
	// dataSourceTypes. This is to ensure any warnings or errors are also
	// returned appropriately when fetching dataSourceTypes.
	dataSourceTypesDiags diag.Diagnostics

	// dataSourceTypesMutex is a mutex to protect concurrent dataSourceTypes
	// access from race conditions.
	dataSourceTypesMutex sync.Mutex

	// providerSchema is the cached Provider Schema for RPCs that need to
	// convert configuration data from the protocol. If not found, it will be
	// fetched from the Provider.GetSchema() method.
	providerSchema fwschema.Schema

	// providerSchemaDiags is the cached Diagnostics obtained while populating
	// providerSchema. This is to ensure any warnings or errors are also
	// returned appropriately when fetching providerSchema.
	providerSchemaDiags diag.Diagnostics

	// providerSchemaMutex is a mutex to protect concurrent providerSchema
	// access from race conditions.
	providerSchemaMutex sync.Mutex

	// providerMetaSchema is the cached Provider Meta Schema for RPCs that need
	// to convert configuration data from the protocol. If not found, it will
	// be fetched from the Provider.GetMetaSchema() method.
	providerMetaSchema fwschema.Schema

	// providerMetaSchemaDiags is the cached Diagnostics obtained while populating
	// providerMetaSchema. This is to ensure any warnings or errors are also
	// returned appropriately when fetching providerMetaSchema.
	providerMetaSchemaDiags diag.Diagnostics

	// providerMetaSchemaMutex is a mutex to protect concurrent providerMetaSchema
	// access from race conditions.
	providerMetaSchemaMutex sync.Mutex

	// providerTypeName is the type name of the provider, if the provider
	// implemented the Metadata method.
	providerTypeName string

	// resourceSchemas is the cached Resource Schemas for RPCs that need to
	// convert configuration data from the protocol. If not found, it will be
	// fetched from the ResourceType.GetSchema() method.
	resourceSchemas map[string]fwschema.Schema

	// resourceSchemasDiags is the cached Diagnostics obtained while populating
	// resourceSchemas. This is to ensure any warnings or errors are also
	// returned appropriately when fetching resourceSchemas.
	resourceSchemasDiags diag.Diagnostics

	// resourceSchemasMutex is a mutex to protect concurrent resourceSchemas
	// access from race conditions.
	resourceSchemasMutex sync.Mutex

	// resourceFuncs is the cached Resource functions for RPCs that need to
	// access resources. If not found, it will be fetched from the
	// Provider.Resources() method.
	resourceFuncs map[string]func() resource.Resource

	// resourceTypes is the cached ResourceTypes for RPCs that need to
	// access resources. If not found, it will be fetched from the
	// Provider.GetResources() method.
	//nolint:staticcheck // Internal implementation
	resourceTypes map[string]provider.ResourceType

	// resourceTypesDiags is the cached Diagnostics obtained while populating
	// resourceTypes. This is to ensure any warnings or errors are also
	// returned appropriately when fetching resourceTypes.
	resourceTypesDiags diag.Diagnostics

	// resourceTypesMutex is a mutex to protect concurrent resourceTypes
	// access from race conditions.
	resourceTypesMutex sync.Mutex
}

// DataSource returns the DataSource for a given type name.
func (s *Server) DataSource(ctx context.Context, typeName string) (datasource.DataSource, diag.Diagnostics) {
	dataSourceFuncs, diags := s.DataSourceFuncs(ctx)

	dataSourceFunc, ok := dataSourceFuncs[typeName]

	if ok {
		return dataSourceFunc(), diags
	}

	dataSourceTypes, diags := s.DataSourceTypes(ctx)

	dataSourceType, ok := dataSourceTypes[typeName]

	if !ok {
		diags.AddError(
			"Data Source Type Not Found",
			fmt.Sprintf("No data source type named %q was found in the provider.", typeName),
		)

		return nil, diags
	}

	logging.FrameworkDebug(ctx, "Calling provider defined DataSourceType NewDataSource")
	dataSource, diags := dataSourceType.NewDataSource(ctx, s.Provider)
	logging.FrameworkDebug(ctx, "Called provider defined DataSourceType NewDataSource")

	return dataSource, diags
}

// DataSourceFuncs returns a map of DataSource functions. The results are cached
// on first use.
func (s *Server) DataSourceFuncs(ctx context.Context) (map[string]func() datasource.DataSource, diag.Diagnostics) {
	logging.FrameworkTrace(ctx, "Checking DataSourceTypes lock")
	s.dataSourceTypesMutex.Lock()
	defer s.dataSourceTypesMutex.Unlock()

	if s.dataSourceFuncs != nil {
		return s.dataSourceFuncs, s.dataSourceTypesDiags
	}

	s.dataSourceFuncs = make(map[string]func() datasource.DataSource)

	providerWithDataSources, ok := s.Provider.(provider.ProviderWithDataSources)

	if !ok {
		return s.dataSourceFuncs, nil
	}

	logging.FrameworkDebug(ctx, "Calling provider defined Provider DataSources")
	dataSourceFuncsSlice := providerWithDataSources.DataSources(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined Provider DataSources")

	for _, dataSourceFunc := range dataSourceFuncsSlice {
		dataSource := dataSourceFunc()

		dataSourceWithMetadata, ok := dataSource.(datasource.DataSourceWithMetadata)

		if !ok {
			s.dataSourceTypesDiags.AddError(
				"Data Source Type Name Missing",
				fmt.Sprintf("The %T DataSource in the provider DataSources method results is missing the Metadata method. ", dataSource)+
					"This is always an issue with the provider and should be reported to the provider developers.",
			)
			continue
		}

		dataSourceTypeNameReq := datasource.MetadataRequest{
			ProviderTypeName: s.providerTypeName,
		}
		dataSourceTypeNameResp := datasource.MetadataResponse{}

		dataSourceWithMetadata.Metadata(ctx, dataSourceTypeNameReq, &dataSourceTypeNameResp)

		if dataSourceTypeNameResp.TypeName == "" {
			s.dataSourceTypesDiags.AddError(
				"Data Source Type Name Missing",
				fmt.Sprintf("The %T DataSource returned an empty string from the Metadata method. ", dataSource)+
					"This is always an issue with the provider and should be reported to the provider developers.",
			)
			continue
		}

		logging.FrameworkTrace(ctx, "Found data source type", map[string]interface{}{logging.KeyDataSourceType: dataSourceTypeNameResp.TypeName})

		if _, ok := s.dataSourceFuncs[dataSourceTypeNameResp.TypeName]; ok {
			s.dataSourceTypesDiags.AddError(
				"Duplicate Data Source Type Defined",
				fmt.Sprintf("The %s data source type name was returned for multiple data sources. ", dataSourceTypeNameResp.TypeName)+
					"Data source type names must be unique. "+
					"This is always an issue with the provider and should be reported to the provider developers.",
			)
			continue
		}

		s.dataSourceFuncs[dataSourceTypeNameResp.TypeName] = dataSourceFunc
	}

	return s.dataSourceFuncs, s.dataSourceTypesDiags
}

// DataSourceSchema returns the Schema associated with the DataSourceType for
// the given type name.
func (s *Server) DataSourceSchema(ctx context.Context, typeName string) (fwschema.Schema, diag.Diagnostics) {
	dataSourceSchemas, diags := s.DataSourceSchemas(ctx)

	dataSourceSchema, ok := dataSourceSchemas[typeName]

	if !ok {
		diags.AddError(
			"Data Source Schema Not Found",
			fmt.Sprintf("No data source type named %q was found in the provider to fetch the schema. ", typeName)+
				"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.",
		)

		return nil, diags
	}

	return dataSourceSchema, diags
}

// DataSourceSchemas returns the map of DataSourceType Schemas. The results are
// cached on first use.
func (s *Server) DataSourceSchemas(ctx context.Context) (map[string]fwschema.Schema, diag.Diagnostics) {
	logging.FrameworkTrace(ctx, "Checking DataSourceSchemas lock")
	s.dataSourceSchemasMutex.Lock()
	defer s.dataSourceSchemasMutex.Unlock()

	if s.dataSourceSchemas != nil {
		return s.dataSourceSchemas, s.dataSourceSchemasDiags
	}

	s.dataSourceSchemas = map[string]fwschema.Schema{}

	dataSourceFuncs, diags := s.DataSourceFuncs(ctx)

	s.dataSourceSchemasDiags = diags

	for dataSourceTypeName, dataSourceFunc := range dataSourceFuncs {
		dataSource := dataSourceFunc()

		dataSourceWithGetSchema, ok := dataSource.(datasource.DataSourceWithGetSchema)

		if !ok {
			s.dataSourceSchemasDiags.AddError(
				"Data Source Schema Missing",
				fmt.Sprintf("The %T DataSource in the provider is missing the GetSchema method. ", dataSource)+
					"This is always an issue with the provider and should be reported to the provider developers.",
			)
			continue
		}

		logging.FrameworkDebug(ctx, "Calling provider defined DataSource GetSchema", map[string]interface{}{logging.KeyDataSourceType: dataSourceTypeName})
		schema, diags := dataSourceWithGetSchema.GetSchema(ctx)
		logging.FrameworkDebug(ctx, "Called provider defined DataSource GetSchema", map[string]interface{}{logging.KeyDataSourceType: dataSourceTypeName})

		s.dataSourceSchemasDiags.Append(diags...)

		if s.dataSourceSchemasDiags.HasError() {
			return s.dataSourceSchemas, s.dataSourceSchemasDiags
		}

		s.dataSourceSchemas[dataSourceTypeName] = schema
	}

	if len(s.dataSourceSchemas) > 0 || s.dataSourceSchemasDiags.HasError() {
		return s.dataSourceSchemas, s.dataSourceSchemasDiags
	}

	dataSourceTypes, diags := s.DataSourceTypes(ctx)

	s.dataSourceSchemasDiags = diags

	if s.dataSourceSchemasDiags.HasError() {
		return s.dataSourceSchemas, s.dataSourceSchemasDiags
	}

	for dataSourceTypeName, dataSourceType := range dataSourceTypes {
		logging.FrameworkTrace(ctx, "Found data source type", map[string]interface{}{logging.KeyDataSourceType: dataSourceTypeName})

		logging.FrameworkDebug(ctx, "Calling provider defined DataSourceType GetSchema", map[string]interface{}{logging.KeyDataSourceType: dataSourceTypeName})
		schema, diags := dataSourceType.GetSchema(ctx)
		logging.FrameworkDebug(ctx, "Called provider defined DataSourceType GetSchema", map[string]interface{}{logging.KeyDataSourceType: dataSourceTypeName})

		s.dataSourceSchemasDiags.Append(diags...)

		if s.dataSourceSchemasDiags.HasError() {
			return s.dataSourceSchemas, s.dataSourceSchemasDiags
		}

		s.dataSourceSchemas[dataSourceTypeName] = &schema
	}

	return s.dataSourceSchemas, s.dataSourceSchemasDiags
}

// DataSourceTypes returns the map of DataSourceTypes. The results are cached
// on first use.
//
//nolint:staticcheck // Internal implementation
func (s *Server) DataSourceTypes(ctx context.Context) (map[string]provider.DataSourceType, diag.Diagnostics) {
	logging.FrameworkTrace(ctx, "Checking DataSourceTypes lock")
	s.dataSourceTypesMutex.Lock()
	defer s.dataSourceTypesMutex.Unlock()

	if s.dataSourceTypes != nil {
		return s.dataSourceTypes, s.dataSourceTypesDiags
	}

	s.dataSourceTypes = make(map[string]provider.DataSourceType)

	providerWithGetDataSources, ok := s.Provider.(provider.ProviderWithGetDataSources) //nolint:staticcheck // Internal usage

	if !ok {
		return s.dataSourceTypes, nil
	}

	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetDataSources")
	s.dataSourceTypes, s.dataSourceTypesDiags = providerWithGetDataSources.GetDataSources(ctx) //nolint:staticcheck // Internal usage
	logging.FrameworkDebug(ctx, "Called provider defined Provider GetDataSources")

	return s.dataSourceTypes, s.dataSourceTypesDiags
}

// ProviderSchema returns the Schema associated with the Provider. The Schema
// and Diagnostics are cached on first use.
func (s *Server) ProviderSchema(ctx context.Context) (fwschema.Schema, diag.Diagnostics) {
	logging.FrameworkTrace(ctx, "Checking ProviderSchema lock")
	s.providerSchemaMutex.Lock()
	defer s.providerSchemaMutex.Unlock()

	if s.providerSchema != nil {
		return s.providerSchema, s.providerSchemaDiags
	}

	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetSchema")
	providerSchema, diags := s.Provider.GetSchema(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined Provider GetSchema")

	s.providerSchema = &providerSchema
	s.providerSchemaDiags = diags

	return s.providerSchema, s.providerSchemaDiags
}

// ProviderMetaSchema returns the Meta Schema associated with the Provider, if
// it implements the ProviderWithMetaSchema interface. The Schema and
// Diagnostics are cached on first use.
func (s *Server) ProviderMetaSchema(ctx context.Context) (fwschema.Schema, diag.Diagnostics) {
	providerWithProviderMeta, ok := s.Provider.(provider.ProviderWithMetaSchema)

	if !ok {
		return nil, nil
	}

	logging.FrameworkTrace(ctx, "Provider implements ProviderWithMetaSchema")
	logging.FrameworkTrace(ctx, "Checking ProviderMetaSchema lock")
	s.providerMetaSchemaMutex.Lock()
	defer s.providerMetaSchemaMutex.Unlock()

	if s.providerMetaSchema != nil {
		return s.providerMetaSchema, s.providerMetaSchemaDiags
	}

	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetMetaSchema")
	providerMetaSchema, diags := providerWithProviderMeta.GetMetaSchema(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined Provider GetMetaSchema")

	s.providerMetaSchema = &providerMetaSchema
	s.providerMetaSchemaDiags = diags

	return s.providerMetaSchema, s.providerMetaSchemaDiags
}

// Resource returns the Resource for a given type name.
func (s *Server) Resource(ctx context.Context, typeName string) (resource.Resource, diag.Diagnostics) {
	resourceFuncs, diags := s.ResourceFuncs(ctx)

	resourceFunc, ok := resourceFuncs[typeName]

	if ok {
		return resourceFunc(), diags
	}

	resourceTypes, diags := s.ResourceTypes(ctx)

	resourceType, ok := resourceTypes[typeName]

	if !ok {
		diags.AddError(
			"Resource Type Not Found",
			fmt.Sprintf("No resource type named %q was found in the provider.", typeName),
		)

		return nil, diags
	}

	logging.FrameworkDebug(ctx, "Calling provider defined ResourceType NewResource")
	resource, diags := resourceType.NewResource(ctx, s.Provider)
	logging.FrameworkDebug(ctx, "Called provider defined ResourceType NewResource")

	return resource, diags
}

// ResourceFuncs returns a map of Resource functions. The results are cached
// on first use.
func (s *Server) ResourceFuncs(ctx context.Context) (map[string]func() resource.Resource, diag.Diagnostics) {
	logging.FrameworkTrace(ctx, "Checking ResourceTypes lock")
	s.resourceTypesMutex.Lock()
	defer s.resourceTypesMutex.Unlock()

	if s.resourceFuncs != nil {
		return s.resourceFuncs, s.resourceTypesDiags
	}

	s.resourceFuncs = make(map[string]func() resource.Resource)

	providerWithResources, ok := s.Provider.(provider.ProviderWithResources)

	if !ok {
		return s.resourceFuncs, nil
	}

	logging.FrameworkDebug(ctx, "Calling provider defined Provider Resources")
	resourceFuncsSlice := providerWithResources.Resources(ctx)
	logging.FrameworkDebug(ctx, "Called provider defined Provider Resources")

	for _, resourceFunc := range resourceFuncsSlice {
		res := resourceFunc()

		resourceWithMetadata, ok := res.(resource.ResourceWithMetadata)

		if !ok {
			s.resourceTypesDiags.AddError(
				"Resource Type Name Missing",
				fmt.Sprintf("The %T Resource in the provider Resources method results is missing the Metadata method. ", res)+
					"This is always an issue with the provider and should be reported to the provider developers.",
			)
			continue
		}

		resourceTypeNameReq := resource.MetadataRequest{
			ProviderTypeName: s.providerTypeName,
		}
		resourceTypeNameResp := resource.MetadataResponse{}

		resourceWithMetadata.Metadata(ctx, resourceTypeNameReq, &resourceTypeNameResp)

		if resourceTypeNameResp.TypeName == "" {
			s.resourceTypesDiags.AddError(
				"Resource Type Name Missing",
				fmt.Sprintf("The %T Resource returned an empty string from the Metadata method. ", res)+
					"This is always an issue with the provider and should be reported to the provider developers.",
			)
			continue
		}

		logging.FrameworkTrace(ctx, "Found resource type", map[string]interface{}{logging.KeyResourceType: resourceTypeNameResp.TypeName})

		if _, ok := s.resourceFuncs[resourceTypeNameResp.TypeName]; ok {
			s.resourceTypesDiags.AddError(
				"Duplicate Resource Type Defined",
				fmt.Sprintf("The %s resource type name was returned for multiple resources. ", resourceTypeNameResp.TypeName)+
					"Resource type names must be unique. "+
					"This is always an issue with the provider and should be reported to the provider developers.",
			)
			continue
		}

		s.resourceFuncs[resourceTypeNameResp.TypeName] = resourceFunc
	}

	return s.resourceFuncs, s.resourceTypesDiags
}

// ResourceSchema returns the Schema associated with the ResourceType for
// the given type name.
func (s *Server) ResourceSchema(ctx context.Context, typeName string) (fwschema.Schema, diag.Diagnostics) {
	resourceSchemas, diags := s.ResourceSchemas(ctx)

	resourceSchema, ok := resourceSchemas[typeName]

	if !ok {
		diags.AddError(
			"Resource Schema Not Found",
			fmt.Sprintf("No resource type named %q was found in the provider to fetch the schema. ", typeName)+
				"This is always an issue in terraform-plugin-framework used to implement the provider and should be reported to the provider developers.",
		)

		return nil, diags
	}

	return resourceSchema, diags
}

// ResourceSchemas returns the map of ResourceType Schemas. The results are
// cached on first use.
func (s *Server) ResourceSchemas(ctx context.Context) (map[string]fwschema.Schema, diag.Diagnostics) {
	logging.FrameworkTrace(ctx, "Checking ResourceSchemas lock")
	s.resourceSchemasMutex.Lock()
	defer s.resourceSchemasMutex.Unlock()

	if s.resourceSchemas != nil {
		return s.resourceSchemas, s.resourceSchemasDiags
	}

	s.resourceSchemas = map[string]fwschema.Schema{}

	resourceFuncs, diags := s.ResourceFuncs(ctx)

	s.resourceSchemasDiags = diags

	for resourceTypeName, resourceFunc := range resourceFuncs {
		res := resourceFunc()

		resourceWithGetSchema, ok := res.(resource.ResourceWithGetSchema)

		if !ok {
			s.resourceSchemasDiags.AddError(
				"Resource Schema Missing",
				fmt.Sprintf("The %T Resource in the provider is missing the GetSchema method. ", res)+
					"This is always an issue with the provider and should be reported to the provider developers.",
			)
			continue
		}

		logging.FrameworkDebug(ctx, "Calling provider defined Resource GetSchema", map[string]interface{}{logging.KeyResourceType: resourceTypeName})
		schema, diags := resourceWithGetSchema.GetSchema(ctx)
		logging.FrameworkDebug(ctx, "Called provider defined Resource GetSchema", map[string]interface{}{logging.KeyResourceType: resourceTypeName})

		s.resourceSchemasDiags.Append(diags...)

		if s.resourceSchemasDiags.HasError() {
			return s.resourceSchemas, s.resourceSchemasDiags
		}

		s.resourceSchemas[resourceTypeName] = schema
	}

	if len(s.resourceSchemas) > 0 || s.resourceSchemasDiags.HasError() {
		return s.resourceSchemas, s.resourceSchemasDiags
	}

	resourceTypes, diags := s.ResourceTypes(ctx)

	s.resourceSchemasDiags = diags

	if s.resourceSchemasDiags.HasError() {
		return s.resourceSchemas, s.resourceSchemasDiags
	}

	for resourceTypeName, resourceType := range resourceTypes {
		logging.FrameworkTrace(ctx, "Found resource type", map[string]interface{}{logging.KeyResourceType: resourceTypeName})

		logging.FrameworkDebug(ctx, "Calling provider defined ResourceType GetSchema", map[string]interface{}{logging.KeyResourceType: resourceTypeName})
		schema, diags := resourceType.GetSchema(ctx)
		logging.FrameworkDebug(ctx, "Called provider defined ResourceType GetSchema", map[string]interface{}{logging.KeyResourceType: resourceTypeName})

		s.resourceSchemasDiags.Append(diags...)

		if s.resourceSchemasDiags.HasError() {
			return s.resourceSchemas, s.resourceSchemasDiags
		}

		s.resourceSchemas[resourceTypeName] = &schema
	}

	return s.resourceSchemas, s.resourceSchemasDiags
}

// ResourceTypes returns the map of ResourceTypes. The results are cached
// on first use.
//
//nolint:staticcheck // Internal implementation
func (s *Server) ResourceTypes(ctx context.Context) (map[string]provider.ResourceType, diag.Diagnostics) {
	logging.FrameworkTrace(ctx, "Checking ResourceTypes lock")
	s.resourceTypesMutex.Lock()
	defer s.resourceTypesMutex.Unlock()

	if s.resourceTypes != nil {
		return s.resourceTypes, s.resourceTypesDiags
	}

	s.resourceTypes = make(map[string]provider.ResourceType)

	providerWithGetResources, ok := s.Provider.(provider.ProviderWithGetResources) //nolint:staticcheck // Internal usage

	if !ok {
		return s.resourceTypes, nil
	}

	logging.FrameworkDebug(ctx, "Calling provider defined Provider GetResources")
	s.resourceTypes, s.resourceTypesDiags = providerWithGetResources.GetResources(ctx) //nolint:staticcheck // Internal usage
	logging.FrameworkDebug(ctx, "Called provider defined Provider GetResources")

	return s.resourceTypes, s.resourceTypesDiags
}
