package meta

import (
	"fmt"
	"strings"

	"github.com/LGUG2Z/story/node"
	"github.com/spf13/afero"
)

type Story struct {
	BlastRadius   map[string]*[]string         `json:"blast-radius,omitempty"`
	Deployables   map[string]bool              `json:"deployables,omitempty"`
	Fs            afero.Fs                     `json:"-"`
	Global        *Manifest                    `json:"-"`
	Name          string                       `json:"story,omitempty"`
	Orgranisation string                       `json:"organisation"`
	PackageJSONs  map[string]*node.PackageJSON `json:"-"`
	Projects      map[string]string            `json:"projects,omitempty"`
}

func (s *Story) Set(name string) {
	s.Name = name
}

func (s *Story) MapBlastRadiusToDeployables() {
	for _, projects := range s.BlastRadius {
		for _, project := range *projects {
			if _, exists := s.Deployables[project]; exists {
				s.Deployables[project] = true
			}

		}
	}
}

func (s *Story) CalculateBlastRadiusForProject(radius IBlastRadius, project string) error {
	if s.BlastRadius == nil {
		s.BlastRadius = make(map[string]*[]string)
	}

	br, err := radius.Calculate(s.Fs, ".", project)
	if err != nil {
		return err
	}

	s.BlastRadius[project] = &br

	return nil
}

func (s *Story) GetDeployables() string {
	var artifacts []string

	for project, _ := range s.Deployables {
		if s.Deployables[project] {
			artifacts = append(artifacts, project)
		}

	}

	return strings.Join(artifacts, " ")
}

func (s *Story) AddToManifest(project string) {
	if s.Projects == nil {
		s.Projects = make(map[string]string)
	}

	s.Projects[project] = fmt.Sprintf("git+ssh://git@github.com:%s/%s.git", s.Orgranisation, project)
}
func (s *Story) RemoveFromManifest(project string) {
	if s.Projects == nil {
		return
	}

	delete(s.Projects, project)
}
