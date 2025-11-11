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
	fmt.Println("Hello, World!")
	return nil
}
