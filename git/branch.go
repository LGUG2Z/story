package git

import (
	"os/exec"
	"github.com/spf13/afero"
	"fmt"
	"strings"
)

type CheckoutBranchOpts struct {
	Branch string
	Create bool
}

func CheckoutBranch(opts CheckoutBranchOpts) (string, error) {
	args := []string{}
	args = append(args, "checkout")
	if opts.Create {
		args = append(args, "-b")

	}
	args = append(args, opts.Branch)

	command := exec.Command("git", args...)
	combinedOutput, err := command.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(combinedOutput), nil
}

type DeleteBranchOpts struct {
	Branch string
	Local  bool
	Remote bool
}

func DeleteBranch(opts DeleteBranchOpts) (string, error) {
	var outputs []string

	if opts.Local {
		var args []string
		args = append(args, "branch")
		args = append(args, "--delete")
		args = append(args, "--force")
		args = append(args, opts.Branch)

		command := exec.Command("git", args...)
		combinedOutput, err := command.CombinedOutput()
		if err != nil {
			return "", err
		}

		outputs = append(outputs, string(combinedOutput))
	}

	if opts.Remote {
		// git push <remote_name> --delete <branch_name>
		var args []string
		args = append(args, "push")
		args = append(args, "origin")
		args = append(args, "--delete")
		args = append(args, opts.Branch)

		command := exec.Command("git", args...)
		combinedOutput, err := command.CombinedOutput()
		if err != nil {
			return "", err
		}

		outputs = append(outputs, string(combinedOutput))
	}


	return strings.Join(outputs, "\n"), nil
}

func GetCurrentBranch(fs afero.Fs, project string) (string, error) {
	b, err := afero.ReadFile(fs, fmt.Sprintf("%s/.git/HEAD", project))
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(string(b), "ref: refs/heads/"), nil
}
