package git_test

import (
	"os"
	"os/exec"
	"strings"

	"github.com/LGUG2Z/story/git"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Add", func() {
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

	Describe("Adding files to be committed", func() {
		It("Should add the files to the staging area for the next commit", func() {
			// Given a repository with an untracked file in a subfolder
			Expect(fs.MkdirAll("sub", os.FileMode(0700))).To(Succeed())
			Expect(afero.WriteFile(fs, "sub/newfile", []byte{}, os.FileMode(0666))).To(Succeed())

			// When I add it to the staging area
			_, err := git.Add(git.AddOpts{Project: "sub", Files: []string{"newfile"}})
			Expect(err).NotTo(HaveOccurred())

			// Then the files appear in the list of files staged for commit
			command := exec.Command("git", "diff", "--cached", "--name-only")

			combinedOutput, err := command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			actual := strings.TrimSpace(string(combinedOutput))
			Expect(actual).To(Equal("sub/newfile"))
		})
	})
})
