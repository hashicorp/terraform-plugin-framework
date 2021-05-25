package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type resourceComputeInstanceType struct{}

func (r resourceComputeInstanceType) GetSchema() schema.Schema {
	return schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"name": {
				Type:        types.StringType,
				Description: "The name to associate with the compute instance.",
				Required:    true,
			},
			"disks": {
				Attributes: schema.ListNestedAttributes(map[string]schema.Attribute{
					"id": {
						Type:        types.StringType,
						Description: "The ID of the pre-existing disk to associate with this compute instance.",
						Required:    true,
					},
					"delete_with_instance": {
						Type:        types.BoolType,
						Description: "Set to true to delete the disk and all its contents when the instance is deleted.",
						Optional:    true,
					},
				}, schema.ListNestedAttributesOptions{}),
				Optional: true,
				Computed: true,
			},
		},
	}
}

// This would actually be `NewResource(p schemaProvider) schemaResource` but we
// don't have those types yet, so... placeholders. Gross.
func (r resourceComputeInstanceType) NewResource(p Provider) Resource {
	return resourceComputeInstance{
		client: p.(*http.Client),
	}
}

type resourceComputeInstance struct {
	client *http.Client
}

type resourceComputeInstanceValues struct {
	Name  types.String `tfsdk:"name"`
	Disks types.List   `tfsdk:"disks"`
}

type resourceComputeInstanceDisksValues struct {
	ID                 types.String `tfsdk:"id"`
	DeleteWithInstance types.Bool   `tfsdk:"delete_with_instance"`
}

func (r resourceComputeInstance) Create(ctx context.Context, req CreateResourceRequest, resp CreateResourceResponse) {
	var values resourceComputeInstanceValues
	err := req.Plan.Get(ctx, &values)
	if err != nil {
		// If this is a tftypes.AttributePathError, we could have it
		// set the attribute path for us automatically :D
		resp.WithError("Error parsing plan", err)
		return
	}
	apiReq := map[string]interface{}{
		// name is required, it will never be null or unknown
		"name": values.Name.Value,
	}
	// disks are optional and computed; they'll never be null, but they may
	// be unknown, in which case we'll fill them in ourselves.
	if !values.Disks.Unknown {
		// we don't want to get these as a types.Object, we want them
		// as a resourceComputeInstanceDisksValues. So let's get them
		// as that.
		path := tftypes.NewAttributePath().WithAttributeName("disks")
		apiReq["disks"] = make([]map[string]interface{}, 0, len(values.Disks.Elems))
		for pos := range values.Disks.Elems {
			var disk resourceComputeInstanceDisksValues
			err := req.Plan.GetAttribute(ctx, path.WithElementKeyInt(int64(pos)), &disk)
			if err != nil {
				resp.WithError("Error parsing disk from plan", err)
				return
			}
			apiReq["disks"] = append(apiReq["disks"].([]map[string]interface{}), map[string]interface{}{
				"id":                   disk.ID,
				"delete_with_instance": disk.DeleteWithInstance,
			})
		}
	}
	apiResp, err := createComputeInstance(ctx, apiReq)
	if err != nil {
		resp.WithError("Error creating instance", err)
		return
	}
	err = resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("name"), values.Name)
	if err != nil {
		resp.WithError("Error setting name in response", err)
		return
	}
	if values.Disks.Unknown {
		// if our disks were unknown, we should set them from the response
		for diskNo := range values.Disks.Elems {
			path := tftypes.NewAttributePath().WithAttributeName("disks").WithElementKeyInt(int64(diskNo))
			// this next line is gross, but wouldn't be if this were real, because API clients are a thing
			disk := apiResp["disks"].([]map[string]interface{})[diskNo]
			err = resp.State.SetAttribute(ctx, path.WithAttributeName("id"), types.String{Value: disk["id"].(string)})
			if err != nil {
				resp.WithError("Error setting disk ID in response", err)
				return
			}
			err = resp.State.SetAttribute(ctx, path.WithAttributeName("delete_with_instance"), types.Bool{Value: disk["delete_with_instance"].(bool)})
			if err != nil {
				resp.WithError("Error setting disk DeleteWithInstance in response", err)
				return
			}
		}
	} else {
		// otherwise, we should set them from what was in the config
		err = resp.State.SetAttribute(ctx, tftypes.NewAttributePath().WithAttributeName("disks"), values.Disks)
		if err != nil {
			resp.WithError("Error setting disks in response", err)
			return
		}
	}
}

// a lazy placeholder client
func createComputeInstance(ctx context.Context, vals map[string]interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(vals)
	if err != nil {
		return nil, err
	}
	apiResp, err := http.Post("https://my.api/resource", "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer apiResp.Body.Close()
	apiRespBody, err := ioutil.ReadAll(apiResp.Body)
	if err != nil {
		return nil, err
	}
	apiRespParsed := map[string]interface{}{}
	err = json.Unmarshal(apiRespBody, &apiRespParsed)
	if err != nil {
		return nil, err
	}
	return apiRespParsed, nil
}
