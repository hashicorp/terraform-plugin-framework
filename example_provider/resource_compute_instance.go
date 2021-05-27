package provider

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"io/ioutil"
// 	"net/http"

// 	"github.com/hashicorp/terraform-plugin-framework/schema"
// 	"github.com/hashicorp/terraform-plugin-framework/types"
// 	"github.com/hashicorp/terraform-plugin-go/tftypes"
// )

// type resourceComputeInstanceType struct{}

// func (r resourceComputeInstanceType) GetSchema() schema.Schema {
// 	return schema.Schema{
// 		Version: 1,
// 		Attributes: map[string]schema.Attribute{
// 			"name": {
// 				Type:        types.StringType,
// 				Description: "The name to associate with the compute instance.",
// 				Required:    true,
// 			},
// 			"disks": {
// 				Attributes: schema.ListNestedAttributes(map[string]schema.Attribute{
// 					"id": {
// 						Type:        types.StringType,
// 						Description: "The ID of the pre-existing disk to associate with this compute instance.",
// 						Required:    true,
// 					},
// 					"delete_with_instance": {
// 						Type:        types.BoolType,
// 						Description: "Set to true to delete the disk and all its contents when the instance is deleted.",
// 						Optional:    true,
// 					},
// 				}, schema.ListNestedAttributesOptions{}),
// 				Optional: true,
// 				Computed: true,
// 			},
// 		},
// 	}
// }

// // This would actually be `NewResource(p tfsdk.Provider) tfsdk.Resource` but we
// // don't have those types yet, so... placeholders. Gross.
// func (r resourceComputeInstanceType) NewResource(p Provider) Resource {
// 	return resourceComputeInstance{
// 		client: p.(*http.Client),
// 	}
// }

// type resourceComputeInstance struct {
// 	client *http.Client
// }

// type resourceComputeInstanceValues struct {
// 	Name  types.String `tfsdk:"name"`
// 	Disks types.List   `tfsdk:"disks"`
// }

// type resourceComputeInstanceDisksValues struct {
// 	ID                 types.String `tfsdk:"id"`
// 	DeleteWithInstance types.Bool   `tfsdk:"delete_with_instance"`
// }

// func (r resourceComputeInstance) Create(ctx context.Context, req CreateResourceRequest, resp CreateResourceResponse) {
// 	var values resourceComputeInstanceValues
// 	err := req.Plan.Get(ctx, &values)
// 	if err != nil {
// 		// If this is a tftypes.AttributePathError, we could have it
// 		// set the attribute path for us automatically :D
// 		resp.WithError("Error parsing plan", err)
// 		return
// 	}
// 	apiReq := map[string]interface{}{
// 		// name is required, it will never be null or unknown
// 		"name": values.Name.Value,
// 	}
// 	// disks are optional and computed; they'll never be null, but they may
// 	// be unknown, in which case we'll fill them in ourselves.
// 	if !values.Disks.Unknown {
// 		// we don't want to get these as a types.Object, we want them
// 		// as a resourceComputeInstanceDisksValues. So let's get them
// 		// as that.
// 		var disks []resourceComputeInstanceDisksValues
// 		err := values.Disks.ElementsAs(ctx, &disks, true)
// 		if err != nil {
// 			resp.WithError("Error parsing disks from plan", err)
// 			return
// 		}
// 		for _, disk := range disks {
// 			// ignore how gross this is, it's because we don't have
// 			// a real API client
// 			apiReq["disks"] = append(apiReq["disks"].([]map[string]interface{}), map[string]interface{}{
// 				"id":                   disk.ID,
// 				"delete_with_instance": disk.DeleteWithInstance,
// 			})
// 		}
// 	}
// 	apiResp, err := createComputeInstance(ctx, apiReq)
// 	if err != nil {
// 		resp.WithError("Error creating instance", err)
// 		return
// 	}
// 	err = resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("name"), values.Name)
// 	if err != nil {
// 		resp.WithError("Error setting name in response", err)
// 		return
// 	}
// 	if values.Disks.Unknown {
// 		// if our disks were unknown, we should set them from the response
// 		var disks []resourceComputeInstanceDisksValues
// 		for _, disk := range apiResp["disks"].([]map[string]interface{}) {
// 			disks = append(disks, resourceComputeInstanceDisksValues{
// 				ID:                 types.String{Value: disk["id"].(string)},
// 				DeleteWithInstance: types.Bool{Value: disk["delete_with_instance"].(bool)},
// 			})
// 		}

// 		// diskVals, err := types.ListOf(disks)
// 		// if err != nil {
// 		// 	resp.WithError("Error converting disks into list", err)
// 		// 	return
// 		// }

// 		// var diskElems []tftypes.Object
// 		// for _, disk := range disks {
// 		// 	diskElems = append(diskElems, tftypes.NewValue(tftypes.Object{
// 		// 		AttributeTypes: map[string]tftypes.Type{
// 		// 			"id":                   tftypes.String,
// 		// 			"delete_with_instance": tftypes.Bool,
// 		// 		},
// 		// 	}, map[string]tftypes.Value{
// 		// 		"id":                   tftypes.NewValue(tftypes.String, disk.ID),
// 		// 		"delete_with_instance": tftypes.NewValue(tftypes.Bool, disk.DeleteWithInstance),
// 		// 	},
// 		// 	))
// 		// }

// 		// diskVals := types.List{
// 		// 	Elems:    diskElems,
// 		// 	ElemType: tftypes.Object,
// 		// }

// 		// err = resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("disks"), diskVals)
// 		// if err != nil {
// 		// 	resp.WithError("Error setting disks in state", err)
// 		// 	return
// 		// }
// 	} else {
// 		// otherwise, we should set them from what was in the config
// 		err = resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("disks"), values.Disks)
// 		if err != nil {
// 			resp.WithError("Error setting disks in response", err)
// 			return
// 		}
// 	}
// }

// // a lazy placeholder client
// func createComputeInstance(ctx context.Context, vals map[string]interface{}) (map[string]interface{}, error) {
// 	b, err := json.Marshal(vals)
// 	if err != nil {
// 		return nil, err
// 	}
// 	apiResp, err := http.Post("https://my.api/resource", "application/json", bytes.NewReader(b))
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer apiResp.Body.Close()
// 	apiRespBody, err := ioutil.ReadAll(apiResp.Body)
// 	if err != nil {
// 		return nil, err
// 	}
// 	apiRespParsed := map[string]interface{}{}
// 	err = json.Unmarshal(apiRespBody, &apiRespParsed)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return apiRespParsed, nil
// }
