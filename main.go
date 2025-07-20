package main

import (
	"fmt"
    "path/filepath"
    "io/fs"
	"github.com/fsnotify/fsnotify"
)

func main() {
	err := startRuntime("./testmd", "./out")
	if err != nil {
        fmt.Printf("MARKI ERROR: %s\n", err.Error())
        main()
	}
}

func startRuntime(inDir string, outDir string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	err = watcher.Add(inDir)
	if err != nil {
		return err
	}
	errChan := make(chan error, 1)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					errChan <- fmt.Errorf("fsnotify: event channels closed")
					return
				}
				err := onChange(inDir, outDir, event)
				if err != nil {
					errChan <- err
					return
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					errChan <- fmt.Errorf("fsnotify: error channels closed")
					return
				}
				errChan <- err
				return
			}
		}
	}()
	if err := <-errChan; err != nil {
		return err
	}
	return nil
}

func onChange(inDir string, outDir string, event fsnotify.Event) error {
    err := filepath.WalkDir(inDir, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        if d.IsDir() {
            return nil
        }
        ext := filepath.Ext(path)
        if ext != ".md" {
            return nil
        }
        mdFile, err := NewMarkdownFile(path)
        if err != nil {
            return err
        }
        return nil
    })
    if err != nil {
        return err
    }
    return nil
}


type MarkdownFile struct {
    path string
}

func NewMarkdownFile(path string) (MarkdownFile, error) {
    mdFile := MarkdownFile {
        path: path,
    }
    return mdFile, nil
}
