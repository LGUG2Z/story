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
		Action: cli.ActionFunc(func(c *cli.Context) error {
			if !isStory {
				return ErrNotWorkingOnAStory
			}

			if c.Args().Present() {
				return ErrCommandTakesNoArguments
			}

			story, err := manifest.LoadStory(fs)
			if err != nil {
				return err
			}

			// Push in all the projects
			for project := range story.Projects {
				output, err := git.Push(git.PushOpts{Remote: "origin", Branch: story.Name, Project: project})
				if err != nil {
					return err
				}

				printGitOutput(output, project)
			}

			// Commit on the metarepo
			output, err := git.Push(git.PushOpts{Branch: story.Name, Remote: "origin"})
			if err != nil {
				return err
			}

			printGitOutput(output, "metarepo")

			return nil
		}),
	}
}
