package cli

import (
	"fmt"

	"sort"

	"strings"

	"os"

	"github.com/LGUG2Z/blastradius/blastradius"
	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/LGUG2Z/story/node"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

var ErrAlreadyWorkingOnAStory = fmt.Errorf("already working on a story")
var ErrCommandRequiresAnArgument = fmt.Errorf("this command requires an argument")
var ErrCommandTakesNoArguments = fmt.Errorf("this command takes no arguments")
var ErrNotWorkingOnAStory = fmt.Errorf("not working on a story")

func CreateCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "create",
		Usage: "Creates a new story",
		Action: func(c *cli.Context) error {
			if isStory {
				return ErrAlreadyWorkingOnAStory
			}

			if !c.Args().Present() {
				return ErrCommandRequiresAnArgument
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
				return ErrAlreadyWorkingOnAStory
			}

			if !c.Args().Present() {
				return ErrCommandRequiresAnArgument
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
				if err := ensureProjectIsCloned(fs, story, project); err != nil {
					return err
				}

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

func BlastRadiusCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:      "blastradius",
		ShortName: "br",
		Usage:     "Shows a list of current story's blast radius",
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

func ArtifactsCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:      "artifacts",
		ShortName: "art",
		Usage:     "Shows a list of artifacts to be built and deployed for the current story",
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

func AddCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:      "add",
		ShortName: "a",
		Usage:     "Adds a project to the current story",
		Flags:     []cli.Flag{cli.BoolFlag{Name: "ci", Usage: "clone without modifying .meta"}},
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
					p.Write(fs, project)
				}
			}

			return nil
		},
	}
}

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
				p := node.PackageJSON{}
				if err := p.Load(fs, project); err != nil {
					return err
				}

				p.SetPrivateDependencyBranchesToCommitHashes(story, projectList...)
				p.Write(fs, project)

				_, err := git.Add(git.AddOpts{Files: []string{"package.json"}, Project: project})
				if err != nil {
					return err
				}

				output, err := git.Diff(project)
				if err != nil {
					return err
				}

				printGitOutput(output, project)
			}

			return nil
		}),
	}
}

func CommitCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:      "commit",
		ShortName: "co",
		Usage:     "Commits code across the current story",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "message, m", Usage: "Commit message"},
		},
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

			// Commit in all the projects
			messages := []string{c.String("message")}
			for project := range story.Projects {
				output, err := git.Commit(git.CommitOpts{Project: project, Messages: messages})
				if err != nil {
					return err
				}

				printGitOutput(output, project)
			}

			// Update the hashes in the meta file and write it out
			hashes, err := story.GetCommitHashes(fs)
			if err != nil {
				return err
			}

			story.Hashes = hashes
			if err := story.Write(fs); err != nil {
				return err
			}

			// Format the hashes to the GitHub format to link to a specific commit
			var hashMessages []string
			for project, hash := range hashes {
				// TODO: switch depending on GitHub or GitLab for now. Maybe more later
				commitUrl := fmt.Sprintf("https://github.com/%s/%s/commit/%s", story.Orgranisation, project, hash)
				hashMessages = append(hashMessages, commitUrl)
			}

			// Add the hashes to the slice for git commit messages
			sort.Strings(hashMessages)
			messages = append(messages, strings.Join(hashMessages, "\n"))

			// Add the blast radius to the slice for git commit messages
			var brMap = make(map[string]bool)

			for _, br := range story.BlastRadius {
				for _, p := range br {
					if !brMap[p] {
						brMap[p] = true
					}
				}
			}

			var brSlice []string
			for project := range brMap {
				brSlice = append(brSlice, project)
			}

			messages = append(messages, fmt.Sprintf("Blast Radius: %s", strings.Join(brSlice, " ")))

			// Stage the story file
			output, err := git.Add(git.AddOpts{Files: []string{".meta"}})
			if err != nil {
				return err
			}

			// Commit on the metarepo
			output, err = git.Commit(git.CommitOpts{Messages: messages})
			if err != nil {
				return err
			}

			printGitOutput(output, "metarepo")

			return nil
		}),
	}
}

func PushCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:      "push",
		ShortName: "p",
		Usage:     "Pushes commits across the current story",
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

			// Push in all the projects
			for project := range story.Projects {
				output, err := git.Push(git.PushOpts{Remote: "origin", Branch: story.Name, Project: project})
				if err != nil {
					return err
				}

				printGitOutput(output, project)
			}

			// Commit on the metarepo
			output, err := git.Push(git.PushOpts{Branch: story.Name, Remote: "origin"})
			if err != nil {
				return err
			}

			printGitOutput(output, "metarepo")

			return nil
		}),
	}
}

// TODO: Add tests
func PrepareCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "prepare",
		Usage: "Prepares a story for merges to trunk",
		Flags: []cli.Flag{cli.BoolFlag{Name: "hash", Usage: "prepare with commit hashes"}},
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

			mergePrepMessage := fmt.Sprintf("Preparing story %s for merge", story.Name)

			// Unpin dependencies in package.json files from branch
			for project := range story.Projects {
				p := node.PackageJSON{}
				if err := p.Load(fs, project); err != nil {
					return err
				}

				if c.Bool("hash") {
					p.ResetPrivateDependencyBranchesToCommitHash(story)
				} else {
					p.ResetPrivateDependencyBranchesToMaster(story.Name)
				}

				p.Write(fs, project)

				// Stage the modified package.json file
				_, err := git.Add(git.AddOpts{Project: project, Files: []string{"package.json"}})
				if err != nil {
					return err
				}

				// Commit the modified package.json file
				output, err := git.Commit(git.CommitOpts{Project: project, Messages: []string{mergePrepMessage}})
				if err != nil {
					return err
				}

				printGitOutput(output, project)
			}

			// Update the story hashes
			hashes, err := story.GetCommitHashes(fs)
			if err != nil {
				return err
			}

			story.Hashes = hashes

			// Create the story folder if it doesn't exist
			exists, err := afero.DirExists(fs, "story")
			if err != nil {
				return err
			}

			if !exists {
				if err := fs.Mkdir("story", os.FileMode(0700)); err != nil {
					return err
				}
			}

			// Write the story .meta to the story folder and stage the file
			storyNameWithoutSlash := strings.Replace(story.Name, "/", "-", 1)
			if err := story.WriteToLocation(fs, fmt.Sprintf("story/%s.json", storyNameWithoutSlash)); err != nil {
				return err
			}

			_, err = git.Add(git.AddOpts{Project: "story", Files: []string{fmt.Sprintf("%s.json", storyNameWithoutSlash)}})
			if err != nil {
				return err
			}

			// Recreate the .meta from the story .meta
			m := manifest.Meta{Projects: story.AllProjects, Artifacts: story.Artifacts, Organisation: story.Orgranisation}
			for artifact := range m.Artifacts {
				m.Artifacts[artifact] = false
			}

			// Write the reconstructed .meta to the metarepo folder and stage the file
			if err := m.Write(fs); err != nil {
				return err
			}

			_, err = git.Add(git.AddOpts{Files: []string{".meta"}})
			if err != nil {
				return err
			}

			// Format the hashes to the GitHub format to link to a specific commit
			var hashMessages []string
			for project, hash := range hashes {
				commitUrl := fmt.Sprintf("https://github.com/%s/%s/commit/%s", story.Orgranisation, project, hash)
				hashMessages = append(hashMessages, commitUrl)
			}

			sort.Strings(hashMessages)

			// Commit on the metarepo
			output, err := git.Commit(git.CommitOpts{Messages: []string{mergePrepMessage, strings.Join(hashMessages, "\n")}})
			if err != nil {
				return err
			}

			printGitOutput(output, "metarepo")

			return nil
		}),
	}
}
