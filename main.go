package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Version information (set by goreleaser)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// AnkiMCPServer wraps the AnkiConnect client and provides MCP tools
type AnkiMCPServer struct {
	ankiClient *AnkiConnect
}

// NewAnkiMCPServer creates a new Anki MCP server
func NewAnkiMCPServer() *AnkiMCPServer {
	// Get AnkiConnect URL from environment variable or use default
	url := os.Getenv("ANKI_CONNECT_URL")
	if url == "" {
		url = defaultAnkiConnectURL
	}

	return &AnkiMCPServer{
		ankiClient: NewAnkiConnectWithURL(url),
	}
}

func main() {
	// Handle version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("anki-mcp %s (commit: %s, built: %s)\n", version, commit, date)
		return
	}

	// Create the Anki MCP server
	ankiServer := NewAnkiMCPServer()

	// Create a new MCP server
	s := server.NewMCPServer(
		"Anki Deck MCP Server",
		version,
		server.WithToolCapabilities(true),
	)

	// Add all tools
	ankiServer.registerTools(s)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

// registerTools registers all Anki tools with the MCP server
func (a *AnkiMCPServer) registerTools(s *server.MCPServer) {
	// Tool: Create Card
	createCardTool := mcp.NewTool("create_card",
		mcp.WithDescription("Create a new Anki card in a specified deck"),
		mcp.WithString("deck_name",
			mcp.Required(),
			mcp.Description("Name of the deck to add the card to"),
		),
		mcp.WithString("front",
			mcp.Required(),
			mcp.Description("Front side content of the card"),
		),
		mcp.WithString("back",
			mcp.Required(),
			mcp.Description("Back side content of the card"),
		),
		mcp.WithString("model_name",
			mcp.Description("Model/note type to use (default: Basic)"),
		),
		mcp.WithArray("tags",
			mcp.Description("Tags to add to the card"),
		),
	)
	s.AddTool(createCardTool, a.handleCreateCard)

	// Tool: List Decks
	listDecksTool := mcp.NewTool("list_decks",
		mcp.WithDescription("List all available Anki decks"),
	)
	s.AddTool(listDecksTool, a.handleListDecks)

	// Tool: Create Deck
	createDeckTool := mcp.NewTool("create_deck",
		mcp.WithDescription("Create a new Anki deck"),
		mcp.WithString("deck_name",
			mcp.Required(),
			mcp.Description("Name of the deck to create"),
		),
	)
	s.AddTool(createDeckTool, a.handleCreateDeck)

	// Tool: Search Cards
	searchCardsTool := mcp.NewTool("search_cards",
		mcp.WithDescription("Search for cards using Anki's search syntax"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query using Anki search syntax (e.g., 'deck:MyDeck', 'tag:important')"),
		),
		mcp.WithString("limit",
			mcp.Description("Maximum number of results to return (default: 10)"),
		),
	)
	s.AddTool(searchCardsTool, a.handleSearchCards)

	// Tool: Add Media
	addMediaTool := mcp.NewTool("add_media",
		mcp.WithDescription("Add a media file to Anki's media collection"),
		mcp.WithString("filename",
			mcp.Required(),
			mcp.Description("Name of the media file"),
		),
		mcp.WithString("data",
			mcp.Required(),
			mcp.Description("Base64 encoded media file data"),
		),
	)
	s.AddTool(addMediaTool, a.handleAddMedia)

	// Tool: Create Card with Media (simplified)
	createCardWithMediaTool := mcp.NewTool("create_card_with_media",
		mcp.WithDescription("Create a new Anki card with media attachments"),
		mcp.WithString("deck_name",
			mcp.Required(),
			mcp.Description("Name of the deck to add the card to"),
		),
		mcp.WithString("front",
			mcp.Required(),
			mcp.Description("Front side content of the card"),
		),
		mcp.WithString("back",
			mcp.Required(),
			mcp.Description("Back side content of the card"),
		),
		mcp.WithString("model_name",
			mcp.Description("Model/note type to use (default: Basic)"),
		),
		mcp.WithArray("tags",
			mcp.Description("Tags to add to the card"),
		),
		mcp.WithString("audio_filename",
			mcp.Description("Audio filename to attach"),
		),
		mcp.WithString("audio_data",
			mcp.Description("Base64 encoded audio data"),
		),
		mcp.WithString("image_filename",
			mcp.Description("Image filename to attach"),
		),
		mcp.WithString("image_data",
			mcp.Description("Base64 encoded image data"),
		),
	)
	s.AddTool(createCardWithMediaTool, a.handleCreateCardWithMedia)

	// Tool: Get Models
	getModelsTool := mcp.NewTool("get_models",
		mcp.WithDescription("Get all available note types/models in Anki"),
	)
	s.AddTool(getModelsTool, a.handleGetModels)

	// Tool: Get Model Fields
	getModelFieldsTool := mcp.NewTool("get_model_fields",
		mcp.WithDescription("Get field names for a specific note type/model"),
		mcp.WithString("model_name",
			mcp.Required(),
			mcp.Description("Name of the model to get fields for"),
		),
	)
	s.AddTool(getModelFieldsTool, a.handleGetModelFields)

	// Tool: Sync
	syncTool := mcp.NewTool("sync",
		mcp.WithDescription("Trigger Anki to sync with AnkiWeb"),
	)
	s.AddTool(syncTool, a.handleSync)

	// Tool: Ping
	pingTool := mcp.NewTool("ping",
		mcp.WithDescription("Check if AnkiConnect is available and responding"),
	)
	s.AddTool(pingTool, a.handlePing)
}

