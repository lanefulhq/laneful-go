package laneful

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestServer returns a send API stub that records each request body.
func newTestServer(t *testing.T, bodies *[]map[string]interface{}) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/email/send" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		var body map[string]interface{}
		if err := json.Unmarshal(raw, &body); err != nil {
			t.Errorf("invalid JSON body: %v", err)
		}
		*bodies = append(*bodies, body)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"accepted"}`))
	}))
	t.Cleanup(server.Close)
	return server
}

func testEmail() Email {
	return Email{
		From:        Address{Email: "sender@example.com"},
		To:          []Address{{Email: "user@example.com"}},
		Subject:     "Hello",
		TextContent: "Hello",
	}
}

func TestSendEmailsWithOptionsSendsIdempotencyKey(t *testing.T) {
	var bodies []map[string]interface{}
	client := NewLanefulClient(newTestServer(t, &bodies).URL, "token")

	resp, err := client.SendEmailsWithOptions(context.Background(), &EmailRequest{
		Emails:       []Email{testEmail()},
		MailSettings: &MailSettings{IdempotencyKey: "order-1234", ReturnMessageIds: true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != "accepted" {
		t.Errorf("status = %q, want accepted", resp.Status)
	}

	settings, ok := bodies[0]["mail_settings"].(map[string]interface{})
	if !ok {
		t.Fatalf("mail_settings missing from request: %v", bodies[0])
	}
	if settings["idempotency_key"] != "order-1234" {
		t.Errorf("idempotency_key = %v, want order-1234", settings["idempotency_key"])
	}
	if settings["return_message_ids"] != true {
		t.Errorf("return_message_ids = %v, want true", settings["return_message_ids"])
	}
}

func TestSendEmailsWithOptionsOmitsUnsetIdempotencyKey(t *testing.T) {
	var bodies []map[string]interface{}
	client := NewLanefulClient(newTestServer(t, &bodies).URL, "token")

	_, err := client.SendEmailsWithOptions(context.Background(), &EmailRequest{
		Emails:       []Email{testEmail()},
		MailSettings: &MailSettings{SandboxMode: true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	settings := bodies[0]["mail_settings"].(map[string]interface{})
	if _, ok := settings["idempotency_key"]; ok {
		t.Errorf("idempotency_key should be omitted when unset: %v", settings)
	}
}

func TestSendEmailsWithOptionsRejectsLongIdempotencyKey(t *testing.T) {
	var bodies []map[string]interface{}
	client := NewLanefulClient(newTestServer(t, &bodies).URL, "token")

	for _, key := range []string{strings.Repeat("k", 37), strings.Repeat("é", 19)} {
		_, err := client.SendEmailsWithOptions(context.Background(), &EmailRequest{
			Emails:       []Email{testEmail()},
			MailSettings: &MailSettings{IdempotencyKey: key},
		})
		if err == nil || !strings.Contains(err.Error(), "between 1 and 36 bytes") {
			t.Errorf("key of %d bytes: err = %v, want a length error", len(key), err)
		}
	}
	if len(bodies) != 0 {
		t.Errorf("no request should be sent for an invalid key, got %d", len(bodies))
	}
}

func TestSendEmailHasNoMailSettings(t *testing.T) {
	var bodies []map[string]interface{}
	client := NewLanefulClient(newTestServer(t, &bodies).URL, "token")

	if _, err := client.SendEmail(context.Background(), testEmail()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := bodies[0]["mail_settings"]; ok {
		t.Errorf("mail_settings should be omitted: %v", bodies[0])
	}
}
