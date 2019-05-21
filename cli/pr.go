package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LGUG2Z/story/manifest"
	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

func PRCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "pr",
		Usage: "Opens pull requests for the current story",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "github-api-token", EnvVar: "GITHUB_API_TOKEN", Usage: "GitHub API token to authenticate with"},
			cli.StringFlag{Name: "issue", Usage: "GitHub issue to link PRs to"},
		},
		Action: func(c *cli.Context) error {
			if len(c.String("github-api-token")) == 0 {
				return ErrGitHubAPITokenRequired
			}

			if len(c.String("issue")) == 0 {
				return ErrIssueURLRequired
			}

			if !isStory {
				return ErrNotWorkingOnAStory
			}

			story, err := manifest.LoadStory(fs)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client := getGitHubClient(ctx, c.String("github-api-token"))

			story.Projects[metarepo] = ""

		ProjectLoop:
			for project := range story.Projects {
				newPR := github.NewPullRequest{
					Title:               github.String(story.Name),
					Head:                github.String(story.Name),
					Base:                github.String(trunk),
					Body:                github.String(c.String("issue")),
					MaintainerCanModify: github.Bool(true),
				}

				// Try to create a new pull request
				pullRequest, _, err := client.PullRequests.Create(ctx, story.Orgranisation, project, &newPR)
				if err != nil {
					// If there is already a pull request for this branch
					if strings.Contains(err.Error(), "A pull request already exists") {
						pullRequest, err := getOpenPullRequest(client, ctx, story, project)
						if err != nil {
							switch err.Error() {
							// This should never actually happen
							case ErrCouldNotFindOpenPullRequest(story.Name).Error():
								fmt.Println(err.Error())
								continue ProjectLoop
							// If there is an error making the call to get all pull requests
							default:
								return err
							}
						}

						// Output the URL of the existing open pull request
						color.Green(project)
						fmt.Println(*pullRequest.HTMLURL)
						continue ProjectLoop
					}

					// If there is a branch with no difference from master
					if strings.Contains(err.Error(), "No commits between") {
						color.Green(project)
						fmt.Println("branch is identical to master, can't open a pull request yet")
						continue ProjectLoop
					}

					// Any other error from the call
					return err
				}

				color.Green(project)
				fmt.Println(*pullRequest.HTMLURL)

				time.Sleep(1 * time.Second)
			}
			return nil
		},
	}
}
