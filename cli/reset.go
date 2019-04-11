package cli

import (
	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

func ResetCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "reset",
		Usage: "Resets all story branches to trunk branches",
		Action: func(c *cli.Context) error {
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

			for project := range story.Projects {
				output, err := git.CheckoutBranch(git.CheckoutBranchOpts{Branch: trunk, Project: project})
				if err != nil {
					return err
				}

				printGitOutput(output, project)
			}

			output, err := git.CheckoutBranch(git.CheckoutBranchOpts{Branch: trunk})
			if err != nil {
				return err
			}

			printGitOutput(output, "metarepo")
			return nil
		},
	}
}
