package cli

import (
	"context"
	"fmt"

	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/spf13/afero"
	"golang.org/x/oauth2"
)

func printGitOutput(output, project string) {
	color.Green(project)
	fmt.Println(output)
}

func ensureProjectIsCloned(fs afero.Fs, story *manifest.Story, project string) error {
	exists, err := afero.DirExists(fs, project)
	if err != nil {
		return err
	}

	if !exists {
		output, err := git.Clone(git.CloneOpts{Repository: story.AllProjects[project]})
		if err != nil {
			return err
		}

		printGitOutput(output, project)
	}

	return nil
}

func getGitHubClient(ctx context.Context, token string) *github.Client {
	return github.NewClient(
		oauth2.NewClient(
			ctx,
			oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			),
		),
	)
}

func getOpenPullRequest(client *github.Client, ctx context.Context, story *manifest.Story, project string) (*github.PullRequest, error) {
	pullRequests, _, err := client.PullRequests.List(ctx, story.Orgranisation, project, &github.PullRequestListOptions{
		State: "open",
		Base:  trunk,
	})

	if err != nil {
		return nil, err
	}

	for _, pr := range pullRequests {
		if *pr.Title == story.Name {
			return pr, nil
		}
	}

	return nil, ErrCouldNotFindOpenPullRequest(story.Name)
}

func getClosedPullRequest(client *github.Client, ctx context.Context, story *manifest.Story, project string) (*github.PullRequest, error) {
	pullRequests, _, err := client.PullRequests.List(ctx, story.Orgranisation, project, &github.PullRequestListOptions{
		State: "closed",
		Base:  trunk,
	})

	if err != nil {
		return nil, err
	}

	for _, pr := range pullRequests {
		if *pr.Title == story.Name {
			return pr, nil
		}
	}

	return nil, ErrCouldNotFindClosedPullRequest(story.Name)
}
