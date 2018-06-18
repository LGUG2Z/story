package manifest

import (
	"fmt"
	"strings"

	"encoding/json"
	"os"

	"sort"

	"github.com/LGUG2Z/blastradius/blastradius"
	"github.com/spf13/afero"
)

type Story struct {
	BlastRadius   map[string][]string `json:"blast-radius,omitempty"`
	Artifacts     map[string]bool     `json:"artifacts,omitempty"`
	Name          string              `json:"story,omitempty"`
	Orgranisation string              `json:"organisation"`
	AllProjects   map[string]string   `json:"all-projects"`
	Projects      map[string]string   `json:"projects,omitempty"`
	Hashes        map[string]string   `json:"hashes,omitempty"`
}

func NewStory(name string, meta *Meta) *Story {
	return &Story{
		Name:          name,
		Artifacts:     meta.Artifacts,
		Orgranisation: meta.Orgranisation,
		AllProjects:   meta.Projects,
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

func (s *Story) MapBlastRadiusToArtifacts() {
	for project := range s.Artifacts {
		s.Artifacts[project] = false
	}

	for _, projects := range s.BlastRadius {
		for _, project := range projects {
			if _, exists := s.Artifacts[project]; exists {
				s.Artifacts[project] = true
			}

		}
	}

	// Ensure relevant projects are marked as artifacts,
	// even if not in the blast radius of others
	for project := range s.Projects {
		if _, exists := s.Artifacts[project]; exists {
			s.Artifacts[project] = true
		}
	}
}

func (s *Story) GetCommitHashes(fs afero.Fs) (map[string]string, error) {
	hashMap := make(map[string]string)
	for project := range s.Projects {
		bytes, err := afero.ReadFile(fs, fmt.Sprintf("%s/.git/refs/heads/%s", project, s.Name))
		if err != nil {
			return nil, err
		}

		hashMap[project] = strings.TrimSpace(string(bytes))
	}

	return hashMap, nil
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

func (s *Story) GetArtifacts() string {
	var artifacts []string

	for project := range s.Artifacts {
		if s.Artifacts[project] {
			artifacts = append(artifacts, project)
		}
	}

	sort.Strings(artifacts)

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
