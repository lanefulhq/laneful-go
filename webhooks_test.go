package laneful

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func sign(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func TestVerifyWebhookSignature(t *testing.T) {
	const secret = "test-secret"
	const payload = `{ "event": "delivery", "message_id": "H-1-abc" }`
	valid := sign(secret, payload)

	tests := []struct {
		name      string
		payload   string
		signature string
		want      bool
	}{
		{"bare hex", payload, valid, true},
		{"sha256= prefix", payload, "sha256=" + valid, true},
		{"wrong signature", payload, sign(secret, "other"), false},
		{"wrong signature with prefix", payload, "sha256=" + sign(secret, "other"), false},
		{"modified payload", payload + " ", valid, false},
		{"empty signature", payload, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VerifyWebhookSignature(secret, tt.payload, tt.signature); got != tt.want {
				t.Errorf("VerifyWebhookSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}
