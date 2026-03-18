package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// dinkRequestPayload represents JSON data about the received request
// Only some fields that are relevant are included in the object
type dinkRequestPayload struct {
	Extra json.RawMessage `json:"extra"`
	Type  string          `json:"type"`

	//Content           string            `json:"content"`
	//PlayerName        string            `json:"playerName"`
	//AccountType       string            `json:"accountType"`
	//SeasonalWorld     bool              `json:"seasonalWorld"`
	//DinkAccountHash   string            `json:"dinkAccountHash"`
	//Embeds            []json.RawMessage `json:"embeds"`
	//World             int               `json:"world,omitempty"`
	//RegionID          int               `json:"regionId,omitempty"`
	//ClanName          string            `json:"clanName,omitempty"`
	//GroupIronClanName string            `json:"groupIronClanName,omitempty"`
	//DiscordUser       *json.RawMessage  `json:"discordUser,omitempty"`
}

func parseDinkRequest(r *http.Request) (*dinkRequestPayload, []byte, error) {
	payload := new(dinkRequestPayload)

	// Read body data to parse it and then restore it for later
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read request body: %w", err)
	}
	r.Body.Close()

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		// An incoming request without an image, its body is directly the incoming request's data
		if err = json.Unmarshal(bodyBytes, payload); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal payload json from body: %w", err)
		}
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		// An incoming request with an image, parse its form data
		// Recreate data with a temporary request with data read earlier
		parseReq, _ := http.NewRequest(r.Method, r.URL.String(), bytes.NewBuffer(bodyBytes))
		parseReq.Header = r.Header // Necessary to keep in check for PostFormValue call (which calls ParseMultipartForm) below

		if err = json.Unmarshal([]byte(parseReq.PostFormValue("payload_json")), payload); err != nil {
			return nil, nil, fmt.Errorf("failed to unmarshal payload json from form: %w", err)
		}
	} else {
		return nil, nil, errors.New("unexpected Content-Type: " + contentType)
	}

	// Restore body data to the original request
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return payload, bodyBytes, nil
}

// Models for currently handled notification types

type dinkPayloadLoot struct {
	Items []lootItem `json:"items"`

	//Source            string    `json:"source"`
	//Party             *[]string `json:"party"` // nullable
	//Category          string    `json:"category"`
	//KillCount         *int      `json:"killCount"`
	//RarestProbability float64   `json:"rarestProbability"`
	//NpcID             *int      `json:"npcId"` // nullable
}

type lootItem struct {
	Quantity  int `json:"quantity"`
	PriceEach int `json:"priceEach"`

	//ID        int      `json:"id"`
	//Name      string   `json:"name"`
	//Criteria  []string `json:"criteria"`
	//Rarity    *float64 `json:"rarity"` // nullable
}

type dinkPayloadKillCount struct {
	Boss           string `json:"boss"`
	Count          int    `json:"count"`
	IsPersonalBest *bool  `json:"isPersonalBest"` // nullable

	//GameMessage    string    `json:"gameMessage"`
	//Time           *string   `json:"time"`           // nullable, is a ISO-8601 duration format: https://en.wikipedia.org/wiki/ISO_8601#Durations
	//PersonalBest   *string   `json:"personalBest"`   // nullable, TODO: investigate correct type
	//Party          *[]string `json:"party"`          // nullable
}
