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

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/MauricioPerera/showme/internal/cli"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "project" || os.Args[2] != "create" {
		fmt.Fprintln(os.Stderr, "usage: showme project create --name <name> --design <path> --knowledge <dir> --deck <path> --out <dir> [--json]")
		os.Exit(1)
	}

	fs := flag.NewFlagSet("project create", flag.ExitOnError)
	name := fs.String("name", "", "project name")
	design := fs.String("design", "", "path to DESIGN.md")
	knowledge := fs.String("knowledge", "", "path to the OKF bundle directory")
	deck := fs.String("deck", "", "path to a deck JSON file")
	out := fs.String("out", "", "directory to save the project")
	asJSON := fs.Bool("json", false, "emit JSON instead of human-readable output")
	_ = fs.Parse(os.Args[3:])

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
		encoded, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(encoded))
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
