# June

June is a super simple single page static site generator for Markdown files, written in Go. It is designed for quick, automated generation of single-page sites, landing pages, or any Markdown-to-HTML workflow — especially in automated deployments.

## Features

- **Markdown to HTML**: Instantly convert Markdown files to static HTML.
- **Content-focused**: Write your content as a regular Markdown file and get HTML out.
- **Frontmatter Support**: Optionally add metadata like title, description, and language with YAML to enhance the output.
- **Simple**: Fully self contained. Includes an embedded default template to let you generate pages with one binary and one Markdown file.
- **Custom Templates & Styles**: Use your own Go HTML templates and CSS, or stick with the built-in defaults.
- **User-Generated Content Mode**: Want to integrate June to publish untrusted Markdown? Enable a single flag to sanitise input.

## CLI Usage

```sh
june generate <input.md> [-o public/output.html] [--style ./custom.css] [--template ./template.gohtml] [--ugc] [--watch]
                                ^ give a default too   ^ switches theme     ^ optional custom template    ^ sanitises markdown as UGC
```

```sh
june version
```

## Example

Given a Markdown file with frontmatter:

```markdown
---
title: "My Page"
description: "A simple static page"
lang: "en"
---
# Hello World

Welcome to my page!
```

Run:

```sh
june generate mypage.md
```

This will produce `public/index.html` using the default template and style.

## Customization

- **Custom CSS**:  
  Use `--style ./your.css` to apply your own CSS file.
- **Custom Template**:  
  Use `--template ./your.gohtml` to use a custom Go HTML template.  
  The template receives all frontmatter fields, `.Content` (HTML), and `.Style` (CSS).

## Frontmatter Fields

- `title`: Sets the HTML `<title>`.
- `description`: Sets the meta description.
- `lang`: Sets the `<html lang="">` attribute.
- `tags`: (optional) Array of tags.

## Sanitization

Use `--ugc` to treat the Markdown as untrusted user content. This strips all HTML and only allows safe Markdown.

## Watch Mode

Use `--watch` to keep June running and regenerate the output HTML whenever the input Markdown file changes.

## Installation

Download a release from [GitHub Releases](https://github.com/kscarlett/june/releases) or build from source:

```sh
go install github.com/kscarlett/june/cmd/june@latest
```

## Contributing

Contributions are very welcome! If you have ideas, bug fixes, or improvements, please open a pull request.  
For questions or suggestions, feel free to open an issue.

## License

MIT

---

For more details, see the [examples](./examples/) directory or open an issue!