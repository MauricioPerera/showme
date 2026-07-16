package main

// main is the thin wiring for the showme MCP server: registers the tools
// from internal/mcpserver and serves them over stdio. It is glue, not
// covered by a frozen oracle -- same criterion as cmd/showme/main.go and
// cmd/showme-web/main.go.
//
// Usage:
//
//	showme-mcp
//
// Register it with an MCP host (e.g. Claude Desktop, Claude Code) pointing
// at this binary; it speaks MCP over stdin/stdout, same transport as the
// KDD infra's own scripts/mcp_server.py.

import (
	"log"

	"github.com/mark3labs/mcp-go/server"

	"github.com/MauricioPerera/showme/internal/mcpserver"
)

func main() {
	s := server.NewMCPServer("showme", "0.1.0")
	s.AddTools(mcpserver.Tools()...)

	if err := server.ServeStdio(s); err != nil {
		log.Fatal(err)
	}
}
