package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nickcecere/terraform-provider-openwebui/internal/client"
)

var _ datasource.DataSource = &toolDataSource{}
var _ datasource.DataSourceWithConfigure = &toolDataSource{}

// toolDataSource exposes tool details.
type toolDataSource struct {
	client *client.Client
}

// toolDataSourceModel maps data source inputs and outputs.
type toolDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Content      types.String `tfsdk:"content"`
	Description  types.String `tfsdk:"description"`
	ManifestJSON types.String `tfsdk:"manifest_json"`
	AccessGrants types.List   `tfsdk:"access_grants"`
	SpecsJSON    types.String `tfsdk:"specs_json"`
	UserID       types.String `tfsdk:"user_id"`
	CreatedAt    types.Int64  `tfsdk:"created_at"`
	UpdatedAt    types.Int64  `tfsdk:"updated_at"`
	WriteAccess  types.Bool   `tfsdk:"write_access"`
}

// NewToolDataSource constructs a new tool data source.
func NewToolDataSource() datasource.DataSource {
	return &toolDataSource{}
}

// Metadata sets the data source type name.
func (d *toolDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tool"
}

// Schema defines the tool data source schema.
func (d *toolDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Identifier of the tool to look up.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Display name for the tool.",
			},
			"content": schema.StringAttribute{
				Computed:    true,
				Description: "Source content for the tool.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Human-readable tool description.",
			},
			"manifest_json": schema.StringAttribute{
				Computed:    true,
				Description: "JSON manifest for the tool.",
			},
			"access_grants": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Access grants controlling who can read or write the tool.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"permission":     schema.StringAttribute{Computed: true, Description: "Permission level: \"read\" or \"write\"."},
						"principal_id":   schema.StringAttribute{Computed: true, Description: "Group name or user email/username."},
						"principal_type": schema.StringAttribute{Computed: true, Description: "Principal type: \"group\" or \"user\"."},
					},
				},
			},
			"specs_json": schema.StringAttribute{
				Computed:    true,
				Description: "Raw JSON specification returned by Open WebUI.",
			},
			"user_id": schema.StringAttribute{
				Computed:    true,
				Description: "Identifier of the user who owns the tool.",
			},
			"created_at": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp of tool creation.",
			},
			"updated_at": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp of the last tool update.",
			},
			"write_access": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the current user has write access.",
			},
		},
	}
}

// Configure assigns the API client.
func (d *toolDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	if client, ok := req.ProviderData.(*client.Client); ok {
		d.client = client
	}
}

// Read retrieves tool details.
func (d *toolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured API client", "Expected provider to configure the Open WebUI client before using the tool data source.")
		return
	}

	var config toolDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ID.IsUnknown() || config.ID.IsNull() || config.ID.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Missing tool identifier",
			"The id argument must be supplied to query an existing tool.",
		)
		return
	}

	access, err := d.client.GetTool(ctx, config.ID.ValueString())
	if err != nil {
		if err == client.ErrNotFound {
			resp.Diagnostics.AddAttributeError(
				path.Root("id"),
				"Tool not found",
				"No Open WebUI tool was found with the supplied id.",
			)
			return
		}
		resp.Diagnostics.AddError("Read tool failed", err.Error())
		return
	}

	content, specs, fetchDiags := fetchToolContent(ctx, d.client, access.ID)
	resp.Diagnostics.Append(fetchDiags...)

	state, diags := toolResponseToModel(ctx, d.client, access, content, specs, types.StringNull())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
