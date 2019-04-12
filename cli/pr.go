package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/LGUG2Z/story/manifest"
	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
	"golang.org/x/oauth2"
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
				return fmt.Errorf("an API token for GitHub is required, either using --github-api-token or $GITHUB_API_TOKEN")
			}

			if len(c.String("issue")) == 0 {
				return fmt.Errorf("an issue URL is required")
			}

			if !isStory {
				return ErrNotWorkingOnAStory
			}

			story, err := manifest.LoadStory(fs)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client := github.NewClient(
				oauth2.NewClient(
					ctx,
					oauth2.StaticTokenSource(
						&oauth2.Token{AccessToken: c.String("github-api-token")},
					),
				),
			)

		ProjectLoop:
			for project := range story.Projects {
				newPR := github.NewPullRequest{
					Title:               github.String(story.Name),
					Head:                github.String(story.Name),
					Base:                github.String("master"),
					Body:                github.String(c.String("issue")),
					MaintainerCanModify: github.Bool(true),
				}

				_, resp, err := client.PullRequests.Create(ctx, story.Orgranisation, project, &newPR)
				if err != nil {
					if strings.Contains(err.Error(), "A pull request already exists") {
						pullRequests, _, err := client.PullRequests.List(ctx, story.Orgranisation, project, &github.PullRequestListOptions{
							State: "open",
							Base:  "master",
						})

						if err != nil {
							return err
						}

						for _, pr := range pullRequests {
							if *pr.Title == story.Name {
								color.Green(project)
								fmt.Println(*pr.HTMLURL)
								continue ProjectLoop
							}
						}

						return fmt.Errorf("a pull request already exists but the url could not be retrieved")
					}

					if strings.Contains(err.Error(), "No commits between master") {
						color.Green(project)
						fmt.Println("branch is identical to master, can't open a pull request yet")
						continue
					}

					color.Red("call error")
					return err
				}
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}

				r := map[string]interface{}{}
				if err := json.Unmarshal(body, r); err != nil {
					return err
				}

				color.Green(project)
				if url, ok := r["html_url"]; ok {
					fmt.Println(url)
				}

			}

			return nil
		},
	}
}
