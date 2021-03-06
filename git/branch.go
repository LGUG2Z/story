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
		return "", fmt.Errorf("%s: %s", err, combinedOutput)
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
		_, err := CheckoutBranch(CheckoutBranchOpts{Branch: "master", Project: opts.Project})
		if err != nil {
			return "", err
		}
	}

	var args []string

	if opts.Local {
		args = append(args, "branch", "--delete", "--force", opts.Branch)
		command := exec.Command("git", args...)
		if opts.Project != "" {
			command.Dir = opts.Project
		}

		combinedOutput, err := command.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("%s: %s", err, combinedOutput)
		}

		outputs = append(outputs, strings.TrimSpace(string(combinedOutput)))
	}

	if opts.Remote {
		args = append(args, "push", "origin", "--delete", opts.Branch)
		command := exec.Command("git", args...)
		if opts.Project != "" {
			command.Dir = opts.Project
		}

		combinedOutput, err := command.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("%s: %s", err, combinedOutput)
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

func HeadsAreEqual(fs afero.Fs, project, b1, b2 string) (bool, error) {
	h1, err := getHead(fs, project, b1)
	if err != nil {
		return false, err
	}

	h2, err := getHead(fs, project, b2)
	if err != nil {
		return false, err
	}

	return h1 == h2, nil
}

func getHead(fs afero.Fs, project, branch string) (string, error) {
	b, err := afero.ReadFile(fs, fmt.Sprintf("%s/.git/refs/heads/%s", project, branch))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(b)), nil
}
