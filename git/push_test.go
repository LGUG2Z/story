package git_test

import (
	"os"
	"os/exec"

	"github.com/LGUG2Z/story/git"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Push", func() {
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
	Describe("Pushing commits", func() {
		It("Should push unpushed commits in a project", func() {
			// Given a repository, with a remote "origin"
			Expect(fs.MkdirAll("remote", os.FileMode(0700))).To(Succeed())
			command := exec.Command("git", "init", "--bare")
			command.Dir = "remote"
			_, err := command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			command = exec.Command("git", "remote", "add", "origin", "./remote")
			_, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			command = exec.Command("git", "push", "--set-upstream", "origin", "master")
			_, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			// And a new branch which is pushed
			expectedBranch := "test-branch"
			_, err = git.CheckoutBranch(git.CheckoutBranchOpts{Create: true, Branch: expectedBranch})
			Expect(err).NotTo(HaveOccurred())

			command = exec.Command("git", "push", "--set-upstream", "origin", "test-branch")
			_, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			// And a change committed but not yet pushed
			Expect(afero.WriteFile(fs, "bla", []byte{}, os.FileMode(0666))).To(Succeed())
			_, err = git.Add(git.AddOpts{Files: []string{"bla"}})
			Expect(err).NotTo(HaveOccurred())
			_, err = git.Commit(git.CommitOpts{Messages: []string{"commit bla"}})
			Expect(err).NotTo(HaveOccurred())

			// When I push
			combinedOutput, err := git.Push(git.PushOpts{Branch: "test-branch", Remote: "origin"})

			// The push should be completed successfully
			Expect(err).NotTo(HaveOccurred())
			Expect(combinedOutput).To(ContainSubstring("test-branch -> test-branch"))
		})

		It("Should not take action if there are no unpushed commits", func() {
			// Given a repository, with a remote "origin"
			Expect(fs.MkdirAll("remote", os.FileMode(0700))).To(Succeed())
			command := exec.Command("git", "init", "--bare")
			command.Dir = "remote"
			_, err := command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			command = exec.Command("git", "remote", "add", "origin", "./remote")
			_, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			command = exec.Command("git", "push", "--set-upstream", "origin", "master")
			_, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			// And a new branch which is pushed
			expectedBranch := "test-branch"
			_, err = git.CheckoutBranch(git.CheckoutBranchOpts{Create: true, Branch: expectedBranch})
			Expect(err).NotTo(HaveOccurred())

			command = exec.Command("git", "push", "--set-upstream", "origin", "test-branch")
			_, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			// And a change committed and pushed
			Expect(afero.WriteFile(fs, "bla", []byte{}, os.FileMode(0666))).To(Succeed())
			_, err = git.Add(git.AddOpts{Files: []string{"bla"}})
			Expect(err).NotTo(HaveOccurred())
			_, err = git.Commit(git.CommitOpts{Messages: []string{"commit bla"}})
			Expect(err).NotTo(HaveOccurred())
			_, err = git.Push(git.PushOpts{Branch: "test-branch", Remote: "origin"})
			Expect(err).NotTo(HaveOccurred())

			// When I push
			combinedOutput, err := git.Push(git.PushOpts{Branch: "test-branch", Remote: "origin"})

			// Then I receive a message telling me there are no unpushed commits
			Expect(err).NotTo(HaveOccurred())
			Expect(combinedOutput).To(ContainSubstring("no unpushed commits"))

		})
	})
})
