package main

import (
	"github.com/spf13/cobra/doc"

	"github.com/harvey-earth/pocdeploy/cmd"
)

func main() {
	title := &doc.GenManHeader{Title: "POCDEPLOY", Section: "1", Source: "harvey-earth", Manual: "pocdeploy Man Page"}
	err := doc.GenManTree(cmd.Root(), title, "./docs")
	if err != nil {
		panic(err)
	}
}
