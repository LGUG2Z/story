package main

import (
	"log"
	"os"

	"fmt"
	"time"

	"github.com/LGUG2Z/story/meta"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

var (
	ErrAlreadyWorkingOnAStory = fmt.Errorf("already working on a story")
	ErrNotWorkingOnAStory     = fmt.Errorf("not working on a story")
	ErrNoArgsRequired         = fmt.Errorf("this command doesn't take any arguments")
	ErrSingleArgRequired      = fmt.Errorf("this command takes a single argument")
	ErrAtLeastOneArgRequired  = fmt.Errorf("this command takes at least one argument")
)

func main() {

	var m meta.Manifest

	app := cli.NewApp()
	app.Name = "story"
	app.Usage = "A workflow tool for implementing stories across a meta-repo"
	app.EnableBashCompletion = true
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		{
			Name:  "J. Iqbal",
			Email: "jade@beamery.com",
		},
	}

	app.Before = func(c *cli.Context) error {
		log.SetFlags(0)
		log.SetOutput(os.Stdout)

		m = meta.Manifest{Fs: afero.NewOsFs()}
		return m.Load(".meta")
	}

	app.Commands = []cli.Command{
		{
			Name:      "set",
			Usage:     "set a story to work on",
			UsageText: "story set new-navigation",
			Action: func(c *cli.Context) error {
				if c.NArg() != 1 {
					return ErrSingleArgRequired
				}

				if m.IsStory() {
					return ErrAlreadyWorkingOnAStory
				}

				return m.SetStory(c.Args().First())
			},
		},
		{
			Name:      "reset",
			Usage:     "reset to master branches",
			UsageText: "story reset",
			Action: func(c *cli.Context) error {
				if c.NArg() != 0 {
					return ErrNoArgsRequired
				}

				if m.IsStory() {
					return m.Reset()
				}

				return ErrNotWorkingOnAStory
			},
		},
		{
			Name:      "add",
			ShortName: "a",
			Usage:     "add projects to a story",
			UsageText: "story add vacancies-service",
			Action: func(c *cli.Context) error {
				if c.NArg() < 1 {
					return ErrAtLeastOneArgRequired
				}

				if m.IsStory() {
					return m.AddProjects(c.Args())
				}

				return ErrNotWorkingOnAStory
			},
		},
		{
			Name:      "remove",
			ShortName: "rm",
			Usage:     "remove projects from a story",
			UsageText: "story remove vacancies-service",
			Action: func(c *cli.Context) error {
				if c.NArg() < 1 {
					return ErrAtLeastOneArgRequired
				}

				if m.IsStory() {
					return m.RemoveProjects(c.Args())
				}

				return ErrNotWorkingOnAStory
			},
		},
		{
			Name:      "prune",
			Usage:     "prune untouched projects from a story",
			UsageText: "story prune",
			Action: func(c *cli.Context) error {
				if c.NArg() != 0 {
					return ErrNoArgsRequired
				}

				if m.IsStory() {
					return m.Prune()
				}

				return ErrNotWorkingOnAStory
			},
		},
		{
			Name:      "blast",
			Usage:     "add projects within the blast radius of a story",
			UsageText: "story blast",
			Action: func(c *cli.Context) error {
				if c.NArg() != 0 {
					return ErrNoArgsRequired
				}

				if m.IsStory() {
					return m.Blast()
				}

				return ErrNotWorkingOnAStory
			},
		},
		{
			Name:      "complete",
			Usage:     "revert branch references in package.json files in preparation for merge to master",
			UsageText: "story complete",
			Action: func(c *cli.Context) error {
				if c.NArg() != 0 {
					return ErrNoArgsRequired
				}

				if m.IsStory() {
					return m.Complete()
				}

				return ErrNotWorkingOnAStory
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
