package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/fsnotify/fsnotify"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func main() {

	args := os.Args
	if len(args) == 1 {
		printHelpScreen()
		return
	}

	indicatingArg := getArg(1)
	if indicatingArg != "-g" {
		printHelpScreen()
		return
	}

	inDir := getArg(2)
	if inDir == "" {
		printHelpScreen()
		return
	}

	outDir := getArg(3)
	if outDir == "" {
		printHelpScreen()
		return
	}

	theme := getArg(4)
	if outDir == "" {
		printHelpScreen()
		return
	}

	watchFlag := getArg(5)

	err := validateInDir(inDir)
	if err != nil {
		perr(err)
	}

	err = validateTheme(theme)
	if err != nil {
		perr(err)
	}

	mkdirIfNotExists(outDir)

	fmt.Println("🤘: marki launch!")

	if watchFlag == "" {
		onChange(inDir, outDir, theme, fsnotify.Event{
			Op: fsnotify.Write,
		})
		return
	}

	fmt.Printf("👁️: watching %s\n", inDir)
	err = startRuntime(inDir, outDir, theme)
	if err != nil {
		perr(err)
		startRuntime(inDir, outDir, theme)
	}
}

func getArg(number int) string {
	if number == 0 {
		return os.Args[0]
	}
	if len(os.Args) <= number {
		return ""
	}
	return os.Args[number]
}

func printHelpScreen() {
	fmt.Println("👋 welcome to marki")
	fmt.Println("")
	fmt.Println("[-g] GENERATE: marki -g <INDIR> <OUTDIR> <THEME>")
	fmt.Println("   iterate through <INDIR>")
	fmt.Println("   convert .md to .html with code <THEME>")
	fmt.Println("   place the .html files in <OUTDIR>")
	fmt.Println("")
	fmt.Println("   OPTIONAL FLAGS:")
	fmt.Println("       [--watch]: watch <INDIR> and re-run on file change")
	fmt.Println("")
	fmt.Println("   EXAMPLES:")
	fmt.Println("        marki -g ./markdown ./html dracula")
	fmt.Println("        marki -g ./markdown ./html dracula --watch")
}

func validateTheme(theme string) error {
	validThemes := []string{"dracula"}
	for _, vTheme := range validThemes {
		if theme == vTheme {
			return nil
		}
	}
	return fmt.Errorf("theme [%s] is not a valid theme", theme)
}

func validateInDir(inDir string) error {
	if !dirExists(inDir) {
		return fmt.Errorf("input directory [%s] does not exist", inDir)
	}
	return nil
}

func dirExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func perr(err error) {
	fmt.Printf("🚨: %s\n", err.Error())
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
	if event.Op != fsnotify.Write {
		return nil
	}
	fmt.Println("📝: writing out..")
	err := dirClear(outDir)
	if err != nil {
		return err
	}
	err = filepath.WalkDir(inDir, func(path string, d fs.DirEntry, err error) error {
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
		mdFile, err := NewMarkdownFile(path, outDir, theme)
		if err != nil {
			return err
		}
		err = SaveMarkdownHtmlToDisk(mdFile)
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
	Path             string
	Text             string
	Theme            string
	Html             string
	Meta             map[string]any
	MetaHtml         string
	FileName         string
	PathWithoutBase  string
	SaveToPath       string
	SaveToDir        string
	ClientMetaScript string
}

func NewMarkdownFile(path string, outDir string, theme string) (MarkdownFile, error) {
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
	mdFile.MetaHtml = ""
	metaTagClassName := "marki-" + randomString(8)
	for key, value := range mdFile.Meta {
		mdFile.MetaHtml = mdFile.MetaHtml + fmt.Sprintf("<meta class='%s' name='%s' content='%s'>\n", metaTagClassName, key, value)
	}
	mdFile.Html = mdFile.MetaHtml + mdFile.Html
	mdFile.FileName = filepath.Base(mdFile.Path)
	mdFile.PathWithoutBase = strings.ReplaceAll(mdFile.Path, mdFile.FileName, "")
	mdFile.SaveToPath = strings.ReplaceAll(strings.ReplaceAll(mdFile.Path, mdFile.PathWithoutBase, outDir+"/"), ".md", ".html")
	mdFile.SaveToDir = filepath.Dir(mdFile.SaveToPath)
	mdFile.ClientMetaScript = `
        <script>
            (() => {
                let metaElements = document.querySelectorAll('` + metaTagClassName + `')
                let head = document.querySelector('head')
                for (let i = 0; i < metaElements.length; i++) {
                    let elm = metaElements[i]
                    head.appendChild(elm)
                }
            })()
        </script>
    `
	mdFile.Html = mdFile.Html + mdFile.ClientMetaScript
	return mdFile, nil
}

func SaveMarkdownHtmlToDisk(mdFile MarkdownFile) error {
	err := os.MkdirAll(mdFile.SaveToDir, 0755)
	if err != nil {
		return err
	}
	htmlFile, err := os.Create(mdFile.SaveToPath)
	if err != nil {
		return err
	}
	defer htmlFile.Close()
	_, err = htmlFile.Write([]byte(mdFile.Html))
	if err != nil {
		return err
	}
	return nil
}

func mkdirIfNotExists(outDir string) error {
	err := os.MkdirAll(outDir, 0755)
	if err != nil {
		return err
	}
	return nil
}

func dirClear(dirName string) error {
	err := filepath.WalkDir(dirName, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		err = os.Remove(path)
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

func randomString(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(result)
}
