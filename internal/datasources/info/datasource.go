// Copyright (c) David Roman
// SPDX-License-Identifier: MPL-2.0

package info

import (
	"context"
	"fmt"

	"github.com/davidroman0O/terraform-provider-turingpi/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &InfoDataSource{}

func NewInfoDataSource() datasource.DataSource {
	return &InfoDataSource{}
}

// InfoDataSource defines the data source implementation.
type InfoDataSource struct {
	client *client.Client
}

// InfoDataSourceModel describes the data source data model.
type InfoDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	IP           types.String `tfsdk:"ip"`
	MAC          types.String `tfsdk:"mac"`
	Version      types.String `tfsdk:"version"`
	APIVersion   types.String `tfsdk:"api_version"`
	BuildTime    types.String `tfsdk:"build_time"`
	BuildVersion types.String `tfsdk:"build_version"`
	Buildroot    types.String `tfsdk:"buildroot"`
}

func (d *InfoDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_info"
}

func (d *InfoDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Retrieves information about the Turing Pi 2 BMC.",
		MarkdownDescription: "Retrieves information about the Turing Pi 2 BMC including IP address, MAC address, firmware version, and build information.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Identifier for this data source.",
				Computed:    true,
			},
			"ip": schema.StringAttribute{
				Description: "IP address of the BMC.",
				Computed:    true,
			},
			"mac": schema.StringAttribute{
				Description: "MAC address of the BMC.",
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "Firmware version.",
				Computed:    true,
			},
			"api_version": schema.StringAttribute{
				Description: "API version.",
				Computed:    true,
			},
			"build_time": schema.StringAttribute{
				Description: "Build timestamp.",
				Computed:    true,
			},
			"build_version": schema.StringAttribute{
				Description: "Build version string.",
				Computed:    true,
			},
			"buildroot": schema.StringAttribute{
				Description: "Buildroot version.",
				Computed:    true,
			},
		},
	}
}

func (d *InfoDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (d *InfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data InfoDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get basic info
	info, err := d.client.Info()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read BMC Info",
			fmt.Sprintf("Could not read BMC info: %s", err.Error()),
		)
		return
	}

	// Get detailed about info
	about, err := d.client.About()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read BMC About",
			fmt.Sprintf("Could not read BMC about info: %s", err.Error()),
		)
		return
	}

	// Map response to model
	data.ID = types.StringValue("bmc-info")
	data.IP = types.StringValue(getMapValue(info, "ip"))
	data.MAC = types.StringValue(getMapValue(info, "mac"))
	data.Version = types.StringValue(getMapValue(about, "version"))
	data.APIVersion = types.StringValue(getMapValue(info, "api"))
	data.BuildTime = types.StringValue(getMapValue(info, "buildtime"))
	data.BuildVersion = types.StringValue(getMapValue(about, "build_version"))
	data.Buildroot = types.StringValue(getMapValue(about, "buildroot"))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// getMapValue safely gets a value from a map, returning empty string if not found.
func getMapValue(m map[string]string, key string) string {
	if v, ok := m[key]; ok {
		return v
	}
	return ""
}
