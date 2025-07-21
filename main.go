package main

import (
	"fmt"
    "path/filepath"
    "io/fs"
	"os"
	"bytes"
	"github.com/fsnotify/fsnotify"
	"github.com/yuin/goldmark"
    "github.com/yuin/goldmark/parser"
    "github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark-meta"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
)

func main() {
	err := startRuntime("./testmd", "./out", "dracula")
	if err != nil {
        fmt.Printf("MARKI ERROR: %s\n", err.Error())
        main()
	}
}

func startRuntime(inDir string, outDir string, theme string) error {
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
				err := onChange(inDir, outDir, theme, event)
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

func onChange(inDir string, outDir string, theme string, event fsnotify.Event) error {
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
		mdFile, err := NewMarkdownFile(path, theme)
        if err != nil {
            return err
        }
		fmt.Println(mdFile.Path)
        return nil
    })
    if err != nil {
        return err
    }
    return nil
}


type MarkdownFile struct {
    Path string
	Text string
	Theme string
	Html string
	Meta map[string]any
}

func NewMarkdownFile(path string, theme string) (MarkdownFile, error) {
	var mdFile MarkdownFile
	mdBytes, err := os.ReadFile(path)
	if err != nil {
		return mdFile, err
	}
    mdFile.Text = string(mdBytes)
	mdFile.Path = path
	mdFile.Theme = theme
	md := goldmark.New(
			goldmark.WithExtensions(
				meta.Meta,
				highlighting.NewHighlighting(
					highlighting.WithStyle(theme),
					highlighting.WithFormatOptions(
						chromahtml.WithLineNumbers(true),
					),
				),
			),
			goldmark.WithParserOptions(
				parser.WithAutoHeadingID(),
			),
			goldmark.WithParserOptions(
				parser.WithAttribute(),
			),
			goldmark.WithRendererOptions(
				html.WithHardWraps(),
				html.WithXHTML(),
				html.WithUnsafe(),
			),
		)
	var buf bytes.Buffer
	context := parser.NewContext()
	if err := md.Convert(mdBytes, &buf, parser.WithContext(context)); err != nil {
		return mdFile, err
	}
	mdFile.Html = buf.String()
	mdFile.Meta = meta.Get(context)
    return mdFile, nil
}
