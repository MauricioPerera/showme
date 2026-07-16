// ProjectManagementTools returns the MCP tools that manage a saved showme
// project as a whole: archive_project, duplicate_project, rename_project,
// review_project and export_project. Each mirrors the matching CLI command
// exactly, same "adapter, not a second implementation" principle as
// Tools().
package mcpserver

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/MauricioPerera/showme/internal/cli"
)

func ProjectManagementTools() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("archive_project",
				mcp.WithDescription("Set a saved showme project's archived state."),
				mcp.WithString("path", mcp.Required(), mcp.Description("Path to the saved project JSON file")),
				mcp.WithBoolean("archived", mcp.Required(), mcp.Description("Desired archived state")),
				mcp.WithString("out_dir", mcp.Required(), mcp.Description("Directory to save the updated project")),
			),
			Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				path, err := request.RequireString("path")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				archived, err := request.RequireBool("archived")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				outDir, err := request.RequireString("out_dir")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				result, err := cli.RunArchiveProjectCommand(cli.ArchiveProjectCommandInput{
					Path:     path,
					Archived: archived,
					OutDir:   outDir,
				})
				if err != nil {
					return mcp.NewToolResultErrorFromErr("archive_project failed", err), nil
				}
				return jsonResult(result), nil
			},
		},
		{
			Tool: mcp.NewTool("duplicate_project",
				mcp.WithDescription("Duplicate a saved showme project under a new name."),
				mcp.WithString("source_path", mcp.Required(), mcp.Description("Path to the saved project JSON file to duplicate")),
				mcp.WithString("new_name", mcp.Required(), mcp.Description("Name for the duplicated project")),
				mcp.WithString("out_dir", mcp.Required(), mcp.Description("Directory to save the duplicated project")),
			),
			Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				sourcePath, err := request.RequireString("source_path")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				newName, err := request.RequireString("new_name")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				outDir, err := request.RequireString("out_dir")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				result, err := cli.RunDuplicateProjectCommand(cli.DuplicateProjectCommandInput{
					SourcePath: sourcePath,
					NewName:    newName,
					OutDir:     outDir,
				})
				if err != nil {
					return mcp.NewToolResultErrorFromErr("duplicate_project failed", err), nil
				}
				return jsonResult(result), nil
			},
		},
		{
			Tool: mcp.NewTool("rename_project",
				mcp.WithDescription("Rename a saved showme project."),
				mcp.WithString("source_path", mcp.Required(), mcp.Description("Path to the saved project JSON file to rename")),
				mcp.WithString("new_name", mcp.Required(), mcp.Description("New name for the project")),
				mcp.WithString("out_dir", mcp.Required(), mcp.Description("Directory the project is saved under")),
			),
			Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				sourcePath, err := request.RequireString("source_path")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				newName, err := request.RequireString("new_name")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				outDir, err := request.RequireString("out_dir")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				result, err := cli.RunRenameProjectCommand(cli.RenameProjectCommandInput{
					SourcePath: sourcePath,
					NewName:    newName,
					OutDir:     outDir,
				})
				if err != nil {
					return mcp.NewToolResultErrorFromErr("rename_project failed", err), nil
				}
				return jsonResult(result), nil
			},
		},
		{
			Tool: mcp.NewTool("review_project",
				mcp.WithDescription("Apply a review decision to one of a saved showme project's slides."),
				mcp.WithString("path", mcp.Required(), mcp.Description("Path to the saved project JSON file")),
				mcp.WithString("slide_id", mcp.Required(), mcp.Description("Id of the slide being reviewed")),
				mcp.WithString("decision", mcp.Required(), mcp.Description("Review decision")),
				mcp.WithString("notes", mcp.Description("Review notes")),
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
				decision, err := request.RequireString("decision")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				outDir, err := request.RequireString("out_dir")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				result, err := cli.RunReviewProjectCommand(cli.ReviewProjectCommandInput{
					Path:     path,
					SlideID:  slideID,
					Decision: decision,
					Notes:    request.GetString("notes", ""),
					OutDir:   outDir,
				})
				if err != nil {
					return mcp.NewToolResultErrorFromErr("review_project failed", err), nil
				}
				return jsonResult(result), nil
			},
		},
		{
			Tool: mcp.NewTool("export_project",
				mcp.WithDescription("Export a saved showme project to an HTML file."),
				mcp.WithString("path", mcp.Required(), mcp.Description("Path to the saved project JSON file")),
				mcp.WithString("out_path", mcp.Required(), mcp.Description("Path to write the exported HTML file")),
			),
			Handler: func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				path, err := request.RequireString("path")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}
				outPath, err := request.RequireString("out_path")
				if err != nil {
					return mcp.NewToolResultError(err.Error()), nil
				}

				result, err := cli.RunExportProjectCommand(cli.ExportProjectCommandInput{
					Path:    path,
					OutPath: outPath,
				})
				if err != nil {
					return mcp.NewToolResultErrorFromErr("export_project failed", err), nil
				}
				return jsonResult(result), nil
			},
		},
	}
}
