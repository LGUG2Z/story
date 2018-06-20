package cli

import (
	"fmt"

	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/fatih/color"
	"github.com/spf13/afero"
)

func printGitOutput(output, project string) {
	color.Green(project)
	fmt.Println(output)
}

func ensureProjectIsCloned(fs afero.Fs, story *manifest.Story, project string) error {
	exists, err := afero.DirExists(fs, project)
	if err != nil {
		return err
	}

	if !exists {
		output, err := git.Clone(git.CloneOpts{Repository: story.AllProjects[project]})
		if err != nil {
			return err
		}

		printGitOutput(output, project)
	}

	return nil
}
