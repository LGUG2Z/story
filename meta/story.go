package meta

import (
	"encoding/json"
	"fmt"
	"github.com/AlexsJones/kepler/commands/node"
	"github.com/spf13/afero"
	"log"
	"os"
)

// Manifest represents a .meta JSON file, enriched with a "story" key
type Manifest struct {
	Fs       afero.Fs          `json:"-"`
	Name     string            `json:"story,omitempty"`
	Projects map[string]string `json:"projects,omitempty"`
	Primary  map[string]bool   `json:"primary,omitempty"`
}

// IsStory checks if the .meta is a story subset or the global .meta file
func (m *Manifest) IsStory() bool {
	return m.Name != ""
}

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

// RemoveProjects will remove a project from a story Meta file
func (m *Manifest) RemoveProjects(projects []string) error {
	global := &Manifest{Fs: m.Fs}
	if err := global.Load(".meta.json"); err != nil {
		return err
	}

	for _, project := range projects {
		log.Printf("removing %s from story %s", project, m.Name)
		delete(m.Projects, project)
		delete(m.Primary, project)
		removePrivateDependencies(global, m, project)

	}

	bytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	if err := afero.WriteFile(m.Fs, ".meta", bytes, os.FileMode(0666)); err != nil {
		return err
	}

	return nil
}

// AddProjects will remove a project from a story Meta file
func (m *Manifest) AddProjects(projects []string) error {
	global := &Manifest{Fs: m.Fs}
	if err := global.Load(".meta.json"); err != nil {
		return fmt.Errorf("not working on a story")
	}

	for _, project := range projects {
		if _, exists := global.Projects[project]; exists {
			if m.Projects == nil {
				m.Projects = make(map[string]string)
			}

			if m.Primary == nil {
				m.Primary = make(map[string]bool)
			}

			m.Primary[project] = true

			if _, exists := m.Projects[project]; !exists {
				log.Printf("adding %s to story %s", project, m.Name)
				m.Projects[project] = fmt.Sprintf("git@github.com:%s/%s.git", os.Getenv("ORGANISATION"), project)
			}

			err := addPrivateDependencies(global, m, project)
			if err != nil {
				log.Printf("problem reading %s/package.json\n", project)
			}
		} else {
			log.Printf("%s is not a part of this metarepo, skipping\n", project)
		}
	}

	bytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	if err := afero.WriteFile(m.Fs, ".meta", bytes, os.FileMode(0666)); err != nil {
		return err
	}

	return nil
}

func removePrivateDependencies(meta, story *Manifest, project string) error {
	bytes, err := afero.ReadFile(story.Fs, fmt.Sprintf("%s/package.json", project))
	if err != nil {
		return err
	}

	p := node.PackageJSON{}
	if err := json.Unmarshal(bytes, &p); err != nil {
		return err
	}

	for dep := range p.Dependencies {
		if _, exists := meta.Projects[dep]; exists {
			if _, exists := story.Projects[dep]; exists {
				log.Printf("removing %s from story %s (dependency of %s)", dep, story.Name, project)
				delete(story.Projects, dep)
			}
		}
	}

	for project := range story.Primary {
		if err := addPrivateDependencies(meta, story, project); err != nil {
			log.Printf("there was a problem reading %s/package.json\n", project)
		}
	}

	return nil
}

func addPrivateDependencies(meta, story *Manifest, project string) error {
	bytes, err := afero.ReadFile(story.Fs, fmt.Sprintf("%s/package.json", project))
	if err != nil {
		return err
	}

	p := node.PackageJSON{}
	if err := json.Unmarshal(bytes, &p); err != nil {
		return err
	}

	for dep := range p.Dependencies {
		if _, exists := meta.Projects[dep]; exists {
			if _, exists := story.Projects[dep]; !exists {
				log.Printf("adding %s to story %s (dependency of %s)", dep, story.Name, project)
				story.Projects[dep] = fmt.Sprintf("git@github.com:%s/%s.git", os.Getenv("ORGANISATION"), dep)
			}
		}
	}
	return nil
}

func (m *Manifest) RestoreGlobal() error {
	if err := m.Fs.Rename(".meta", fmt.Sprintf(".meta.%s", m.Name)); err != nil {
		return err
	}

	if err := m.Fs.Rename(".meta.json", ".meta"); err != nil {
		return err
	}

	return nil
}

func (m *Manifest) SetStory(story string) error {
	if err := m.Fs.Rename(".meta", ".meta.json"); err != nil {
		return err
	}

	s := Manifest{Fs: m.Fs}
	if err := s.Load(fmt.Sprintf(".meta.%s", story)); err != nil {
		log.Printf("no existing story meta found for story %s, creating...\n", story)
		s.Name = story
	}

	bytes, err := json.MarshalIndent(s, "", "  ")
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
