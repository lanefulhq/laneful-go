package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/lanefulhq/laneful-go"
)

func main() {
	// Deliverability analytics lives on the organization API host.
	endpoint := flag.String("e", "https://api.laneful.net", "API endpoint URL")
	domain := flag.String("d", "", "optional sending domain filter")
	ip := flag.String("ip", "", "optional sending IP filter (SNDS)")
	flag.Parse()

	apiKey := os.Getenv("LANEFUL_API_KEY")
	client := laneful.NewLanefulClient(*endpoint, apiKey)
	ctx := context.Background()

	// Look at the last 7 days (UTC).
	end := time.Now().UTC()
	start := end.AddDate(0, 0, -7)

	// 1. Domain spam ratio radar: domains whose spam complaint ratio reached
	//    a critical level (>= 0.1%) at a mailbox provider.
	radar, err := client.ListDomainSpamRatioRadar(ctx, &laneful.ListDomainSpamRatioRadarParams{
		Domain:    *domain,
		StartDate: start,
		EndDate:   end,
		Limit:     50,
	})
	if err != nil {
		log.Fatalf("radar: %v", err)
	}
	log.Printf("spam ratio radar: %d entr(y/ies)", len(radar.Radar))
	for _, e := range radar.Radar {
		log.Printf("  %s  %s @ %s  spam_ratio=%.3f%%", e.Date, e.Domain, e.Esp, e.SpamRatio)
	}

	// 2. Google Postmaster spam reports for Gmail.
	postmaster, err := client.ListGooglePostmasterSpamReports(ctx, &laneful.ListGooglePostmasterSpamReportsParams{
		Domain:    *domain,
		StartDate: start,
		EndDate:   end,
		Limit:     50,
	})
	if err != nil {
		log.Fatalf("postmaster: %v", err)
	}
	log.Printf("google postmaster: %d report(s)", len(postmaster.SpamReports))
	for _, r := range postmaster.SpamReports {
		log.Printf("  %s  %s  spam_ratio=%.3f%%", r.Date, r.Domain, r.SpamRatio)
	}

	// 3. Microsoft SNDS reports for the organization's sending IPs.
	snds, err := client.ListSndsReports(ctx, &laneful.ListSndsReportsParams{
		IP:        *ip,
		StartDate: start,
		EndDate:   end,
		Limit:     50,
	})
	if err != nil {
		log.Fatalf("snds: %v", err)
	}
	log.Printf("microsoft snds: %d report(s)", len(snds.SndsReports))
	for _, r := range snds.SndsReports {
		log.Printf("  %s  %s  filter=%s complaint_rate=%.3f%% traps=%d",
			r.Date, r.IP, r.FilterResult, r.ComplaintRate, r.TrapHits)
	}
}
