package meta

import (
	"encoding/json"
	"fmt"
	"os"
	"github.com/AlexsJones/kepler/commands/node"
	"github.com/spf13/afero"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4"
)

// CheckoutBranch is a helper to check out either an existing or a new branch in a git repository
func CheckoutBranch(branch string, repository *git.Repository) error {
	if os.Getenv("TEST") == "1" {
		return nil
	}

	workTree, err := repository.Worktree()
	if err != nil {
		return fmt.Errorf("getting worktree: %s", err)
	}

	head, err := repository.Head()
	ref := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch))

	if branch == "master" {
		return workTree.Checkout(&git.CheckoutOptions{})
	}

	err = workTree.Checkout(&git.CheckoutOptions{
		Branch: ref,
		Hash:   head.Hash(),
		Create: true,
	})

	if err != nil && err.Error() != fmt.Sprintf(`a branch named "refs/heads/%s" already exists`, ref.String()) {
		err = workTree.Checkout(&git.CheckoutOptions{Branch: ref})
		return err
	}

	if err != nil {
		return fmt.Errorf("creating branch: %s", err)
	}

	return err
}

// DeleteBranch is a helper for deleting local branches in a git repository
func DeleteBranch(story string, repository *git.Repository) error {
	if os.Getenv("TEST") == "1" {
		return nil
	}

	storyReference := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", story))
	workTree, err := repository.Worktree()
	if err != nil {
		return err
	}

	if err = workTree.Checkout(&git.CheckoutOptions{}); err != nil {
		return err
	}

	if repository.Storer.RemoveReference(storyReference); err != nil {
		return err
	}

	return nil
}


func getPackageJSON(fs afero.Fs, project string) (*node.PackageJSON, error) {
	packageJSON := &node.PackageJSON{}
	bytes, err := afero.ReadFile(fs, fmt.Sprintf("%s/package.json", project))
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bytes, packageJSON); err != nil {
		return nil, err
	}

	return packageJSON, nil
}

