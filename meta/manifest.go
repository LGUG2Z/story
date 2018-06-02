package meta

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/AlexsJones/kepler/commands/node"
	"github.com/LGUG2Z/blastradius/blastradius"
	"github.com/spf13/afero"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
)

// Manifest represents a .meta JSON file, enriched with a "story" key
type Manifest struct {
	Fs          afero.Fs          `json:"-"`
	Global      *Manifest         `json:"-"`
	Name        string            `json:"story,omitempty"`
	Primaries   map[string]bool   `json:"primaries,omitempty"`
	Projects    map[string]string `json:"projects,omitempty"`
	BlastRadius map[string]bool   `json:"blast-radius,omitempty"`
}

// IsStory checks if the .meta is a story subset or the global .meta file
func (m *Manifest) IsStory() bool {
	return m.Name != ""
}

// Write takes the current Manifest struct and writes it to disk
func (m *Manifest) Write() error {
	bytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	if err := afero.WriteFile(m.Fs, ".meta", bytes, os.FileMode(0666)); err != nil {
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

func (m *Manifest) Blast() error {
	var blastRadius []string
	blastMap := make(map[string]bool)

	for project := range m.Projects {
		blastMap[project] = true
	}

	for project := range m.Projects {
		// TODO: Need to update blastradius to use afero.FS
		calculated, err := blastradius.Calculate(m.Fs, ".", project)
		if err != nil {
			return err
		}

		for _, prj := range calculated {
			if !blastMap[prj] {
				blastRadius = append(blastRadius, prj)
				blastMap[prj] = true
			}
		}
	}

	for _, project := range blastRadius {
		repository, err := getRepository(project)
		if err != nil {
			return err
		}

		err = CheckoutBranch(m.Name, repository)
		if err != nil {
			return fmt.Errorf("%s: %s", project, err)
		}

		if _, exists := m.Projects[project]; !exists {
			m.Projects[project] = fmt.Sprintf("git@github.com:%s/%s.git", os.Getenv("ORGANISATION"), project)
		}

		if m.BlastRadius == nil {
			m.BlastRadius = make(map[string]bool)
		}

		if !m.BlastRadius[project] {
			m.BlastRadius[project] = true
		}

		//UpdatePackageJSON(project, m.Projects)
	}

	if err := m.Write(); err != nil {
		return err
	}

	fmt.Println(strings.Join(blastRadius, ", "))

	return nil
}

// Prune removes from the story all projects where the head of the current story branch
// and the master branch are the same, and reverts any changes made to the package.json
// files of the primary projects that they were included from
func (m *Manifest) Prune() error {
	if err := m.Load(".meta"); err != nil {
		return err
	}

	m.Global = &Manifest{Fs: m.Fs}
	if err := m.Global.Load(".meta.json"); err != nil {
		return err
	}

	var pruned []string

	for project := range m.Projects {
		repository, err := getRepository(project)
		if err != nil {
			return fmt.Errorf("%s: %s", project, err)
		}

		storyRevision := plumbing.Revision(fmt.Sprintf("refs/heads/%s", m.Name))
		masterRevision := plumbing.Revision("master")

		storyHash, err := repository.ResolveRevision(storyRevision)
		if err != nil {
			return err
		}

		masterHash, err := repository.ResolveRevision(masterRevision)
		if err != nil {
			return err
		}

		if storyHash.String() == masterHash.String() {
			if err := m.RemoveProjects([]string{project}); err != nil {
				return err
			}

			pruned = append(pruned, project)
		}
	}

	if len(pruned) > 0 {
		fmt.Printf("pruned unchanged projects and their dependencies:\n  %s\n", strings.Join(pruned, "\n  "))
	}

	if err := m.Write(); err != nil {
		return err
	}

	return nil
}

// Load reads the contents of a global meta or story meta file into
// a Manifest object
func (m *Manifest) Load(filename string) error {
	bytes, err := afero.ReadFile(m.Fs, filename)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, &m); err != nil {
		return err
	}

	return nil
}

// RemoveProjects removes one or more projects from the current Manifest object
// and then writes that updated object back to disk
func (m *Manifest) RemoveProjects(projects []string) error {
	m.Global = &Manifest{Fs: m.Fs}
	if err := m.Global.Load(".meta.json"); err != nil {
		return err
	}

	for _, project := range projects {
		if _, exists := m.Projects[project]; exists {
			delete(m.Projects, project)

			repository, err := getRepository(project)
			if err != nil {
				return err
			}

			if err := DeleteBranch(m.Name, repository); err != nil {
				return err
			}

			fmt.Printf("removed: %s\n", project)
		}

		if _, exists := m.Primaries[project]; exists {
			delete(m.Primaries, project)
			removed, err := removePrivateDependencies(m.Global, m, project)
			if err != nil {
				return err
			}

			fmt.Printf("removed as dependencies of %s:\n  %s\n", project, strings.Join(removed, "\n  "))

		} else {
			for prj := range m.Primaries {
				packageJSON := fmt.Sprintf("%s/package.json", prj)
				bytes, err := afero.ReadFile(m.Fs, packageJSON)
				if err != nil {
					return err
				}

				p := node.PackageJSON{}
				if err = json.Unmarshal(bytes, &p); err != nil {
					return err
				}

				storyBranch := fmt.Sprintf("#%s", m.Name)
				if _, exists := p.Dependencies[project]; exists {
					if strings.HasSuffix(p.Dependencies[project], storyBranch) {
						p.Dependencies[project] = strings.TrimSuffix(p.Dependencies[project], storyBranch)
					}
				}

				bytes, err = json.MarshalIndent(p, "", "  ")
				if err != nil {
					return err
				}

				if err := afero.WriteFile(m.Fs, packageJSON, bytes, os.FileMode(0666)); err != nil {
					return err
				}
			}
		}
	}

	err := m.Write()
	return err
}

// AddProjects adds one or more projects to the current Manifest object
// and then writes that updated object back to disk
func (m *Manifest) AddProjects(projects []string) error {
	m.Global = &Manifest{Fs: m.Fs}
	if err := m.Global.Load(".meta.json"); err != nil {
		return err
	}

	var skipped []string

	for _, project := range projects {
		if _, exists := m.Global.Projects[project]; exists {
			if m.Projects == nil {
				m.Projects = make(map[string]string)
			}

			if m.Primaries == nil {
				m.Primaries = make(map[string]bool)
			}

			if !m.Primaries[project] {
				m.Primaries[project] = true
			}

			if _, exists := m.Projects[project]; !exists {
				m.Projects[project] = fmt.Sprintf("git@github.com:%s/%s.git", os.Getenv("ORGANISATION"), project)

				repository, err := getRepository(project)
				if err != nil {
					return err
				}

				if err := CheckoutBranch(m.Name, repository); err != nil {
					return fmt.Errorf("%s: %s", project, err)
				}

				fmt.Printf("added: %s\n", project)
			}

			added, err := addPrivateDependencies(m.Global, m, project)
			if err != nil {
				fmt.Printf("problem reading %s/package.json\n", project)
			}

			if len(added) > 0 {
				fmt.Printf("added as dependencies of %s:\n  %s\n", project, strings.Join(added, "\n  "))
			}
		} else {
			skipped = append(skipped, project)
			fmt.Printf("skipped: %s\n", strings.Join(skipped, ", "))
		}
	}

	err := m.Write()
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

// RestoreGlobal moves the current story file to a backup file and restores the
// global .meta.json file
func (m *Manifest) RestoreGlobal() error {
	if err := m.Load(".meta"); err != nil {
		return err
	}

	for project := range m.Projects {
		repository, err := getRepository(project)
		if err != nil {
			return err
		}

		if err := CheckoutBranch("master", repository); err != nil {
			return fmt.Errorf("%s: %s", project, err)
		}
	}

	repository, err := getRepository(".")
	if err != nil {
		return err
	}

	if err := CheckoutBranch("master", repository); err != nil {
		return fmt.Errorf("%s: %s", "meta-repo", err)
	}

	return nil
}

// SetStory moves the current global meta file to a backup file and initialises
// a new .meta file for the given story
func (m *Manifest) SetStory(story string) error {
	repository, err := getRepository(".")
	if err != nil {
		return err
	}

	if err := CheckoutBranch(story, repository); err != nil {
		return fmt.Errorf("%s: %s", "meta-repo", err)
	}

	_, err = m.Fs.Stat(".meta.json")
	if err != nil {
		if err := m.Fs.Rename(".meta", ".meta.json"); err != nil {
			return err
		}

		m.Name = story
		m.Projects = nil
		m.Primaries = nil
		m.BlastRadius = nil

		bytes, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return err
		}

		return afero.WriteFile(m.Fs, ".meta", bytes, os.FileMode(0666))
	}

	m = &Manifest{Fs: m.Fs}
	if err := m.Load(".meta"); err != nil {
		return err
	}

	if m.Projects != nil {
		for project := range m.Projects {
			repository, err := getRepository(project)
			if err != nil {
				return err
			}

			if err := CheckoutBranch(m.Name, repository); err != nil {
				return fmt.Errorf("%s: %s", project, err)
			}
		}
	}

	return nil
}
