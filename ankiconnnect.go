package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func ankiDo(client *http.Client, action string, params any) (json.RawMessage, error) {
	body, _ := json.Marshal(ankiRequest{Action: action, Version: 6, Params: params})
	resp, err := client.Post(ankiConnectURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ankiconnect %s: %w", action, err)
	}
	defer resp.Body.Close()

	var ar ankiResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return nil, fmt.Errorf("ankiconnect %s decode: %w", action, err)
	}
	if ar.Error != nil {
		return nil, fmt.Errorf("ankiconnect %s: %s", action, *ar.Error)
	}
	return ar.Result, nil
}

// EnsureModel creates VocabPro if it doesn't exist, then always syncs the
// latest template, styling, and any missing fields.
func EnsureModel(client *http.Client) error {
	raw, err := ankiDo(client, "modelNames", nil)
	if err != nil {
		return err
	}

	var names []string
	json.Unmarshal(raw, &names)

	exists := false
	for _, n := range names {
		if n == modelName {
			exists = true
			break
		}
	}

	if !exists {
		params := map[string]any{
			"modelName":     modelName,
			"inOrderFields": wantedFields,
			"css":           cardCSS,
			"cardTemplates": []map[string]string{
				{"Name": "Fill in the blank", "Front": cardFront, "Back": cardBack},
			},
		}
		_, err = ankiDo(client, "createModel", params)
		return err
	}

	// Model exists — add any missing fields then update template + styling.
	raw, err = ankiDo(client, "modelFieldNames", map[string]any{"modelName": modelName})
	if err != nil {
		return err
	}
	var existing []string
	json.Unmarshal(raw, &existing)
	has := make(map[string]bool, len(existing))
	for _, f := range existing {
		has[f] = true
	}
	for _, f := range wantedFields {
		if !has[f] {
			fmt.Printf("[MODEL] adding missing field %q\n", f)
			if _, err := ankiDo(client, "modelFieldAdd", map[string]any{"modelName": modelName, "fieldName": f}); err != nil {
				return err
			}
		}
	}

	if _, err := ankiDo(client, "updateModelTemplates", map[string]any{
		"model": map[string]any{
			"name": modelName,
			"templates": map[string]any{
				"Fill in the blank": map[string]string{"Front": cardFront, "Back": cardBack},
			},
		},
	}); err != nil {
		return err
	}

	_, err = ankiDo(client, "updateModelStyling", map[string]any{
		"model": map[string]any{"name": modelName, "css": cardCSS},
	})
	return err
}

// PushToAnki uploads audio and creates notes for all records, then syncs.
func PushToAnki(client *http.Client, records []AnkiRecord) {
	if err := EnsureModel(client); err != nil {
		log.Fatal("ensure model: ", err)
	}

	type audioEntry struct {
		Data     string   `json:"data"`
		Filename string   `json:"filename"`
		Fields   []string `json:"fields"`
	}

	type noteOptions struct {
		AllowDuplicate bool   `json:"allowDuplicate"`
		DuplicateScope string `json:"duplicateScope"`
	}

	type note struct {
		DeckName  string            `json:"deckName"`
		ModelName string            `json:"modelName"`
		Fields    map[string]string `json:"fields"`
		Options   noteOptions       `json:"options"`
		Tags      []string          `json:"tags"`
		Audio     []audioEntry      `json:"audio,omitempty"`
	}

	var notes []note
	for _, r := range records {
		n := note{
			DeckName:  "vocabulary",
			ModelName: modelName,
			Fields: map[string]string{
				"Word":           r.Word,
				"IPA":            r.Phonetic,
				"Part_of_Speech": r.PartOfSpeech,
				"Definition_VI":  r.Definition,
				"Sentence_Front": r.SentenceFront,
				"Synonym_Hint":   r.SynonymHint,
				"Structure_Hint": r.StructureHint,
				"Audio":          "",
			},
			Options: noteOptions{AllowDuplicate: true, DuplicateScope: "deck"},
			Tags:    parseTags(r.Tags),
		}

		if r.AudioFilePath != "" {
			if data, err := os.ReadFile(r.AudioFilePath); err == nil {
				n.Audio = []audioEntry{
					{
						Data:     base64.StdEncoding.EncodeToString(data),
						Filename: r.AudioFileName,
						Fields:   []string{"Audio"},
					},
				}
			} else {
				fmt.Printf("[AUDIO] skipping %s: %v\n", r.AudioFilePath, err)
			}
		}

		notes = append(notes, n)
	}

	raw, err := ankiDo(client, "addNotes", map[string]any{"notes": notes})
	if err != nil {
		log.Fatal("addNotes: ", err)
	}

	// addNotes returns an array of note IDs (null for duplicates/failures)
	var ids []any
	json.Unmarshal(raw, &ids)
	added := 0
	for _, id := range ids {
		if id != nil {
			added++
		}
	}
	fmt.Printf("[ANKI] %d/%d notes created\n", added, len(notes))

	if _, err := ankiDo(client, "sync", nil); err != nil {
		log.Println("sync warning:", err) // non-fatal — user can sync manually
	} else {
		fmt.Println("[ANKI] Synced to AnkiWeb")
	}
}

func parseTags(raw string) []string {
	var tags []string
	for _, t := range strings.Fields(raw) {
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}
