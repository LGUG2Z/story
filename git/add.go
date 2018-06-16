package git

import (
	"os/exec"
	"strings"
)

type AddOpts struct {
	Files   []string
	Project string
}

func Add(opts AddOpts) (string, error) {
	var args []string
	args = append(args, "add")

	for _, file := range opts.Files {
		args = append(args, file)
	}

	command := exec.Command("git", args...)
	if opts.Project != "" {
		command.Dir = opts.Project
	}

	combinedOutput, err := command.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(combinedOutput)), nil
}
