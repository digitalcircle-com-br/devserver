package main

import (
	"log"

	"github.com/digitalcircle-com-br/caroot"
	"github.com/digitalcircle-com-br/devserver/lib/config"
	"github.com/digitalcircle-com-br/devserver/lib/server"
	"github.com/digitalcircle-com-br/devserver/lib/tray"
)

func Start() error {

	err:=config.Init()
	if err!=nil{
		return err
	}

	err = caroot.InitCA("caroot", func(ca string) {
		log.Printf("Initiating CA: %s", ca)
	})

	if err!=nil{
		return err
	}

	server.StartHttpsServer()

	tray.Run()
	return nil
}

func main() {
	err:=Start()
	if err!=nil{
		panic(err)
	}
}
