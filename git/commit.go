package git

import (
	"fmt"
	"os/exec"
	"strings"
)

type CommitOpts struct {
	Messages []string
	Project  string
}

func hasStagedChanges(project string) (bool, error) {
	command := exec.Command("git", "diff", "--cached", "--name-only")
	if project != "" {
		command.Dir = project
	}

	combinedOutput, err := command.CombinedOutput()
	if err != nil {
		return false, err
	}

	trimmed := strings.TrimSpace(string(combinedOutput))

	return len(trimmed) > 0, nil
}

func Commit(opts CommitOpts) (string, error) {
	changes, err := hasStagedChanges(opts.Project)
	if err != nil {
		return "", err
	}

	if !changes {
		return "no staged changes to commit", err
	}

	var args []string
	args = append(args, "commit")

	for _, message := range opts.Messages {
		args = append(args, "--message")
		args = append(args, message)
	}

	command := exec.Command("git", args...)
	if opts.Project != "" {
		command.Dir = opts.Project
	}

	combinedOutput, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, combinedOutput)
	}

	return strings.TrimSpace(string(combinedOutput)), nil
}
