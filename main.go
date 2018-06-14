package main

import (
	"log"
	"os"

	"github.com/LGUG2Z/story/cmd"
)

func main() {
	if err := cmd.Execute(os.Args...); err != nil {
		log.Fatal(err)
	}
}
