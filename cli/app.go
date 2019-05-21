package cli

import (
	"os"
	"path/filepath"
	"time"

	"fmt"

	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

var isStory bool
var trunk string
var metarepo string
var ignore map[string]bool

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
			Name:   "trunk",
			Value:  "master",
			EnvVar: "STORY_TRUNK",
		},
	}

	app.Before = func(c *cli.Context) error {
		trunk = c.String("trunk")
		branch, err := git.GetCurrentBranch(fs, ".")
		if err != nil {
			return err
		}

		isStory = branch != trunk

		path, err := os.Getwd()
		if err != nil {
			return err
		}

		metarepo = filepath.Base(path)

		ignore, err = manifest.LoadStoryIgnore(path)
		return err
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
		UnpinCmd(fs),
		PinCmd(fs),
		PrepareCmd(fs),
		UpdateCmd(fs),
		MergeCmd(fs),
		PRCmd(fs),
	}

	return app
}
