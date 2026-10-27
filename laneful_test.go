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

func TestSendEmailsWithOptionsChecksWebhookDataBytes(t *testing.T) {
	var bodies []map[string]interface{}
	client := NewLanefulClient(newTestServer(t, &bodies).URL, "token")

	send := func(data map[string]string) error {
		email := testEmail()
		email.WebhookData = data
		_, err := client.SendEmailsWithOptions(context.Background(), &EmailRequest{Emails: []Email{email}})
		return err
	}

	// "é" is 2 bytes in UTF-8, so 25 of them fill a 50-byte key and 50 fill a 100-byte value.
	if err := send(map[string]string{strings.Repeat("é", 25): strings.Repeat("é", 50)}); err != nil {
		t.Errorf("key of 50 bytes and value of 100 bytes: err = %v, want nil", err)
	}
	if len(bodies) != 1 {
		t.Fatalf("valid webhook data should be sent, got %d requests", len(bodies))
	}

	tooMany := map[string]string{}
	for i := 0; i < 21; i++ {
		tooMany[strings.Repeat("k", i+1)] = "v"
	}
	for _, tc := range []struct {
		data map[string]string
		want string
	}{
		{map[string]string{strings.Repeat("é", 26): "v"}, "keys cannot exceed 50 bytes"},
		{map[string]string{"k": strings.Repeat("é", 51)}, "values cannot exceed 100 bytes"},
		{tooMany, "more than 20 keys"},
	} {
		if err := send(tc.data); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("err = %v, want %q", err, tc.want)
		}
	}
	if len(bodies) != 1 {
		t.Errorf("no request should be sent for invalid webhook data, got %d", len(bodies)-1)
	}
}
