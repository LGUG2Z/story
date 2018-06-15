package manifest

import (
	"encoding/json"

	"os"

	"github.com/spf13/afero"
)

type Meta struct {
	Deployables   map[string]bool   `json:"deployables,omitempty"`
	Orgranisation string            `json:"organisation,omitempty"`
	Projects      map[string]string `json:"projects,omitempty"`
}

func LoadMetaOnTrunk(fs afero.Fs) (*Meta, error) {
	bytes, err := afero.ReadFile(fs, ".meta")
	if err != nil {
		return nil, err
	}

	m := &Meta{}

	if err := json.Unmarshal(bytes, &m); err != nil {
		return nil, err
	}

	return m, nil
}

func LoadMetaOnBranch(fs afero.Fs) (*Meta, error) {
	bytes, err := afero.ReadFile(fs, ".meta.json")
	if err != nil {
		return nil, err
	}

	m := &Meta{}

	if err := json.Unmarshal(bytes, &m); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Meta) MoveForStory(fs afero.Fs) error {
	bytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	if err := fs.Remove(".meta"); err != nil {
		return err
	}

	return afero.WriteFile(fs, ".meta.json", bytes, os.FileMode(0666))
}


