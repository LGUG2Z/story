package cli

import (
	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

func CreateCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "create",
		Usage: "Creates a new story",
		Action: func(c *cli.Context) error {
			if isStory {
				return ErrAlreadyWorkingOnAStory
			}

			if !c.Args().Present() {
				return ErrCommandRequiresAnArgument
			}

			name := c.Args().First()

			meta, err := manifest.LoadMetaOnTrunk(fs)
			if err != nil {
				return err
			}

			story := manifest.NewStory(name, meta)
			output, err := git.CheckoutBranch(git.CheckoutBranchOpts{Branch: story.Name, Create: true})
			if err != nil {
				return err
			}

			printGitOutput(output, metarepo)

			return story.Write(fs)
		},
	}
}
