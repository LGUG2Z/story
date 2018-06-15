package main

import (
	"log"
	"os"

	"github.com/LGUG2Z/story/cli"
)

func main() {
	if err := cli.App().Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
