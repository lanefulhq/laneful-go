package laneful

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// dateLayout is the wire format for calendar dates used by the analytics API.
const dateLayout = "2006-01-02"

// Date is a calendar date (no time component) serialized as YYYY-MM-DD.
// It embeds time.Time, so all the usual time.Time accessors are available;
// use the embedded Time field when you need the underlying value.
type Date struct {
	time.Time
}

// MarshalJSON renders the date as a YYYY-MM-DD string.
func (d Date) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Format(dateLayout))
}

// UnmarshalJSON parses a YYYY-MM-DD string. An empty string yields the zero date.
func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		d.Time = time.Time{}
		return nil
	}
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return fmt.Errorf("invalid date %q: %w", s, err)
	}
	d.Time = t
	return nil
}

// String returns the YYYY-MM-DD representation of the date.
func (d Date) String() string {
	return d.Format(dateLayout)
}

// SndsFilterResult is the Microsoft SNDS spam-filter verdict for a day.
type SndsFilterResult string

const (
	// SndsFilterUnknown indicates the verdict was not available for the day.
	SndsFilterUnknown SndsFilterResult = ""
	SndsFilterGreen   SndsFilterResult = "GREEN"
	SndsFilterYellow  SndsFilterResult = "YELLOW"
	SndsFilterRed     SndsFilterResult = "RED"
)

// DomainSpamRatioRadar reports a sending domain whose spam complaint ratio
// reached a critical level (0.1% or more of delivered messages) at a specific
// mailbox provider on a specific day.
type DomainSpamRatioRadar struct {
	// WorkspaceID is the workspace the domain belongs to.
	WorkspaceID int64 `json:"workspace_id"`
	// Domain is the sending domain the radar entry refers to.
	Domain string `json:"domain"`
	// Esp is the mailbox provider the spam complaints were reported by
	// (e.g. "Gmail", "Outlook", "Yahoo"), "UnsetProvider" when unknown.
	Esp string `json:"esp"`
	// SpamRatio is spam complaints as a percentage of delivered messages for
	// the day (0.265 = 0.265%). Only values of at least 0.1 (0.1%) are reported.
	SpamRatio float64 `json:"spam_ratio"`
	// Date is the UTC day the radar entry refers to.
	Date Date `json:"date"`
}

// GooglePostmasterSpamReport is a daily spam-rate report for a domain from
// Google Postmaster Tools.
type GooglePostmasterSpamReport struct {
	// WorkspaceID is the workspace the domain belongs to.
	WorkspaceID int64 `json:"workspace_id"`
	// Domain is the sending domain the report is for.
	Domain string `json:"domain"`
	// Date is the UTC day of the report.
	Date Date `json:"date"`
	// SpamRatio is the percentage of delivered mail that Gmail users marked as
	// spam (0.05 = 0.05%). Google recommends keeping this below 0.1%.
	SpamRatio float64 `json:"spam_ratio"`
}

// SndsReport is a daily report for a sending IP from Microsoft SNDS (Smart
// Network Data Services), covering Outlook.com traffic volume, spam filter
// verdict, complaint rate and spam trap hits.
type SndsReport struct {
	// IP is the sending IP address the report is for.
	IP string `json:"ip"`
	// Date is the UTC day of the report.
	Date Date `json:"date"`
	// RcptCommands is the number of RCPT commands seen from the IP.
	RcptCommands int64 `json:"rcpt_commands"`
	// DataCommands is the number of DATA commands seen from the IP.
	DataCommands int64 `json:"data_commands"`
	// MessageRecipients is the number of message recipients.
	MessageRecipients int64 `json:"message_recipients"`
	// FilterResult is the spam filter verdict for the day.
	FilterResult SndsFilterResult `json:"filter_result"`
	// ComplaintRate is the percentage of recipients that reported the mail as
	// junk (0.1 = 0.1%).
	ComplaintRate float64 `json:"complaint_rate"`
	// TrapHits is the number of messages sent to spam trap addresses.
	TrapHits int64 `json:"trap_hits"`
}

// ListDomainSpamRatioRadarParams holds the optional filters for
// ListDomainSpamRatioRadar. The zero value requests the default range (the
// current UTC day and the 7 days before it) across all workspaces.
type ListDomainSpamRatioRadarParams struct {
	// WorkspaceIDs filters results to specific workspaces. Empty means all
	// workspaces of the organization.
	WorkspaceIDs []int64
	// Domain filters results to a specific sending domain.
	Domain string
	// StartDate is the inclusive start of the range (UTC). Zero means 7 days ago.
	StartDate time.Time
	// EndDate is the inclusive end of the range (UTC). Zero means today.
	EndDate time.Time
	// Cursor is the pagination cursor from a previous page.
	Cursor string
	// Limit is the maximum number of results to return (default 50, max 200).
	Limit int
}

// ListDomainSpamRatioRadarResponse is one page of domain spam ratio radar entries.
type ListDomainSpamRatioRadarResponse struct {
	// Radar lists domains with concerning spam ratios, ordered by date (newest first).
	Radar []DomainSpamRatioRadar `json:"radar"`
	// NextCursor fetches the next page. Nil when there are no more results.
	NextCursor *string `json:"next_cursor,omitempty"`
}

