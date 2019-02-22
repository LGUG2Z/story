package cli

import (
	"time"

	"fmt"

	"github.com/LGUG2Z/story/git"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

var isStory bool
var trunk string

var (
	Version string
	Commit  string
)

func App() *cli.App {
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("story version %s (commit %s)\n", c.App.Version, Commit)
	}

	fs := afero.NewOsFs()
	app := cli.NewApp()

	app.Name = "story"
	app.Usage = "A workflow tool for implementing stories across a node meta-repo"
	app.EnableBashCompletion = true
	app.Compiled = time.Now()
	app.Version = Version
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
		trunk = c.String("trunk")
		branch, err := git.GetCurrentBranch(fs, ".")
		if err != nil {
			return err
		}

		isStory = branch != trunk

		return nil
	}

	app.Commands = []cli.Command{
		CreateCmd(fs),
		LoadCmd(fs),
		ResetCmd(fs),
		AddCmd(fs),
		RemoveCmd(fs),
		ListCmd(fs),
		BlastRadiusCmd(fs),
		ArtifactsCmd(fs),
		CommitCmd(fs),
		PushCmd(fs),
		PrepareCmd(fs),
	}

	return app
}
