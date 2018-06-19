package git

import (
	"fmt"
	"os/exec"
	"strings"
)

type CloneOpts struct {
	Repository string
	Directory  string
}

func Clone(opts CloneOpts) (string, error) {
	var args []string
	args = append(args, "clone", opts.Repository)

	command := exec.Command("git", args...)
	if opts.Directory != "" {
		command.Args = append(command.Args, opts.Directory)
	}

	combinedOutput, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, combinedOutput)
	}

	return strings.TrimSpace(string(combinedOutput)), nil
}
