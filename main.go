package main

import (
	"fmt"

	"github.com/phillip-england/whip"
)

func main() {

	cli, err := whip.New(NewDefaultCmd)
	if err != nil {
		fmt.Println(err.Error())
	}

	cli.At("convert", NewConvertCmd)

	err = cli.Run()
	if err != nil {
		fmt.Println(err.Error())
	}

}
