package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nickcecere/terraform-provider-openwebui/internal/client"
)

type accessGrantModel struct {
	Permission    types.String `tfsdk:"permission"`
	PrincipalID   types.String `tfsdk:"principal_id"`
	PrincipalType types.String `tfsdk:"principal_type"`
}

var accessGrantAttrTypes = map[string]attr.Type{
	"permission":     types.StringType,
	"principal_id":   types.StringType,
	"principal_type": types.StringType,
}

var accessGrantObjectType = types.ObjectType{AttrTypes: accessGrantAttrTypes}

// groupResolver fetches all groups once and provides efficient bidirectional ID↔name lookup.
type groupResolver struct {
	byName    map[string]string // lower(name) → canonical ID
	byID      map[string]string // lower(ID) → canonical ID
	idToName  map[string]string // canonical ID → name
	apiClient *client.Client
}

func newGroupResolver(ctx context.Context, apiClient *client.Client, diags *diag.Diagnostics) *groupResolver {
	groups, err := apiClient.ListGroups(ctx)
	if err != nil {
		diags.AddError("Unable to list groups",
			fmt.Sprintf("Failed to retrieve groups from Open WebUI: %v", err))
		return nil
	}

	r := &groupResolver{
		byName:    make(map[string]string, len(groups)),
		byID:      make(map[string]string, len(groups)),
		idToName:  make(map[string]string, len(groups)),
		apiClient: apiClient,
	}
	for _, g := range groups {
		r.byName[strings.ToLower(g.Name)] = g.ID
		r.byID[strings.ToLower(g.ID)] = g.ID
		r.idToName[g.ID] = g.Name
	}
	return r
}

func (r *groupResolver) toID(ctx context.Context, identifier string, attribute path.Path, diags *diag.Diagnostics) string {
	key := strings.ToLower(strings.TrimSpace(identifier))
	if id, ok := r.byName[key]; ok {
		return id
	}
	if id, ok := r.byID[key]; ok {
		return id
	}
	group, err := r.apiClient.GetGroup(ctx, identifier)
	if err != nil {
		diags.AddAttributeError(attribute, "Unknown group reference",
			fmt.Sprintf("No Open WebUI group was found for %q.", identifier))
		return ""
	}
	r.byName[strings.ToLower(group.Name)] = group.ID
	r.byID[strings.ToLower(group.ID)] = group.ID
	r.idToName[group.ID] = group.Name
	return group.ID
}

func (r *groupResolver) toName(id string) string {
	if name, ok := r.idToName[id]; ok {
		return name
	}
	return id
}

// userResolver caches user lookups to avoid repeated API calls for the same user.
type userResolver struct {
	idToLabel map[string]string // user ID → best human-readable label (email/username/name)
	labelToID map[string]string // lower(label) → user ID
	apiClient *client.Client
}

func newUserResolver(apiClient *client.Client) *userResolver {
	return &userResolver{
		idToLabel: make(map[string]string),
		labelToID: make(map[string]string),
		apiClient: apiClient,
	}
}

func (r *userResolver) toID(ctx context.Context, identifier string, attribute path.Path, diags *diag.Diagnostics) string {
	key := strings.ToLower(identifier)
	if id, ok := r.labelToID[key]; ok {
		return id
	}
	id, err := lookupUserID(ctx, r.apiClient, identifier)
	if err != nil {
		diags.AddAttributeError(attribute, "Unable to resolve user",
			fmt.Sprintf("Failed to find user %q: %v", identifier, err))
		return ""
	}
	r.labelToID[key] = id
	return id
}

func (r *userResolver) toLabel(ctx context.Context, id string) string {
	if label, ok := r.idToLabel[id]; ok {
		return label
	}
	user, err := r.apiClient.GetUser(ctx, id)
	if err != nil {
		return id
	}
	label := user.Email
	if label == "" && user.Username != nil && *user.Username != "" {
		label = *user.Username
	} else if label == "" {
		label = user.Name
	}
	if label == "" {
		label = id
	}
	r.idToLabel[id] = label
	return label
}

// expandAccessGrants resolves group names and user emails/usernames to IDs for the API payload.
func expandAccessGrants(ctx context.Context, apiClient *client.Client, list types.List, attribute path.Path, diags *diag.Diagnostics) []client.AccessGrant {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var models []accessGrantModel
	diags.Append(list.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil
	}

	var (
		gr        *groupResolver
		grFetched bool
		ur        *userResolver
	)

	result := make([]client.AccessGrant, 0, len(models))
	for _, m := range models {
		principalID := m.PrincipalID.ValueString()
		principalType := m.PrincipalType.ValueString()

		switch principalType {
		case "group":
			if !grFetched {
				grFetched = true
				gr = newGroupResolver(ctx, apiClient, diags)
			}
			if gr != nil {
				if resolved := gr.toID(ctx, principalID, attribute, diags); resolved != "" {
					principalID = resolved
				}
			}
		case "user":
			if principalID != "*" {
				if ur == nil {
					ur = newUserResolver(apiClient)
				}
				if resolved := ur.toID(ctx, principalID, attribute, diags); resolved != "" {
					principalID = resolved
				}
			}
		}

		result = append(result, client.AccessGrant{
			Permission:    m.Permission.ValueString(),
			PrincipalID:   principalID,
			PrincipalType: principalType,
		})
	}

	return result
}

// flattenAccessGrants converts API grants to Terraform state, resolving IDs to human-readable labels.
func flattenAccessGrants(ctx context.Context, apiClient *client.Client, grants []client.AccessGrant) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	var (
		gr        *groupResolver
		grFetched bool
		ur        *userResolver
	)

	models := make([]accessGrantModel, 0, len(grants))
	for _, g := range grants {
		principalLabel := g.PrincipalID

		switch g.PrincipalType {
		case "group":
			if !grFetched {
				grFetched = true
				gr = newGroupResolver(ctx, apiClient, &diags)
			}
			if gr != nil {
				principalLabel = gr.toName(g.PrincipalID)
			}
		case "user":
			if g.PrincipalID != "*" {
				if ur == nil {
					ur = newUserResolver(apiClient)
				}
				principalLabel = ur.toLabel(ctx, g.PrincipalID)
			}
		}

		models = append(models, accessGrantModel{
			Permission:    types.StringValue(g.Permission),
			PrincipalID:   types.StringValue(principalLabel),
			PrincipalType: types.StringValue(g.PrincipalType),
		})
	}

	list, listDiags := types.ListValueFrom(ctx, accessGrantObjectType, models)
	diags.Append(listDiags...)
	return list, diags
}
