package fromproto6

import (
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func ReadDataSourceClientCapabilities(in *tfprotov6.ReadDataSourceClientCapabilities) *datasource.ReadClientCapabilities {
	if in == nil {
		return nil
	}

	resp := &datasource.ReadClientCapabilities{
		DeferralAllowed: in.DeferralAllowed,
	}

	return resp
}

func ReadResourceClientCapabilities(in *tfprotov6.ReadResourceClientCapabilities) *resource.ReadClientCapabilities {
	if in == nil {
		return nil
	}

	resp := &resource.ReadClientCapabilities{
		DeferralAllowed: in.DeferralAllowed,
	}

	return resp
}

func ModifyPlanClientCapabilities(in *tfprotov6.PlanResourceChangeClientCapabilities) *resource.ModifyPlanClientCapabilities {
	if in == nil {
		return nil
	}

	resp := &resource.ModifyPlanClientCapabilities{
		DeferralAllowed: in.DeferralAllowed,
	}

	return resp
}

func ImportStateClientCapabilities(in *tfprotov6.ImportResourceStateClientCapabilities) *resource.ImportStateClientCapabilities {
	if in == nil {
		return nil
	}

	resp := &resource.ImportStateClientCapabilities{
		DeferralAllowed: in.DeferralAllowed,
	}

	return resp
}
