package watch

import (
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kscarlett/june/internal/generate"
)

func Run(input, output string, ugc bool, stylePath, templatePath string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error setting up watcher: %w", err)
	}
	defer watcher.Close()

	err = watcher.Add(input)
	if err != nil {
		return fmt.Errorf("error adding file %s to watcher: %w", input, err)
	}

	fmt.Println("Watching for changes. Press Ctrl+C to stop.")
	if errGen := generate.Generate(generate.GenerateConfig{
		Input:    input,
		Output:   output,
		Style:    stylePath,
		Template: templatePath,
		Ugc:      ugc,
	}); errGen != nil {
		fmt.Println("Initial generation error:", errGen)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				// Watcher channel closed
				return nil
			}
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				fmt.Println("File changed, regenerating...")
				time.Sleep(100 * time.Millisecond) // debounce
				if errGen := generate.Generate(generate.GenerateConfig{
					Input:    input,
					Output:   output,
					Style:    stylePath,
					Template: templatePath,
					Ugc:      ugc,
				}); errGen != nil {
					fmt.Println("Generation error:", errGen)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				// Watcher error channel closed
				return nil
			}
			// Log watcher errors but continue running, as they might be transient
			// or related to specific files that can't be watched.
			fmt.Println("Watcher error:", err)
		}
	}
	// Unreachable in the current loop structure, but good form if loop could exit.
	// return nil
}
