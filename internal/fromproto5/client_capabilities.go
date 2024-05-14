// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fromproto5

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func ConfigureProviderClientCapabilities(in *tfprotov5.ConfigureProviderClientCapabilities) provider.ConfigureProviderClientCapabilities {
	if in == nil {
		return provider.ConfigureProviderClientCapabilities{}
	}

	resp := provider.ConfigureProviderClientCapabilities{
		DeferralAllowed: in.DeferralAllowed,
	}

	return resp
}

func ReadDataSourceClientCapabilities(in *tfprotov5.ReadDataSourceClientCapabilities) *datasource.ReadClientCapabilities {
	if in == nil {
		return nil
	}

	resp := &datasource.ReadClientCapabilities{
		DeferralAllowed: in.DeferralAllowed,
	}

	return resp
}

func ReadResourceClientCapabilities(in *tfprotov5.ReadResourceClientCapabilities) *resource.ReadClientCapabilities {
	if in == nil {
		return nil
	}

	resp := &resource.ReadClientCapabilities{
		DeferralAllowed: in.DeferralAllowed,
	}

	return resp
}

func ModifyPlanClientCapabilities(in *tfprotov5.PlanResourceChangeClientCapabilities) *resource.ModifyPlanClientCapabilities {
	if in == nil {
		return nil
	}

	resp := &resource.ModifyPlanClientCapabilities{
		DeferralAllowed: in.DeferralAllowed,
	}

	return resp
}

func ImportStateClientCapabilities(in *tfprotov5.ImportResourceStateClientCapabilities) *resource.ImportStateClientCapabilities {
	if in == nil {
		return nil
	}

	resp := &resource.ImportStateClientCapabilities{
		DeferralAllowed: in.DeferralAllowed,
	}

	return resp
}
