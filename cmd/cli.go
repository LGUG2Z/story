package cmd

import (
	"fmt"

	"time"

	"github.com/LGUG2Z/blastradius/blastradius"
	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/LGUG2Z/story/node"
	"github.com/fatih/color"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

var fs afero.Fs
var isStory bool
var trunk string

func Execute(args ...string) error {
	app := cli.NewApp()
	app.Name = "story"
	app.Usage = "A workflow tool for implementing stories across a meta-repo"
	app.EnableBashCompletion = true
	app.Compiled = time.Now()
	app.Authors = []cli.Author{{
		Name:  "J. Iqbal",
		Email: "jade@beamery.com",
	}}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "trunk",
			Value: "master",
		},
	}

	app.Before = func(c *cli.Context) error {
		fs = afero.NewOsFs()
		trunk = c.String("trunk")
		branch, err := git.GetCurrentBranch(fs, ".")
		if err != nil {
			return err
		}

		isStory = branch != trunk

		return nil
	}

	set := cli.Command{
		Name: "set",
		Action: cli.ActionFunc(func(c *cli.Context) error {
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
			output, err := git.CheckoutBranchWithCreateIfRequired(name)
			if err != nil {
				return err
			}

			printGitOutput(output, "metarepo")

			if err := meta.MoveForStory(fs); err != nil {
				return err
			}

			if err := story.Write(fs); err != nil {
				return err
			}

			return nil
		}),
	}

	reset := cli.Command{
		Name: "reset",
		Action: cli.ActionFunc(func(c *cli.Context) error {
			if !isStory {
				return fmt.Errorf("not working on a story")
			}

			if c.Args().Present() {
				return fmt.Errorf("this command takes no arguments")
			}

			story, err := manifest.LoadStory(fs)
			for _, project := range story.Projects {
				output, err := git.CheckoutBranch(git.CheckoutBranchOpts{Branch: "master", Project: project})
				if err != nil {
					return err
				}

				printGitOutput(output, project)
			}

			output, err := git.CheckoutBranch(git.CheckoutBranchOpts{Branch: "master"})
			if err != nil {
				return err
			}

			printGitOutput(output, "metarepo")
			return nil
		}),
	}

	add := cli.Command{
		Name: "add",
		Action: cli.ActionFunc(func(c *cli.Context) error {
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
			for project, _ := range story.Projects {
				projectList = append(projectList, project)
			}

			// Update all of the package.json files where any other added project is used
			for project, _ := range story.Projects {
				p := node.PackageJSON{}
				if err := p.Load(fs, project); err != nil {
					return err
				}

				p.SetPrivateDependencyBranchesToStory(story.Name, projectList...)
				p.Write(fs, project)
			}

			return nil
		}),
	}

	remove := cli.Command{
		Name: "remove",
		Action: cli.ActionFunc(func(c *cli.Context) error {
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
			for project, _ := range story.Projects {
				projectList = append(projectList, project)
			}

			// Update all of the package.json files where any removed project is used
			for project, _ := range story.Projects {
				p := node.PackageJSON{}
				if err := p.Load(fs, project); err != nil {
					return err
				}

				for _, toReset := range c.Args() {
					p.ResetPrivateDependencyBranches(toReset, story.Name, projectList...)
					p.Write(fs, project)
				}
			}

			return nil
		}),
	}
	commands := []cli.Command{
		set,
		reset,
		add,
		remove,
	}

	app.Commands = commands

	return app.Run(args)
}

func printGitOutput(output, project string) {
	color.Green(project)
	fmt.Println(output)
}
