package cli

import (
	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/LGUG2Z/story/node"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

func RemoveCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "remove",
		Usage: "Removes a project from the current story",
		Action: func(c *cli.Context) error {
			if !isStory {
				return ErrNotWorkingOnAStory
			}

			if !c.Args().Present() {
				return ErrCommandRequiresAnArgument
			}

			story, err := manifest.LoadStory(fs)
			if err != nil {
				return err
			}

			for _, project := range c.Args() {
				// Remove project from manifest
				story.RemoveFromManifest(project)

				// Remove it from the blast radius map
				delete(story.BlastRadius, project)

				// Delete the branch
				output, err := git.DeleteBranch(git.DeleteBranchOpts{
					Branch:  story.Name,
					Local:   true,
					Project: project,
				})

				if err != nil {
					return err
				}

				printGitOutput(output, project)
			}

			// Use the Blast Radius to update artifacts
			story.MapBlastRadiusToArtifacts()

			// Set the latest commit hashes for current projects
			hashes, err := story.GetCommitHashes(fs)
			story.Hashes = hashes

			// Update the manifest
			if err := story.Write(fs); err != nil {
				return err
			}

			var projectList []string
			for project := range story.Projects {
				projectList = append(projectList, project)
			}

			// Update all of the package.json files where any removed project is used
			for project := range story.Projects {
				p := node.PackageJSON{}
				if err := p.Load(fs, project); err != nil {
					return err
				}

				for _, toReset := range c.Args() {
					p.ResetPrivateDependencyBranches(toReset, story.Name)
					if err := p.Write(fs, project); err != nil {
						return err
					}
				}
			}

			return nil
		},
	}
}
