// Copyright (c) David Roman
// SPDX-License-Identifier: MPL-2.0

package node_power

import (
	"context"
	"fmt"

	"github.com/davidroman0O/terraform-provider-turingpi/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &NodePowerResource{}
var _ resource.ResourceWithImportState = &NodePowerResource{}

func NewNodePowerResource() resource.Resource {
	return &NodePowerResource{}
}

// NodePowerResource defines the resource implementation.
type NodePowerResource struct {
	client *client.Client
}

// NodePowerResourceModel describes the resource data model.
type NodePowerResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Node    types.Int64  `tfsdk:"node"`
	PowerOn types.Bool   `tfsdk:"power_on"`
}

func (r *NodePowerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_power"
}

func (r *NodePowerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages the power state of a Turing Pi 2 node.",
		MarkdownDescription: "Manages the power state of a Turing Pi 2 node. This resource is idempotent - it will only change the power state if it differs from the desired state.",
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
			"power_on": schema.BoolAttribute{
				Description:         "Desired power state. true = powered on, false = powered off.",
				MarkdownDescription: "Desired power state. `true` = powered on, `false` = powered off.",
				Required:            true,
			},
		},
	}
}

func (r *NodePowerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NodePowerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NodePowerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	node := int(plan.Node.ValueInt64())
	desiredState := plan.PowerOn.ValueBool()

	// Check current state for idempotency
	status, err := r.client.PowerStatus()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Power Status",
			fmt.Sprintf("Could not read power status: %s", err.Error()),
		)
		return
	}

	currentState := status[node]

	// Only change if different (idempotent)
	if currentState != desiredState {
		if desiredState {
			err = r.client.PowerOn(node)
		} else {
			err = r.client.PowerOff(node)
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Power State",
				fmt.Sprintf("Could not set power state for node %d: %s", node, err.Error()),
			)
			return
		}
	}

	plan.ID = types.StringValue(fmt.Sprintf("node-%d-power", node))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NodePowerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NodePowerResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	node := int(state.Node.ValueInt64())

	status, err := r.client.PowerStatus()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Power Status",
			fmt.Sprintf("Could not read power status: %s", err.Error()),
		)
		return
	}

	state.PowerOn = types.BoolValue(status[node])

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NodePowerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NodePowerResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	node := int(plan.Node.ValueInt64())
	desiredState := plan.PowerOn.ValueBool()

	// Check current state for idempotency
	status, err := r.client.PowerStatus()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Power Status",
			fmt.Sprintf("Could not read power status: %s", err.Error()),
		)
		return
	}

	currentState := status[node]

	// Only change if different
	if currentState != desiredState {
		if desiredState {
			err = r.client.PowerOn(node)
		} else {
			err = r.client.PowerOff(node)
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Power State",
				fmt.Sprintf("Could not set power state for node %d: %s", node, err.Error()),
			)
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NodePowerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Power resources don't need cleanup - just remove from state
	// Optionally, we could power off the node on delete, but that might be unexpected behavior
}

func (r *NodePowerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
