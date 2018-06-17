package cli_test

import (
	"testing"

	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var fs afero.Fs

func TestCli(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cli Suite")
}

var _ = Describe("Setup", func() {
	BeforeSuite(func() {
		fs = afero.NewOsFs()
	})
})

func initialiseMetarepo(directory string) error {
	command := exec.Command("git", "init")
	command.Dir = directory
	out, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", out)
	}

	m := manifest.Meta{
		Orgranisation: "test-org",
		Artifacts:     map[string]bool{"one": false},
		Projects:      map[string]string{"one": "git@github.com:test-org/one.git"},
	}

	b, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		return err
	}

	if err := afero.WriteFile(fs, fmt.Sprintf("%s/.meta", directory), b, os.FileMode(0666)); err != nil {
		return err
	}

	_, err = git.Add(git.AddOpts{Files: []string{".meta"}})
	if err != nil {
		return err
	}

	command = exec.Command("git", "commit", "-m", "Initialise metarepo")
	command.Dir = directory
	_, err = command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", out)
	}

	return nil
}
