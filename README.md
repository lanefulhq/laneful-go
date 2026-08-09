[![Go Reference](https://pkg.go.dev/badge/github.com/lanefulhq/laneful-go/v1.svg)](https://pkg.go.dev/github.com/lanefulhq/laneful-go)
---

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
- Domain management (list, create, verify, update email track, delete)

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

## Domain Management

Manage a workspace's sending domains. These are workspace-scoped endpoints on the
organization API host, so point the client at it:

```go
client := laneful.NewLanefulClient("https://api.laneful.net", authToken)
```

### List, create, inspect, verify, and delete

```go
// List (paginated: pass Cursor from the previous response's Pagination.NextCursor)
list, err := client.ListDomains(ctx, workspaceID, &laneful.ListDomainsParams{Limit: 50})

// Create
d, err := client.CreateDomain(ctx, workspaceID, &laneful.CreateDomainRequest{
    Domain:     "mydomain.com",
    Tracking:   "tracking",
    ReturnPath: "return-path",
})

// Get one
d, err = client.GetDomain(ctx, workspaceID, "mydomain.com")

// Trigger DNS verification
d, err = client.VerifyDomain(ctx, workspaceID, "mydomain.com")

// Delete
_, err = client.DeleteDomain(ctx, workspaceID, "mydomain.com")
```

### Update the email track

`UpdateDomain` changes a domain's email track after creation, without deleting and
recreating the domain (which would require re-verifying DNS). `EmailTrackID` is a
`*string` so you can distinguish three cases:

```go
// Set the track
trackID := "e59f0a35-05bc-4516-b585-c06f69c3e67e"
d, err := client.UpdateDomain(ctx, workspaceID, "mydomain.com", &laneful.UpdateDomainRequest{
    EmailTrackID: &trackID,
})

// Clear the track (fall back to the default) — send an empty string
empty := ""
d, err = client.UpdateDomain(ctx, workspaceID, "mydomain.com", &laneful.UpdateDomainRequest{
    EmailTrackID: &empty,
})

// Leave the track unchanged — pass nil (the zero value)
```

See [`examples/domains`](examples/domains) for a runnable sample.

## Deliverability Analytics

Organization-level endpoints for monitoring deliverability. These live on the
organization API host, so point the client at it:

```go
client := laneful.NewLanefulClient("https://api.laneful.net", authToken)
```

All three are paginated: pass a `cursor` (from the previous response's
`NextCursor`) and an optional `Limit` (default 50, max 200). Dates are
`time.Time` values sent as UTC calendar days; leave them zero to use the API
defaults. See [`examples/analytics`](examples/analytics) for a runnable sample.

### Domain Spam Ratio Radar

Domains whose spam complaint ratio reached a critical level (≥ 0.1% of delivered
messages) at a mailbox provider on a given day.

```go
res, err := client.ListDomainSpamRatioRadar(ctx, &laneful.ListDomainSpamRatioRadarParams{
    WorkspaceIDs: []int64{1, 2}, // optional; empty = all workspaces
    Domain:       "example.com", // optional
    StartDate:    time.Now().AddDate(0, 0, -7),
    EndDate:      time.Now(),
})
for _, e := range res.Radar {
    fmt.Printf("%s %s @ %s: %.3f%%\n", e.Date, e.Domain, e.Esp, e.SpamRatio)
}
```

### Google Postmaster Spam Reports

Daily Gmail spam-rate reports from Google Postmaster Tools.

```go
res, err := client.ListGooglePostmasterSpamReports(ctx, &laneful.ListGooglePostmasterSpamReportsParams{
    Domain: "example.com", // optional
})
for _, r := range res.SpamReports {
    fmt.Printf("%s %s: %.3f%%\n", r.Date, r.Domain, r.SpamRatio)
}
```

### Microsoft SNDS Reports

Daily Microsoft SNDS reports for the organization's sending IPs, including the
spam-filter verdict (`SndsFilterGreen`/`Yellow`/`Red`, empty when unknown).

```go
res, err := client.ListSndsReports(ctx, &laneful.ListSndsReportsParams{
    IP: "203.0.113.5", // optional
})
for _, r := range res.SndsReports {
    fmt.Printf("%s %s: filter=%s complaint=%.3f%% traps=%d\n",
        r.Date, r.IP, r.FilterResult, r.ComplaintRate, r.TrapHits)
}
```
