package laneful

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Domain represents a sending domain and its verification state.
type Domain struct {
	Domain             string `json:"domain"`
	Tracking           string `json:"tracking"`
	ReturnPath         string `json:"return_path"`
	Verified           bool   `json:"verified"`
	TrackingVerified   bool   `json:"tracking_verified"`
	ReturnPathVerified bool   `json:"return_path_verified"`
	Dkim1Verified      bool   `json:"dkim1_verified"`
	Dkim2Verified      bool   `json:"dkim2_verified"`
	// RequireTLS reports whether all mail from this domain must be sent over TLS.
	RequireTLS bool `json:"require_tls"`
	// EmailTrackID is the track used when sending from this domain. Empty means
	// the domain falls back to the API key's default track.
	EmailTrackID string `json:"email_track_id"`
}

// CreateDomainRequest is the request body for creating a domain.
type CreateDomainRequest struct {
	Domain     string `json:"domain"`
	Tracking   string `json:"tracking"`
	ReturnPath string `json:"return_path"`
	// RequireTLS, when true, requires all mail from this domain to be sent over TLS.
	RequireTLS bool `json:"require_tls,omitempty"`
	// EmailTrackID optionally pins the domain to a specific track. Empty uses the
	// API key's default track.
	EmailTrackID string `json:"email_track_id,omitempty"`
}

// UpdateDomainRequest updates a domain's mutable settings. Currently only the
// email track can be changed after creation (without deleting and recreating
// the domain, which would require re-verifying DNS).
type UpdateDomainRequest struct {
	// EmailTrackID sets the domain's email track. Point it at a track ID to set
	// the track, or at an empty string ("") to clear it (the domain then falls
	// back to the default track). Leave it nil to leave the current track
	// unchanged.
	EmailTrackID *string `json:"email_track_id,omitempty"`
}

// ListDomainsParams holds the optional filters for ListDomains. The zero value
// requests the first page with the server default page size.
type ListDomainsParams struct {
	// Cursor is the pagination cursor from a previous page.
	Cursor string
	// Limit is the maximum number of domains to return (default 50, max 1000).
	Limit int
	// FilterDomain filters results by hostname (maps to the filter[domain] query
	// parameter).
	FilterDomain string
}

// DomainsPagination holds pagination details for a domains listing.
type DomainsPagination struct {
	// NextCursor fetches the next page. Nil when there are no more results.
	NextCursor *string `json:"next_cursor,omitempty"`
}

// ListDomainsResponse is one page of domains.
type ListDomainsResponse struct {
	Domains    []Domain           `json:"domains"`
	Pagination *DomainsPagination `json:"pagination,omitempty"`
}

// SuccessResponse is a generic success message returned by mutating endpoints
// that do not return a resource body (for example, delete).
type SuccessResponse struct {
	Message string `json:"message"`
}

// ListDomains retrieves a paginated list of domains for a workspace.
func (c *LanefulClient) ListDomains(ctx context.Context, workspaceID int64, params *ListDomainsParams) (*ListDomainsResponse, error) {
	endpoint := fmt.Sprintf("%s/v1/workspaces/%d/domains", c.baseURL, workspaceID)
	if params != nil {
		q := url.Values{}
		if params.Cursor != "" {
			q.Set("cursor", params.Cursor)
		}
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if params.FilterDomain != "" {
			q.Set("filter[domain]", params.FilterDomain)
		}
		if encoded := q.Encode(); encoded != "" {
			endpoint += "?" + encoded
		}
	}

	var result ListDomainsResponse
	if err := c.doJSONRequest(ctx, http.MethodGet, endpoint, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetDomain retrieves a single domain by name.
func (c *LanefulClient) GetDomain(ctx context.Context, workspaceID int64, domain string) (*Domain, error) {
	endpoint := fmt.Sprintf("%s/v1/workspaces/%d/domains/%s", c.baseURL, workspaceID, url.PathEscape(domain))
	var result Domain
	if err := c.doJSONRequest(ctx, http.MethodGet, endpoint, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateDomain creates a new domain in the workspace.
func (c *LanefulClient) CreateDomain(ctx context.Context, workspaceID int64, request *CreateDomainRequest) (*Domain, error) {
	endpoint := fmt.Sprintf("%s/v1/workspaces/%d/domains", c.baseURL, workspaceID)
	var result Domain
	if err := c.doJSONRequest(ctx, http.MethodPost, endpoint, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UpdateDomain updates a domain's mutable settings (currently the email track).
// See UpdateDomainRequest for the set/clear/leave-unchanged semantics.
func (c *LanefulClient) UpdateDomain(ctx context.Context, workspaceID int64, domain string, request *UpdateDomainRequest) (*Domain, error) {
	endpoint := fmt.Sprintf("%s/v1/workspaces/%d/domains/%s", c.baseURL, workspaceID, url.PathEscape(domain))
	var result Domain
	if err := c.doJSONRequest(ctx, http.MethodPatch, endpoint, request, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// VerifyDomain triggers DNS verification for a domain and returns its updated
// verification state.
func (c *LanefulClient) VerifyDomain(ctx context.Context, workspaceID int64, domain string) (*Domain, error) {
	endpoint := fmt.Sprintf("%s/v1/workspaces/%d/domains/%s/verify", c.baseURL, workspaceID, url.PathEscape(domain))
	var result Domain
	if err := c.doJSONRequest(ctx, http.MethodPost, endpoint, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteDomain removes a domain from the workspace.
func (c *LanefulClient) DeleteDomain(ctx context.Context, workspaceID int64, domain string) (*SuccessResponse, error) {
	endpoint := fmt.Sprintf("%s/v1/workspaces/%d/domains/%s", c.baseURL, workspaceID, url.PathEscape(domain))
	var result SuccessResponse
	if err := c.doJSONRequest(ctx, http.MethodDelete, endpoint, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
