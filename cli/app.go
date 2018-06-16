package cli

import (
	"fmt"

	"time"

	"github.com/LGUG2Z/story/git"
	"github.com/fatih/color"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

var fs afero.Fs
var isStory bool
var trunk string

func App() *cli.App {
	app := cli.NewApp()
	app.Name = "story"
	app.Usage = "A workflow tool for implementing stories across a meta-repo"
	app.EnableBashCompletion = true
	app.Compiled = time.Now()
	app.Authors = []cli.Author{{
		Name:  "J. Iqbal",
		Email: "jade@beamery.com",
	}}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "trunk",
			Value: "master",
		},
	}

	app.Before = func(c *cli.Context) error {
		fs = afero.NewOsFs()
		trunk = c.String("trunk")
		branch, err := git.GetCurrentBranch(fs, ".")
		if err != nil {
			return err
		}

		isStory = branch != trunk

		return nil
	}

	app.Commands = []cli.Command{
		CreateCmd(),
		LoadCmd(),
		ResetCmd(),
		AddCmd(),
		RemoveCmd(),
		ListCmd(),
		ArtifactsCmd(),
	}

	return app
}

func printGitOutput(output, project string) {
	color.Green(project)
	fmt.Println(output)
}
