# Technical Debt

This document outlines known areas of technical debt in the `june` project. Addressing these items can improve maintainability, testability, and robustness.

## 1. Integration Testing for `internal/generate.Generate` function

The `internal/generate.Generate` function is a core part of this application. It orchestrates several key operations:
- Reading the input Markdown file.
- Parsing the Markdown content (which includes handling frontmatter).
- Loading the HTML template (either user-provided or embedded).
- Loading the CSS style (either user-provided or embedded).
- Executing the template with the parsed content and metadata.
- Writing the final HTML to an output file.

Currently, the `internal/generate.parseMarkdown` sub-component has its own unit tests. However, the `Generate` function itself, which ties everything together, lacks comprehensive tests. Specifically, there are no tests that cover:
- The full end-to-end generation process.
- File I/O operations (reading input, creating output directories, writing output files).
- Interactions with the template loading and execution system (`internal/template`).
- Correct handling of various `GenerateConfig` options in the context of the full generation.

**Suggestion:**
Implement integration tests for `generate.Generate`. These tests would involve:
- Creating temporary input Markdown files with various valid and invalid configurations.
- Running `generate.Generate` with different `GenerateConfig` settings.
- Verifying the content of the output HTML files.
- Checking that errors are correctly propagated from file operations or template processing.

## 2. Advanced Testability for `internal/watch.Run` function

The `internal/watch.Run` function provides the file watching capability for live regeneration of HTML. As analyzed in `internal/watch/watch_test.go`, its current implementation presents significant challenges for unit testing:
- **Direct `fsnotify` Dependency:** The function directly uses the `fsnotify` package, which interacts with the operating system's file event system. This makes it difficult to simulate file events or watcher errors reliably in a unit test environment.
- **Inline Event Loop:** The main `for { select { ... } }` loop is part of the `Run` function, making it hard to test event handling logic in isolation or control the loop's execution during tests.
- **Concrete Types:** `fsnotify.Watcher` is a concrete type, preventing easy substitution with mocks or fakes.

**Suggestion:**
Refactor `internal/watch.Run` to improve its testability, incorporating suggestions from the analysis in `internal/watch/watch_test.go`:
- **Introduce a `FileWatcher` interface:** This interface would abstract the necessary methods from `fsnotify.Watcher` (e.g., `Events()`, `Errors()`, `Add()`, `Close()`). `watch.Run` would then depend on this interface, allowing mock implementations to be injected during tests.
- **Extract event handling logic:** The core logic within the `select` statement that processes file events and triggers regeneration could be moved to a separate, testable function or method. This function could take parameters like the event details, generation configuration, and a way to trigger the generation.
- **Use `context.Context`:** Introduce `context.Context` into `watch.Run` to allow for graceful cancellation and better control over the watch loop's lifecycle, especially in test scenarios.
- **Decouple Logging and Generation Calls:** Instead of directly calling `fmt.Println` and `generate.Generate`, these dependencies could be injected (e.g., via a logger interface and a function type for the generation call), making the watch logic more focused and easier to test in isolation.

## 3. Structured Logging

Currently, the application uses `fmt.Println` for informational messages (e.g., successful writes, watcher status) and `fmt.Fprintln(os.Stderr, ...)` for error reporting in `main.go`. While functional for a simple CLI, this approach has limitations:
- **Lack of Levels:** No distinction between debug, info, warning, error levels.
- **Inconsistent Formatting:** Output format is ad-hoc.
- **Difficulty in Filtering/Parsing:** Plain text output is harder to parse for monitoring or redirect to different sinks (file, syslog, etc.).

This is particularly relevant for the `watch` command, which runs as a longer-lived process where better logging would aid in monitoring and debugging.

**Suggestion:**
Implement a more structured logging mechanism. Options include:
- Using the standard Go `log` package with custom output formatting or flags.
- Adopting a third-party logging library (e.g., Logrus, Zap, Zerolog) that provides features like leveled logging, structured output (JSON, key-value pairs), and easier configuration for different output destinations.
This would improve debuggability, make it easier to manage log output, and align with common practices for application logging.

## 4. Configuration Management for `GenerateConfig`

The `GenerateConfig` struct, which holds settings for the HTML generation process (input/output paths, template, style, UGC mode), is currently populated directly from command-line flags parsed by the `kong` library in `cmd/june/main.go`.

While this is suitable for basic command-line usage, it has limitations:
- **Limited Flexibility:** All configuration must be passed via CLI flags.
- **Scalability:** As more configuration options are added, the list of CLI flags could become unwieldy.
- **No Persistent Configuration:** Users cannot save a common set of configurations for a project.

**Suggestion:**
Consider implementing a more robust configuration loading mechanism. This could involve:
- **Configuration File:** Allow loading `GenerateConfig` settings from a configuration file (e.g., YAML, TOML, JSON) in the project directory or a user-specified path.
- **Layered Configuration:** Implement a system where configurations can be overridden (e.g., defaults < config file < CLI flags).
- **Dedicated Configuration Package:** Potentially extract configuration loading and management into its own internal package if complexity grows.
This would provide users with more flexibility and make the application more versatile for different project setups and future enhancements.