// handleCreateCard creates a new Anki card
func (a *AnkiMCPServer) handleCreateCard(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	deckName, ok := args["deck_name"].(string)
	if !ok {
		return errorResult("deck_name is required and must be a string"), nil
	}

	front, ok := args["front"].(string)
	if !ok {
		return errorResult("front is required and must be a string"), nil
	}

	back, ok := args["back"].(string)
	if !ok {
		return errorResult("back is required and must be a string"), nil
	}

	modelName := "Basic"
	if model, ok := args["model_name"].(string); ok && model != "" {
		modelName = model
	}

	var tags []string
	if tagsInterface, ok := args["tags"].([]interface{}); ok {
		for _, tag := range tagsInterface {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
	}

	note := Note{
		DeckName:  deckName,
		ModelName: modelName,
		Fields: map[string]string{
			"Front": front,
			"Back":  back,
		},
		Tags: tags,
		Options: map[string]interface{}{
			"allowDuplicate": false,
		},
	}

	noteID, err := a.ankiClient.AddNote(note)
	if err != nil {
		return errorResult(fmt.Sprintf("Failed to create card: %v", err)), nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Successfully created card with ID: %d", noteID),
			},
		},
	}, nil
}

// handleListDecks lists all available Anki decks
func (a *AnkiMCPServer) handleListDecks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	decks, err := a.ankiClient.GetDeckNames()
	if err != nil {
		return errorResult(fmt.Sprintf("Failed to get deck names: %v", err)), nil
	}

	deckList := strings.Join(decks, "\n")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Available decks (%d):\n%s", len(decks), deckList),
			},
		},
	}, nil
}

// handleCreateDeck creates a new Anki deck
func (a *AnkiMCPServer) handleCreateDeck(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	deckName, ok := args["deck_name"].(string)
	if !ok {
		return errorResult("deck_name is required and must be a string"), nil
	}

	err := a.ankiClient.CreateDeck(deckName)
	if err != nil {
		return errorResult(fmt.Sprintf("Failed to create deck: %v", err)), nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Successfully created deck: %s", deckName),
			},
		},
	}, nil
}

// handleSearchCards searches for cards using Anki's search syntax
func (a *AnkiMCPServer) handleSearchCards(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	query, ok := args["query"].(string)
	if !ok {
		return errorResult("query is required and must be a string"), nil
	}

	limit := 10
	if limitStr, ok := args["limit"].(string); ok && limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	noteIDs, err := a.ankiClient.FindNotes(query)
	if err != nil {
		return errorResult(fmt.Sprintf("Failed to search cards: %v", err)), nil
	}

	if len(noteIDs) == 0 {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: "No cards found matching the query.",
				},
			},
		}, nil
	}

	// Limit results
	if len(noteIDs) > limit {
		noteIDs = noteIDs[:limit]
	}

	// Get detailed information about the notes
	notesInfo, err := a.ankiClient.GetNotesInfo(noteIDs)
	if err != nil {
		return errorResult(fmt.Sprintf("Failed to get note details: %v", err)), nil
	}

	var results []string
	for i, noteInfo := range notesInfo {
		noteID := noteIDs[i]
		fields, ok := noteInfo["fields"].(map[string]interface{})
		if !ok {
			continue
		}

		var front, back string
		if frontField, ok := fields["Front"].(map[string]interface{}); ok {
			if value, ok := frontField["value"].(string); ok {
				front = value
			}
		}
		if backField, ok := fields["Back"].(map[string]interface{}); ok {
			if value, ok := backField["value"].(string); ok {
				back = value
			}
		}

		var tags []string
		if tagsInterface, ok := noteInfo["tags"].([]interface{}); ok {
			for _, tag := range tagsInterface {
				if tagStr, ok := tag.(string); ok {
					tags = append(tags, tagStr)
				}
			}
		}

		result := fmt.Sprintf("ID: %d\nFront: %s\nBack: %s", noteID, front, back)
		if len(tags) > 0 {
			result += fmt.Sprintf("\nTags: %s", strings.Join(tags, ", "))
		}
		results = append(results, result)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Found %d cards (showing %d):\n\n%s", len(noteIDs), len(results), strings.Join(results, "\n\n---\n\n")),
			},
		},
	}, nil
}

