// SlideTools returns the MCP tools that edit a project's slides/deck in
// place: add_slide, remove_slide, update_slide, reorder_slides and
// update_deck_info. Each mirrors the matching CLI command exactly, same
// "adapter, not a second implementation" principle as Tools().
package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/MauricioPerera/showme/internal/cli"
)

func SlideTools() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("add_slide",
				mcp.WithDescription("Add a new slide to a saved showme project's deck."),
				mcp.WithString("path", mcp.Required(), mcp.Description("Path to the saved project JSON file")),
				mcp.WithString("slide_id", mcp.Required(), mcp.Description("Unique id for the new slide")),
				mcp.WithString("title", mcp.Required(), mcp.Description("Slide title")),
				mcp.WithString("intent", mcp.Description("Slide intent")),
				mcp.WithString("content", mcp.Description("Slide content")),
				mcp.WithString("status", mcp.Description("Slide status")),
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
				title, err := request.RequireString("title")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				outDir, err := request.RequireString("out_dir")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				result, err := cli.RunAddSlideCommand(cli.AddSlideCommandInput{
					Path:    path,
					SlideID: slideID,
					Title:   title,
					Intent:  request.GetString("intent", ""),
					Content: request.GetString("content", ""),
					Status:  request.GetString("status", ""),
					OutDir:  outDir,
				})
				if err != nil {
					return mcp.NewToolResultErrorFromErr("add_slide failed", err), nil
				}
				return jsonResult(result), nil
			},
		},
		{
			Tool: mcp.NewTool("remove_slide",
				mcp.WithDescription("Remove a slide from a saved showme project's deck."),
				mcp.WithString("path", mcp.Required(), mcp.Description("Path to the saved project JSON file")),
				mcp.WithString("slide_id", mcp.Required(), mcp.Description("Id of the slide to remove")),
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
				outDir, err := request.RequireString("out_dir")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				result, err := cli.RunRemoveSlideCommand(cli.RemoveSlideCommandInput{
					Path:    path,
					SlideID: slideID,
					OutDir:  outDir,
				})
				if err != nil {
					return mcp.NewToolResultErrorFromErr("remove_slide failed", err), nil
				}
				return jsonResult(result), nil
			},
		},
		{
			Tool: mcp.NewTool("update_slide",
				mcp.WithDescription("Replace one of a saved showme project's slides."),
				mcp.WithString("path", mcp.Required(), mcp.Description("Path to the saved project JSON file")),
				mcp.WithString("slide_id", mcp.Required(), mcp.Description("Id of the slide to update")),
				mcp.WithString("title", mcp.Required(), mcp.Description("Slide title")),
				mcp.WithString("intent", mcp.Description("Slide intent")),
				mcp.WithString("content", mcp.Description("Slide content")),
				mcp.WithString("status", mcp.Description("Slide status")),
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
				title, err := request.RequireString("title")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				outDir, err := request.RequireString("out_dir")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				result, err := cli.RunUpdateSlideCommand(cli.UpdateSlideCommandInput{
					Path:    path,
					SlideID: slideID,
					Title:   title,
					Intent:  request.GetString("intent", ""),
					Content: request.GetString("content", ""),
					Status:  request.GetString("status", ""),
					OutDir:  outDir,
				})
				if err != nil {
					return mcp.NewToolResultErrorFromErr("update_slide failed", err), nil
				}
				return jsonResult(result), nil
			},
		},
		{
			Tool: mcp.NewTool("reorder_slides",
				mcp.WithDescription("Reorder a saved showme project's slides."),
				mcp.WithString("path", mcp.Required(), mcp.Description("Path to the saved project JSON file")),
				mcp.WithArray("order", mcp.Required(), mcp.Description("Slide ids in the desired order")),
				mcp.WithString("out_dir", mcp.Required(), mcp.Description("Directory to save the updated project")),
			),
			Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				path, err := request.RequireString("path")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				order, err := request.RequireStringSlice("order")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				outDir, err := request.RequireString("out_dir")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				result, err := cli.RunReorderSlidesCommand(cli.ReorderSlidesCommandInput{
					Path:   path,
					Order:  order,
					OutDir: outDir,
				})
				if err != nil {
					return mcp.NewToolResultErrorFromErr("reorder_slides failed", err), nil
				}
				return jsonResult(result), nil
			},
		},
		{
			Tool: mcp.NewTool("update_deck_info",
				mcp.WithDescription("Update a saved showme project's deck title and audience."),
				mcp.WithString("path", mcp.Required(), mcp.Description("Path to the saved project JSON file")),
				mcp.WithString("title", mcp.Required(), mcp.Description("Deck title")),
				mcp.WithString("audience", mcp.Description("Deck audience")),
				mcp.WithString("out_dir", mcp.Required(), mcp.Description("Directory to save the updated project")),
			),
			Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				path, err := request.RequireString("path")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				title, err := request.RequireString("title")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				outDir, err := request.RequireString("out_dir")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				result, err := cli.RunUpdateDeckInfoCommand(cli.UpdateDeckInfoCommandInput{
					Path:     path,
					Title:    title,
					Audience: request.GetString("audience", ""),
					OutDir:   outDir,
				})
				if err != nil {
					return mcp.NewToolResultErrorFromErr("update_deck_info failed", err), nil
				}
				return jsonResult(result), nil
			},
		},
	}
}
