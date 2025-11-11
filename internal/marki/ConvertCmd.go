package marki

import (
	"github.com/phillip-england/wherr"
	"github.com/phillip-england/whip"
)

type ConvertCmd struct {
	Src       string
	SrcIsFile bool
	Out       string
	Theme     string
}

func NewConvertCmd(cli *whip.Cli) (whip.Cmd, error) {
	src, err := cli.ArgGetByPositionForce(2, "missing <SOURCE> in 'marki convert'")
	if err != nil {
		return ConvertCmd{}, wherr.Consume(wherr.Here(), err, "")
	}
	srcIsFile := whip.IsFile(src)
	srcIsDir := whip.IsDir(src)
	if !srcIsDir && !srcIsFile {
		return ConvertCmd{}, wherr.Err(wherr.Here(), "<SOURCE> in 'marki convert' must be either a file or dir on your system")
	}
	if srcIsFile {
		if !whip.FileExists(src) {
			return ConvertCmd{}, wherr.Err(wherr.Here(), "file %s does not exist", src)
		}
	} else {
		if !whip.DirExists(src) {
			return ConvertCmd{}, wherr.Err(wherr.Here(), "dir %s does not exist", src)
		}
	}
	out, err := cli.ArgGetByPositionForce(3, "missing <OUT> in 'marki convert'")
	if err != nil {
		return ConvertCmd{}, wherr.Consume(wherr.Here(), err, "")
	}
	if srcIsFile {
		if !whip.IsFile(out) {
			return ConvertCmd{}, wherr.Err(wherr.Here(), "<OUT> must be a properly formatted file path if <SOURCE> is a file")
		}
	} else {
		if !whip.IsDir(out) {
			return ConvertCmd{}, wherr.Err(wherr.Here(), "<OUT> must be a properly formatted dir path if <SOURCE> is a dir")
		}
	}
	theme, err := cli.ArgGetByPositionForce(4, "<THEME> ")
	if err != nil {
		theme = "dracula"
	}
	err = isValidTheme(theme)
	if err != nil {
		return ConvertCmd{}, wherr.Consume(wherr.Here(), err, "")
	}
	return ConvertCmd{
		Src:       src,
		SrcIsFile: srcIsFile,
		Out:       out,
		Theme:     theme,
	}, nil
}

func (cmd ConvertCmd) Execute(app *whip.Cli) error {
	if cmd.SrcIsFile {
		err := handleFile(cmd, app)
		if err != nil {
			return wherr.Consume(wherr.Here(), err, "")
		}
		return nil
	}
	err := handleDir(cmd, app)
	if err != nil {
		return wherr.Consume(wherr.Here(), err, "")
	}
	return nil
}

func handleFile(cmd ConvertCmd, app *whip.Cli) error {
	mdFile, err := NewMarkdownFile(cmd.Src, cmd.Theme)
	if err != nil {
		return wherr.Consume(wherr.Here(), err, "")
	}
	err = SaveMarkdownHtmlToDisk(mdFile, cmd.Out)
	if err != nil {
		return wherr.Consume(wherr.Here(), err, "")
	}
	return nil
}

func handleDir(cmd ConvertCmd, app *whip.Cli) error {
	return nil
}
