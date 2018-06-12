package git

import (
	"github.com/spf13/afero"
	"fmt"
)

func getHead(fs afero.Fs, project, branch string) (string, error) {
	b, err := afero.ReadFile(fs, fmt.Sprintf("%s/.git/refs/heads/%s", project, branch))
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func HeadsAreEqual(fs afero.Fs, project, b1, b2 string) (bool, error) {
	h1, err := getHead(fs, project, b1)
	if err != nil {
		return false, err
	}

	h2, err := getHead(fs, project, b2)
	if err != nil {
		return false, err
	}

	return h1 == h2, nil
}
