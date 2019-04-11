package cli

import (
	"fmt"

	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/fatih/color"
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

			if c.Args().Present() {
				return ErrCommandTakesNoArguments
			}

			currentBranch, err := git.GetCurrentBranch(fs, ".")
			if err != nil {
				return err
			}

			story, err := manifest.LoadStoryFromBranchName(fs, currentBranch)
			if err != nil {
				return err
			}

			messages := []string{fmt.Sprintf("[story merge] Merge branch '%s'", story.Name)}

			// Checkout master and merge story in all projects
			for project := range story.Projects {
				color.Green(project)
				checkoutBranchOutput, err := git.CheckoutBranch(git.CheckoutBranchOpts{
					Branch:  "master",
					Project: project,
					Create:  false,
				})

				if err != nil {
					return err
				}

				fmt.Printf("%s\n\n", checkoutBranchOutput)

				mergeOutput, err := git.Merge(git.MergeOpts{
					SourceBranch:      story.Name,
					DestinationBranch: "master",
					Project:           project,
					Squash:            true,
				})

				if err != nil {
					return err
				}

				fmt.Printf("%s\n\n", mergeOutput)

				commitOutput, err := git.Commit(
					git.CommitOpts{
						Project:  project,
						Messages: messages,
					},
				)

				if err != nil {
					return err
				}

				fmt.Println(commitOutput)
			}

			color.Green("metarepo")
			checkoutBranchOutput, err := git.CheckoutBranch(git.CheckoutBranchOpts{
				Branch: "master",
				Create: false,
			})

			if err != nil {
				return err
			}

			fmt.Printf("%s\n\n", checkoutBranchOutput)

			// Merge story into master on the metarepo
			mergeOutput, err := git.Merge(git.MergeOpts{
				SourceBranch:      story.Name,
				DestinationBranch: "master",
				Squash:            true,
			})

			if err != nil {
				return err
			}

			fmt.Printf("%s\n\n", mergeOutput)

			commitOutput, err := git.Commit(git.CommitOpts{Messages: messages})
			if err != nil {
				return err
			}

			fmt.Println(commitOutput)

			return nil
		}),
	}
}
