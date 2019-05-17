package cli

import (
	"github.com/LGUG2Z/story/manifest"
	"github.com/LGUG2Z/story/node"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

func PinCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "pin",
		Usage: "Pins code in the current story",
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

			var projectList []string
			for project := range story.Projects {
				projectList = append(projectList, project)
			}

			// Update all of the package.json files where any other added project is used
			for project := range story.Projects {
				if ignore[project] {
					continue
				}

				p := node.PackageJSON{}
				if err := p.Load(fs, project); err != nil {
					return err
				}

				p.SetPrivateDependencyBranchesToCommitHashes(story, projectList...)
				if err := p.Write(fs, project); err != nil {
					return err
				}

				printGitOutput("package.json updated", project)
			}

			return nil
		}),
	}
}
