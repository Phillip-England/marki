package main

import (
    "testing"
)

func TestNewMarkdownFile(t *testing.T) {
    mdFile, err := NewMarkdownFile("./tmp/index.md", "./tmp-out", "dracula")
    if err != nil {
        panic(err)
    }
    err = SaveMarkdownHtmlToDisk(mdFile, outDir)
}
