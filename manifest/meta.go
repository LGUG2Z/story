package manifest

import (
	"encoding/json"

	"github.com/spf13/afero"
)

type Meta struct {
	Artifacts     map[string]bool   `json:"artifacts,omitempty"`
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
