package cli

import (
	"fmt"

	"github.com/LGUG2Z/story/manifest"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

func BlastRadiusCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "blastradius",
		Usage: "Shows a list of current story's blast radius",
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

			var brMap = make(map[string]bool)

			for _, br := range story.BlastRadius {
				for _, p := range br {
					if !brMap[p] {
						brMap[p] = true
					}
				}
			}

			for project := range brMap {
				fmt.Println(project)
			}

			return nil
		},
	}
}
