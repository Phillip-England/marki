---
metaContent: "a readme about marki"
---

# marki
A runtime for content-driven developers who just want to turn `.md` into `.html`. Run marki in the background, write your content, and use the generated html. Dead simple.

## Installation
```bash
go install github.com/phillip-england/marki@v0.0.5
```

## Usage
```bash
marki convert <SRC> <OUT> <THEME> <FLAGS>
marki convert ./indir ./outdir dracula
marki convert ./indir ./outdir dracula --watch
marki convert ./infile.md ./outfile.html dracula
marki convert ./infile.md ./outfile.html dracula --watch
```

## Themes
Marki uses [Goldmark](https://github.com/yuin/goldmark) for converting markdown into html. 

Goldmark uses [Chroma](https://github.com/alecthomas/chroma) for syntax highlighting. All the available themes for chroma can be found in the `.xml` files listed [here](https://github.com/alecthomas/chroma/tree/master/styles).

The first theme is `abap.xml`, so to use it with marki call:

```bash
marki convert <SRC> <OUT> abap --watch
```

## Metadata
Use YAML-style frontmatter in your markdown to generate HTML `<meta>` tags for your content. For example, the following markdown:

```md
---
metaDescription: "my description"
---
# Content
some markdown content
```

will result in the following HTML:
```html
<meta name='metaDescription' content='my description'>
<!-- MARKI SPLIT --><h1 id="content">Content</h1>
<p>some markdown content</p>
```

You can then split off the HTML by splitting the string by `<!-- MARKI SPLIT -->`, making it easy to parse out meta content from UI content.