package main

import (
    "fmt"
    "testing"
    "strings"
)

func TestNewMarkdownFile(t *testing.T) {

    path := "./tmp/index.md"
    theme := "dracula"

    mdFile, err := NewMarkdownFile(path, theme)
    if err != nil {
        panic(err)
    }

    fmt.Println(mdFile.FileName)



}
