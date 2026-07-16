// Package mcpserver exposes showme's core use cases as MCP tools, reusing
// internal/cli's already-contracted command functions directly -- this
// package only translates MCP tool-call arguments into their inputs and
// their results into JSON-encoded tool results, the same "adapter, not a
// second implementation" principle DEFINITION.md requires of every
// adapter (CLI, webapp, MCP) over the core.
//
// This is the one package in the repo whose deps_allowed is not empty: it
// imports github.com/mark3labs/mcp-go, the project's first external Go
// dependency, to speak the MCP protocol. Every other contract stays
// dependency-free; this exception is scoped to MCP wiring only.
package mcpserver

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/MauricioPerera/showme/internal/cli"
)

func jsonResult(v any) *mcp.CallToolResult {
	encoded, err := json.Marshal(v)
	if err != nil {
		return mcp.NewToolResultErrorFromErr("failed to encode result", err)
	}
	return mcp.NewToolResultText(string(encoded))
}

// Tools returns the MCP tools showme exposes to agents: create_project,
// list_projects and show_project. Each handler validates its required
// arguments, delegates to the matching cli.RunXCommand, and returns the
// result JSON-encoded (mirroring the CLI's --json output) or, on a
// file-system error from the command, an MCP tool error.
func Tools() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("create_project",
				mcp.WithDescription("Create a showme project from a DESIGN.md, an OKF knowledge bundle and a deck JSON file."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Project name")),
				mcp.WithString("design_path", mcp.Required(), mcp.Description("Path to DESIGN.md")),
				mcp.WithString("knowledge_root", mcp.Required(), mcp.Description("Path to the OKF bundle directory")),
				mcp.WithString("deck_path", mcp.Required(), mcp.Description("Path to a deck JSON file")),
				mcp.WithString("out_dir", mcp.Required(), mcp.Description("Directory to save the project")),
			),
			Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				name, err := request.RequireString("name")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				designPath, err := request.RequireString("design_path")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				knowledgeRoot, err := request.RequireString("knowledge_root")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				deckPath, err := request.RequireString("deck_path")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				outDir, err := request.RequireString("out_dir")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				result, err := cli.RunCreateProjectCommand(cli.CreateProjectCommandInput{
					Name:          name,
					DesignPath:    designPath,
					KnowledgeRoot: knowledgeRoot,
					DeckPath:      deckPath,
					OutDir:        outDir,
				})
				if err != nil {
					return mcp.NewToolResultErrorFromErr("create_project failed", err), nil
				}
				return jsonResult(result), nil
			},
		},
		{
			Tool: mcp.NewTool("list_projects",
				mcp.WithDescription("List the showme projects saved under a directory, with their name, path and archived state."),
				mcp.WithString("dir", mcp.Required(), mcp.Description("Directory containing saved projects")),
			),
			Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				dir, err := request.RequireString("dir")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				result, err := cli.RunListProjectsCommand(cli.ListProjectsCommandInput{Dir: dir})
				if err != nil {
					return mcp.NewToolResultErrorFromErr("list_projects failed", err), nil
				}
				return jsonResult(result), nil
			},
		},
		{
			Tool: mcp.NewTool("show_project",
				mcp.WithDescription("Load and return a saved showme project in full, given its file path."),
				mcp.WithString("path", mcp.Required(), mcp.Description("Path to a saved project JSON file")),
			),
			Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				path, err := request.RequireString("path")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				result, err := cli.RunShowProjectCommand(cli.ShowProjectCommandInput{Path: path})
				if err != nil {
					return mcp.NewToolResultErrorFromErr("show_project failed", err), nil
				}
				return jsonResult(result), nil
			},
		},
	}
}
