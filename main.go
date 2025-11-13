package main

import (
	"fmt"

	"github.com/phillip-england/marki/marki"
	"github.com/phillip-england/whip"
)

func main() {

	cli, err := whip.New(marki.NewDefaultCmd)
	if err != nil {
		fmt.Println(err.Error())
	}

	cli.At("convert", marki.NewConvertCmd)

	err = cli.Run()
	if err != nil {
		fmt.Println(err.Error())
	}

}