// ListDomainSpamRatioRadar retrieves domains of the organization whose spam
// complaint ratio reached a critical level (0.1% or more of delivered messages)
// at a mailbox provider on a day within the requested range.
func (c *LanefulClient) ListDomainSpamRatioRadar(ctx context.Context, params *ListDomainSpamRatioRadarParams) (*ListDomainSpamRatioRadarResponse, error) {
	endpoint := c.baseURL + "/v1/analytics/radar/domain-spam-ratio"
	if params != nil {
		q := url.Values{}
		for _, id := range params.WorkspaceIDs {
			q.Add("workspace_ids", strconv.FormatInt(id, 10))
		}
		if params.Domain != "" {
			q.Set("domain", params.Domain)
		}
		if !params.StartDate.IsZero() {
			q.Set("start_date", params.StartDate.Format(dateLayout))
		}
		if !params.EndDate.IsZero() {
			q.Set("end_date", params.EndDate.Format(dateLayout))
		}
		if params.Cursor != "" {
			q.Set("cursor", params.Cursor)
		}
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if encoded := q.Encode(); encoded != "" {
			endpoint += "?" + encoded
		}
	}

	var result ListDomainSpamRatioRadarResponse
	if err := c.doJSONRequest(ctx, http.MethodGet, endpoint, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListGooglePostmasterSpamReportsParams holds the optional filters for
// ListGooglePostmasterSpamReports. The zero value returns reports for all
// domains across all workspaces of the organization.
type ListGooglePostmasterSpamReportsParams struct {
	// WorkspaceIDs filters results to domains in specific workspaces. Empty
	// means all workspaces of the organization.
	WorkspaceIDs []int64
	// Domain filters results to a specific sending domain.
	Domain string
	// StartDate returns reports on or after this date (UTC, inclusive).
	StartDate time.Time
	// EndDate returns reports on or before this date (UTC, inclusive).
	EndDate time.Time
	// Cursor is the pagination cursor from a previous page.
	Cursor string
	// Limit is the maximum number of reports to return (default 50, max 200).
	Limit int
}

// ListGooglePostmasterSpamReportsResponse is one page of Google Postmaster spam reports.
type ListGooglePostmasterSpamReportsResponse struct {
	// SpamReports lists daily spam-rate reports ordered by domain and date.
	SpamReports []GooglePostmasterSpamReport `json:"spam_reports"`
	// NextCursor fetches the next page. Nil when there are no more results.
	NextCursor *string `json:"next_cursor,omitempty"`
}

// ListGooglePostmasterSpamReports retrieves daily Google Postmaster Tools
// spam-rate reports for the domains in the organization.
func (c *LanefulClient) ListGooglePostmasterSpamReports(ctx context.Context, params *ListGooglePostmasterSpamReportsParams) (*ListGooglePostmasterSpamReportsResponse, error) {
	endpoint := c.baseURL + "/v1/analytics/google-postmaster/spam-reports"
	if params != nil {
		q := url.Values{}
		for _, id := range params.WorkspaceIDs {
			q.Add("workspace_ids", strconv.FormatInt(id, 10))
		}
		if params.Domain != "" {
			q.Set("domain", params.Domain)
		}
		if !params.StartDate.IsZero() {
			q.Set("start_date", params.StartDate.Format(dateLayout))
		}
		if !params.EndDate.IsZero() {
			q.Set("end_date", params.EndDate.Format(dateLayout))
		}
		if params.Cursor != "" {
			q.Set("cursor", params.Cursor)
		}
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if encoded := q.Encode(); encoded != "" {
			endpoint += "?" + encoded
		}
	}

	var result ListGooglePostmasterSpamReportsResponse
	if err := c.doJSONRequest(ctx, http.MethodGet, endpoint, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListSndsReportsParams holds the optional filters for ListSndsReports. The
// zero value returns reports for all sending IPs used by the organization's lanes.
type ListSndsReportsParams struct {
	// IP filters results to a specific sending IP address.
	IP string
	// StartDate returns reports on or after this date (UTC, inclusive).
	StartDate time.Time
	// EndDate returns reports on or before this date (UTC, inclusive).
	EndDate time.Time
	// Cursor is the pagination cursor from a previous page.
	Cursor string
	// Limit is the maximum number of reports to return (default 50, max 200).
	Limit int
}

// ListSndsReportsResponse is one page of Microsoft SNDS reports.
type ListSndsReportsResponse struct {
	// SndsReports lists daily SNDS reports ordered by IP and date.
	SndsReports []SndsReport `json:"snds_reports"`
	// NextCursor fetches the next page. Nil when there are no more results.
	NextCursor *string `json:"next_cursor,omitempty"`
}

// ListSndsReports retrieves daily Microsoft SNDS (Smart Network Data Services)
// reports for the sending IPs used by the organization's lanes.
func (c *LanefulClient) ListSndsReports(ctx context.Context, params *ListSndsReportsParams) (*ListSndsReportsResponse, error) {
	endpoint := c.baseURL + "/v1/analytics/microsoft-snds/reports"
	if params != nil {
		q := url.Values{}
		if params.IP != "" {
			q.Set("ip", params.IP)
		}
		if !params.StartDate.IsZero() {
			q.Set("start_date", params.StartDate.Format(dateLayout))
		}
		if !params.EndDate.IsZero() {
			q.Set("end_date", params.EndDate.Format(dateLayout))
		}
		if params.Cursor != "" {
			q.Set("cursor", params.Cursor)
		}
		if params.Limit > 0 {
			q.Set("limit", strconv.Itoa(params.Limit))
		}
		if encoded := q.Encode(); encoded != "" {
			endpoint += "?" + encoded
		}
	}

	var result ListSndsReportsResponse
	if err := c.doJSONRequest(ctx, http.MethodGet, endpoint, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
