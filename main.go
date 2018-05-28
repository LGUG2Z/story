package main

import (
	"os"
	"github.com/LGUG2Z/story/meta"
	"log"
	"github.com/spf13/afero"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	meta := meta.Manifest{Fs: afero.NewOsFs()}
	if err := meta.Load(".meta"); err != nil {
		log.Fatal("this folder does not contain a .meta file", err)
	}

	switch os.Args[1] {
	case "reset":
		if !meta.IsStory() {
			log.Fatal("not working on a story")
		}

		meta.RestoreGlobal()
	case "add":
		if meta.IsStory() {
			if err := meta.AddProjects(os.Args[2:]); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal("not working on a story")
		}
	case "remove":
		if meta.IsStory() {
			if err := meta.RemoveProjects(os.Args[2:]); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal("not working on a story")
		}
	case "set":
		if meta.IsStory() {
			log.Fatal("already working on a story")
		}

		err := meta.SetStory(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
	}
}
