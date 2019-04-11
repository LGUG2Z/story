package cli

import (
	"fmt"

	"github.com/LGUG2Z/story/manifest"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

func ArtifactsCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "artifacts",
		Usage: "Shows a list of artifacts to be built and deployed for the current story",
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

			for project := range story.Artifacts {
				if story.Artifacts[project] {
					fmt.Println(project)
				}
			}

			return nil
		},
	}
}
