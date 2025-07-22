// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	identityschema "github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ LinkedResourceType = LinkedResource{}
	_ RawLinkedResource  = RawV5LinkedResource{}
	_ RawLinkedResource  = RawV6LinkedResource{}
)

// TODO:Actions: docs
type LinkedResourceType interface {
	isLinkedResourceType()

	GetTypeName() string
	GetDescription() string
}

// TODO:Actions: docs
type RawLinkedResource interface {
	LinkedResourceType

	GetSchema() fwschema.Schema
	GetIdentitySchema() fwschema.Schema
}

type LinkedResources []LinkedResource

// TODO:Actions: docs
type LinkedResource struct {
	TypeName    string
	Description string
}

func (l LinkedResource) isLinkedResourceType() {}

func (l LinkedResource) GetTypeName() string {
	return l.TypeName
}

func (l LinkedResource) GetDescription() string {
	return l.Description
}

// TODO:Actions: docs
type RawV5LinkedResource struct {
	TypeName    string
	Description string

	// TODO:Actions: It feels likely that we'd want to receive these as functions, in-case the provider schema is rather large :)
	Schema         *tfprotov5.Schema
	IdentitySchema *tfprotov5.ResourceIdentitySchema
}

func (l RawV5LinkedResource) isLinkedResourceType() {}

func (l RawV5LinkedResource) GetTypeName() string {
	return l.TypeName
}

func (l RawV5LinkedResource) GetDescription() string {
	return l.Description
}

func (l RawV5LinkedResource) GetSchema() fwschema.Schema {
	// TODO:Actions: This logic should probably live in an internal package, maybe fromproto
	attrs := make(map[string]resourceschema.Attribute, len(l.Schema.Block.Attributes))
	for _, attr := range l.Schema.Block.Attributes {
		switch {
		case attr.Type.Is(tftypes.Bool):
			attrs[attr.Name] = resourceschema.BoolAttribute{
				Required:  attr.Required,
				Optional:  attr.Optional,
				Computed:  attr.Computed,
				WriteOnly: attr.WriteOnly,
				Sensitive: attr.Sensitive,
				// TODO:Actions: Do we need to set more than these? Probs not.
			}
		case attr.Type.Is(tftypes.Number):
			attrs[attr.Name] = resourceschema.NumberAttribute{
				Required:  attr.Required,
				Optional:  attr.Optional,
				Computed:  attr.Computed,
				WriteOnly: attr.WriteOnly,
				Sensitive: attr.Sensitive,
				// TODO:Actions: Do we need to set more than these? Probs not.
			}
		case attr.Type.Is(tftypes.String):
			attrs[attr.Name] = resourceschema.StringAttribute{
				Required:  attr.Required,
				Optional:  attr.Optional,
				Computed:  attr.Computed,
				WriteOnly: attr.WriteOnly,
				Sensitive: attr.Sensitive,
				// TODO:Actions: Do we need to set more than these? Probs not.
			}
			// TODO:Actions: All other types (collections/structural/dynamic)
			// TODO:Actions: This should essentially be the inverse of toproto schema mapping logic
		}

		// TODO:Actions: Block mapping
	}
	return resourceschema.Schema{
		Attributes: attrs,
		Blocks:     map[string]resourceschema.Block{},
		// TODO:Actions: Do we need to set more than these? Probs not.
	}
}

func (l RawV5LinkedResource) GetIdentitySchema() fwschema.Schema {
	// It's valid for a managed resource to not support identity, we return nil to indicate to
	// other pieces of framework logic that there is no identity support for this resource.
	if l.IdentitySchema == nil {
		return nil
	}

	// TODO:Actions: This logic should probably live in an internal package, maybe fromproto
	attrs := make(map[string]identityschema.Attribute, len(l.IdentitySchema.IdentityAttributes))
	for _, attr := range l.IdentitySchema.IdentityAttributes {
		switch {
		case attr.Type.Is(tftypes.Bool):
			attrs[attr.Name] = identityschema.BoolAttribute{
				RequiredForImport: attr.RequiredForImport,
				OptionalForImport: attr.OptionalForImport,
				// TODO:Actions: Do we need to set more than these? Probs not.
			}
		case attr.Type.Is(tftypes.Number):
			attrs[attr.Name] = identityschema.NumberAttribute{
				RequiredForImport: attr.RequiredForImport,
				OptionalForImport: attr.OptionalForImport,
				// TODO:Actions: Do we need to set more than these? Probs not.
			}
		case attr.Type.Is(tftypes.String):
			attrs[attr.Name] = identityschema.StringAttribute{
				RequiredForImport: attr.RequiredForImport,
				OptionalForImport: attr.OptionalForImport,
				// TODO:Actions: Do we need to set more than these? Probs not.
			}
			// TODO:Actions: All other types
			// TODO:Actions: This should essentially be the inverse of toproto schema mapping logic
		}
	}
	return identityschema.Schema{
		Attributes: attrs,
		// TODO:Actions: Do we need to set more than these? Probs not.
	}
}

