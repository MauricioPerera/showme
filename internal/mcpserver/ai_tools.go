// AITools returns the MCP tools that call showme's AI provider:
// generate_slide_content, generate_storyboard and generate_all_slides.
// Each mirrors the matching CLI command exactly, same "adapter, not a
// second implementation" principle as Tools(); base_url/model are always
// per-call arguments, never server-side configuration.
package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/MauricioPerera/showme/internal/cli"
)

func AITools() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("generate_slide_content",
				mcp.WithDescription("Generate one slide's content with AI, using the project's OKF knowledge as context."),
				mcp.WithString("path", mcp.Required(), mcp.Description("Path to the saved project JSON file")),
				mcp.WithString("slide_id", mcp.Required(), mcp.Description("Id of the slide to generate content for")),
				mcp.WithString("base_url", mcp.Required(), mcp.Description("OpenAI-compatible base URL")),
				mcp.WithString("model", mcp.Required(), mcp.Description("Model name")),
				mcp.WithString("out_dir", mcp.Required(), mcp.Description("Directory to save the updated project")),
			),
			Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				path, err := request.RequireString("path")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				slideID, err := request.RequireString("slide_id")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				baseURL, err := request.RequireString("base_url")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				model, err := request.RequireString("model")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				outDir, err := request.RequireString("out_dir")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				result, err := cli.RunGenerateSlideContentCommand(cli.GenerateSlideContentCommandInput{
					Path:    path,
					SlideID: slideID,
					BaseURL: baseURL,
					Model:   model,
					OutDir:  outDir,
				})
				if err != nil {
					return mcp.NewToolResultErrorFromErr("generate_slide_content failed", err), nil
				}
				return jsonResult(result), nil
			},
		},
		{
			Tool: mcp.NewTool("generate_storyboard",
				mcp.WithDescription("Propose a storyboard with AI and write it as a deck JSON file, ready for create_project."),
				mcp.WithString("objective", mcp.Required(), mcp.Description("Presentation objective")),
				mcp.WithString("audience", mcp.Description("Presentation audience")),
				mcp.WithString("knowledge_root", mcp.Description("Path to an OKF bundle directory for context")),
				mcp.WithString("base_url", mcp.Required(), mcp.Description("OpenAI-compatible base URL")),
				mcp.WithString("model", mcp.Required(), mcp.Description("Model name")),
				mcp.WithString("deck_title", mcp.Required(), mcp.Description("Title for the generated deck")),
				mcp.WithNumber("count", mcp.Required(), mcp.Description("Number of slides to propose")),
				mcp.WithString("out_path", mcp.Required(), mcp.Description("Path to write the generated deck JSON file")),
			),
			Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				objective, err := request.RequireString("objective")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				baseURL, err := request.RequireString("base_url")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				model, err := request.RequireString("model")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				deckTitle, err := request.RequireString("deck_title")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				count, err := request.RequireInt("count")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				outPath, err := request.RequireString("out_path")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				result, err := cli.RunGenerateStoryboardCommand(cli.GenerateStoryboardCommandInput{
					Objective:     objective,
					Audience:      request.GetString("audience", ""),
					KnowledgeRoot: request.GetString("knowledge_root", ""),
					BaseURL:       baseURL,
					Model:         model,
					DeckTitle:     deckTitle,
					Count:         count,
					OutPath:       outPath,
				})
				if err != nil {
					return mcp.NewToolResultErrorFromErr("generate_storyboard failed", err), nil
				}
				return jsonResult(result), nil
			},
		},
		{
			Tool: mcp.NewTool("generate_all_slides",
				mcp.WithDescription("Generate content with AI for every slide of a saved project that has none yet."),
				mcp.WithString("path", mcp.Required(), mcp.Description("Path to the saved project JSON file")),
				mcp.WithString("base_url", mcp.Required(), mcp.Description("OpenAI-compatible base URL")),
				mcp.WithString("model", mcp.Required(), mcp.Description("Model name")),
				mcp.WithString("out_dir", mcp.Required(), mcp.Description("Directory to save the updated project")),
			),
			Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				path, err := request.RequireString("path")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				baseURL, err := request.RequireString("base_url")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				model, err := request.RequireString("model")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				outDir, err := request.RequireString("out_dir")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				result, err := cli.RunGenerateAllSlidesCommand(cli.GenerateAllSlidesCommandInput{
					Path:    path,
					BaseURL: baseURL,
					Model:   model,
					OutDir:  outDir,
				})
				if err != nil {
					return mcp.NewToolResultErrorFromErr("generate_all_slides failed", err), nil
				}
				return jsonResult(result), nil
			},
		},
	}
}