// handleAddMedia adds a media file to Anki's media collection
func (a *AnkiMCPServer) handleAddMedia(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	filename, ok := args["filename"].(string)
	if !ok {
		return errorResult("filename is required and must be a string"), nil
	}

	dataStr, ok := args["data"].(string)
	if !ok {
		return errorResult("data is required and must be a string"), nil
	}

	// Decode base64 data
	data, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		return errorResult(fmt.Sprintf("Failed to decode base64 data: %v", err)), nil
	}

	err = a.ankiClient.StoreMediaFile(filename, data)
	if err != nil {
		return errorResult(fmt.Sprintf("Failed to store media file: %v", err)), nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Successfully stored media file: %s", filename),
			},
		},
	}, nil
}

// handleCreateCardWithMedia creates a new Anki card with media attachments
func (a *AnkiMCPServer) handleCreateCardWithMedia(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	deckName, ok := args["deck_name"].(string)
	if !ok {
		return errorResult("deck_name is required and must be a string"), nil
	}

	front, ok := args["front"].(string)
	if !ok {
		return errorResult("front is required and must be a string"), nil
	}

	back, ok := args["back"].(string)
	if !ok {
		return errorResult("back is required and must be a string"), nil
	}

	modelName := "Basic"
	if model, ok := args["model_name"].(string); ok && model != "" {
		modelName = model
	}

	var tags []string
	if tagsInterface, ok := args["tags"].([]interface{}); ok {
		for _, tag := range tagsInterface {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
	}

	note := Note{
		DeckName:  deckName,
		ModelName: modelName,
		Fields: map[string]string{
			"Front": front,
			"Back":  back,
		},
		Tags: tags,
		Options: map[string]interface{}{
			"allowDuplicate": false,
		},
	}

	// Process audio file
	if audioFilename, ok := args["audio_filename"].(string); ok && audioFilename != "" {
		if audioData, ok := args["audio_data"].(string); ok && audioData != "" {
			note.Audio = append(note.Audio, MediaFile{
				Filename: audioFilename,
				Data:     audioData,
				Fields:   []string{"Front"},
			})
		}
	}

	// Process image file
	if imageFilename, ok := args["image_filename"].(string); ok && imageFilename != "" {
		if imageData, ok := args["image_data"].(string); ok && imageData != "" {
			note.Picture = append(note.Picture, MediaFile{
				Filename: imageFilename,
				Data:     imageData,
				Fields:   []string{"Front"},
			})
		}
	}

	noteID, err := a.ankiClient.AddNote(note)
	if err != nil {
		return errorResult(fmt.Sprintf("Failed to create card with media: %v", err)), nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Successfully created card with media, ID: %d", noteID),
			},
		},
	}, nil
}

// handleGetModels gets all available note types/models
func (a *AnkiMCPServer) handleGetModels(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	models, err := a.ankiClient.GetModelNames()
	if err != nil {
		return errorResult(fmt.Sprintf("Failed to get models: %v", err)), nil
	}

	modelList := strings.Join(models, "\n")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Available models (%d):\n%s", len(models), modelList),
			},
		},
	}, nil
}

// handleGetModelFields gets field names for a specific model
func (a *AnkiMCPServer) handleGetModelFields(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	modelName, ok := args["model_name"].(string)
	if !ok {
		return errorResult("model_name is required and must be a string"), nil
	}

	fields, err := a.ankiClient.GetModelFieldNames(modelName)
	if err != nil {
		return errorResult(fmt.Sprintf("Failed to get model fields: %v", err)), nil
	}

	fieldList := strings.Join(fields, "\n")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Fields for model '%s' (%d):\n%s", modelName, len(fields), fieldList),
			},
		},
	}, nil
}

// handleSync triggers Anki to sync with AnkiWeb
func (a *AnkiMCPServer) handleSync(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	err := a.ankiClient.Sync()
	if err != nil {
		return errorResult(fmt.Sprintf("Failed to sync: %v", err)), nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: "Successfully triggered sync with AnkiWeb",
			},
		},
	}, nil
}

// handlePing checks if AnkiConnect is available
func (a *AnkiMCPServer) handlePing(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	err := a.ankiClient.Ping()
	if err != nil {
		return errorResult(fmt.Sprintf("AnkiConnect is not available: %v", err)), nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: "AnkiConnect is available and responding",
			},
		},
	}, nil
}

// errorResult creates an error result
func errorResult(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Error: %s", message),
			},
		},
		IsError: true,
	}
}
