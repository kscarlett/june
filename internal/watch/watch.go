package watch

import (
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kscarlett/june/internal/generate"
)

func Run(input, output string, ugc bool, stylePath, templatePath string) {
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
	generate.Generate(generate.GenerateConfig{
		Input:    input,
		Output:   output,
		Style:    stylePath,
		Template: templatePath,
		Ugc:      ugc,
	}) // Initial generation

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				fmt.Println("File changed, regenerating...")
				time.Sleep(100 * time.Millisecond) // debounce
				generate.Generate(generate.GenerateConfig{
					Input:    input,
					Output:   output,
					Style:    stylePath,
					Template: templatePath,
					Ugc:      ugc,
				})
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("Watcher error:", err)
		}
	}
}
