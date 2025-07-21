# Anki MCP Server

A Model Context Protocol (MCP) server that provides tools for managing Anki flashcards through AnkiConnect. This server allows AI assistants to create, search, and manage Anki cards and decks programmatically.

## Features

- **Card Management**: Create new flashcards with front/back content
- **Deck Operations**: List, create, and manage Anki decks
- **Search Functionality**: Search cards using Anki's powerful search syntax
- **Media Support**: Add images and audio files to cards
- **Model Support**: Work with different note types/models
- **Sync Integration**: Trigger AnkiWeb synchronization
- **Configurable**: Customizable AnkiConnect URL

## Prerequisites

1. **Anki Desktop**: Install [Anki](https://apps.ankiweb.net/) desktop application
2. **AnkiConnect**: Install the [AnkiConnect](https://ankiweb.net/shared/info/2055492159) addon in Anki
3. **Go**: Go 1.21 or later for building from source

## Installation

### From Source

```bash
git clone https://github.com/ezynda3/anki-mcp.git
cd anki-mcp
go build -o anki-mcp .
```

### Using Go Install

```bash
go install github.com/ezynda3/anki-mcp@latest
```

## Configuration

The server can be configured using environment variables:

- `ANKI_CONNECT_URL`: AnkiConnect server URL (default: `http://localhost:8765`)

## Usage

### With Claude Desktop

Add the following to your Claude Desktop configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "anki": {
      "command": "/path/to/anki-mcp",
      "env": {
        "ANKI_CONNECT_URL": "http://localhost:8765"
      }
    }
  }
}
```

### With MCP Inspector

For development and testing:

```bash
# Install MCP Inspector
npm install -g @modelcontextprotocol/inspector

# Run with inspector
mcp-inspector /path/to/anki-mcp
```

### Direct Usage

```bash
# Start the server
./anki-mcp

# The server communicates via stdio using the MCP protocol
```

## Available Tools

### `ping`
Check if AnkiConnect is available and responding.

**Parameters**: None

**Example**:
```
Use the ping tool to check if Anki is running.
```

### `list_decks`
List all available Anki decks.

**Parameters**: None

**Example**:
```
Show me all my Anki decks.
```

### `create_deck`
Create a new Anki deck.

**Parameters**:
- `deck_name` (required): Name of the deck to create

**Example**:
```
Create a new deck called "Spanish Vocabulary".
```

### `create_card`
Create a new flashcard in a specified deck.

**Parameters**:
- `deck_name` (required): Name of the deck to add the card to
- `front` (required): Front side content of the card
- `back` (required): Back side content of the card
- `model_name` (optional): Note type to use (default: "Basic")
- `tags` (optional): Array of tags to add to the card

**Example**:
```
Create a card in my "Spanish Vocabulary" deck with front "Hola" and back "Hello" with tags "greetings" and "basic".
```

### `search_cards`
Search for cards using Anki's search syntax.

**Parameters**:
- `query` (required): Search query using Anki search syntax
- `limit` (optional): Maximum number of results to return (default: 10)

**Examples**:
```
Search for cards in the "Spanish Vocabulary" deck.
Find cards tagged with "difficult".
Search for cards containing "hello" in any field.
```

**Search Syntax Examples**:
- `deck:"Spanish Vocabulary"` - Cards in a specific deck
- `tag:difficult` - Cards with a specific tag
- `front:hello` - Cards with "hello" in the front field
- `is:due` - Cards that are due for review
- `added:1` - Cards added in the last day

### `add_media`
Add a media file to Anki's media collection.

**Parameters**:
- `filename` (required): Name of the media file
- `data` (required): Base64 encoded media file data

**Example**:
```
Add an audio file "pronunciation.mp3" to Anki's media collection.
```

### `create_card_with_media`
Create a new flashcard with media attachments.

**Parameters**:
- `deck_name` (required): Name of the deck to add the card to
- `front` (required): Front side content of the card
- `back` (required): Back side content of the card
- `model_name` (optional): Note type to use (default: "Basic")
- `tags` (optional): Array of tags to add to the card
- `audio_filename` (optional): Audio filename to attach
- `audio_data` (optional): Base64 encoded audio data
- `image_filename` (optional): Image filename to attach
- `image_data` (optional): Base64 encoded image data

**Example**:
```
Create a card with an image showing a Spanish flag and audio pronunciation.
```

### `get_models`
Get all available note types/models in Anki.

**Parameters**: None

**Example**:
```
Show me all available note types in Anki.
```

### `get_model_fields`
Get field names for a specific note type/model.

**Parameters**:
- `model_name` (required): Name of the model to get fields for

**Example**:
```
What fields are available for the "Cloze" note type?
```

### `sync`
Trigger Anki to sync with AnkiWeb.

**Parameters**: None

**Example**:
```
Sync my Anki collection with AnkiWeb.
```

## Error Handling

The server provides detailed error messages for common issues:

- **AnkiConnect not available**: Ensure Anki is running and AnkiConnect addon is installed
- **Deck not found**: Check deck name spelling and existence
- **Invalid parameters**: Verify required parameters are provided
- **Media encoding errors**: Ensure media data is properly base64 encoded

## Development

### Building

```bash
go build -v .
```

### Testing

```bash
go test -v ./...
```

### Code Structure

- `main.go`: MCP server implementation and tool handlers
- `ankiconnect.go`: AnkiConnect client wrapper
- `go.mod`: Go module dependencies

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- Built using [MCP-Go](https://github.com/mark3labs/mcp-go)
- Integrates with [AnkiConnect](https://foosoft.net/projects/anki-connect/)
- Inspired by the [go-anki-deck](https://github.com/ezynda3/go-anki-deck) library

## Troubleshooting

### Common Issues

**"AnkiConnect is not available"**
- Ensure Anki desktop is running
- Verify AnkiConnect addon is installed and enabled
- Check that AnkiConnect is listening on the correct port (default: 8765)

**"Failed to create deck/card"**
- Verify deck names don't contain invalid characters
- Check that required fields are provided
- Ensure AnkiConnect permissions allow the operation

**Media files not working**
- Verify media data is properly base64 encoded
- Check file extensions are supported by Anki
- Ensure media files aren't too large

### Debug Mode

Set environment variable for verbose logging:
```bash
export MCP_DEBUG=1
./anki-mcp
```

## API Reference

For detailed information about the Model Context Protocol, see:
- [MCP Specification](https://modelcontextprotocol.io/)
- [MCP-Go Documentation](https://github.com/mark3labs/mcp-go)

For AnkiConnect API details, see:
- [AnkiConnect Documentation](https://foosoft.net/projects/anki-connect/)