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

const (
	version   = "1.3.0"
	userAgent = "laneful-go/" + version
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
	Opens                bool    `json:"opens"`
	Clicks               bool    `json:"clicks"`
	Unsubscribes         bool    `json:"unsubscribes"`
	UnsubscribeGroupID   *int64  `json:"unsubscribe_group_id,omitempty"`
	UnsubscribeGroupName *string `json:"unsubscribe_group_name,omitempty"`
}

// MailSettings represents additional settings for the email
type MailSettings struct {
	SandboxMode      bool `json:"sandbox_mode,omitempty"`       // When enabled, messages are not persisted or sent (sandbox only)
	ReturnMessageIds bool `json:"return_message_ids,omitempty"` // When enabled, the API response will include message IDs for each email sent
	// IdempotencyKey protects against duplicate sends (1 to 36 bytes; empty means unset).
	// Laneful accepts a key once per organization for one hour: a repeated request with
	// the same key is answered with status "accepted" (and no message IDs) and sends nothing.
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}

// Email represents a single email to be sent
type Email struct {
	FromHeader   Address                `json:"from_header,omitempty"`
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
	// WebhookData is returned in this email's webhook events: at most 20 keys, keys up to
	// 50 bytes and values up to 100 bytes (UTF-8). SendEmailsWithOptions checks these limits.
	WebhookData map[string]string `json:"webhook_data,omitempty"`
	Tag         string            `json:"tag,omitempty"`
	Tracking    *TrackingSettings `json:"tracking,omitempty"`
}

// EmailRequest represents the request body for sending emails
type EmailRequest struct {
	Emails       []Email       `json:"emails"`
	MailSettings *MailSettings `json:"mail_settings,omitempty"`
}

// ApiResponse represents a successful API response
type ApiResponse struct {
	Status     string   `json:"status"`
	MessageIds []string `json:"message_ids,omitempty"`
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
func (c *LanefulClient) SendEmailsWithOptions(ctx context.Context, request *EmailRequest) (*ApiResponse, error) {
	if request.MailSettings != nil && len(request.MailSettings.IdempotencyKey) > 36 {
		return nil, fmt.Errorf("idempotency_key must contain between 1 and 36 bytes")
	}
	for _, email := range request.Emails {
		if err := validateWebhookData(email.WebhookData); err != nil {
			return nil, err
		}
	}

	reqBody := *request

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
	req.Header.Set("User-Agent", userAgent)

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

// validateWebhookData applies the send API's webhook_data limits: at most 20 keys,
// keys up to 50 bytes and values up to 100 bytes (UTF-8, so "é" counts as 2).
func validateWebhookData(data map[string]string) error {
	if len(data) > 20 {
		return fmt.Errorf("webhook data cannot have more than 20 keys")
	}
	for key, value := range data {
		if len(key) > 50 {
			return fmt.Errorf("webhook data keys cannot exceed 50 bytes (UTF-8)")
		}
		if len(value) > 100 {
			return fmt.Errorf("webhook data values cannot exceed 100 bytes (UTF-8)")
		}
	}
	return nil
}

func (c *LanefulClient) SendEmails(ctx context.Context, emails []Email) (*ApiResponse, error) {
	return c.SendEmailsWithOptions(ctx, &EmailRequest{
		Emails: emails,
	})
}

// SendEmail is a convenience method to send a single email
func (c *LanefulClient) SendEmail(ctx context.Context, email Email) (*ApiResponse, error) {
	return c.SendEmails(ctx, []Email{email})
}
