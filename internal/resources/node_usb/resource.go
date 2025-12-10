// Copyright (c) David Roman
// SPDX-License-Identifier: MPL-2.0

package node_usb

import (
	"context"
	"fmt"

	"github.com/davidroman0O/terraform-provider-turingpi/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &NodeUsbResource{}
var _ resource.ResourceWithImportState = &NodeUsbResource{}

func NewNodeUsbResource() resource.Resource {
	return &NodeUsbResource{}
}

// NodeUsbResource defines the resource implementation.
type NodeUsbResource struct {
	client *client.Client
}

// NodeUsbResourceModel describes the resource data model.
type NodeUsbResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Node types.Int64  `tfsdk:"node"`
	Mode types.String `tfsdk:"mode"`
	BMC  types.Bool   `tfsdk:"bmc"`
}

func (r *NodeUsbResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_usb"
}

func (r *NodeUsbResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages the USB configuration of a Turing Pi 2 node.",
		MarkdownDescription: "Manages the USB configuration of a Turing Pi 2 node including the USB mode and routing.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Resource identifier.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"node": schema.Int64Attribute{
				Description:         "Node number (1-4).",
				MarkdownDescription: "Node number (`1`-`4`).",
				Required:            true,
				Validators: []validator.Int64{
					int64validator.Between(1, 4),
				},
			},
			"mode": schema.StringAttribute{
				Description:         "USB mode: host, device, or flash.",
				MarkdownDescription: "USB mode: `host`, `device`, or `flash`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("host", "device", "flash"),
				},
			},
			"bmc": schema.BoolAttribute{
				Description:         "Route USB through BMC. When false, routes through USB-A connector. Default: false.",
				MarkdownDescription: "Route USB through BMC. When `false`, routes through USB-A connector. Default: `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *NodeUsbResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *NodeUsbResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NodeUsbResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	node := int(plan.Node.ValueInt64())
	mode := plan.Mode.ValueString()
	bmc := plan.BMC.ValueBool()

	err := r.setUsbMode(node, mode, bmc)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Setting USB Mode",
			fmt.Sprintf("Could not set USB mode for node %d: %s", node, err.Error()),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("node-%d-usb", node))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NodeUsbResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NodeUsbResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	status, err := r.client.UsbGetStatus()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading USB Status",
			fmt.Sprintf("Could not read USB status: %s", err.Error()),
		)
		return
	}

	// Map the status mode to our lowercase mode values
	mode := mapModeToLower(status.Mode)
	state.Mode = types.StringValue(mode)

	// Map route to bmc boolean
	state.BMC = types.BoolValue(status.Route == "BMC")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NodeUsbResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NodeUsbResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	node := int(plan.Node.ValueInt64())
	mode := plan.Mode.ValueString()
	bmc := plan.BMC.ValueBool()

	err := r.setUsbMode(node, mode, bmc)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Setting USB Mode",
			fmt.Sprintf("Could not set USB mode for node %d: %s", node, err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NodeUsbResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// USB configuration doesn't need cleanup - just remove from state
}

func (r *NodeUsbResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// setUsbMode sets the USB mode for the specified node.
func (r *NodeUsbResource) setUsbMode(node int, mode string, bmc bool) error {
	switch mode {
	case "host":
		return r.client.UsbSetHost(node, bmc)
	case "device":
		return r.client.UsbSetDevice(node, bmc)
	case "flash":
		return r.client.UsbSetFlash(node, bmc)
	default:
		return fmt.Errorf("unknown USB mode: %s", mode)
	}
}

// mapModeToLower converts the API mode response to lowercase.
func mapModeToLower(mode string) string {
	switch mode {
	case "Host":
		return "host"
	case "Device":
		return "device"
	case "Flash":
		return "flash"
	default:
		return mode
	}
}
