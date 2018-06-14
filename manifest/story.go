package manifest

import (
	"fmt"
	"strings"

	"encoding/json"
	"os"

	"github.com/LGUG2Z/blastradius/blastradius"
	"github.com/LGUG2Z/story/node"
	"github.com/spf13/afero"
)

type Story struct {
	BlastRadius   map[string][]string          `json:"blast-radius,omitempty"`
	Deployables   map[string]bool              `json:"deployables,omitempty"`
	Name          string                       `json:"story,omitempty"`
	Orgranisation string                       `json:"organisation"`
	PackageJSONs  map[string]*node.PackageJSON `json:"-"`
	Projects      map[string]string            `json:"projects,omitempty"`
}

func NewStory(name string, meta *Meta) *Story {
	return &Story{
		Name:          name,
		Deployables:   meta.Deployables,
		Orgranisation: meta.Orgranisation,
	}
}

func LoadStory(fs afero.Fs) (*Story, error) {
	bytes, err := afero.ReadFile(fs, ".meta")
	if err != nil {
		return nil, err
	}

	s := &Story{}
	if err := json.Unmarshal(bytes, &s); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Story) Write(fs afero.Fs) error {
	bytes, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return afero.WriteFile(fs, ".meta", bytes, os.FileMode(0666))
}

func (s *Story) Set(name string) {
	s.Name = name
}

func (s *Story) MapBlastRadiusToDeployables() {
	for project := range s.Deployables {
		s.Deployables[project] = false
	}

	for _, projects := range s.BlastRadius {
		for _, project := range projects {
			if _, exists := s.Deployables[project]; exists {
				s.Deployables[project] = true
			}

		}
	}
}

func (s *Story) CalculateBlastRadiusForProject(fs afero.Fs, blaster blastradius.RadiusCalculator, project string) error {
	if s.BlastRadius == nil {
		s.BlastRadius = make(map[string][]string)
	}

	br, err := blaster.Calculate(fs, ".", project)
	if err != nil {
		return err
	}

	s.BlastRadius[project] = br

	return nil
}

func (s *Story) GetDeployables() string {
	var artifacts []string

	for project := range s.Deployables {
		if s.Deployables[project] {
			artifacts = append(artifacts, project)
		}
	}

	return strings.Join(artifacts, " ")
}

func (s *Story) AddToManifest(allProjects map[string]string, project string) error {
	if _, exists := allProjects[project]; !exists {
		return fmt.Errorf("this project is not in the metarepo")
	}

	if s.Projects == nil {
		s.Projects = make(map[string]string)
	}

	s.Projects[project] = fmt.Sprintf("git@github.com:%s/%s.git", s.Orgranisation, project)

	return nil
}

func (s *Story) RemoveFromManifest(project string) {
	if s.Projects == nil {
		return
	}

	delete(s.Projects, project)
}
