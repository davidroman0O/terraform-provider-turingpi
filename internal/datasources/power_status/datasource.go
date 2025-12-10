// Copyright (c) David Roman
// SPDX-License-Identifier: MPL-2.0

package power_status

import (
	"context"
	"fmt"

	"github.com/davidroman0O/terraform-provider-turingpi/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &PowerStatusDataSource{}

func NewPowerStatusDataSource() datasource.DataSource {
	return &PowerStatusDataSource{}
}

// PowerStatusDataSource defines the data source implementation.
type PowerStatusDataSource struct {
	client *client.Client
}

// PowerStatusDataSourceModel describes the data source data model.
type PowerStatusDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Node1 types.Bool   `tfsdk:"node1"`
	Node2 types.Bool   `tfsdk:"node2"`
	Node3 types.Bool   `tfsdk:"node3"`
	Node4 types.Bool   `tfsdk:"node4"`
}

func (d *PowerStatusDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_power_status"
}

func (d *PowerStatusDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Retrieves the power status of all nodes on the Turing Pi 2.",
		MarkdownDescription: "Retrieves the power status of all nodes on the Turing Pi 2. Each node attribute indicates whether the node is powered on (`true`) or off (`false`).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this data source.",
				Computed:    true,
			},
			"node1": schema.BoolAttribute{
				Description: "Power status of node 1 (true = on, false = off).",
				Computed:    true,
			},
			"node2": schema.BoolAttribute{
				Description: "Power status of node 2 (true = on, false = off).",
				Computed:    true,
			},
			"node3": schema.BoolAttribute{
				Description: "Power status of node 3 (true = on, false = off).",
				Computed:    true,
			},
			"node4": schema.BoolAttribute{
				Description: "Power status of node 4 (true = on, false = off).",
				Computed:    true,
			},
		},
	}
}

func (d *PowerStatusDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PowerStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PowerStatusDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := d.client.PowerStatus()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Power Status",
			fmt.Sprintf("Could not read power status: %s", err.Error()),
		)
		return
	}

	data.ID = types.StringValue("power-status")
	data.Node1 = types.BoolValue(status[1])
	data.Node2 = types.BoolValue(status[2])
	data.Node3 = types.BoolValue(status[3])
	data.Node4 = types.BoolValue(status[4])

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
