package cli

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/spf13/afero"
	"github.com/urfave/cli"
)

func MergeCmd(fs afero.Fs) cli.Command {
	return cli.Command{
		Name:  "merge",
		Usage: "Merges prepared code to master branches across the current story",
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "github", Usage: "Use the GitHub squash and merge implementation via API"},
			cli.StringFlag{Name: "github-api-token", EnvVar: "GITHUB_API_TOKEN", Usage: "GitHub API token to authenticate with"},
		},
		Action: cli.ActionFunc(func(c *cli.Context) error {
			if !isStory {
				return ErrNotWorkingOnAStory
			}

			if c.Args().Present() {
				return ErrCommandTakesNoArguments
			}

			currentBranch, err := git.GetCurrentBranch(fs, ".")
			if err != nil {
				return err
			}

			story, err := manifest.LoadStoryFromBranchName(fs, currentBranch)
			if err != nil {
				return err
			}

			messages := []string{fmt.Sprintf("[story merge] Merge branch '%s'", story.Name)}

			if c.Bool("github") {
				if len(c.String("github-api-token")) == 0 {
					return ErrGitHubAPITokenRequired
				}

				ctx := context.Background()
				client := getGitHubClient(ctx, c.String("github-api-token"))

				story.Projects[metarepo] = ""

				for project := range story.Projects {
					color.Green(project)
					// Get the open pull request
					openPullRequest, err := getOpenPullRequest(client, ctx, story, project)
					if err != nil {
						switch err.Error() {
						// If there is no open pull request
						case ErrCouldNotFindOpenPullRequest(story.Name).Error():
							// Get the closed pull request
							closedPullRequest, err := getClosedPullRequest(client, ctx, story, project)
							if err != nil {
								switch err.Error() {
								// If there is no closed pull request, a pull request was never opened
								case ErrCouldNotFindClosedPullRequest(story.Name).Error():
									fmt.Printf("could not find an open or closed pull request for %s\n", story.Name)
									continue
								// Something else went wrong here with the call to list closed pull requests
								default:
									fmt.Println(err)
									continue
								}
							}

							// Report that the pull request has already been closed with a link
							fmt.Println("pull request has already been closed")
							fmt.Println(*closedPullRequest.HTMLURL)
							continue
						// Something else went wrong here with the call to list open pull requests
						default:
							return err
						}
					}

					_, resp, err := client.PullRequests.Merge(ctx, story.Orgranisation, project, *openPullRequest.Number, "", &github.PullRequestOptions{
						CommitTitle: messages[0],
						MergeMethod: "squash",
						SHA:         story.Hashes[project],
					})

					if err != nil {
						return err
					}

					switch resp.StatusCode {
					case http.StatusOK:
						fmt.Println("pull request successfully merged")
					case http.StatusMethodNotAllowed:
						fmt.Println("pull request is not mergeable")
					case http.StatusConflict:
						fmt.Println("head branch was modified, review and try the merge again")
					}

					fmt.Println(*openPullRequest.HTMLURL)
					time.Sleep(1 * time.Second)
				}

				return nil
			}

			// Checkout master and merge story in all projects
			for project := range story.Projects {
				color.Green(project)
				checkoutBranchOutput, err := git.CheckoutBranch(git.CheckoutBranchOpts{
					Branch:  trunk,
					Project: project,
					Create:  false,
				})

				if err != nil {
					return err
				}

				fmt.Printf("%s\n\n", checkoutBranchOutput)

				mergeOutput, err := git.Merge(git.MergeOpts{
					SourceBranch:      story.Name,
					DestinationBranch: trunk,
					Project:           project,
					Squash:            true,
				})

				if err != nil {
					return err
				}

				fmt.Printf("%s\n\n", mergeOutput)

				commitOutput, err := git.Commit(
					git.CommitOpts{
						Project:  project,
						Messages: messages,
					},
				)

				if err != nil {
					return err
				}

				fmt.Println(commitOutput)
			}

			color.Green(metarepo)
			checkoutBranchOutput, err := git.CheckoutBranch(git.CheckoutBranchOpts{
				Branch: trunk,
				Create: false,
			})

			if err != nil {
				return err
			}

			fmt.Printf("%s\n\n", checkoutBranchOutput)

			// Merge story into master on the metarepo
			mergeOutput, err := git.Merge(git.MergeOpts{
				SourceBranch:      story.Name,
				DestinationBranch: trunk,
				Squash:            true,
			})

			if err != nil {
				return err
			}

			fmt.Printf("%s\n\n", mergeOutput)

			commitOutput, err := git.Commit(git.CommitOpts{Messages: messages})
			if err != nil {
				return err
			}

			fmt.Println(commitOutput)

			return nil
		}),
	}
}
