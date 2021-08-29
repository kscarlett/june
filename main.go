package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/alecthomas/kong"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var CLI struct {
	Generate struct {
		Input  string `arg name:"file" help:"Input file to generate from." type:"existingfile"`
		Output string `optional help:"Where to output the file." short:"o" default:"public/index.html" type:"path"`
		Ugc    bool   `optional help:"Whether to treat the markdown as untrusted."`
	} `cmd help:"Generate HTML output from Markdown file."`
	Version struct {
	} `cmd help:"Show the current version"`
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
		Generate(CLI.Generate.Input, CLI.Generate.Output, CLI.Generate.Ugc)
	case "version":
		Version()
	default:
		//fmt.Printf("no match: %v\n", ctx.Command())
	}
}

func Generate(input string, output string, ugc bool) {
	source, err := ioutil.ReadFile(input)
	if err != nil {
		panic(err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		os.MkdirAll(path.Dir(output), 0700) // Create your file
	}

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Typographer, extension.Footnote),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
	var buf bytes.Buffer
	if err = md.Convert(source, &buf); err != nil {
		panic(err)
	}

	fmt.Printf("ugc: %v\n", ugc)
	if ugc {
		// use bluemonday to sanitise
	}

	err = ioutil.WriteFile(output, buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}
}

func Version() {
	fmt.Printf("june %s - %s\n" /*version, build*/, "v0.0.0", "dev")
}
