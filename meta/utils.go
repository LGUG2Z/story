package meta

import (
	"encoding/json"
	"fmt"

	"github.com/LGUG2Z/story/node"
	"github.com/spf13/afero"
)

func getPackageJSON(fs afero.Fs, project string) (*node.PackageJSON, error) {
	packageJSON := &node.PackageJSON{}
	bytes, err := afero.ReadFile(fs, fmt.Sprintf("%s/package.json", project))
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(bytes, packageJSON); err != nil {
		return nil, err
	}

	return packageJSON, nil
}
