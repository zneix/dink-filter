package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

type dinkRequest struct {
	image       *multipart.File
	imageHeader *multipart.FileHeader
	payload     *dinkRequestPayload
}

// dinkRequestPayload represents JSON data about the received request
// Only some fields that are relevant are included in the object
type dinkRequestPayload struct {
	Content     string          `json:"content"`
	Extra       json.RawMessage `json:"extra"`
	Type        string          `json:"type"`
	PlayerName  string          `json:"playerName"`
	AccountType string          `json:"accountType"`

	//SeasonalWorld     bool              `json:"seasonalWorld"`
	//DinkAccountHash   string            `json:"dinkAccountHash"`
	//Embeds            []json.RawMessage `json:"embeds"`
	//World             int               `json:"world,omitempty"`
	//RegionID          int               `json:"regionId,omitempty"`
	//ClanName          string            `json:"clanName,omitempty"`
	//GroupIronClanName string            `json:"groupIronClanName,omitempty"`
	//DiscordUser       *json.RawMessage  `json:"discordUser,omitempty"`
}

func (p *dinkRequestPayload) String() string {
	return fmt.Sprintf("&api.dinkRequestPayload{Content:%#v, Extra:%v, Type:%v, PlayerName:%v, AccountType:%v}",
		p.Content, string(p.Extra), p.Type, p.PlayerName, p.AccountType,
	)
}

func parseDinkRequest(r *http.Request) (*dinkRequest, error) {
	dr := new(dinkRequest)

	// Read body data to parse it and then restore it for later
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}
	r.Body.Close()

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") {
		// An incoming upload without an image with body being directly the incoming request's data
		if err = json.Unmarshal(bodyBytes, &dr.payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload json from body: %w", err)
		}
	} else if strings.HasPrefix(contentType, "multipart/form-data") {
		// An incoming request with an image, parse its form data
		// Recreate data with a temporary request with data read earlier
		parseReq, _ := http.NewRequest(r.Method, r.URL.String(), bytes.NewBuffer(bodyBytes))
		parseReq.Header = r.Header
		imageFile, imageFileHeader, err := parseReq.FormFile("file")
		if err != nil {
			return nil, fmt.Errorf("failed to get form file: %w", err)
		}
		dr.image = &imageFile
		dr.imageHeader = imageFileHeader

		if err = json.Unmarshal([]byte(parseReq.PostFormValue("payload_json")), &dr.payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload json from form: %w", err)
		}
	} else {
		return nil, errors.New("unexpected Content-Type: " + contentType)
	}

	// Restore body data to the original request
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return dr, nil
}

// Models for currently handled notification types

type dinkPayloadLoot struct {
	Items             []lootItem `json:"items"`
	Source            string     `json:"source"`
	Party             *[]string  `json:"party"` // nullable
	Category          string     `json:"category"`
	KillCount         *int       `json:"killCount"`
	RarestProbability float64    `json:"rarestProbability"`
	NpcID             *int       `json:"npcId"` // nullable
}

type lootItem struct {
	ID        int      `json:"id"`
	Quantity  int      `json:"quantity"`
	PriceEach int      `json:"priceEach"`
	Name      string   `json:"name"`
	Criteria  []string `json:"criteria"`
	Rarity    *float64 `json:"rarity"` // nullable
}

type dinkPayloadKillCount struct {
	Boss           string    `json:"boss"`
	Count          int       `json:"count"`
	GameMessage    string    `json:"gameMessage"`
	Time           *string   `json:"time"`           // nullable, is a ISO-8601 duration format: https://en.wikipedia.org/wiki/ISO_8601#Durations
	IsPersonalBest *bool     `json:"isPersonalBest"` // nullable
	PersonalBest   *string   `json:"personalBest"`   // nullable, TODO: investigate correct type
	Party          *[]string `json:"party"`          // nullable
}
