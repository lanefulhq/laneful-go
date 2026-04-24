package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/lanefulhq/laneful-go"
)

func main() {
	from := flag.String("f", "sender@example.com", "from email address")
	to := flag.String("t", "recipient@example.com", "to email address")
	endpoint := flag.String("e", "https://custom-endpoint.send.laneful.net", "API endpoint URL")
	flag.Parse()

	apiKey := os.Getenv("LANEFUL_API_KEY")
	client := laneful.NewLanefulClient(*endpoint, apiKey)

	email := laneful.Email{
		From: laneful.Address{
			Email: *from,
			Name:  "From Name",
		},
		To: []laneful.Address{
			{Email: *to, Name: "Recipient Name"},
		},
		Subject:     "Hello from Laneful",
		TextContent: "This is a test email.",
		HTMLContent: "<h1>This is a test email.</h1>",
	}

	resp, err := client.SendEmailsWithOptions(context.Background(), &laneful.EmailRequest{
		Emails: []laneful.Email{email},
		MailSettings: &laneful.MailSettings{
			ReturnMessageIds: true,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Email sent successfully: %v", resp)
}
