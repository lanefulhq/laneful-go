package laneful

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// UnsubscribeGroup represents an unsubscribe group
type UnsubscribeGroup struct {
	UnsubscribeGroupID int64  `json:"unsubscribe_group_id"`
	Name               string `json:"name"`
	CreatedAt          int64  `json:"created_at"`
}

// CreateUnsubscribeGroupRequest represents the request body for creating an unsubscribe group
type CreateUnsubscribeGroupRequest struct {
	Name string `json:"name"`
}

// UpdateUnsubscribeGroupRequest represents the request body for updating an unsubscribe group
type UpdateUnsubscribeGroupRequest struct {
	Name string `json:"name"`
}

// UnsubscribeGroupResponse represents a single unsubscribe group response
type UnsubscribeGroupResponse struct {
	UnsubscribeGroup *UnsubscribeGroup `json:"unsubscribe_group,omitempty"`
}

// ListUnsubscribeGroupsParams represents query parameters for listing unsubscribe groups
type ListUnsubscribeGroupsParams struct {
	Cursor string
	Limit  int
	Search string
}

// ListUnsubscribeGroupsResponse represents the response for listing unsubscribe groups
type ListUnsubscribeGroupsResponse struct {
	UnsubscribeGroups []UnsubscribeGroup `json:"unsubscribe_groups"`
	NextCursor        *string            `json:"next_cursor,omitempty"`
}

// ListUnsubscribeGroups retrieves a paginated list of unsubscribe groups for a workspace
func (c *LanefulClient) ListUnsubscribeGroups(ctx context.Context, workspaceID int64, params *ListUnsubscribeGroupsParams) (*ListUnsubscribeGroupsResponse, error) {
	endpoint := fmt.Sprintf("%s/v1/workspaces/%d/unsubscribe-groups", c.baseURL, workspaceID)
	if params != nil {
		q := url.Values{}
		if params.Cursor != "" {
			q.Set("cursor", params.Cursor)
		}
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.Search != "" {
			q.Set("search", params.Search)
		}
		if encoded := q.Encode(); encoded != "" {
			endpoint += "?" + encoded
		}
	}

	var result ListUnsubscribeGroupsResponse
	if err := c.doJSONRequest(ctx, http.MethodGet, endpoint, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateUnsubscribeGroup creates a new unsubscribe group in the workspace
func (c *LanefulClient) CreateUnsubscribeGroup(ctx context.Context, workspaceID int64, request *CreateUnsubscribeGroupRequest) (*UnsubscribeGroupResponse, error) {
	endpoint := fmt.Sprintf("%s/v1/workspaces/%d/unsubscribe-groups", c.baseURL, workspaceID)
	var result UnsubscribeGroupResponse
	if err := c.doJSONRequest(ctx, http.MethodPost, endpoint, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateUnsubscribeGroup updates an existing unsubscribe group
func (c *LanefulClient) UpdateUnsubscribeGroup(ctx context.Context, workspaceID, unsubscribeGroupID int64, request *UpdateUnsubscribeGroupRequest) (*UnsubscribeGroupResponse, error) {
	endpoint := fmt.Sprintf("%s/v1/workspaces/%d/unsubscribe-groups/%d", c.baseURL, workspaceID, unsubscribeGroupID)
	var result UnsubscribeGroupResponse
	if err := c.doJSONRequest(ctx, http.MethodPatch, endpoint, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *LanefulClient) doJSONRequest(ctx context.Context, method, endpoint string, body, out interface{}) error {
	var reqBody *bytes.Buffer
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(data)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+c.authToken)
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp ApiErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("API error: status %d", resp.StatusCode)
		}
		return fmt.Errorf("API error: %s", errResp.Error)
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
}
