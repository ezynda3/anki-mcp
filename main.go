package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
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
		"Simple Anki MCP Server",
		version,
		server.WithToolCapabilities(true),
	)

	// Add simplified tools
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
		mcp.WithDescription("Create a Basic Anki card. Images appear above text, audio references below text. Supports separate audio for front and back."),
		mcp.WithString("deck",
			mcp.Required(),
			mcp.Description("Name of the deck"),
		),
		mcp.WithString("front",
			mcp.Required(),
			mcp.Description("Front text content"),
		),
		mcp.WithString("back",
			mcp.Required(),
			mcp.Description("Back text content"),
		),
		mcp.WithString("image_path",
			mcp.Description("Optional: Path to an image file to include"),
		),
		mcp.WithString("front_audio_path",
			mcp.Description("Optional: Path to an audio file for the front of the card"),
		),
		mcp.WithString("back_audio_path",
			mcp.Description("Optional: Path to an audio file for the back of the card"),
		),
		mcp.WithArray("tags",
			mcp.Description("Optional: Tags for the card"),
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
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the deck to create"),
		),
	)
	s.AddTool(createDeckTool, a.handleCreateDeck)
}

// handleCreateCard creates a new Anki card with standardized formatting
func (a *AnkiMCPServer) handleCreateCard(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	deckName, ok := args["deck"].(string)
	if !ok {
		return errorResult("deck is required"), nil
	}

	frontText, ok := args["front"].(string)
	if !ok {
		return errorResult("front is required"), nil
	}

	backText, ok := args["back"].(string)
	if !ok {
		return errorResult("back is required"), nil
	}

	var tags []string
	if tagsInterface, ok := args["tags"].([]interface{}); ok {
		for _, tag := range tagsInterface {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
	}

	// Process optional image
	var imageName string
	if imagePath, ok := args["image_path"].(string); ok && imagePath != "" {
		imageData, err := fileToBase64(imagePath)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to read image file: %v", err)), nil
		}

		// Generate filename from path
		parts := strings.Split(imagePath, "/")
		if len(parts) == 0 {
			parts = strings.Split(imagePath, "\\")
		}
		imageName = parts[len(parts)-1]

		// Store the image file
		decodedData, err := base64.StdEncoding.DecodeString(imageData)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to decode image data: %v", err)), nil
		}

		err = a.ankiClient.StoreMediaFile(imageName, decodedData)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to store image: %v", err)), nil
		}
	}

	// Process optional front audio
	var frontAudioName string
	if audioPath, ok := args["front_audio_path"].(string); ok && audioPath != "" {
		audioData, err := fileToBase64(audioPath)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to read front audio file: %v", err)), nil
		}

		// Generate filename from path
		parts := strings.Split(audioPath, "/")
		if len(parts) == 0 {
			parts = strings.Split(audioPath, "\\")
		}
		frontAudioName = parts[len(parts)-1]

		// Store the audio file
		decodedData, err := base64.StdEncoding.DecodeString(audioData)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to decode front audio data: %v", err)), nil
		}

		err = a.ankiClient.StoreMediaFile(frontAudioName, decodedData)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to store front audio: %v", err)), nil
		}
	}

	// Process optional back audio
	var backAudioName string
	if audioPath, ok := args["back_audio_path"].(string); ok && audioPath != "" {
		audioData, err := fileToBase64(audioPath)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to read back audio file: %v", err)), nil
		}

		// Generate filename from path
		parts := strings.Split(audioPath, "/")
		if len(parts) == 0 {
			parts = strings.Split(audioPath, "\\")
		}
		backAudioName = parts[len(parts)-1]

		// Store the audio file
		decodedData, err := base64.StdEncoding.DecodeString(audioData)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to decode back audio data: %v", err)), nil
		}

		err = a.ankiClient.StoreMediaFile(backAudioName, decodedData)
		if err != nil {
			return errorResult(fmt.Sprintf("Failed to store back audio: %v", err)), nil
		}
	}

	// Build formatted content
	frontContent := formatContent(frontText, imageName, frontAudioName)
	backContent := formatContent(backText, "", backAudioName)
	note := Note{
		DeckName:  deckName,
		ModelName: "Basic",
		Fields: map[string]string{
			"Front": frontContent,
			"Back":  backContent,
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
				Text: fmt.Sprintf("Created card (ID: %d)", noteID),
			},
		},
	}, nil
}

// formatContent formats the card content with media in standardized positions
func formatContent(text, imageName, audioName string) string {
	var content strings.Builder

	// Image goes first (above text)
	if imageName != "" {
		content.WriteString(fmt.Sprintf(`<img src="%s"><br><br>`, imageName))
	}

	// Text content
	content.WriteString(text)

	// Audio goes last (below text)
	if audioName != "" {
		content.WriteString(fmt.Sprintf(`<br><br>[sound:%s]`, audioName))
	}

	return content.String()
}

// handleListDecks lists all available Anki decks
func (a *AnkiMCPServer) handleListDecks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	decks, err := a.ankiClient.GetDeckNames()
	if err != nil {
		return errorResult(fmt.Sprintf("Failed to get decks: %v", err)), nil
	}

	deckList := strings.Join(decks, "\n")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Decks (%d):\n%s", len(decks), deckList),
			},
		},
	}, nil
}

// handleCreateDeck creates a new Anki deck
func (a *AnkiMCPServer) handleCreateDeck(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	deckName, ok := args["name"].(string)
	if !ok {
		return errorResult("name is required"), nil
	}

	err := a.ankiClient.CreateDeck(deckName)
	if err != nil {
		return errorResult(fmt.Sprintf("Failed to create deck: %v", err)), nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Created deck: %s", deckName),
			},
		},
	}, nil
}

// fileToBase64 reads a file and returns its contents as a base64 string
func fileToBase64(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", filePath, err)
	}
	return base64.StdEncoding.EncodeToString(data), nil
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
