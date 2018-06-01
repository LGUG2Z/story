package main

import (
	"github.com/LGUG2Z/story/meta"
	"github.com/spf13/afero"
	"log"
	"os"
)

func main() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	meta := meta.Manifest{Fs: afero.NewOsFs()}
	if err := meta.Load(".meta"); err != nil {
		log.Fatal(err)
	}

	switch os.Args[1] {
	case "blast":
		if meta.IsStory() {
			if err := meta.Blast(); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal("not working on a story")
		}
	case "prune":
		if meta.IsStory() {
			if err := meta.Prune(); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal("not working on a story")
		}
	case "reset":
		if meta.IsStory() {
			if err := meta.RestoreGlobal(); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal("not working on a story")
		}
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
