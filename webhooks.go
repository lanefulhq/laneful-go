package laneful

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// VerifyWebhookSignature verifies the HMAC-SHA256 signature of a raw webhook payload.
// The signature is the x-webhook-signature header value, as bare hex or "sha256=<hex>".
// It returns false when secret or signature is empty.
func VerifyWebhookSignature(secret, payload, signature string) bool {
	if secret == "" || signature == "" {
		return false
	}
	signature = strings.TrimPrefix(signature, "sha256=")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}
