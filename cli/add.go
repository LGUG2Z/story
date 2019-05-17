package cli

import (
	"github.com/LGUG2Z/blastradius/blastradius"
	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/LGUG2Z/story/node"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

func AddCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "add",
		Usage: "Adds a project to the current story",
		Flags: []cli.Flag{cli.BoolFlag{Name: "ci", Usage: "clone without modifying .meta"}},
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

			if c.Bool("ci") {
				for _, project := range c.Args() {
					if err := ensureProjectIsCloned(fs, story, project); err != nil {
						return err
					}
				}

				return nil
			}

			for _, project := range c.Args() {
				// Add to manifest
				if err := story.AddToManifest(story.AllProjects, project); err != nil {
					return err
				}

				// Calculate the blast radius for the project and add to story
				b := blastradius.NewCalculator()
				if err := story.CalculateBlastRadiusForProject(fs, b, project); err != nil {
					return err
				}

				// Checkout the branch
				output, err := git.CheckoutBranch(git.CheckoutBranchOpts{
					Branch:  story.Name,
					Create:  true,
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

			// Update all of the package.json files where any other added project is used
			for project := range story.Projects {
				if ignore[project] {
					continue
				}

				p := node.PackageJSON{}
				if err := p.Load(fs, project); err != nil {
					return err
				}

				p.SetPrivateDependencyBranchesToStory(story.Name, projectList...)
				if err := p.Write(fs, project); err != nil {
					return err
				}
			}

			return nil
		},
	}
}
