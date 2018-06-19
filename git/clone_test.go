package git_test

import (
	"os"
	"os/exec"

	"github.com/LGUG2Z/story/git"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Clone", func() {
	BeforeEach(func() {
		if err := fs.MkdirAll("test", os.FileMode(0700)); err != nil {
			Fail(err.Error())
		}

		if err := os.Chdir("test"); err != nil {
			Fail(err.Error())
		}

		if err := initialiseRepository("."); err != nil {
			Fail(err.Error())
		}
	})

	AfterEach(func() {
		if err := os.Chdir(".."); err != nil {
			Fail(err.Error())
		}

		if err := fs.RemoveAll("test"); err != nil {
			Fail(err.Error())
		}
	})

	Describe("Cloning repositories", func() {
		It("Should clone a repository into a given folder", func() {
			// Given a remote / bare repository
			Expect(fs.MkdirAll("remote", os.FileMode(0700))).To(Succeed())
			command := exec.Command("git", "init", "--bare")
			command.Dir = "remote"
			_, err := command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			// When I clone that remote repository
			_, err = git.Clone(git.CloneOpts{Repository: "remote", Directory: "cloned"})
			Expect(err).NotTo(HaveOccurred())

			// Then the cloned directory should exist on the fs
			clonedExists, err := afero.DirExists(fs, "cloned")
			Expect(err).NotTo(HaveOccurred())
			clonedDotGitExists, err := afero.DirExists(fs, "cloned/.git")
			Expect(err).NotTo(HaveOccurred())

			Expect(clonedExists).To(BeTrue())
			Expect(clonedDotGitExists).To(BeTrue())
		})
	})
})
