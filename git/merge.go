package git

import (
	"fmt"
	"os/exec"
	"strings"
)

type FetchOpts struct {
	Branch  string
	Remote  string
	Project string
}

func Fetch(opts FetchOpts) (string, error) {
	var args []string
	args = append(args, "fetch", opts.Remote, fmt.Sprintf("%s:%s", opts.Branch, opts.Branch))

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

type MergeOpts struct {
	SourceBranch      string
	DestinationBranch string
	Project           string
	Squash            bool
}

func Merge(opts MergeOpts) (string, error) {
	var args []string
	args = append(args, "merge", opts.SourceBranch)
	if opts.Squash {
		args = append(args, "--squash")
	}

	args = append(args, opts.SourceBranch)

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
