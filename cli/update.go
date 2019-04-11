package cli

import (
	"fmt"

	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

func UpdateCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "update",
		Usage: "Updates code from the upstream master branch across the current story",
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

			// Pull and merge master in all the projects
			for project := range story.Projects {
				fmt.Println("running fetch", project)
				fetchOutput, err := git.Fetch(git.FetchOpts{
					Branch:  "master",
					Remote:  "origin",
					Project: project,
				})

				if err != nil {
					return err
				}

				printGitOutput(fetchOutput, project)

				fmt.Println("running merge", project)
				mergeOutput, err := git.Merge(git.MergeOpts{
					SourceBranch:      "master",
					DestinationBranch: story.Name,
					Project:           project,
				})

				if err != nil {
					return err
				}

				printGitOutput(mergeOutput, project)
			}

			return nil
		}),
	}
}
