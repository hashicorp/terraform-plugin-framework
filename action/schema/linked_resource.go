// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var (
	_ LinkedResourceType = LinkedResource{}
	_ LinkedResourceType = RawV5LinkedResource{}
	_ LinkedResourceType = RawV6LinkedResource{}
)

// LinkedResourceType is the interface that a linked resource type must implement. Linked resource
// types are statically defined by framework, so this interface is not meant to be implemented
// outside of this package.
//
// LinkedResourceType implementations allow a provider to describe to Terraform the managed resource types that
// can be modified by lifecycle and linked actions. The most common implementation to use is [LinkedResource], however,
// additional implementations exist to allow providers to write actions that modify managed resources that are
// written in SDKs other than framework. [RawV5LinkedResource] can be used for protocol v5 managed resources
// (terraform-plugin-sdk/v2 or terraform-plugin-go) and [RawV6LinkedResource] can be used for protocol v6 managed
// resources (terraform-plugin-go).
//
// Regardless of which linked resource type is used in the schema, the methods for accessing and setting the data during
// the action plan and invoke are the same.
type LinkedResourceType interface {
	// Linked resource types are statically defined by framework, so this
	// interface is not meant to be implemented outside of this package
	isLinkedResourceType()

	// GetTypeName returns the full name of the managed resource which can have it's resource state changed by the action.
	GetTypeName() string

	// GetDescription returns the human-readable description of the linked resource.
	GetDescription() string
}

// LinkedResource describes to Terraform a managed resource that can be modified by a lifecycle or linked action. This
// implementation only needs the TypeName of the managed resource.
//
// The linked resource must be defined on the same framework provider server as the action.
type LinkedResource struct {
	// TypeName is the name of the managed resource which can have it's resource state changed by the action.
	// The name should be prefixed with the provider shortname and an underscore.
	//
	// The linked resource must be defined in the same provider as the action is defined.
	TypeName string

	// Description is a human-readable description of the linked resource.
	Description string
}

func (l LinkedResource) isLinkedResourceType() {}

func (l LinkedResource) GetTypeName() string {
	return l.TypeName
}

func (l LinkedResource) GetDescription() string {
	return l.Description
}

// RawV5LinkedResource describes to Terraform a managed resource that can be modified by a lifecycle or linked action. This
// implementation needs the TypeName, Schema and (if supported) the resource identity schema of the managed resource. The most common
// scenario for using this linked resource type is when defining an action that modifies a resource implemented with terraform-plugin-sdk/v2.
//
// If the linked resource is already defined in framework, use [LinkedResource]. If the linked resource is implemented with
// protocol v6, use [RawV6LinkedResource].
type RawV5LinkedResource struct {
	// TypeName is the name of the managed resource which can have it's resource state changed by the action.
	// The name should be prefixed with the provider shortname and an underscore.
	//
	// The linked resource must be defined in the same provider as the action is defined.
	TypeName string

	// Description is a human-readable description of the linked resource.
	Description string

	// Schema is a function that returns the protocol v5 schema of the linked resource.
	Schema func() *tfprotov5.Schema

	// IdentitySchema is a function returns the protocol v5 identity schema of the linked resource. This field
	// is only needed if the managed resource supports resource identity.
	IdentitySchema func() *tfprotov5.ResourceIdentitySchema
}

func (l RawV5LinkedResource) isLinkedResourceType() {}

func (l RawV5LinkedResource) GetTypeName() string {
	return l.TypeName
}

func (l RawV5LinkedResource) GetDescription() string {
	return l.Description
}

func (l RawV5LinkedResource) GetSchema() *tfprotov5.Schema {
	if l.Schema == nil {
		return nil
	}
	return l.Schema()
}

func (l RawV5LinkedResource) GetIdentitySchema() *tfprotov5.ResourceIdentitySchema {
	if l.IdentitySchema == nil {
		return nil
	}
	return l.IdentitySchema()
}

// RawV6LinkedResource describes to Terraform a managed resource that can be modified by a lifecycle or linked action. This
// implementation needs the TypeName, Schema and (if supported) the resource identity schema of the managed resource. The most common
// scenario for using this linked resource type is when defining an action that modifies a resource implemented with terraform-plugin-go.
//
// If the linked resource is already defined in framework, use [LinkedResource]. If the linked resource is implemented with
// protocol v5, use [RawV5LinkedResource].
type RawV6LinkedResource struct {
	// TypeName is the name of the managed resource which can have it's resource state changed by the action.
	// The name should be prefixed with the provider shortname and an underscore.
	//
	// The linked resource must be defined in the same provider as the action is defined.
	TypeName string

	// Description is a human-readable description of the linked resource.
	Description string

	// Schema is a function returns the protocol v6 schema of the linked resource.
	Schema func() *tfprotov6.Schema

	// IdentitySchema is a function returns the protocol v6 identity schema of the linked resource. This field
	// is only needed if the managed resource supports resource identity.
	IdentitySchema func() *tfprotov6.ResourceIdentitySchema
}

func (l RawV6LinkedResource) isLinkedResourceType() {}

func (l RawV6LinkedResource) GetTypeName() string {
	return l.TypeName
}

func (l RawV6LinkedResource) GetDescription() string {
	return l.Description
}

func (l RawV6LinkedResource) GetSchema() *tfprotov6.Schema {
	if l.Schema == nil {
		return nil
	}
	return l.Schema()
}

func (l RawV6LinkedResource) GetIdentitySchema() *tfprotov6.ResourceIdentitySchema {
	if l.IdentitySchema == nil {
		return nil
	}
	return l.IdentitySchema()
}
