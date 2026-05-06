package client

import (
	"context"
	"net/http"
	"net/url"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"encoding/json"
)

// ModelForm represents the payload for creating or updating models.
type ModelForm struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Meta          map[string]any `json:"meta"`
	Params        map[string]any `json:"params"`
	BaseModelID   *string        `json:"base_model_id,omitempty"`
	IsActive      *bool          `json:"is_active,omitempty"`
	AccessGrants  []AccessGrant  `json:"access_grants"`
}

// ModelResponse captures details returned by the model endpoints.
type ModelResponse struct {
	ID            string         `json:"id"`
	UserID        string         `json:"user_id"`
	Name          string         `json:"name"`
	Meta          map[string]any `json:"meta"`
	Params        map[string]any `json:"params"`
	BaseModelID   *string        `json:"base_model_id,omitempty"`
	IsActive      bool           `json:"is_active"`
	AccessGrants  []AccessGrant  `json:"access_grants"`
	CreatedAt     int64          `json:"created_at"`
	UpdatedAt     int64          `json:"updated_at"`
}

// CreateModel registers a new model.
func (c *Client) CreateModel(ctx context.Context, form ModelForm) (*ModelResponse, error) {
	var resp ModelResponse
	if err := c.do(ctx, http.MethodPost, "models/create", nil, form, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetModel obtains model details by identifier.
func (c *Client) GetModel(ctx context.Context, id string) (*ModelResponse, error) {
	var resp ModelResponse
	query := url.Values{"id": []string{id}}
	if err := c.do(ctx, http.MethodGet, "models/model", query, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// UpdateModel updates a model by identifier.
func (c *Client) UpdateModel(ctx context.Context, id string, form ModelForm) (*ModelResponse, error) {
	var resp ModelResponse
	data, err0 := json.Marshal(form)
	if err0 != nil {
		tflog.Info(ctx, "Got error while serializing json")
	} else {
		tflog.Info(ctx, "JSON of model: ", map[string]any{"model": string(data)})
	}
	query := url.Values{"id": []string{id}}
	if err := c.do(ctx, http.MethodPost, "models/model/update", query, form, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// DeleteModel removes a model by identifier.
func (c *Client) DeleteModel(ctx context.Context, id string) error {
	query := url.Values{"id": []string{id}}
	return c.do(ctx, http.MethodPost, "models/model/delete", query, nil, nil)
}
