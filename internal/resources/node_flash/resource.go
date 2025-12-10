// Copyright (c) David Roman
// SPDX-License-Identifier: MPL-2.0

package node_flash

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/davidroman0O/terraform-provider-turingpi/internal/client"
	tpi "github.com/davidroman0O/tpi/client"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Default timeouts for flash operations.
const (
	defaultCreateTimeout = 3 * time.Hour
	defaultUpdateTimeout = 3 * time.Hour
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &NodeFlashResource{}
var _ resource.ResourceWithImportState = &NodeFlashResource{}

func NewNodeFlashResource() resource.Resource {
	return &NodeFlashResource{}
}

// NodeFlashResource defines the resource implementation.
type NodeFlashResource struct {
	client *client.Client
}

// NodeFlashResourceModel describes the resource data model.
type NodeFlashResourceModel struct {
	ID          types.String   `tfsdk:"id"`
	Node        types.Int64    `tfsdk:"node"`
	ImageURL    types.String   `tfsdk:"image_url"`
	ImagePath   types.String   `tfsdk:"image_path"`
	SHA256      types.String   `tfsdk:"sha256"`
	Cache       types.String   `tfsdk:"cache"`
	SkipCRC     types.Bool     `tfsdk:"skip_crc"`
	FlashStatus types.String   `tfsdk:"flash_status"`
	LastFlashed types.String   `tfsdk:"last_flashed"`
	Timeouts    timeouts.Value `tfsdk:"timeouts"`
}

func (r *NodeFlashResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_node_flash"
}

func (r *NodeFlashResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Flashes an OS image to a Turing Pi 2 node.",
		MarkdownDescription: `Flashes an OS image to a Turing Pi 2 node.

This resource supports downloading images from URLs with automatic decompression for .xz, .gz, and .zip formats.
Images can be cached locally or on the BMC to speed up subsequent flashes.

## Example Usage

` + "```hcl" + `
resource "turingpi_node_flash" "ubuntu" {
  node      = 1
  image_url = "https://cdimage.ubuntu.com/releases/24.04/release/ubuntu-24.04-preinstalled-server-arm64+raspi.img.xz"
  cache     = "bmc"

  timeouts {
    create = "3h"
  }
}
` + "```" + `
`,
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
			"image_url": schema.StringAttribute{
				Description:         "URL to download the OS image from. Supports .xz, .gz, and .zip compression.",
				MarkdownDescription: "URL to download the OS image from. Supports `.xz`, `.gz`, and `.zip` compression.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("image_url"),
						path.MatchRoot("image_path"),
					),
				},
			},
			"image_path": schema.StringAttribute{
				Description: "Local file path to the OS image.",
				Optional:    true,
			},
			"sha256": schema.StringAttribute{
				Description:         "SHA256 checksum for image verification and cache key. If not provided, it will be calculated automatically.",
				MarkdownDescription: "SHA256 checksum for image verification and cache key. If not provided, it will be calculated automatically.",
				Optional:            true,
				Computed:            true,
			},
			"cache": schema.StringAttribute{
				Description:         "Cache strategy: 'local' (local filesystem), 'bmc' (BMC filesystem via SFTP), or 'none' (no caching). Default: 'none'.",
				MarkdownDescription: "Cache strategy: `local` (local filesystem), `bmc` (BMC filesystem via SFTP), or `none` (no caching). Default: `none`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("none"),
				Validators: []validator.String{
					stringvalidator.OneOf("local", "bmc", "none"),
				},
			},
			"skip_crc": schema.BoolAttribute{
				Description:         "Skip CRC verification during flash. Default: false.",
				MarkdownDescription: "Skip CRC verification during flash. Default: `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"flash_status": schema.StringAttribute{
				Description: "Status of the last flash operation (success, failed, or pending).",
				Computed:    true,
			},
			"last_flashed": schema.StringAttribute{
				Description: "Timestamp of the last successful flash in RFC3339 format.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
			}),
		},
	}
}

