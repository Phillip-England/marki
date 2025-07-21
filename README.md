# marki
A runtime for content-driven developers who just want to turn `.md` into `.html` with styled code blocks. Run marki in the background, write your content, and use the generate html. Dead simple.

## Installation
```bash
go install github.com/Phillip-England/marki@latest
```

## Usage
```bash
marki <INDIR> <OUTDIR> <THEME> <FLAGS>
marki ./in ./out dracula
marki ./in ./out dracula --watch
```

## Themes
Marki uses [Goldmark](https://github.com/yuin/goldmark) for converting markdown into html. 

Goldmark uses [Chroma](https://github.com/alecthomas/chroma) for syntax highlighting. All the available themes for chroma can be found in the `.xml` files listed [here](https://github.com/alecthomas/chroma/tree/master/styles).

The first theme is `abap.xml`, so to use it with marki call:

```bash
marki <INDIR> <OUTDIR> abap --watch
```