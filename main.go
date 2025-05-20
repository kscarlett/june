package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"os"
	"path"
	"time"

	"github.com/alecthomas/kong"
	"github.com/fsnotify/fsnotify"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/frontmatter"
)

var (
	// These are automatically filled in by goreleaser
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// access templates and styles as embeddedFiles.ReadFile("files/styles/light.css")
//
//go:embed files/*
var embeddedFiles embed.FS

type PageMeta struct {
	Title string   `yaml:"title"`
	Desc  string   `yaml:"description"`
	Lang  string   `yaml:"lang"`
	Tags  []string `yaml:"tags"`
}

var CLI struct {
	Generate struct {
		Input    string `arg name:"file" help:"Input file to generate from." type:"existingfile"`
		Output   string `optional help:"Where to output the file." short:"o" default:"public/index.html" type:"path"`
		Ugc      bool   `optional help:"Whether to treat the markdown as untrusted."`
		Watch    bool   `optional help:"Watches for changes to your markdown and updates the html."`
		Style    string `optional help:"Path to a CSS/SCSS file for styling." default:"embedded style"`
		Template string `optional help:"Path to a gohtml template file." default:"embedded template"`
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
		Run(
			CLI.Generate.Input,
			CLI.Generate.Output,
			CLI.Generate.Ugc,
			CLI.Generate.Watch,
			CLI.Generate.Style,
			CLI.Generate.Template,
		)
	case "version":
		Version()
	default:
		//fmt.Printf("no match: %v\n", ctx.Command())
	}
}

func Run(input, output string, ugc, watch bool, stylePath, templatePath string) {
	if watch {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			panic(err)
		}
		defer watcher.Close()

		err = watcher.Add(input)
		if err != nil {
			panic(err)
		}

		fmt.Println("Watching for changes. Press Ctrl+C to stop.")
		Generate(input, output, ugc, stylePath, templatePath) // Initial generation

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					fmt.Println("File changed, regenerating...")
					time.Sleep(100 * time.Millisecond) // debounce
					Generate(input, output, ugc, stylePath, templatePath)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("Watcher error:", err)
			}
		}
	} else {
		Generate(input, output, ugc, stylePath, templatePath)
	}
}

func Generate(input, output string, ugc bool, stylePath, templatePath string) {
	fmt.Printf("watch: %v\n", false)

	source, err := os.ReadFile(input)
	if err != nil {
		panic(err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		os.MkdirAll(path.Dir(output), 0700)
	}

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM,
			extension.Typographer,
			extension.Footnote,
			&frontmatter.Extender{}),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	var buf bytes.Buffer
	ctx := parser.NewContext()
	if err = md.Convert(source, &buf, parser.WithContext(ctx)); err != nil {
		panic(err)
	}

	var metadata PageMeta
	d := frontmatter.Get(ctx)
	if err := d.Decode(&metadata); err != nil {
		fmt.Printf("error decoding frontmatter: %v\n", err)
		return
	}
	if metadata.Lang == "" {
		metadata.Lang = "en"
	}

	generated := buf.Bytes()
	if ugc {
		p := bluemonday.UGCPolicy()
		generated = p.SanitizeBytes(generated)
	}

	// --- Load template ---
	var tmplBytes []byte
	if _, err := os.Stat(templatePath); err == nil {
		tmplBytes, err = os.ReadFile(templatePath)
		if err != nil {
			panic(fmt.Sprintf("failed to read template: %v", err))
		}
	} else {
		tmplBytes, err = embeddedFiles.ReadFile("files/templates/basic.gohtml")
		if err != nil {
			panic(fmt.Sprintf("failed to read embedded template: %v", err))
		}
	}

	// --- Load style ---
	var styleBytes []byte
	if _, err := os.Stat(stylePath); err == nil {
		styleBytes, err = os.ReadFile(stylePath)
		if err != nil {
			panic(fmt.Sprintf("failed to read style: %v", err))
		}
	} else {
		styleBytes, err = embeddedFiles.ReadFile("files/styles/simple.css")
		if err != nil {
			panic(fmt.Sprintf("failed to read embedded style: %v", err))
		}
	}
	css := string(styleBytes)

	tmpl, err := template.New("page").Parse(string(tmplBytes))
	if err != nil {
		panic(fmt.Sprintf("failed to parse template: %v", err))
	}

	// Prepare data for template
	data := struct {
		PageMeta
		Content template.HTML
		Style   template.CSS
	}{
		PageMeta: metadata,
		Content:  template.HTML(generated),
		Style:    template.CSS(css),
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		panic(fmt.Sprintf("failed to execute template: %v", err))
	}

	err = os.WriteFile(output, out.Bytes(), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Successfully wrote to %s\n", output)
}

func Version() {
	fmt.Printf("june version %s - commit %s (built at %s)\n", version, commit, date)
}
