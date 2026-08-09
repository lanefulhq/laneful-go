package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/lanefulhq/laneful-go"
)

func main() {
	// Domain management lives on the organization API host.
	endpoint := flag.String("e", "https://api.laneful.net", "organization API endpoint URL")
	workspaceID := flag.Int64("w", 0, "workspace ID")
	domain := flag.String("d", "", "domain name to inspect")
	track := flag.String("track", "", "email track ID to assign to the domain")
	clearTrack := flag.Bool("clear-track", false, "clear the domain's email track (fall back to the default track)")
	create := flag.Bool("create", false, "create the domain before inspecting it")
	del := flag.Bool("delete", false, "delete the domain at the end")
	flag.Parse()

	if *workspaceID == 0 || *domain == "" {
		log.Fatal("usage: -w <workspace_id> -d <domain> [-create] [-track <id> | -clear-track] [-delete]")
	}

	apiKey := os.Getenv("LANEFUL_API_KEY")
	client := laneful.NewLanefulClient(*endpoint, apiKey)
	ctx := context.Background()

	// Optionally create the domain.
	if *create {
		created, err := client.CreateDomain(ctx, *workspaceID, &laneful.CreateDomainRequest{
			Domain:     *domain,
			Tracking:   "tracking",
			ReturnPath: "return-path",
		})
		if err != nil {
			log.Fatalf("create: %v", err)
		}
		log.Printf("created domain %q (verified=%t)", created.Domain, created.Verified)
	}

	// List domains in the workspace.
	list, err := client.ListDomains(ctx, *workspaceID, &laneful.ListDomainsParams{Limit: 50})
	if err != nil {
		log.Fatalf("list: %v", err)
	}
	log.Printf("workspace %d has %d domain(s) on this page", *workspaceID, len(list.Domains))

	// Fetch the specific domain.
	d, err := client.GetDomain(ctx, *workspaceID, *domain)
	if err != nil {
		log.Fatalf("get: %v", err)
	}
	log.Printf("domain %q: verified=%t track=%q", d.Domain, d.Verified, d.EmailTrackID)

	// Update the email track after creation: set it to a track, or clear it back
	// to the default. Note the *string field on UpdateDomainRequest:
	//   - a track ID   -> set the track
	//   - &""          -> clear the track (fall back to the default)
	//   - nil          -> leave the track unchanged
	switch {
	case *clearTrack:
		empty := ""
		updated, err := client.UpdateDomain(ctx, *workspaceID, *domain, &laneful.UpdateDomainRequest{
			EmailTrackID: &empty,
		})
		if err != nil {
			log.Fatalf("clear track: %v", err)
		}
		log.Printf("cleared track; now %q", updated.EmailTrackID)
	case *track != "":
		updated, err := client.UpdateDomain(ctx, *workspaceID, *domain, &laneful.UpdateDomainRequest{
			EmailTrackID: track,
		})
		if err != nil {
			log.Fatalf("set track: %v", err)
		}
		log.Printf("set track to %q", updated.EmailTrackID)
	}

	// Trigger DNS verification and report the per-record results.
	verified, err := client.VerifyDomain(ctx, *workspaceID, *domain)
	if err != nil {
		log.Fatalf("verify: %v", err)
	}
	log.Printf("verification: verified=%t (dkim1=%t dkim2=%t tracking=%t return_path=%t)",
		verified.Verified, verified.Dkim1Verified, verified.Dkim2Verified,
		verified.TrackingVerified, verified.ReturnPathVerified)

	// Optionally delete the domain.
	if *del {
		res, err := client.DeleteDomain(ctx, *workspaceID, *domain)
		if err != nil {
			log.Fatalf("delete: %v", err)
		}
		log.Printf("delete: %s", res.Message)
	}
}
