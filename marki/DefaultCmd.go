package marki

import (
	"fmt"

	"github.com/phillip-england/whip"
)

type DefaultCmd struct{}

func NewDefaultCmd(app *whip.Cli) (whip.Cmd, error) {
	return DefaultCmd{}, nil
}

func (cmd DefaultCmd) Execute(app *whip.Cli) error {
	fmt.Println(`marki - a runtime for converting .md into .html

convert:
	marki convert <SRC> <OUT> <THEME> <FLAGS>
	marki convert ./README.md ./README.html dracula --watch
	marki convert ./indir ./outdir dracula --watch

Don't forget to give me a star ⭐ at https://github.com/phillip-england/marki
	`)
	return nil
}
