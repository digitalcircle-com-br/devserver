package main

import (
	"github.com/digitalcircle-com-br/devserver"
)

func main() {
	err := devserver.Start()
	if err != nil {
		panic(err.Error())
	}
}
