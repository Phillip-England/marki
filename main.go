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

	args, err := ArgsGenerateNew()
	if err != nil {
		printHelpScreen()
		printError(err)
		return
	}

	mkdirIfNotExists(args.OutDir)
	fmt.Println("🤘: marki launch!")

	// handling normal generation
	err = onChange(args.InDir, args.OutDir, args.Theme, fsnotify.Event{
		Op: fsnotify.Write,
	})
	if err != nil {
		printError(err)
	}

	// exit if we are not watching
	if args.FlagWatch == "" {
		return
	}

	// handling runtime generation with --watch
	fmt.Printf("👁️: watching %s\n", args.InDir)
	err = startRuntime(args.InDir, args.OutDir, args.Theme)
	if err != nil {
		printError(err)
		startRuntime(args.InDir, args.OutDir, args.Theme)
	}

}

type ArgsGenerate struct {
	InDir     string
	OutDir    string
	Theme     string
	FlagWatch string
}

func ArgsGenerateNew() (ArgsGenerate, error) {
	args := &ArgsGenerate{}
	args.InDir = getArg(1)
	args.OutDir = getArg(2)
	args.Theme = getArg(3)
	args.FlagWatch = getArg(4)
	err := args.validateTheme()
	if err != nil {
		return *args, err
	}
	err = args.validateInDir()
	if err != nil {
		return *args, err
	}
	return *args, nil
}

func (args *ArgsGenerate) validateTheme() error {
	validThemes := []string{
		"abap", "algol", "algol_nu", "arduino", "autumn", "average", "base16-snazzy",
		"borland", "bw", "catppuccin-frappe", "catppuccin-latte", "catppuccin-macchiato",
		"catppuccin-mocha", "colorful", "doom-one", "doom-one2", "dracula", "emacs",
		"evergarden", "friendly", "fruity", "github-dark", "github", "gruvbox-light",
		"gruvbox", "hr_high_contrast", "hrdark", "igor", "lovelace", "manni", "modus-operandi",
		"modus-vivendi", "monokai", "monokailight", "murphy", "native", "nord", "nordic",
		"onedark", "onesenterprise", "paraiso-dark", "paraiso-light", "pastie", "perldoc",
		"pygments", "rainbow_dash", "rose-pine-dawn", "rose-pine-moon", "rose-pine", "rpgle",
		"rrt", "solarized-dark", "solarized-dark256", "solarized-light", "swapoff", "tango",
		"tokyonight-day", "tokyonight-moon", "tokyonight-night", "tokyonight-storm", "trac",
		"vim", "vs", "vulcan", "witchhazel", "xcode-dark", "xcode",
	}
	themeList := ""
	for _, vTheme := range validThemes {
		themeList = themeList + vTheme + "\n"
		if args.Theme == vTheme {
			return nil
		}
	}
	return fmt.Errorf("theme [%s] is not a valid theme\nhere is a list of valid themes:\n%s", args.Theme, themeList)
}

func (args *ArgsGenerate) validateInDir() error {
	if !dirExists(args.InDir) {
		return fmt.Errorf("input directory [%s] does not exist", args.InDir)
	}
	return nil
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
	fmt.Println("USAGE: marki <INDIR> <OUTDIR> <THEME>")
	fmt.Println("   iterate through <INDIR>")
	fmt.Println("   convert .md to .html with code <THEME>")
	fmt.Println("   place the .html files in <OUTDIR>")
	fmt.Println("")
	fmt.Println("   OPTIONAL FLAGS:")
	fmt.Println("       [--watch]: watch <INDIR> and re-run on file change")
	fmt.Println("")
	fmt.Println("   EXAMPLES:")
	fmt.Println("        marki ./markdown ./html dracula")
	fmt.Println("        marki ./markdown ./html dracula --watch")
}

func dirExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func printError(err error) {
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
