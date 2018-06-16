package cli

import (
	"fmt"

	"sort"

	"os"
	"strings"

	"github.com/LGUG2Z/blastradius/blastradius"
	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/LGUG2Z/story/node"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

func CreateCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "create",
		Usage: "Creates a new story",
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

func LoadCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "load",
		Usage: "Loads an existing story",
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

func ResetCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "reset",
		Usage: "Resets all story branches to trunk branches",
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

func ListCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:      "list",
		ShortName: "ls",
		Usage:     "Shows a list of projects added to the current story",
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
				fmt.Println(project)
			}

			return nil
		},
	}
}

func ArtifactsCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:      "artifacts",
		ShortName: "art",
		Usage:     "Shows a list of artifacts to be built and deployed for the current story",
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

			for project := range story.Artifacts {
				if story.Artifacts[project] {
					fmt.Println(project)
				}
			}

			return nil
		},
	}
}

func AddCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:      "add",
		ShortName: "a",
		Usage:     "Adds a project to the current story",
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

			// Use the Blast Radius to update artifacts
			story.MapBlastRadiusToArtifacts()

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

func RemoveCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:      "remove",
		ShortName: "rm",
		Usage:     "Removes a project from the current story",
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

			// Use the Blast Radius to update artifacts
			story.MapBlastRadiusToArtifacts()

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

func CommitCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:      "commit",
		ShortName: "co",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "message, m", Usage: "Commit message"},
		},
		Action: cli.ActionFunc(func(c *cli.Context) error {
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

			messages := []string{c.String("message")}
			for project := range story.Projects {
				output, err := git.Commit(git.CommitOpts{Project: project, Messages: messages})
				if err != nil {
					return err
				}

				printGitOutput(output, project)
			}

			var hashMessages []string
			var storyHashes []string

			hashes, err := story.GetCommitHashes(fs)
			for project, hash := range hashes {
				commitUrl := fmt.Sprintf("https://github.com/%s/%s/commit/%s", story.Orgranisation, project, hash)
				projectHash := fmt.Sprintf("%s: %s", project, hash)
				hashMessages = append(hashMessages, commitUrl)
				storyHashes = append(storyHashes, projectHash)
			}

			sort.Strings(hashMessages)
			sort.Strings(storyHashes)

			b := []byte(strings.Join(storyHashes, "\n"))
			if err := afero.WriteFile(fs, ".storyhash", b, os.FileMode(0666)); err != nil {
				return err
			}

			output, err := git.Add(git.AddOpts{Files: []string{".storyhash"}})
			if err != nil {
				return err
			}

			messages = append(messages, strings.Join(hashMessages, "\n"))

			output, err = git.Commit(git.CommitOpts{Messages: messages})
			if err != nil {
				return err
			}

			printGitOutput(output, "metarepo")

			return nil
		}),
	}
}
