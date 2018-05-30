package meta

import (
	"encoding/json"
	"fmt"
	"github.com/AlexsJones/kepler/commands/node"
	"github.com/spf13/afero"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"os"
	"strings"
)

// Manifest represents a .meta JSON file, enriched with a "story" key
type Manifest struct {
	Fs        afero.Fs          `json:"-"`
	Global    *Manifest         `json:"-"`
	Name      string            `json:"story,omitempty"`
	Primaries map[string]bool   `json:"primaries,omitempty"`
	Projects  map[string]string `json:"projects,omitempty"`
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
		repository, err := git.PlainOpen(project)
		if err != nil {
			return err
		}

		storyHash, err := repository.ResolveRevision(plumbing.Revision(fmt.Sprintf("refs/heads/%s", m.Name)))
		if err != nil {
			return err
		}

		masterHash, err := repository.ResolveRevision(plumbing.Revision("master"))
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
	if err := m.Fs.Rename(".meta", fmt.Sprintf(".meta.%s", m.Name)); err != nil {
		return err
	}

	if err := m.Fs.Rename(".meta.json", ".meta"); err != nil {
		return err
	}

	return nil
}

// SetStory moves the current global meta file to a backup file and initialises
// a new .meta file for the given story
func (m *Manifest) SetStory(story string) error {
	if err := m.Fs.Rename(".meta", ".meta.json"); err != nil {
		return err
	}

	if err := m.Load(fmt.Sprintf(".meta.%s", story)); err != nil {
		m.Name = story
		m.Projects = nil
		m.Primaries = nil
	}

	bytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	if err := afero.WriteFile(m.Fs, ".meta", bytes, os.FileMode(0666)); err != nil {
		return err
	}

	if err := m.Fs.Remove(fmt.Sprintf(".meta.%s", story)); err != nil {
		//
	}

	return nil
}
