package cli

import (
	"fmt"
)

var ErrAlreadyWorkingOnAStory = fmt.Errorf("already working on a story")
var ErrCommandRequiresAnArgument = fmt.Errorf("this command requires an argument")
var ErrCommandTakesNoArguments = fmt.Errorf("this command takes no arguments")
var ErrNotWorkingOnAStory = fmt.Errorf("not working on a story")
var ErrGitHubAPITokenRequired = fmt.Errorf("an API token for GitHub is required, either using --github-api-token or $GITHUB_API_TOKEN")
var ErrIssueURLRequired = fmt.Errorf("an issue URL is required")

func ErrCouldNotFindOpenPullRequest(story string) error {
	return fmt.Errorf("could not find an open pull request for %s", story)
}

func ErrCouldNotFindClosedPullRequest(story string) error {
	return fmt.Errorf("could not find a closed pull request for %s", story)
}
