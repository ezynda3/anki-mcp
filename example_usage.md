# Example Usage

This document shows practical examples of using the Anki MCP Server with Claude or other AI assistants.

## Basic Card Creation

**User**: "Create a Spanish vocabulary card with 'Hola' on the front and 'Hello' on the back in my Spanish deck."

**AI Response**: I'll create a Spanish vocabulary card for you.

*Uses `create_card` tool with:*
- deck_name: "Spanish"
- front: "Hola"
- back: "Hello"
- tags: ["vocabulary", "greetings"]

## Deck Management

**User**: "Show me all my Anki decks and create a new one called 'French Vocabulary'."

**AI Response**: Let me first show you your existing decks, then create the new French vocabulary deck.

*Uses `list_decks` tool, then `create_deck` tool with deck_name: "French Vocabulary"*

## Searching Cards

**User**: "Find all cards in my Spanish deck that are tagged with 'difficult'."

**AI Response**: I'll search for difficult Spanish cards for you.

*Uses `search_cards` tool with query: `deck:"Spanish" tag:difficult`*

## Advanced Search Examples

**User**: "Show me cards I added yesterday that are due for review."

*Uses `search_cards` tool with query: `added:1 is:due`*

**User**: "Find cards containing 'conjugation' in any field."

*Uses `search_cards` tool with query: `conjugation`*

## Media Cards

**User**: "Create a pronunciation card with audio for the Spanish word 'Gracias'."

**AI Response**: I'll create a pronunciation card with audio for you.

*Uses `create_card_with_media` tool with:*
- deck_name: "Spanish"
- front: "How do you pronounce 'Gracias'?"
- back: "GRAH-see-ahs (Thank you)"
- audio_filename: "gracias.mp3"
- audio_data: "[base64 encoded audio data]"

## Working with Note Types

**User**: "What note types are available in my Anki collection?"

*Uses `get_models` tool*

**User**: "What fields does the 'Cloze' note type have?"

*Uses `get_model_fields` tool with model_name: "Cloze"*

## Batch Operations

**User**: "Create 5 basic math cards for addition facts."

**AI Response**: I'll create 5 addition fact cards for you.

*Uses `create_card` tool multiple times with different math problems*

## Maintenance

**User**: "Check if Anki is running and sync my collection."

**AI Response**: Let me check your Anki connection and sync your collection.

*Uses `ping` tool to check connectivity, then `sync` tool to synchronize*

## Complex Workflows

**User**: "I'm studying for a Spanish exam. Create a deck called 'Spanish Exam Prep', add 3 vocabulary cards, and show me what's already in my Spanish deck."

**AI Response**: I'll help you prepare for your Spanish exam by creating the deck, adding vocabulary cards, and showing your existing Spanish cards.

*Sequence of operations:*
1. `create_deck` with "Spanish Exam Prep"
2. Multiple `create_card` calls for vocabulary
3. `search_cards` with query: `deck:"Spanish"`

## Error Handling Examples

**User**: "Create a card in my 'NonExistent' deck."

**AI Response**: I tried to create the card, but the deck 'NonExistent' doesn't exist. Would you like me to create the deck first?

*Error from `create_card` tool, followed by suggestion to use `create_deck`*

## Tips for AI Assistants

When using these tools:

1. **Always check connectivity first** with `ping` if users report issues
2. **List decks** before creating cards to suggest existing decks
3. **Use appropriate search queries** - Anki's search syntax is powerful
4. **Handle media carefully** - ensure base64 encoding is correct
5. **Provide helpful error messages** when operations fail
6. **Suggest related operations** (e.g., creating missing decks)

## Common Search Patterns

- `deck:"Deck Name"` - Cards in specific deck
- `tag:tagname` - Cards with specific tag
- `is:due` - Cards due for review
- `is:new` - New cards not yet studied
- `added:1` - Cards added in last day
- `front:word` - Cards with "word" in front field
- `note:"Note Type"` - Cards of specific note type
- `prop:due>-1` - Cards due yesterday or earlier