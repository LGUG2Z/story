package git_test

import (
	"testing"

	"fmt"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var fs afero.Fs

func TestGit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Git Suite")
}

var _ = Describe("Setup", func() {
	BeforeSuite(func() {
		fs = afero.NewOsFs()
	})
})

func initialiseRepository(directory string) error {
	command := exec.Command("git", "init")
	command.Dir = directory
	out, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", out)
	}

	if err := afero.WriteFile(fs, fmt.Sprintf("%s/blank", directory), []byte{}, os.FileMode(0666)); err != nil {
		return err
	}

	command = exec.Command("git", "add", ".")
	command.Dir = directory
	out, err = command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", out)
	}

	command = exec.Command("git", "commit", "-m", "initial")
	command.Dir = directory
	_, err = command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", out)
	}

	return nil
}
