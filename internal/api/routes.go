package api

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/zneix/dink-filter/internal/config"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "moo\n")
}

func handleFilter(w http.ResponseWriter, r *http.Request, cfg *config.Config) {
	// 0. Check if incoming request is authorized to forward requests
	if subtle.ConstantTimeCompare([]byte(r.URL.Query().Get("password")), []byte(cfg.Password)) == 0 {
		log.Println("[API:filter] Unauthorized request from", r.UserAgent())
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// 1. Parse incomding request's body
	// 2. Figure out whether or not the request should be forwarded to other hosts (e.g. check kc/loot threshold against defined values)
	// 3. If successful, make requests to all defined hosts that succeeded
	log.Println("[API:filter] Received a filter request")

	dinkPayload, err := parseDinkRequest(r)
	if err != nil {
		log.Println("[API:filter] Failed to parse incoming request:", err)
		http.Error(w, "failed to parse incoming request", http.StatusInternalServerError)
		return
	}
	log.Printf("[API:filter] Parsed request data %v\n", dinkPayload)

	switch dinkPayload.Type {
	case "KILL_COUNT":
		extra := new(dinkPayloadKillCount)
		if err = json.Unmarshal(dinkPayload.Extra, extra); err != nil {
			log.Println("[API:filter:KILL_COUNT] Failed to parse extra field")
			http.Error(w, "failed to parse extra field in incoming request", http.StatusBadRequest)
			return
		}

		// Filter the request based on defined kill count thresholds
		log.Printf("[API:filter:KILL_COUNT] boss %q kc %d\n", extra.Boss, extra.Count)
		for destURL, destFilter := range cfg.Destinations {
			// First check if we should care about this destination to begin with
			if destFilter == nil || !destFilter.EnableKillCount {
				continue
			}

			if extra.IsPersonalBest != nil {
				if *extra.IsPersonalBest {
					if !destFilter.EnableKillCountPBs {
						continue
					}
				} else {
					if !destFilter.EnableKillCountRegular {
						continue
					}
				}
			}

			kcInterval := cfg.GetKillCountInterval(destURL, extra.Boss)
			log.Printf("[API:filter:KILL_COUNT] host %q wants interval %d\n", destURL, kcInterval)
			// Always notify on the first kill or on a PB
			if extra.Count == 1 || (extra.IsPersonalBest != nil && *extra.IsPersonalBest) || extra.Count%kcInterval == 0 {
				// Send the requests that satisfy filter criteria
				go forwardRequestToDestination(r, destURL, dinkPayload)
			}
		}
	case "LOOT":
		extra := new(dinkPayloadLoot)
		if err = json.Unmarshal(dinkPayload.Extra, extra); err != nil {
			log.Println("[API:filter:LOOT] Failed to parse extra field")
			http.Error(w, "failed to parse extra field in incoming request", http.StatusBadRequest)
			return
		}

		// Filter the request based on total value of all looted items
		var totalValue int
		for _, item := range extra.Items {
			totalValue += item.PriceEach * item.Quantity
		}
		log.Println("[API:filter:LOOT] total value:", totalValue)

		for destURL, destFilter := range cfg.Destinations {
			// First check if we should care about this destination to begin with
			if destFilter == nil || !destFilter.EnableLoot {
				continue
			}

			valueTreshold := cfg.GetLootTreshold(destURL)
			log.Printf("[API:filter:LOOT] host %q wants threshold %d\n", destURL, valueTreshold)
			if totalValue >= valueTreshold {
				// Send the requests that satisfy filter criteria
				go forwardRequestToDestination(r, destURL, dinkPayload)
			}
		}
	default:
		log.Printf("[API:filter] Received unhandled dink request type %q\n", dinkPayload.Type)
	}

	w.WriteHeader(http.StatusOK)
}
