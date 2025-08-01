package laneful

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Address represents an email address with an optional name
type Address struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

// Attachment represents an email attachment
type Attachment struct {
	FileName    string `json:"file_name,omitempty"`
	Content     string `json:"content,omitempty"`
	ContentType string `json:"content_type"`
	InlineID    string `json:"inline_id,omitempty"`
}

// TrackingSettings controls email tracking and unsubscribe settings
type TrackingSettings struct {
	Opens              bool   `json:"opens,omitempty"`
	Clicks             bool   `json:"clicks,omitempty"`
	Unsubscribes       bool   `json:"unsubscribes,omitempty"`
	UnsubscribeGroupID *int64 `json:"unsubscribe_group_id,omitempty"`
}

// Email represents a single email to be sent
type Email struct {
	From         Address                `json:"from"`
	To           []Address              `json:"to,omitempty"`
	CC           []Address              `json:"cc,omitempty"`
	BCC          []Address              `json:"bcc,omitempty"`
	Subject      string                 `json:"subject,omitempty"`
	TextContent  string                 `json:"text_content,omitempty"`
	HTMLContent  string                 `json:"html_content,omitempty"`
	TemplateID   string                 `json:"template_id,omitempty"`
	TemplateData map[string]interface{} `json:"template_data,omitempty"`
	Attachments  []Attachment           `json:"attachments,omitempty"`
	Headers      map[string]string      `json:"headers,omitempty"`
	ReplyTo      *Address               `json:"reply_to,omitempty"`
	SendTime     int64                  `json:"send_time,omitempty"`
	WebhookData  map[string]string      `json:"webhook_data,omitempty"`
	Tag          string                 `json:"tag,omitempty"`
	Tracking     *TrackingSettings      `json:"tracking,omitempty"`
}

// EmailRequest represents the request body for sending emails
type EmailRequest struct {
	Emails []Email `json:"emails"`
}

// ApiResponse represents a successful API response
type ApiResponse struct {
	Status string `json:"status"`
}

// ApiErrorResponse represents an error API response
type ApiErrorResponse struct {
	Error string `json:"error"`
}

// LanefulClient handles communication with the email service
type LanefulClient struct {
	baseURL    string
	httpClient *http.Client
	authToken  string
}

// NewLanefulClient creates a new instance of LanefulClient
func NewLanefulClient(baseURL, authToken string) *LanefulClient {
	insecure := false
	return &LanefulClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: insecure,
				},
			},
		},
		authToken: authToken,
	}
}

// SendEmails sends one or more emails through the email service
func (c *LanefulClient) SendEmails(ctx context.Context, emails []Email) (*ApiResponse, error) {
	reqBody := EmailRequest{
		Emails: emails,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/email/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.authToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ApiErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("failed to decode error response: %w", err)
		}
		return nil, fmt.Errorf("API error: %s", errResp.Error)
	}

	var apiResp ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &apiResp, nil
}

// SendEmail is a convenience method to send a single email
func (c *LanefulClient) SendEmail(ctx context.Context, email Email) (*ApiResponse, error) {
	return c.SendEmails(ctx, []Email{email})
}
