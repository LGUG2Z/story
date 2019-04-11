package cli

import (
	"fmt"

	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

func MergeCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "merge",
		Usage: "Merges prepared code to master branches across the current story",
		Action: cli.ActionFunc(func(c *cli.Context) error {
			if !isStory {
				return ErrNotWorkingOnAStory
			}

			if !c.Args().Present() {
				return ErrCommandRequiresAnArgument
			}

			currentBranch, err := git.GetCurrentBranch(fs, ".")
			if err != nil {
				return err
			}

			story, err := manifest.LoadStoryFromBranchName(fs, currentBranch)
			if err != nil {
				return err
			}

			// Checkout master and merge story in all projects
			for project := range story.Projects {
				checkoutMasterOutput, err := git.CheckoutBranch(git.CheckoutBranchOpts{
					Branch: "master",
					Create: false,
				})

				if err != nil {
					return err
				}

				printGitOutput(checkoutMasterOutput, project)

				fmt.Println("running merge", project)
				mergeOutput, err := git.Merge(git.MergeOpts{
					SourceBranch:      currentBranch,
					DestinationBranch: "master",
					Project:           project,
				})

				if err != nil {
					return err
				}

				printGitOutput(mergeOutput, project)
			}

			// Merge story into master on the metarepo
			output, err := git.Merge(git.MergeOpts{
				SourceBranch:      currentBranch,
				DestinationBranch: "master",
			})
			if err != nil {
				return err
			}

			printGitOutput(output, "metarepo")

			return nil
		}),
	}
}
