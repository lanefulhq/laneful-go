package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/lanefulhq/laneful-go"
)

func main() {
	addr := flag.String("addr", ":8000", "address to listen on")
	flag.Parse()

	secret := os.Getenv("LANEFUL_WEBHOOK_SECRET")
	if secret == "" {
		log.Fatal("Missing LANEFUL_WEBHOOK_SECRET")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Verify the raw body: re-serializing parsed JSON can change the bytes.
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 10<<20))
		if err != nil {
			http.Error(w, "invalid body", http.StatusBadRequest)
			return
		}
		signature := r.Header.Get("x-webhook-signature")
		if signature == "" || !laneful.VerifyWebhookSignature(secret, string(body), signature) {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		// Batch mode sends a JSON array of events; otherwise the body is a single event.
		// UseNumber keeps numbers such as timestamps exactly as sent.
		var events []map[string]interface{}
		trimmed := bytes.TrimSpace(body)
		decoder := json.NewDecoder(bytes.NewReader(trimmed))
		decoder.UseNumber()
		if len(trimmed) > 0 && trimmed[0] == '[' {
			err = decoder.Decode(&events)
		} else {
			var event map[string]interface{}
			err = decoder.Decode(&event)
			events = append(events, event)
		}
		if err != nil {
			log.Printf("invalid webhook payload: %v", err)
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		for _, event := range events {
			// Fields every event carries; event-specific fields (url, is_hard, ...) are top level too.
			log.Printf("[%v] id=%v message_id=%v email=%v tag=%q metadata=%v event=%v",
				event["event"], event["id"], event["message_id"], event["email"], event["tag"], event["metadata"], event)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Webhook example listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