// TODO:Actions: docs
type RawV6LinkedResource struct {
	TypeName    string
	Description string

	// TODO:Actions: It feels likely that we'd want to receive these as functions, in-case the provider schema is rather large :)
	Schema         *tfprotov6.Schema
	IdentitySchema *tfprotov6.ResourceIdentitySchema
}

func (l RawV6LinkedResource) isLinkedResourceType() {}

func (l RawV6LinkedResource) GetTypeName() string {
	return l.TypeName
}

func (l RawV6LinkedResource) GetDescription() string {
	return l.Description
}

// TODO:Actions: Would it be invalid to use a v6 linked resource in a v5 action? My initial thought is that
// this would never happen (since the provider must all be the same protocol version at the end of the day to Terraform,
// and providers can't build actions for other providers), but I can't think of a reason why we couldn't do this?
//
// The data is all the same under the hood, but perhaps there are some validations that might break down when attempting to prevent
// setting data in nested computed attributes? :shrug:
//
// We can very easily validate this in the proto5server/proto6server in our type switch, just need to determine if that restriction is reasonable.
func (l RawV6LinkedResource) GetSchema() fwschema.Schema {
	// TODO:Actions: This logic should probably live in an internal package, maybe fromproto
	attrs := make(map[string]resourceschema.Attribute, len(l.Schema.Block.Attributes))
	for _, attr := range l.Schema.Block.Attributes {
		switch {
		case attr.Type.Is(tftypes.Bool):
			attrs[attr.Name] = resourceschema.BoolAttribute{
				Required:  attr.Required,
				Optional:  attr.Optional,
				Computed:  attr.Computed,
				WriteOnly: attr.WriteOnly,
				Sensitive: attr.Sensitive,
				// TODO:Actions: Do we need to set more than these? Probs not.
			}
		case attr.Type.Is(tftypes.Number):
			attrs[attr.Name] = resourceschema.NumberAttribute{
				Required:  attr.Required,
				Optional:  attr.Optional,
				Computed:  attr.Computed,
				WriteOnly: attr.WriteOnly,
				Sensitive: attr.Sensitive,
				// TODO:Actions: Do we need to set more than these? Probs not.
			}
		case attr.Type.Is(tftypes.String):
			attrs[attr.Name] = resourceschema.StringAttribute{
				Required:  attr.Required,
				Optional:  attr.Optional,
				Computed:  attr.Computed,
				WriteOnly: attr.WriteOnly,
				Sensitive: attr.Sensitive,
				// TODO:Actions: Do we need to set more than these? Probs not.
			}
			// TODO:Actions: All other types (collections/structural/dynamic)
			// TODO:Actions: This should essentially be the inverse of toproto schema mapping logic
		}

		// TODO:Actions: Block mapping
	}
	return resourceschema.Schema{
		Attributes: attrs,
		Blocks:     map[string]resourceschema.Block{},
		// TODO:Actions: Do we need to set more than these? Probs not.
	}
}

func (l RawV6LinkedResource) GetIdentitySchema() fwschema.Schema {
	// It's valid for a managed resource to not support identity, we return nil to indicate to
	// other pieces of framework logic that there is no identity support for this resource.
	if l.IdentitySchema == nil {
		return nil
	}

	// TODO:Actions: This logic should probably live in an internal package, maybe fromproto
	attrs := make(map[string]identityschema.Attribute, len(l.IdentitySchema.IdentityAttributes))
	for _, attr := range l.IdentitySchema.IdentityAttributes {
		switch {
		case attr.Type.Is(tftypes.Bool):
			attrs[attr.Name] = identityschema.BoolAttribute{
				RequiredForImport: attr.RequiredForImport,
				OptionalForImport: attr.OptionalForImport,
				// TODO:Actions: Do we need to set more than these? Probs not.
			}
		case attr.Type.Is(tftypes.Number):
			attrs[attr.Name] = identityschema.NumberAttribute{
				RequiredForImport: attr.RequiredForImport,
				OptionalForImport: attr.OptionalForImport,
				// TODO:Actions: Do we need to set more than these? Probs not.
			}
		case attr.Type.Is(tftypes.String):
			attrs[attr.Name] = identityschema.StringAttribute{
				RequiredForImport: attr.RequiredForImport,
				OptionalForImport: attr.OptionalForImport,
				// TODO:Actions: Do we need to set more than these? Probs not.
			}
			// TODO:Actions: All other types
			// TODO:Actions: This should essentially be the inverse of toproto schema mapping logic
		}
	}
	return identityschema.Schema{
		Attributes: attrs,
		// TODO:Actions: Do we need to set more than these? Probs not.
	}
}
