FROM scratch

COPY anki-mcp /anki-mcp

ENTRYPOINT ["/anki-mcp"]