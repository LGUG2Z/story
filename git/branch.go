package git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/afero"
)

type CheckoutBranchOpts struct {
	Branch  string
	Create  bool
	Project string
}

func CheckoutBranchWithCreateIfRequired(branch string) (string, error) {
	output, err := CheckoutBranch(CheckoutBranchOpts{Branch: branch})
	if err == nil {
		return output, nil
	}

	if err != nil && err.Error() == fmt.Sprintf("error: pathspec '%s' did not match any file(s) known to git.", branch) {
		err = nil
	}

	if err != nil {
		return "", err
	}

	output, err = CheckoutBranch(CheckoutBranchOpts{Branch: branch, Create: true})
	if err != nil {
		return "", err
	}

	return output, nil
}

func CheckoutBranch(opts CheckoutBranchOpts) (string, error) {
	var args []string
	args = append(args, "checkout")
	if opts.Create {
		args = append(args, "-b")

	}

	args = append(args, opts.Branch)

	command := exec.Command("git", args...)
	if opts.Project != "" {
		command.Dir = opts.Project
	}

	combinedOutput, err := command.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s", err.Error())
	}

	return strings.TrimSpace(string(combinedOutput)), nil
}

type DeleteBranchOpts struct {
	Branch  string
	Local   bool
	Project string
	Remote  bool
}

func DeleteBranch(opts DeleteBranchOpts) (string, error) {
	var outputs []string

	if opts.Local {
		combinedOutput, err := CheckoutBranch(CheckoutBranchOpts{Branch: "master", Project: opts.Project})
		if err != nil {
			return "", fmt.Errorf("%s", strings.TrimSpace(string(combinedOutput)))
		}
	}

	if opts.Local {
		var args []string
		args = append(args, "branch")
		args = append(args, "--delete")
		args = append(args, "--force")
		args = append(args, opts.Branch)

		command := exec.Command("git", args...)
		if opts.Project != "" {
			command.Dir = opts.Project
		}

		combinedOutput, err := command.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("%s", strings.TrimSpace(string(combinedOutput)))
		}

		outputs = append(outputs, strings.TrimSpace(string(combinedOutput)))
	}

	if opts.Remote {
		var args []string
		args = append(args, "push")
		args = append(args, "origin")
		args = append(args, "--delete")
		args = append(args, opts.Branch)

		command := exec.Command("git", args...)
		if opts.Project != "" {
			command.Dir = opts.Project
		}

		combinedOutput, err := command.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("%s", strings.TrimSpace(string(combinedOutput)))
		}

		outputs = append(outputs, strings.TrimSpace(string(combinedOutput)))
	}

	return strings.Join(outputs, "\n"), nil
}

func GetCurrentBranch(fs afero.Fs, project string) (string, error) {
	b, err := afero.ReadFile(fs, fmt.Sprintf("%s/.git/HEAD", project))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(strings.TrimPrefix(string(b), "ref: refs/heads/")), nil
}