func (r *NodeFlashResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NodeFlashResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NodeFlashResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get timeout (default: 3 hours)
	createTimeout, diags := plan.Timeouts.Create(ctx, defaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	tflog.Info(ctx, "Starting flash operation", map[string]interface{}{
		"node":  plan.Node.ValueInt64(),
		"cache": plan.Cache.ValueString(),
	})

	// Execute flash
	result, err := r.executeFlash(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Flash Operation Failed",
			fmt.Sprintf("Failed to flash node %d: %s", plan.Node.ValueInt64(), err.Error()),
		)
		plan.FlashStatus = types.StringValue("failed")
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	// Update model with results
	node := plan.Node.ValueInt64()
	plan.ID = types.StringValue(fmt.Sprintf("node-%d-flash-%s", node, result.SHA256[:8]))
	plan.SHA256 = types.StringValue(result.SHA256)
	plan.FlashStatus = types.StringValue("success")
	plan.LastFlashed = types.StringValue(time.Now().UTC().Format(time.RFC3339))

	tflog.Info(ctx, "Flash operation completed successfully", map[string]interface{}{
		"node":   node,
		"sha256": result.SHA256[:16] + "...",
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NodeFlashResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NodeFlashResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Flash resources are stateless on the BMC side - there's no API to query
	// "what image is currently flashed". We just preserve the state as-is.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NodeFlashResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NodeFlashResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get timeout (default: 3 hours)
	updateTimeout, diags := plan.Timeouts.Update(ctx, defaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	tflog.Info(ctx, "Re-flashing node due to configuration change", map[string]interface{}{
		"node": plan.Node.ValueInt64(),
	})

	// Re-flash on changes to image_url, image_path, or sha256
	result, err := r.executeFlash(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Flash Update Failed",
			fmt.Sprintf("Failed to re-flash node %d: %s", plan.Node.ValueInt64(), err.Error()),
		)
		plan.FlashStatus = types.StringValue("failed")
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	// Update model with results
	plan.SHA256 = types.StringValue(result.SHA256)
	plan.FlashStatus = types.StringValue("success")
	plan.LastFlashed = types.StringValue(time.Now().UTC().Format(time.RFC3339))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NodeFlashResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Flash operations are not reversible - just remove from state
	tflog.Info(ctx, "Removing flash resource from state (node content is not affected)")
}

func (r *NodeFlashResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// FlashResult contains the result of a flash operation.
type FlashResult struct {
	SHA256 string
}

// executeFlash handles the actual flash operation with caching.
func (r *NodeFlashResource) executeFlash(ctx context.Context, plan *NodeFlashResourceModel) (*FlashResult, error) {
	node := int(plan.Node.ValueInt64())
	cacheLocation := plan.Cache.ValueString()

	var imagePath string
	var sha256 string
	var tempFile string // Track temp file for cleanup

	// Initialize cache
	cache, err := client.NewImageCache(r.client)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	// Handle image source
	if !plan.ImageURL.IsNull() && plan.ImageURL.ValueString() != "" {
		// Download from URL
		tflog.Info(ctx, "Downloading image from URL", map[string]interface{}{
			"url": plan.ImageURL.ValueString(),
		})

		// Check if expected SHA256 is provided
		var expectedSHA256 string
		if !plan.SHA256.IsNull() && plan.SHA256.ValueString() != "" {
			expectedSHA256 = plan.SHA256.ValueString()

			// Check cache first
			cachedPath, err := cache.GetCachedImagePath(expectedSHA256, cacheLocation)
			if err != nil {
				tflog.Warn(ctx, "Failed to check cache", map[string]interface{}{
					"error": err.Error(),
				})
			} else if cachedPath != "" {
				tflog.Info(ctx, "Using cached image", map[string]interface{}{
					"path": cachedPath,
				})
				imagePath = cachedPath
				sha256 = expectedSHA256
			}
		}

		// Download if not cached
		if imagePath == "" {
			result, err := client.DownloadImage(ctx, plan.ImageURL.ValueString(), &client.DownloadOptions{
				ExpectedSHA256: expectedSHA256,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to download image: %w", err)
			}
			imagePath = result.Path
			sha256 = result.SHA256
			tempFile = result.Path // Mark for cleanup later

			// Cache the downloaded image if caching is enabled
			if cacheLocation != client.CacheLocationNone {
				cachedPath, err := cache.CacheImage(imagePath, sha256, cacheLocation)
				if err != nil {
					tflog.Warn(ctx, "Failed to cache image", map[string]interface{}{
						"error": err.Error(),
					})
				} else {
					tflog.Info(ctx, "Image cached", map[string]interface{}{
						"path": cachedPath,
					})
				}
			}
		}
	} else {
		// Use local file
		imagePath = plan.ImagePath.ValueString()

		// Calculate SHA256 if not provided
		if !plan.SHA256.IsNull() && plan.SHA256.ValueString() != "" {
			sha256 = plan.SHA256.ValueString()
		} else {
			calculatedSHA256, err := client.CalculateFileSHA256(imagePath)
			if err != nil {
				return nil, fmt.Errorf("failed to calculate SHA256: %w", err)
			}
			sha256 = calculatedSHA256
		}

		// Cache the local file if caching is enabled
		if cacheLocation != client.CacheLocationNone {
			cachedPath, err := cache.CacheImage(imagePath, sha256, cacheLocation)
			if err != nil {
				tflog.Warn(ctx, "Failed to cache image", map[string]interface{}{
					"error": err.Error(),
				})
			} else {
				imagePath = cachedPath
				tflog.Info(ctx, "Image cached", map[string]interface{}{
					"path": cachedPath,
				})
			}
		}
	}

	// Cleanup temp file if we created one and it's not the cache path
	defer func() {
		if tempFile != "" && tempFile != imagePath {
			os.Remove(tempFile)
		}
	}()

	// Perform flash operation
	tflog.Info(ctx, "Starting flash to node", map[string]interface{}{
		"node":   node,
		"image":  imagePath,
		"sha256": sha256[:16] + "...",
	})

	// Use FlashNodeLocal if the image is on BMC
	if cacheLocation == client.CacheLocationBMC {
		err = r.client.FlashNodeLocal(node, imagePath)
	} else {
		opts := &tpi.FlashOptions{
			ImagePath: imagePath,
			SHA256:    sha256,
			SkipCRC:   plan.SkipCRC.ValueBool(),
		}
		err = r.client.FlashNode(node, opts)
	}

	if err != nil {
		return nil, fmt.Errorf("flash operation failed: %w", err)
	}

	return &FlashResult{
		SHA256: sha256,
	}, nil
}
