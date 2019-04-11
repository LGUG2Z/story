package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/LGUG2Z/story/node"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

// TODO: Add tests
func UnpinCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "unpin",
		Usage: "Unpins code in the current story",
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

			messages := []string{fmt.Sprintf("[story unpin] Unpinning package.json dependencies from '%s' [skip ci]", story.Name)}

			// Unpin dependencies in package.json files from branch
			for project := range story.Projects {
				p := node.PackageJSON{}
				if err := p.Load(fs, project); err != nil {
					return err
				}

				p.ResetPrivateDependencyBranchesToMaster(story.Name)
				if err := p.Write(fs, project); err != nil {
					return err
				}

				// Stage the modified package.json file
				_, err := git.Add(git.AddOpts{Project: project, Files: []string{"package.json"}})
				if err != nil {
					return err
				}

				// Commit the modified package.json file
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
