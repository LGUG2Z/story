package cli

import (
	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

func PushCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "push",
		Usage: "Pushes commits across the current story",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "from-manifest", Usage: "Use a manifest from a specific story to determine which repos to push"},
		},
		Action: cli.ActionFunc(func(c *cli.Context) error {
			if len(c.String("from-manifest")) == 0 {
				if !isStory {
					return ErrNotWorkingOnAStory
				}
			}

			if c.Args().Present() {
				return ErrCommandTakesNoArguments
			}

			var story *manifest.Story
			var err error
			var branch string

			if len(c.String("from-manifest")) > 0 {
				story, err = manifest.LoadStoryFromBranchName(fs, c.String("from-manifest"))
				if err != nil {
					return err
				}

				branch = "master"
			} else {
				story, err = manifest.LoadStory(fs)
				if err != nil {
					return err
				}

				branch = story.Name
			}

			// Push in all the projects
			for project := range story.Projects {
				output, err := git.Push(git.PushOpts{Remote: "origin", Branch: branch, Project: project})
				if err != nil {
					return err
				}

				printGitOutput(output, project)
			}

			// Commit on the metarepo
			output, err := git.Push(git.PushOpts{Branch: branch, Remote: "origin"})
			if err != nil {
				return err
			}

			printGitOutput(output, "metarepo")

			return nil
		}),
	}
}
