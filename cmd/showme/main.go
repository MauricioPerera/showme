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
//	showme project duplicate --source <path/to/project.json> --name <new-name> --out <dir> [--json]
//	showme project review --path <path/to/project.json> --slide <id> --decision <accepted|edited|rejected> [--notes <text>] --out <dir> [--json]
//	showme project add-slide --path <path/to/project.json> --slide <id> --title <title> [--intent <text>] [--content <text>] [--status <status>] --out <dir> [--json]

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/MauricioPerera/showme/internal/cli"
)

const usage = "usage: showme project create --name <name> --design <path> --knowledge <dir> --deck <path> --out <dir> [--json]\n" +
	"       showme project list --dir <dir> [--json]\n" +
	"       showme project show --path <path> [--json]\n" +
	"       showme project duplicate --source <path> --name <new-name> --out <dir> [--json]\n" +
	"       showme project review --path <path> --slide <id> --decision <decision> [--notes <text>] --out <dir> [--json]\n" +
	"       showme project add-slide --path <path> --slide <id> --title <title> [--intent <text>] [--content <text>] [--status <status>] --out <dir> [--json]"

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
	case "duplicate":
		runDuplicate(os.Args[3:])
	case "review":
		runReview(os.Args[3:])
	case "add-slide":
		runAddSlide(os.Args[3:])
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
		fmt.Printf("%s\t%s\n", p.Name, p.Path)
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
	fmt.Printf("%s (v%d)\n", proj.Name, proj.Version)
	fmt.Printf("design: %s\n", proj.DesignPath)
	fmt.Printf("knowledge: %s\n", proj.KnowledgePath)
	for _, s := range proj.Deck.Slides {
		fmt.Printf("- [%s] %s (%s)\n", s.ID, s.Title, s.Status)
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

func printJSON(v any) {
	encoded, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(encoded))
}
