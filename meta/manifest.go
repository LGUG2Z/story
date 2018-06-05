package meta

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/AlexsJones/kepler/commands/node"
	"github.com/LGUG2Z/blastradius/blastradius"
	"github.com/spf13/afero"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Manifest represents a .meta JSON file, enriched with a "story" key
type Manifest struct {
	Fs          afero.Fs          `json:"-"`
	Global      *Manifest         `json:"-"`
	Name        string            `json:"story,omitempty"`
	Deployables map[string]bool   `json:"deployables,omitempty"`
	Primaries   map[string]bool   `json:"primaries,omitempty"`
	Projects    map[string]string `json:"projects,omitempty"`
	BlastRadius map[string]bool   `json:"blast-radius,omitempty"`
}

// IsStory checks if the .meta is a story subset or the global .meta file
func (m *Manifest) IsStory() bool {
	return m.Name != ""
}

// Load reads the contents of a global meta or story meta file into
// a Manifest object
func (m *Manifest) Load(filename string) error {
	bytes, err := afero.ReadFile(m.Fs, filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, &m)
}

// Write takes the current Manifest struct and writes it to disk
func (m *Manifest) Write() error {
	bytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return afero.WriteFile(m.Fs, ".meta", bytes, os.FileMode(0666))
}

// SetStory moves the current global meta file to a backup file and initialises
// a new .meta file for the given story
func (m *Manifest) SetStory(story string) error {
	repository, err := getRepository(".")
	if err != nil {
		return err
	}

	m.Global = &Manifest{Fs: m.Fs}
	if err := m.Global.Load(".meta"); err != nil {
		return err
	}

	if err = CheckoutBranch(story, repository); err != nil {
		return fmt.Errorf("%s: %s", "meta-repo", err)
	}

	_, err = m.Fs.Stat(".meta.json")
	if err != nil {
		// Create story
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

	// Load existing story
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

// Reset moves the current story file to a backup file and restores the
// global .meta.json file
func (m *Manifest) Reset() error {
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

			// TODO: add test case
			if _, exists := m.Deployables[project]; exists {
				m.Deployables[project] = true
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

		if _, exists := m.BlastRadius[project]; exists {
			delete(m.BlastRadius, project)
		}

		if _, exists := m.Primaries[project]; exists {
			delete(m.Primaries, project)
			removed, err := removePrivateDependencies(m.Global, m, project)
			if err != nil {
				return err
			}

			// TODO: add test case
			if _, exists := m.Deployables[project]; exists {
				m.Deployables[project] = false
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

// TODO: Make a PackageJSON type with functions on it for this kind of stuff
func (m *Manifest) updatePackageJSONFiles(packageJSONs map[string]*node.PackageJSON, dep string) {
	for _, pkg := range packageJSONs {
		if _, exists := pkg.Dependencies[dep]; exists {
			if strings.HasSuffix(pkg.Dependencies[dep], ".git") {
				pkg.Dependencies[dep] = fmt.Sprintf("%s#%s", pkg.Dependencies[dep], m.Name)
			}
		}
	}
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

	return m.Write()
}

// Blast calculates the blast radius of projects within the current story and then adds those projects
// to the story, checking out the appropriate branches and updating the appropriate package.json files
// in the process
func (m *Manifest) Blast() error {
	var blastRadius []string

	packageJSONs := make(map[string]*node.PackageJSON)
	blastMap := make(map[string]bool)

	for project := range m.Projects {
		blastMap[project] = true
		packageJSON, err := getPackageJSON(m.Fs, project)
		if err != nil {
			return err
		}

		packageJSONs[project] = packageJSON
	}

	for project := range m.Projects {
		calculated, err := blastradius.Calculate(m.Fs, ".", project)
		if err != nil {
			return err
		}

		for _, prj := range calculated {
			if !blastMap[prj] {
				blastRadius = append(blastRadius, prj)
				blastMap[prj] = true

				packageJSON, err := getPackageJSON(m.Fs, prj)
				if err != nil {
					return err
				}

				packageJSONs[prj] = packageJSON
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

	}

	for project := range m.Projects {
		m.updatePackageJSONFiles(packageJSONs, project)
	}

	for prj, pkg := range packageJSONs {
		bytes, err := json.MarshalIndent(pkg, "", "  ")
		if err != nil {
			return err
		}

		if err := afero.WriteFile(m.Fs, fmt.Sprintf("%s/package.json", prj), bytes, os.FileMode(0666)); err != nil {
			return err
		}
	}

	// TODO: add test case
	for deployable, _ := range m.Deployables {
		if _, exists := m.BlastRadius[deployable]; exists {
			m.Deployables[deployable] = true
		}
	}


	if err := m.Write(); err != nil {
		return err
	}

	fmt.Printf("added projects within the blast radius of this story:\n  %s\n", strings.Join(blastRadius, "\n  "))

	return nil
}

// Complete prepares a story for merging into a master branch by reverting all references to the story branch in all
// package.json files
func (m *Manifest) Complete() error {
	modified := make(map[string]*node.PackageJSON)

	for project := range m.Projects {
		packageJSON, err := getPackageJSON(m.Fs, project)
		if err != nil {
			return err
		}

		storyBranch := fmt.Sprintf("#%s", m.Name)

		for pkg, src := range packageJSON.Dependencies {
			if strings.HasSuffix(src, storyBranch) {
				packageJSON.Dependencies[pkg] = strings.TrimSuffix(src, storyBranch)
				if _, exists := modified[project]; !exists {
					modified[project] = packageJSON
				}
			}
		}
	}

	var saved []string
	for project, packageJSON := range modified {
		bytes, err := json.MarshalIndent(packageJSON, "", "  ")
		if err != nil {
			return err
		}

		filename := fmt.Sprintf("%s/package.json", project)
		if err := afero.WriteFile(m.Fs, filename, bytes, os.FileMode(0666)); err != nil {
			return err
		}
		saved = append(saved, project)
	}

	fmt.Printf("references to branches of this story removed in projects:\n  %s\n", strings.Join(saved, "\n  "))

	return nil
}
