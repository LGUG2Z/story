package git_test

import (
	"os"

	"github.com/LGUG2Z/story/git"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Commit", func() {
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

	Describe("Committing with staged files", func() {
		It("Should commit the staged files", func() {
			// Given a repository with files staged for commit
			Expect(fs.MkdirAll("sub", os.FileMode(0700))).To(Succeed())
			Expect(afero.WriteFile(fs, "sub/newfile", []byte{}, os.FileMode(0666))).To(Succeed())
			_, err := git.Add(git.AddOpts{Project: "sub", Files: []string{"newfile"}})
			Expect(err).NotTo(HaveOccurred())

			// When I commit those files
			output, err := git.Commit(git.CommitOpts{Project: "sub", Messages: []string{"Some msg"}})
			Expect(err).NotTo(HaveOccurred())

			// Then I see output confirming the commit
			Expect(output).To(ContainSubstring("1 file changed"))
		})
	})

	Describe("Committing with no staged files", func() {
		It("Should return a message informing that there is nothing to commit", func() {
			// Given a repository without files staged for commit

			// When I try to commit
			output, err := git.Commit(git.CommitOpts{Messages: []string{}})
			Expect(err).NotTo(HaveOccurred())

			// Then I see output confirming the commit
			Expect(output).To(Equal("no staged changes to commit"))
		})
	})
})
