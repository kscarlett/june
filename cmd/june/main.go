package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/kscarlett/june/internal/generate"
	"github.com/kscarlett/june/internal/watch"
)

var (
	// These are automatically filled in by goreleaser
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var CLI struct {
	Generate struct {
		Input    string `arg name:"file" help:"Input file to generate from." type:"existingfile"`
		Output   string `optional help:"Where to output the file." short:"o" default:"public/index.html" type:"path"`
		Ugc      bool   `optional help:"Whether to treat the markdown as untrusted."`
		Watch    bool   `optional help:"Watches for changes to your markdown and updates the html."`
		Style    string `optional help:"Path to a CSS file for styling." default:"embedded style"`
		Template string `optional help:"Path to a gohtml template file." default:"embedded template"`
	} `cmd help:"Generate HTML output from Markdown file."`
	Version struct{} `cmd help:"Show the current version"`
}

func main() {
	ctx := kong.Parse(&CLI,
		kong.Name("june"),
		kong.Description("A super simple static page generator."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))

	switch ctx.Command() {
	case "generate <file>":
		if CLI.Generate.Watch {
			// Set up context that cancels on interrupt signal (Ctrl+C)
			ctx, cancel := signal.NotifyContext(
				context.Background(),
				os.Interrupt, syscall.SIGTERM,
			)
			defer cancel()
			if err := watch.Run(
				ctx,
				CLI.Generate.Input,
				CLI.Generate.Output,
				CLI.Generate.Ugc,
				CLI.Generate.Style,
				CLI.Generate.Template,
			); err != nil {
				fmt.Fprintln(os.Stderr, "Error starting watcher:", err)
				os.Exit(1)
			}
		} else {
			if err := generate.Generate(generate.GenerateConfig{
				Input:    CLI.Generate.Input,
				Output:   CLI.Generate.Output,
				Style:    CLI.Generate.Style,
				Template: CLI.Generate.Template,
				Ugc:      CLI.Generate.Ugc,
			}); err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
		}
	case "version":
		fmt.Println(generate.VersionString())
	default:
	}
}
