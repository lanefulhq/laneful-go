# Laneful Go Client

A Go client library for the Laneful API.

## Installation

```bash
go get github.com/lanefulhq/laneful-go
```

## Quick Start

```go
package main

import (
    "context"
    "log"

    "github.com/lanefulhq/laneful-go"
)

func main() {
    client := laneful.NewLanefulClient("https://custom-endpoint.send.laneful.net", "your-auth-token")

    email := laneful.Email{
        From: laneful.Address{
            Email: "sender@example.com",
            Name:  "Your Name",
        },
        To: []laneful.Address{
            {Email: "recipient@example.com", Name: "Recipient Name"},
        },
        Subject:     "Hello from Laneful",
        TextContent: "This is a test email.",
        HTMLContent: "<h1>This is a test email.</h1>",
    }

    resp, err := client.SendEmail(context.Background(), email)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Email sent successfully: %s", resp.Status)
}
```

## Features

- Send single or multiple emails
- Support for plain text and HTML content
- Email templates with dynamic data
- File attachments
- Email tracking (opens, clicks, unsubscribes)
- Custom headers
- Scheduled sending
- Webhook data
- Reply-to addresses

## API Reference

### Creating a Client

```go
client := laneful.NewLanefulClient(baseURL, authToken)
```

### Sending Emails

#### Single Email

```go
resp, err := client.SendEmail(ctx, email)
```

#### Multiple Emails

```go
resp, err := client.SendEmails(ctx, []laneful.Email{email1, email2})
```

## Examples

### Template Email

```go
email := laneful.Email{
    From: laneful.Address{Email: "sender@example.com"},
    To: []laneful.Address{{Email: "user@example.com"}},
    TemplateID: "welcome-template",
    TemplateData: map[string]interface{}{
        "name": "John Doe",
        "company": "Acme Corp",
    },
}
```

### Email with Attachments

```go
email := laneful.Email{
    From: laneful.Address{Email: "sender@example.com"},
    To: []laneful.Address{{Email: "user@example.com"}},
    Subject: "Document Attached",
    TextContent: "Please find the document attached.",
    Attachments: []laneful.Attachment{
        {
            FileName:    "document.pdf",
            Content:     "base64-encoded-content",
            ContentType: "application/pdf",
        },
    },
}
```

### Scheduled Email

```go
email := laneful.Email{
    From: laneful.Address{Email: "sender@example.com"},
    To: []laneful.Address{{Email: "user@example.com"}},
    Subject: "Scheduled Email",
    TextContent: "This email was scheduled.",
    SendTime: time.Now().Add(24 * time.Hour).Unix(),
}
```

### Email with Tracking

```go
email := laneful.Email{
    From: laneful.Address{Email: "sender@example.com"},
    To: []laneful.Address{{Email: "user@example.com"}},
    Subject: "Tracked Email",
    HTMLContent: "<p>This email is tracked.</p>",
    Tracking: &laneful.TrackingSettings{
        Opens:  true,
        Clicks: true,
    },
}
```
