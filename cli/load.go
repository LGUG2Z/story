package cli

import (
	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

func LoadCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "load",
		Usage: "Loads an existing story",
		Action: func(c *cli.Context) error {
			if isStory {
				return ErrAlreadyWorkingOnAStory
			}

			if !c.Args().Present() {
				return ErrCommandRequiresAnArgument
			}

			name := c.Args().First()
			output, err := git.CheckoutBranch(git.CheckoutBranchOpts{Branch: name})
			if err != nil {
				return err
			}

			printGitOutput(output, "metarepo")

			story, err := manifest.LoadStory(fs)
			if err != nil {
				return err
			}

			for project := range story.Projects {
				if err := ensureProjectIsCloned(fs, story, project); err != nil {
					return err
				}

				output, err := git.CheckoutBranch(git.CheckoutBranchOpts{Branch: name, Project: project})
				if err != nil {
					return err
				}

				printGitOutput(output, project)
			}

			return nil
		},
	}
}
