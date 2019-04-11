package cli

import (
	"fmt"

	"github.com/LGUG2Z/story/manifest"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

func ListCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "list",
		Usage: "Shows a list of projects added to the current story",
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
				fmt.Println(project)
			}

			return nil
		},
	}
}
