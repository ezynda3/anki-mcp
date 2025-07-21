package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultAnkiConnectURL = "http://localhost:8765"
	ankiConnectVersion    = 6
)

// AnkiConnect represents a client for communicating with AnkiConnect addon
type AnkiConnect struct {
	URL     string
	Version int
	client  *http.Client
}

// ankiRequest represents a request to AnkiConnect API
type ankiRequest struct {
	Action  string      `json:"action"`
	Version int         `json:"version"`
	Params  interface{} `json:"params,omitempty"`
}

// ankiResponse represents a response from AnkiConnect API
type ankiResponse struct {
	Result interface{} `json:"result"`
	Error  string      `json:"error"`
}

// NewAnkiConnect creates a new AnkiConnect client with default settings
func NewAnkiConnect() *AnkiConnect {
	return &AnkiConnect{
		URL:     defaultAnkiConnectURL,
		Version: ankiConnectVersion,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewAnkiConnectWithURL creates a new AnkiConnect client with custom URL
func NewAnkiConnectWithURL(url string) *AnkiConnect {
	ac := NewAnkiConnect()
	ac.URL = url
	return ac
}

// invoke makes a request to AnkiConnect API
func (ac *AnkiConnect) invoke(action string, params interface{}) (interface{}, error) {
	req := ankiRequest{
		Action:  action,
		Version: ac.Version,
		Params:  params,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := ac.client.Post(ac.URL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to AnkiConnect: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result ankiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("AnkiConnect error: %s", result.Error)
	}

	return result.Result, nil
}

// Ping checks if AnkiConnect is available
func (ac *AnkiConnect) Ping() error {
	_, err := ac.invoke("version", nil)
	return err
}

// GetDeckNames returns all deck names in Anki
func (ac *AnkiConnect) GetDeckNames() ([]string, error) {
	result, err := ac.invoke("deckNames", nil)
	if err != nil {
		return nil, err
	}

	names, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	deckNames := make([]string, len(names))
	for i, name := range names {
		deckNames[i], ok = name.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected deck name type")
		}
	}

	return deckNames, nil
}

// CreateDeck creates a new deck in Anki
func (ac *AnkiConnect) CreateDeck(name string) error {
	params := map[string]string{"deck": name}
	_, err := ac.invoke("createDeck", params)
	return err
}

// DeleteDeck deletes a deck and all its cards
func (ac *AnkiConnect) DeleteDeck(name string) error {
	params := map[string]interface{}{
		"decks":    []string{name},
		"cardsToo": true,
	}
	_, err := ac.invoke("deleteDecks", params)
	return err
}

// Note represents a note in AnkiConnect format
type Note struct {
	DeckName  string                 `json:"deckName"`
	ModelName string                 `json:"modelName"`
	Fields    map[string]string      `json:"fields"`
	Tags      []string               `json:"tags,omitempty"`
	Audio     []MediaFile            `json:"audio,omitempty"`
	Picture   []MediaFile            `json:"picture,omitempty"`
	Video     []MediaFile            `json:"video,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// MediaFile represents media attachment in AnkiConnect format
type MediaFile struct {
	Path     string   `json:"path,omitempty"`
	Filename string   `json:"filename,omitempty"`
	Fields   []string `json:"fields,omitempty"`
	Data     string   `json:"data,omitempty"`
}

// AddNote adds a single note to Anki
func (ac *AnkiConnect) AddNote(note Note) (int64, error) {
	params := map[string]interface{}{"note": note}
	result, err := ac.invoke("addNote", params)
	if err != nil {
		return 0, err
	}

	// AnkiConnect returns note ID as float64
	if id, ok := result.(float64); ok {
		return int64(id), nil
	}

	return 0, fmt.Errorf("unexpected note ID type")
}

// FindNotes searches for notes matching a query
func (ac *AnkiConnect) FindNotes(query string) ([]int64, error) {
	params := map[string]string{"query": query}
	result, err := ac.invoke("findNotes", params)
	if err != nil {
		return nil, err
	}

	ids, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	noteIDs := make([]int64, len(ids))
	for i, id := range ids {
		if fid, ok := id.(float64); ok {
			noteIDs[i] = int64(fid)
		} else {
			return nil, fmt.Errorf("unexpected note ID type")
		}
	}

	return noteIDs, nil
}

// UpdateNoteFields updates fields of an existing note
func (ac *AnkiConnect) UpdateNoteFields(noteID int64, fields map[string]string) error {
	params := map[string]interface{}{
		"note": map[string]interface{}{
			"id":     noteID,
			"fields": fields,
		},
	}
	_, err := ac.invoke("updateNoteFields", params)
	return err
}

// StoreMediaFile stores a media file in Anki's media folder
func (ac *AnkiConnect) StoreMediaFile(filename string, data []byte) error {
	// AnkiConnect expects base64 encoded data
	encodedData := base64.StdEncoding.EncodeToString(data)
	params := map[string]interface{}{
		"filename": filename,
		"data":     encodedData,
	}
	_, err := ac.invoke("storeMediaFile", params)
	return err
}

// Sync triggers Anki to sync with AnkiWeb
func (ac *AnkiConnect) Sync() error {
	_, err := ac.invoke("sync", nil)
	return err
}

// GetNotesInfo retrieves detailed information about notes
func (ac *AnkiConnect) GetNotesInfo(noteIDs []int64) ([]map[string]interface{}, error) {
	params := map[string]interface{}{"notes": noteIDs}
	result, err := ac.invoke("notesInfo", params)
	if err != nil {
		return nil, err
	}

	notes, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	notesInfo := make([]map[string]interface{}, len(notes))
	for i, note := range notes {
		noteMap, ok := note.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected note type")
		}
		notesInfo[i] = noteMap
	}

	return notesInfo, nil
}

// GetModelNames returns all model names in Anki
func (ac *AnkiConnect) GetModelNames() ([]string, error) {
	result, err := ac.invoke("modelNames", nil)
	if err != nil {
		return nil, err
	}

	names, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	modelNames := make([]string, len(names))
	for i, name := range names {
		modelNames[i], ok = name.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected model name type")
		}
	}

	return modelNames, nil
}

// GetModelFieldNames returns field names for a given model
func (ac *AnkiConnect) GetModelFieldNames(modelName string) ([]string, error) {
	params := map[string]string{"modelName": modelName}
	result, err := ac.invoke("modelFieldNames", params)
	if err != nil {
		return nil, err
	}

	names, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response type")
	}

	fieldNames := make([]string, len(names))
	for i, name := range names {
		fieldNames[i], ok = name.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected field name type")
		}
	}

	return fieldNames, nil
}
