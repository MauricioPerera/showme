package main

// main is the thin wiring for the showme CLI: parses flags/subcommands and
// prints the result of internal/cli's pure command functions. It is glue,
// not covered by a frozen oracle -- same criterion as tui/main.go.
//
// Usage:
//
//	showme project create --name <name> --design <path/to/DESIGN.md> \
//	    --knowledge <path/to/okf/bundle> --deck <path/to/deck.json> \
//	    --out <output/dir> [--json]
//	showme project list --dir <output/dir> [--json]
//	showme project show --path <path/to/project.json> [--json]
//	showme project export --path <path/to/project.json> --out <path/to/output.html> [--json]
//	showme project duplicate --source <path/to/project.json> --name <new-name> --out <dir> [--json]
//	showme project rename --source <path/to/project.json> --name <new-name> --out <dir> [--json]
//	showme project review --path <path/to/project.json> --slide <id> --decision <accepted|edited|rejected> [--notes <text>] --out <dir> [--json]
//	showme project add-slide --path <path/to/project.json> --slide <id> --title <title> [--intent <text>] [--content <text>] [--status <status>] --out <dir> [--json]
//	showme project remove-slide --path <path/to/project.json> --slide <id> --out <dir> [--json]
//	showme project update-slide --path <path/to/project.json> --slide <id> --title <title> [--intent <text>] [--content <text>] [--status <status>] --out <dir> [--json]
//	showme project reorder-slides --path <path/to/project.json> --order <id1,id2,...> --out <dir> [--json]
//	showme project update-info --path <path/to/project.json> --title <title> [--audience <text>] --out <dir> [--json]
//	showme project generate-slide --path <path/to/project.json> --slide <id> --base-url <ai-base-url> --model <model-name> --out <dir> [--json]
//	showme project archive --path <path/to/project.json> --out <dir> [--json]
//	showme project unarchive --path <path/to/project.json> --out <dir> [--json]

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/MauricioPerera/showme/internal/cli"
)

const usage = "usage: showme project create --name <name> --design <path> --knowledge <dir> --deck <path> --out <dir> [--json]\n" +
	"       showme project list --dir <dir> [--json]\n" +
	"       showme project show --path <path> [--json]\n" +
	"       showme project export --path <path> --out <output.html> [--json]\n" +
	"       showme project duplicate --source <path> --name <new-name> --out <dir> [--json]\n" +
	"       showme project rename --source <path> --name <new-name> --out <dir> [--json]\n" +
	"       showme project review --path <path> --slide <id> --decision <decision> [--notes <text>] --out <dir> [--json]\n" +
	"       showme project add-slide --path <path> --slide <id> --title <title> [--intent <text>] [--content <text>] [--status <status>] --out <dir> [--json]\n" +
	"       showme project remove-slide --path <path> --slide <id> --out <dir> [--json]\n" +
	"       showme project update-slide --path <path> --slide <id> --title <title> [--intent <text>] [--content <text>] [--status <status>] --out <dir> [--json]\n" +
	"       showme project reorder-slides --path <path> --order <id1,id2,...> --out <dir> [--json]\n" +
	"       showme project update-info --path <path> --title <title> [--audience <text>] --out <dir> [--json]\n" +
	"       showme project generate-slide --path <path> --slide <id> --base-url <url> --model <name> --out <dir> [--json]\n" +
	"       showme project archive --path <path> --out <dir> [--json]\n" +
	"       showme project unarchive --path <path> --out <dir> [--json]"

func main() {
	if len(os.Args) < 3 || os.Args[1] != "project" {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}

	switch os.Args[2] {
	case "create":
		runCreate(os.Args[3:])
	case "list":
		runList(os.Args[3:])
	case "show":
		runShow(os.Args[3:])
	case "export":
		runExport(os.Args[3:])
	case "duplicate":
		runDuplicate(os.Args[3:])
	case "rename":
		runRename(os.Args[3:])
	case "review":
		runReview(os.Args[3:])
	case "add-slide":
		runAddSlide(os.Args[3:])
	case "remove-slide":
		runRemoveSlide(os.Args[3:])
	case "update-slide":
		runUpdateSlide(os.Args[3:])
	case "reorder-slides":
		runReorderSlides(os.Args[3:])
	case "update-info":
		runUpdateDeckInfo(os.Args[3:])
	case "generate-slide":
		runGenerateSlide(os.Args[3:])
	case "archive":
		runArchive(os.Args[3:], true)
	case "unarchive":
		runArchive(os.Args[3:], false)
	default:
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(1)
	}
}

