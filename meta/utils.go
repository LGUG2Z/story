package meta

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/AlexsJones/kepler/commands/node"
	"github.com/spf13/afero"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
)

func DeleteBranch(story string, repository *git.Repository) error {
	if os.Getenv("TEST") == "1" {
		return nil
	}

	storyReference := plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", story))
	workTree, err := repository.Worktree()
	if err != nil {
		return err
	}

	if err := workTree.Checkout(&git.CheckoutOptions{}); err != nil {
		return err
	}

	if repository.Storer.RemoveReference(storyReference); err != nil {
		return err
	}

	return nil
}

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
		err := workTree.Checkout(&git.CheckoutOptions{})
		return err
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

func removePrivateDependencies(meta, story *Manifest, project string) ([]string, error) {
	packageJSON := fmt.Sprintf("%s/package.json", project)
	bytes, err := afero.ReadFile(story.Fs, packageJSON)
	if err != nil {
		return nil, err
	}

	p := node.PackageJSON{}
	if err = json.Unmarshal(bytes, &p); err != nil {
		return nil, err
	}

	var removed []string

	for dep := range p.Dependencies {
		if _, exists := meta.Projects[dep]; exists {
			if _, exists := story.Projects[dep]; exists {
				delete(story.Projects, dep)
				removed = append(removed, dep)

				repository, err := getRepository(dep)
				if err != nil {
					return nil, err
				}

				if err := DeleteBranch(story.Name, repository); err != nil {
					return nil, err
				}

				storyBranch := fmt.Sprintf("#%s", story.Name)
				if strings.HasSuffix(p.Dependencies[dep], storyBranch) {
					p.Dependencies[dep] = strings.TrimSuffix(p.Dependencies[dep], storyBranch)
				}
			}
		}
	}

	bytes, err = json.MarshalIndent(p, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := afero.WriteFile(story.Fs, packageJSON, bytes, os.FileMode(0666)); err != nil {
		return nil, err
	}

	for project := range story.Primaries {
		_, err := addPrivateDependencies(meta, story, project)
		if err != nil {
			fmt.Printf("there was a problem reading %s/package.json\n", project)
		}
	}

	return removed, nil
}

func addPrivateDependencies(meta, story *Manifest, project string) ([]string, error) {
	packageJSON := fmt.Sprintf("%s/package.json", project)

	bytes, err := afero.ReadFile(story.Fs, packageJSON)
	if err != nil {
		return nil, err
	}

	p := node.PackageJSON{}
	if err = json.Unmarshal(bytes, &p); err != nil {
		return nil, err
	}

	var added []string

	for dep := range p.Dependencies {
		if _, exists := meta.Projects[dep]; exists {
			if _, exists := story.Projects[dep]; !exists {
				story.Projects[dep] = fmt.Sprintf("git@github.com:%s/%s.git", os.Getenv("ORGANISATION"), dep)
				added = append(added, dep)

				repository, err := getRepository(dep)
				if err != nil {
					return nil, err
				}

				if err := CheckoutBranch(story.Name, repository); err != nil {
					return nil, fmt.Errorf("%s: %s", project, err)
				}
			}

			if strings.HasSuffix(p.Dependencies[dep], ".git") {
				p.Dependencies[dep] = fmt.Sprintf("%s#%s", p.Dependencies[dep], story.Name)
			}
		}
	}

	bytes, err = json.MarshalIndent(p, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := afero.WriteFile(story.Fs, packageJSON, bytes, os.FileMode(0666)); err != nil {
		return nil, err
	}

	return added, nil
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

func getRepository(project string) (*git.Repository, error) {
	if os.Getenv("TEST") == "1" {
		return nil, nil
	}

	projectDotGit := fmt.Sprintf("%s/.git", project)

	s, err := filesystem.NewStorage(osfs.New(projectDotGit))
	if err != nil {
		return nil, err
	}

	wt, err := filesystem.NewStorage(osfs.New(project))
	if err != nil {
		return nil, err
	}

	repository, err := git.Open(s, wt.Filesystem())
	if err != nil {
		return nil, err
	}

	return repository, nil
}
