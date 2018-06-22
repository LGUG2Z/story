package manifest

import (
	"encoding/json"

	"os"

	"github.com/spf13/afero"
)

type Meta struct {
	Artifacts    map[string]bool   `json:"artifacts,omitempty"`
	Organisation string            `json:"organisation,omitempty"`
	Projects     map[string]string `json:"projects,omitempty"`
}

// TODO: Add Test
func (m *Meta) Write(fs afero.Fs) error {
	bytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}

	return afero.WriteFile(fs, ".meta", bytes, os.FileMode(0666))
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