func runCreate(args []string) {
	fs := flag.NewFlagSet("project create", flag.ExitOnError)
	name := fs.String("name", "", "project name")
	design := fs.String("design", "", "path to DESIGN.md")
	knowledge := fs.String("knowledge", "", "path to the OKF bundle directory")
	deck := fs.String("deck", "", "path to a deck JSON file")
	out := fs.String("out", "", "directory to save the project")
	asJSON := fs.Bool("json", false, "emit JSON instead of human-readable output")
	_ = fs.Parse(args)

	result, err := cli.RunCreateProjectCommand(cli.CreateProjectCommandInput{
		Name:          *name,
		DesignPath:    *design,
		KnowledgeRoot: *knowledge,
		DeckPath:      *deck,
		OutDir:        *out,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *asJSON {
		printJSON(result)
	} else if result.OK {
		fmt.Printf("OK: saved to %s\n", result.Path)
		for _, w := range result.Warnings {
			fmt.Printf("WARNING: %s\n", w)
		}
	} else {
		for _, e := range result.Errors {
			fmt.Printf("ERROR: %s\n", e)
		}
	}

	if !result.OK {
		os.Exit(1)
	}
}

func runList(args []string) {
	fs := flag.NewFlagSet("project list", flag.ExitOnError)
	dir := fs.String("dir", "", "directory containing saved projects")
	asJSON := fs.Bool("json", false, "emit JSON instead of human-readable output")
	_ = fs.Parse(args)

	result, err := cli.RunListProjectsCommand(cli.ListProjectsCommandInput{Dir: *dir})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *asJSON {
		printJSON(result)
		return
	}

	for _, p := range result.Projects {
		archivedMark := ""
		if p.Archived {
			archivedMark = " [archived]"
		}
		fmt.Printf("%s\t%s%s\n", p.Name, p.Path, archivedMark)
	}
	for _, e := range result.Errors {
		fmt.Printf("ERROR: %s\n", e)
	}
}

func runShow(args []string) {
	fs := flag.NewFlagSet("project show", flag.ExitOnError)
	path := fs.String("path", "", "path to a saved project JSON file")
	asJSON := fs.Bool("json", false, "emit JSON instead of human-readable output")
	_ = fs.Parse(args)

	result, err := cli.RunShowProjectCommand(cli.ShowProjectCommandInput{Path: *path})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *asJSON {
		printJSON(result)
		return
	}

	proj := result.Project
	archivedMark := ""
	if proj.Archived {
		archivedMark = " [archived]"
	}
	fmt.Printf("%s (v%d)%s\n", proj.Name, proj.Version, archivedMark)
	fmt.Printf("design: %s\n", proj.DesignPath)
	fmt.Printf("knowledge: %s\n", proj.KnowledgePath)
	for _, s := range proj.Deck.Slides {
		fmt.Printf("- [%s] %s (%s)\n", s.ID, s.Title, s.Status)
	}
}

func runExport(args []string) {
	fs := flag.NewFlagSet("project export", flag.ExitOnError)
	path := fs.String("path", "", "path to a saved project JSON file")
	out := fs.String("out", "", "path to the HTML file to write")
	asJSON := fs.Bool("json", false, "emit JSON instead of human-readable output")
	_ = fs.Parse(args)

	result, err := cli.RunExportProjectCommand(cli.ExportProjectCommandInput{
		Path:    *path,
		OutPath: *out,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *asJSON {
		printJSON(result)
	} else {
		fmt.Printf("OK: exported to %s\n", result.Path)
	}
}

func runDuplicate(args []string) {
	fs := flag.NewFlagSet("project duplicate", flag.ExitOnError)
	source := fs.String("source", "", "path to the project to duplicate")
	name := fs.String("name", "", "name for the duplicated project")
	out := fs.String("out", "", "directory to save the duplicate")
	asJSON := fs.Bool("json", false, "emit JSON instead of human-readable output")
	_ = fs.Parse(args)

	result, err := cli.RunDuplicateProjectCommand(cli.DuplicateProjectCommandInput{
		SourcePath: *source,
		NewName:    *name,
		OutDir:     *out,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *asJSON {
		printJSON(result)
	} else if result.OK {
		fmt.Printf("OK: saved to %s\n", result.Path)
	} else {
		for _, e := range result.Errors {
			fmt.Printf("ERROR: %s\n", e)
		}
	}

	if !result.OK {
		os.Exit(1)
	}
}

func runRename(args []string) {
	fs := flag.NewFlagSet("project rename", flag.ExitOnError)
	source := fs.String("source", "", "path to the project to rename")
	name := fs.String("name", "", "new name for the project")
	out := fs.String("out", "", "directory to save the renamed project")
	asJSON := fs.Bool("json", false, "emit JSON instead of human-readable output")
	_ = fs.Parse(args)

	result, err := cli.RunRenameProjectCommand(cli.RenameProjectCommandInput{
		SourcePath: *source,
		NewName:    *name,
		OutDir:     *out,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *asJSON {
		printJSON(result)
	} else if result.OK {
		fmt.Printf("OK: saved to %s\n", result.Path)
	} else {
		for _, e := range result.Errors {
			fmt.Printf("ERROR: %s\n", e)
		}
	}

	if !result.OK {
		os.Exit(1)
	}
}

func runReview(args []string) {
	fs := flag.NewFlagSet("project review", flag.ExitOnError)
	path := fs.String("path", "", "path to the project to review")
	slide := fs.String("slide", "", "slide id being reviewed")
	decision := fs.String("decision", "", "accepted, edited or rejected")
	notes := fs.String("notes", "", "optional free-text notes")
	out := fs.String("out", "", "directory to save the updated project")
	asJSON := fs.Bool("json", false, "emit JSON instead of human-readable output")
	_ = fs.Parse(args)

	result, err := cli.RunReviewProjectCommand(cli.ReviewProjectCommandInput{
		Path:     *path,
		SlideID:  *slide,
		Decision: *decision,
		Notes:    *notes,
		OutDir:   *out,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *asJSON {
		printJSON(result)
	} else if result.OK {
		fmt.Printf("OK: saved to %s\n", result.Path)
	} else {
		for _, e := range result.Errors {
			fmt.Printf("ERROR: %s\n", e)
		}
	}

	if !result.OK {
		os.Exit(1)
	}
}

func runAddSlide(args []string) {
	fs := flag.NewFlagSet("project add-slide", flag.ExitOnError)
	path := fs.String("path", "", "path to the project to update")
	slide := fs.String("slide", "", "id for the new slide")
	title := fs.String("title", "", "title for the new slide")
	intent := fs.String("intent", "", "optional intent for the new slide")
	content := fs.String("content", "", "optional content for the new slide")
	status := fs.String("status", "", "optional status for the new slide (defaults to draft)")
	out := fs.String("out", "", "directory to save the updated project")
	asJSON := fs.Bool("json", false, "emit JSON instead of human-readable output")
	_ = fs.Parse(args)

	result, err := cli.RunAddSlideCommand(cli.AddSlideCommandInput{
		Path:    *path,
		SlideID: *slide,
		Title:   *title,
		Intent:  *intent,
		Content: *content,
		Status:  *status,
		OutDir:  *out,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *asJSON {
		printJSON(result)
	} else if result.OK {
		fmt.Printf("OK: saved to %s\n", result.Path)
	} else {
		for _, e := range result.Errors {
			fmt.Printf("ERROR: %s\n", e)
		}
	}

	if !result.OK {
		os.Exit(1)
	}
}

func runRemoveSlide(args []string) {
	fs := flag.NewFlagSet("project remove-slide", flag.ExitOnError)
	path := fs.String("path", "", "path to the project to update")
	slide := fs.String("slide", "", "id of the slide to remove")
	out := fs.String("out", "", "directory to save the updated project")
	asJSON := fs.Bool("json", false, "emit JSON instead of human-readable output")
	_ = fs.Parse(args)

	result, err := cli.RunRemoveSlideCommand(cli.RemoveSlideCommandInput{
		Path:    *path,
		SlideID: *slide,
		OutDir:  *out,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *asJSON {
		printJSON(result)
	} else if result.OK {
		fmt.Printf("OK: saved to %s\n", result.Path)
	} else {
		for _, e := range result.Errors {
			fmt.Printf("ERROR: %s\n", e)
		}
	}

	if !result.OK {
		os.Exit(1)
	}
}

func runUpdateSlide(args []string) {
	fs := flag.NewFlagSet("project update-slide", flag.ExitOnError)
	path := fs.String("path", "", "path to the project to update")
	slide := fs.String("slide", "", "id of the slide to update")
	title := fs.String("title", "", "new title for the slide")
	intent := fs.String("intent", "", "new intent for the slide")
	content := fs.String("content", "", "new content for the slide")
	status := fs.String("status", "", "new status for the slide (leave empty to keep the current one)")
	out := fs.String("out", "", "directory to save the updated project")
	asJSON := fs.Bool("json", false, "emit JSON instead of human-readable output")
	_ = fs.Parse(args)

	result, err := cli.RunUpdateSlideCommand(cli.UpdateSlideCommandInput{
		Path:    *path,
		SlideID: *slide,
		Title:   *title,
		Intent:  *intent,
		Content: *content,
		Status:  *status,
		OutDir:  *out,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *asJSON {
		printJSON(result)
	} else if result.OK {
		fmt.Printf("OK: saved to %s\n", result.Path)
	} else {
		for _, e := range result.Errors {
			fmt.Printf("ERROR: %s\n", e)
		}
	}

	if !result.OK {
		os.Exit(1)
	}
}

func runReorderSlides(args []string) {
	fs := flag.NewFlagSet("project reorder-slides", flag.ExitOnError)
	path := fs.String("path", "", "path to the project to update")
	order := fs.String("order", "", "comma-separated list of every slide id in the desired order")
	out := fs.String("out", "", "directory to save the updated project")
	asJSON := fs.Bool("json", false, "emit JSON instead of human-readable output")
	_ = fs.Parse(args)

	var orderList []string
	if *order != "" {
		orderList = strings.Split(*order, ",")
	}

	result, err := cli.RunReorderSlidesCommand(cli.ReorderSlidesCommandInput{
		Path:   *path,
		Order:  orderList,
		OutDir: *out,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *asJSON {
		printJSON(result)
	} else if result.OK {
		fmt.Printf("OK: saved to %s\n", result.Path)
	} else {
		for _, e := range result.Errors {
			fmt.Printf("ERROR: %s\n", e)
		}
	}

	if !result.OK {
		os.Exit(1)
	}
}

func runUpdateDeckInfo(args []string) {
	fs := flag.NewFlagSet("project update-info", flag.ExitOnError)
	path := fs.String("path", "", "path to the project to update")
	title := fs.String("title", "", "new title for the deck")
	audience := fs.String("audience", "", "new audience for the deck")
	out := fs.String("out", "", "directory to save the updated project")
	asJSON := fs.Bool("json", false, "emit JSON instead of human-readable output")
	_ = fs.Parse(args)

	result, err := cli.RunUpdateDeckInfoCommand(cli.UpdateDeckInfoCommandInput{
		Path:     *path,
		Title:    *title,
		Audience: *audience,
		OutDir:   *out,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *asJSON {
		printJSON(result)
	} else if result.OK {
		fmt.Printf("OK: saved to %s\n", result.Path)
	} else {
		for _, e := range result.Errors {
			fmt.Printf("ERROR: %s\n", e)
		}
	}

	if !result.OK {
		os.Exit(1)
	}
}

func runGenerateSlide(args []string) {
	fs := flag.NewFlagSet("project generate-slide", flag.ExitOnError)
	path := fs.String("path", "", "path to the project to update")
	slide := fs.String("slide", "", "id of the slide to generate content for")
	baseURL := fs.String("base-url", "", "base URL of the OpenAI-compatible AI provider")
	model := fs.String("model", "", "model name to request from the AI provider")
	out := fs.String("out", "", "directory to save the updated project")
	asJSON := fs.Bool("json", false, "emit JSON instead of human-readable output")
	_ = fs.Parse(args)

	result, err := cli.RunGenerateSlideContentCommand(cli.GenerateSlideContentCommandInput{
		Path:    *path,
		SlideID: *slide,
		BaseURL: *baseURL,
		Model:   *model,
		OutDir:  *out,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *asJSON {
		printJSON(result)
	} else if result.OK {
		fmt.Printf("OK: saved to %s\n\n%s\n", result.Path, result.Content)
	} else {
		for _, e := range result.Errors {
			fmt.Printf("ERROR: %s\n", e)
		}
	}

	if !result.OK {
		os.Exit(1)
	}
}

func runArchive(args []string, archived bool) {
	name := "project archive"
	if !archived {
		name = "project unarchive"
	}
	fs := flag.NewFlagSet(name, flag.ExitOnError)
	path := fs.String("path", "", "path to the project to update")
	out := fs.String("out", "", "directory to save the updated project")
	asJSON := fs.Bool("json", false, "emit JSON instead of human-readable output")
	_ = fs.Parse(args)

	result, err := cli.RunArchiveProjectCommand(cli.ArchiveProjectCommandInput{
		Path:     *path,
		Archived: archived,
		OutDir:   *out,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *asJSON {
		printJSON(result)
	} else if result.OK {
		fmt.Printf("OK: saved to %s\n", result.Path)
	} else {
		for _, e := range result.Errors {
			fmt.Printf("ERROR: %s\n", e)
		}
	}

	if !result.OK {
		os.Exit(1)
	}
}

func printJSON(v any) {
	encoded, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(encoded))
}
