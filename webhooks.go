package laneful

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// VerifyWebhookSignature verifies the signature of a webhook payload
func VerifyWebhookSignature(secret, payload, signature string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expected))
}
