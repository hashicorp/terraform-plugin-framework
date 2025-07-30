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

// TODO:Actions: docs
type LinkedResourceType interface {
	isLinkedResourceType()

	GetTypeName() string
	GetDescription() string
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
	TypeName       string
	Description    string
	Schema         func() *tfprotov5.Schema
	IdentitySchema func() *tfprotov5.ResourceIdentitySchema
}

func (l RawV5LinkedResource) isLinkedResourceType() {}

func (l RawV5LinkedResource) GetTypeName() string {
	return l.TypeName
}

func (l RawV5LinkedResource) GetDescription() string {
	return l.Description
}

// TODO:Actions: docs
type RawV6LinkedResource struct {
	TypeName       string
	Description    string
	Schema         func() *tfprotov6.Schema
	IdentitySchema func() *tfprotov6.ResourceIdentitySchema
}

func (l RawV6LinkedResource) isLinkedResourceType() {}

func (l RawV6LinkedResource) GetTypeName() string {
	return l.TypeName
}

func (l RawV6LinkedResource) GetDescription() string {
	return l.Description
}
