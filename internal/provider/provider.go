// Copyright (c) David Roman
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/davidroman0O/terraform-provider-turingpi/internal/client"
	"github.com/davidroman0O/terraform-provider-turingpi/internal/datasources/info"
	"github.com/davidroman0O/terraform-provider-turingpi/internal/datasources/power_status"
	"github.com/davidroman0O/terraform-provider-turingpi/internal/datasources/usb_status"
	"github.com/davidroman0O/terraform-provider-turingpi/internal/resources/node_flash"
	"github.com/davidroman0O/terraform-provider-turingpi/internal/resources/node_power"
	"github.com/davidroman0O/terraform-provider-turingpi/internal/resources/node_usb"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure TuringPiProvider satisfies various provider interfaces.
var _ provider.Provider = &TuringPiProvider{}

// TuringPiProvider defines the provider implementation.
type TuringPiProvider struct {
	version string
}

// TuringPiProviderModel describes the provider data model.
type TuringPiProviderModel struct {
	Host        types.String `tfsdk:"host"`
	Username    types.String `tfsdk:"username"`
	Password    types.String `tfsdk:"password"`
	SSHUser     types.String `tfsdk:"ssh_user"`
	SSHPassword types.String `tfsdk:"ssh_password"`
	SSHPort     types.Int64  `tfsdk:"ssh_port"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TuringPiProvider{
			version: version,
		}
	}
}

func (p *TuringPiProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "turingpi"
	resp.Version = p.version
}

func (p *TuringPiProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provider for managing Turing Pi 2 BMC resources including node flashing, power management, and USB configuration.",
		MarkdownDescription: `
The Turing Pi provider allows you to manage Turing Pi 2 cluster board resources.

## Features

- **Flash OS images** to compute modules (supports URL download with auto-decompression)
- **Power management** for individual nodes
- **USB configuration** for node connectivity
- **BMC information** retrieval

## Example Usage

` + "```hcl" + `
provider "turingpi" {
  host     = "192.168.1.90"
  username = "root"
  password = "turing"
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description:         "BMC hostname or IP address. Can also be set via TURINGPI_HOST environment variable.",
				MarkdownDescription: "BMC hostname or IP address. Can also be set via `TURINGPI_HOST` environment variable.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				Description:         "BMC username for authentication. Can also be set via TURINGPI_USERNAME environment variable. Default: root",
				MarkdownDescription: "BMC username for authentication. Can also be set via `TURINGPI_USERNAME` environment variable. Default: `root`",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				Description:         "BMC password for authentication. Can also be set via TURINGPI_PASSWORD environment variable. Default: turing",
				MarkdownDescription: "BMC password for authentication. Can also be set via `TURINGPI_PASSWORD` environment variable. Default: `turing`",
				Optional:            true,
				Sensitive:           true,
			},
			"ssh_user": schema.StringAttribute{
				Description:         "SSH username for BMC file operations (used for BMC caching). Defaults to username if not set.",
				MarkdownDescription: "SSH username for BMC file operations (used for BMC caching). Defaults to `username` if not set.",
				Optional:            true,
			},
			"ssh_password": schema.StringAttribute{
				Description:         "SSH password for BMC file operations (used for BMC caching). Defaults to password if not set.",
				MarkdownDescription: "SSH password for BMC file operations (used for BMC caching). Defaults to `password` if not set.",
				Optional:            true,
				Sensitive:           true,
			},
			"ssh_port": schema.Int64Attribute{
				Description:         "SSH port for BMC file operations. Default: 22",
				MarkdownDescription: "SSH port for BMC file operations. Default: `22`",
				Optional:            true,
			},
		},
	}
}

func (p *TuringPiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config TuringPiProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Get host from config or environment
	host := config.Host.ValueString()
	if host == "" {
		host = os.Getenv("TURINGPI_HOST")
	}
	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Turing Pi BMC Host",
			"The provider cannot create the Turing Pi client as there is a missing or empty value for the BMC host. "+
				"Set the host value in the configuration or use the TURINGPI_HOST environment variable.",
		)
	}

	// Get username from config or environment, default to "root"
	username := config.Username.ValueString()
	if username == "" {
		username = os.Getenv("TURINGPI_USERNAME")
	}
	if username == "" {
		username = "root"
	}

	// Get password from config or environment, default to "turing"
	password := config.Password.ValueString()
	if password == "" {
		password = os.Getenv("TURINGPI_PASSWORD")
	}
	if password == "" {
		password = "turing"
	}

	// Get SSH credentials, defaulting to BMC credentials
	sshUser := config.SSHUser.ValueString()
	if sshUser == "" {
		sshUser = username
	}

	sshPassword := config.SSHPassword.ValueString()
	if sshPassword == "" {
		sshPassword = password
	}

	sshPort := int(config.SSHPort.ValueInt64())
	if sshPort == 0 {
		sshPort = 22
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create client wrapper
	clientWrapper, err := client.NewClient(host, username, password, sshUser, sshPassword, sshPort)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Turing Pi Client",
			"An unexpected error occurred when creating the Turing Pi client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return
	}

	// Make the client available to resources and data sources
	resp.DataSourceData = clientWrapper
	resp.ResourceData = clientWrapper
}

func (p *TuringPiProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		node_flash.NewNodeFlashResource,
		node_power.NewNodePowerResource,
		node_usb.NewNodeUsbResource,
	}
}

func (p *TuringPiProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		info.NewInfoDataSource,
		power_status.NewPowerStatusDataSource,
		usb_status.NewUsbStatusDataSource,
	}
}
