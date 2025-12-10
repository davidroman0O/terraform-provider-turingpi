// Copyright (c) David Roman
// SPDX-License-Identifier: MPL-2.0

package usb_status

import (
	"context"
	"fmt"

	"github.com/davidroman0O/terraform-provider-turingpi/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &UsbStatusDataSource{}

func NewUsbStatusDataSource() datasource.DataSource {
	return &UsbStatusDataSource{}
}

// UsbStatusDataSource defines the data source implementation.
type UsbStatusDataSource struct {
	client *client.Client
}

// UsbStatusDataSourceModel describes the data source data model.
type UsbStatusDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Node  types.String `tfsdk:"node"`
	Mode  types.String `tfsdk:"mode"`
	Route types.String `tfsdk:"route"`
}

func (d *UsbStatusDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_usb_status"
}

func (d *UsbStatusDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Retrieves the current USB configuration of the Turing Pi 2.",
		MarkdownDescription: "Retrieves the current USB configuration of the Turing Pi 2 including which node is connected, the USB mode, and the routing.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this data source.",
				Computed:    true,
			},
			"node": schema.StringAttribute{
				Description: "The node currently connected via USB.",
				Computed:    true,
			},
			"mode": schema.StringAttribute{
				Description: "The current USB mode (Host, Device, or Flash).",
				Computed:    true,
			},
			"route": schema.StringAttribute{
				Description: "The current USB routing (USB-A or BMC).",
				Computed:    true,
			},
		},
	}
}

func (d *UsbStatusDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *UsbStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UsbStatusDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := d.client.UsbGetStatus()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read USB Status",
			fmt.Sprintf("Could not read USB status: %s", err.Error()),
		)
		return
	}

	data.ID = types.StringValue("usb-status")
	data.Node = types.StringValue(status.Node)
	data.Mode = types.StringValue(status.Mode)
	data.Route = types.StringValue(status.Route)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
