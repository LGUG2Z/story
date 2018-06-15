package cli

import (
	"fmt"

	"github.com/LGUG2Z/blastradius/blastradius"
	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/LGUG2Z/story/node"
	"github.com/urfave/cli"
)

func CreateCmd() cli.Command {
	return cli.Command{
		Name: "create",
		Action: func(c *cli.Context) error {
			if isStory {
				return fmt.Errorf("already working on a story")
			}

			if !c.Args().Present() {
				return fmt.Errorf("this command requires an argument")
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

			printGitOutput(output, "metarepo")

			if err := meta.MoveForStory(fs); err != nil {
				return err
			}

			return story.Write(fs)
		},
	}
}

func LoadCmd() cli.Command {
	return cli.Command{
		Name: "load",
		Action: func(c *cli.Context) error {
			if isStory {
				return fmt.Errorf("already working on a story")
			}

			if !c.Args().Present() {
				return fmt.Errorf("this command requires an argument")
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

func ResetCmd() cli.Command {
	return cli.Command{
		Name: "reset",
		Action: func(c *cli.Context) error {
			if !isStory {
				return fmt.Errorf("not working on a story")
			}

			if c.Args().Present() {
				return fmt.Errorf("this command takes no arguments")
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

func AddCmd() cli.Command {
	return cli.Command{
		Name: "add",
		Action: func(c *cli.Context) error {
			if !isStory {
				return fmt.Errorf("not working on a story")
			}

			if !c.Args().Present() {
				return fmt.Errorf("this command requires at least one argument")
			}

			story, err := manifest.LoadStory(fs)
			if err != nil {
				return err
			}

			meta, err := manifest.LoadMetaOnBranch(fs)
			if err != nil {
				return err
			}

			for _, project := range c.Args() {
				// Add to manifest
				if err := story.AddToManifest(meta.Projects, project); err != nil {
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

			// Use the Blast Radius to update deployables
			story.MapBlastRadiusToDeployables()

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
				p := node.PackageJSON{}
				if err := p.Load(fs, project); err != nil {
					return err
				}

				p.SetPrivateDependencyBranchesToStory(story.Name, projectList...)
				p.Write(fs, project)
			}

			return nil
		},
	}
}

func RemoveCmd() cli.Command {
	return cli.Command{
		Name: "remove",
		Action: func(c *cli.Context) error {
			if !isStory {
				return fmt.Errorf("not working on a story")
			}

			if !c.Args().Present() {
				return fmt.Errorf("this command requires at least one argument")
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

			// Use the Blast Radius to update deployables
			story.MapBlastRadiusToDeployables()

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
					p.Write(fs, project)
				}
			}

			return nil
		},
	}
}
