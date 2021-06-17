package main

import (
	"log"
	"os"

	"./ctl"
)

func main() {
	app := ctl.NewCliApp()
	if err := app.Run(os.Args); err != nil {
		log.Fatal("App启动失败：", err)
	}
}
