package git

import (
	"fmt"
	"os/exec"
	"strings"
)

type PushOpts struct {
	Project string
	Remote  string
	Branch  string
}

func hasUnpushedCommits(project string) (bool, error) {
	command := exec.Command("git", "log", "--branches", "--not", "--remotes")
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

func Push(opts PushOpts) (string, error) {
	changes, err := hasUnpushedCommits(opts.Project)
	if err != nil {
		return "", err
	}

	if !changes {
		return "no unpushed commits", err
	}

	var args []string
	args = append(args, "push", "-u", opts.Remote, opts.Branch)

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
