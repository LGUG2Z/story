package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func Diff(project string) (string, error) {
	var args []string
	args = append(args,"--no-pager", "diff", "-U0", "HEAD~1")

	command := exec.Command("git", args...)
	command.Dir = project

	combinedOutput, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, combinedOutput)
	}

	return strings.TrimSpace(string(combinedOutput)), nil
}
