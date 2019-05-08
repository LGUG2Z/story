package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

// TODO: Add tests
func PrepareCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "prepare",
		Usage: "Prepares a story for merges to trunk",
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

			mergePrepMessage := fmt.Sprintf("[story prepare] Preparing '%s' for merge [skip ci]", story.Name)

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
			storyNameWithoutSlash := strings.ReplaceAll(story.Name, "/", "-")
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

			printGitOutput(output, metarepo)

			return nil
		}),
	}
}
