package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
)

func main() {
	err := startRuntime("./tmp", "./out")
	if err != nil {
		panic(err)
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
	return nil
}
