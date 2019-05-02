package cli

import (
	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

func UpdateCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "update",
		Usage: "Updates code from the upstream master branch across the current story",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "from-branch", Usage: "Update from a specific branch", Value: "master"},
		},
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

			sourceBranch := c.String("from-branch")

			// Pull and merge master in all the projects
			for project := range story.Projects {
				if _, err := git.Fetch(git.FetchOpts{
					Branch:  sourceBranch,
					Remote:  "origin",
					Project: project,
				}); err != nil {
					return err
				}

				mergeOutput, err := git.Merge(git.MergeOpts{
					SourceBranch:      sourceBranch,
					DestinationBranch: story.Name,
					Project:           project,
				})

				if err != nil {
					return err
				}

				printGitOutput(mergeOutput, project)
			}

			// Pull and merge master in the metarepo
			if _, err := git.Fetch(git.FetchOpts{
				Branch: sourceBranch,
				Remote: "origin",
			}); err != nil {
				return err
			}

			mergeOutput, err := git.Merge(git.MergeOpts{
				SourceBranch:      sourceBranch,
				DestinationBranch: story.Name,
			})

			if err != nil {
				return err
			}

			printGitOutput(mergeOutput, metarepo)

			return nil
		}),
	}
}
