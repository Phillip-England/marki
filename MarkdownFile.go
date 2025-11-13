package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/phillip-england/wherr"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type MarkdownFile struct {
	Path            string
	Text            string
	Theme           string
	Html            string
	Meta            map[string]any
	MetaHtml        string
	FileName        string
	PathWithoutBase string
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
	mdFile.MetaHtml = ""
	for key, value := range mdFile.Meta {
		if !strings.HasPrefix(key, "meta") {
			continue
		}
		mdFile.MetaHtml = mdFile.MetaHtml + fmt.Sprintf("<meta name='%s' content='%s'>\n", key, value)
	}
	mdFile.Html = mdFile.MetaHtml + "<!-- MARKI SPLIT -->" + mdFile.Html
	mdFile.FileName = filepath.Base(mdFile.Path)
	mdFile.PathWithoutBase = strings.ReplaceAll(mdFile.Path, mdFile.FileName, "")
	return mdFile, nil
}

func SaveMarkdownHtmlToDisk(mdFile MarkdownFile, saveTo string) error {
	err := os.MkdirAll(filepath.Dir(saveTo), 0755)
	if err != nil {
		return wherr.Consume(wherr.Here(), err, "")
	}
	err = os.WriteFile(saveTo, []byte(mdFile.Html), 0644)
	if err != nil {
		return wherr.Consume(wherr.Here(), err, "")
	}
	return nil
}
