package main

import (
	"fmt"

	"github.com/phillip-england/marki/internal/marki"
	"github.com/phillip-england/whip"
)

func main() {

	app, err := whip.New(marki.NewDefaultCmd)
	if err != nil {
		fmt.Println(err.Error())
	}

	app.At("convert", marki.NewConvertCmd)

	err = app.Run()
	if err != nil {
		fmt.Println(err.Error())
	}

}
