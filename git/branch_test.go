package git_test

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/LGUG2Z/story/git"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

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

var fs afero.Fs

var _ = Describe("Branch", func() {
	BeforeSuite(func() {
		fs = afero.NewOsFs()
	})

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

	Describe("Checking out branches", func() {
		It("Should create a new branch", func() {
			// Given a repository

			// When I check out a new branch
			expectedBranch := "test-branch"
			_, err := git.CheckoutBranch(git.CheckoutBranchOpts{Create: true, Branch: expectedBranch})
			Expect(err).NotTo(HaveOccurred())

			// Then that branch should be created and set as the current branch
			actualBranch, err := git.GetCurrentBranch(fs, ".")
			Expect(err).NotTo(HaveOccurred())
			Expect(actualBranch).To(Equal(expectedBranch))
		})

		It("Should fail to create a new branch if a branch of the same name already exists", func() {
			// Given a repository, with a new branch, reverted back to master
			expectedBranch := "test-branch"
			_, err := git.CheckoutBranch(git.CheckoutBranchOpts{Create: true, Branch: expectedBranch})
			Expect(err).NotTo(HaveOccurred())
			_, err = git.CheckoutBranch(git.CheckoutBranchOpts{Branch: "master"})
			Expect(err).NotTo(HaveOccurred())

			// When I create a branch of the same name
			_, err = git.CheckoutBranch(git.CheckoutBranchOpts{Create: true, Branch: expectedBranch})

			// Then I should receive an error
			Expect(err).To(HaveOccurred())
		})

		It("Should checkout an existing branch", func() {
			// Given a repository, with a new branch, reverted back to master
			expectedBranch := "test-branch"
			_, err := git.CheckoutBranch(git.CheckoutBranchOpts{Create: true, Branch: expectedBranch})
			Expect(err).NotTo(HaveOccurred())
			_, err = git.CheckoutBranch(git.CheckoutBranchOpts{Branch: "master"})
			Expect(err).NotTo(HaveOccurred())

			// When I checkout the branch
			_, err = git.CheckoutBranch(git.CheckoutBranchOpts{Branch: expectedBranch})
			Expect(err).NotTo(HaveOccurred())

			// Then that branch should be created and set as the current branch
			actualBranch, err := git.GetCurrentBranch(fs, ".")
			Expect(err).NotTo(HaveOccurred())
			Expect(actualBranch).To(Equal(expectedBranch))
		})

		It("Should fail to checkout a branch that doesn't exist", func() {
			// Given a repository

			// When I check out a branch that hasn't been created
			expectedBranch := "test-branch"
			_, err := git.CheckoutBranch(git.CheckoutBranchOpts{Branch: expectedBranch})

			// Then I should receive an error
			Expect(err).To(HaveOccurred())

		})

		It("Should checkout a branch and create it if it doesn't exist", func() {
			// Given a repository

			// When I check out a branch that hasn't been created
			expectedBranch := "test-branch"
			_, err := git.CheckoutBranchWithCreateIfRequired(expectedBranch)
			Expect(err).NotTo(HaveOccurred())

			// Then that branch should be created and set as the current branch
			actualBranch, err := git.GetCurrentBranch(fs, ".")
			Expect(err).NotTo(HaveOccurred())
			Expect(actualBranch).To(Equal(expectedBranch))
		})

		It("Should checkout a branch if it exists", func() {
			// Given a repository, with a new branch, reverted back to master
			expectedBranch := "test-branch"
			_, err := git.CheckoutBranch(git.CheckoutBranchOpts{Create: true, Branch: expectedBranch})
			Expect(err).NotTo(HaveOccurred())
			_, err = git.CheckoutBranch(git.CheckoutBranchOpts{Branch: "master"})
			Expect(err).NotTo(HaveOccurred())

			// When I check out a branch that hasn't been created
			_, err = git.CheckoutBranchWithCreateIfRequired(expectedBranch)
			Expect(err).NotTo(HaveOccurred())

			// Then that branch should be created and set as the current branch
			actualBranch, err := git.GetCurrentBranch(fs, ".")
			Expect(err).NotTo(HaveOccurred())
			Expect(actualBranch).To(Equal(expectedBranch))
		})
	})

	Describe("Deleting branches", func() {
		It("Should delete a local branch", func() {
			// Given a repository, with a new branch, reverted back to master
			expectedBranch := "test-branch"
			_, err := git.CheckoutBranch(git.CheckoutBranchOpts{Create: true, Branch: expectedBranch})
			Expect(err).NotTo(HaveOccurred())
			_, err = git.CheckoutBranch(git.CheckoutBranchOpts{Branch: "master"})
			Expect(err).NotTo(HaveOccurred())

			// When I delete the branch
			_, err = git.DeleteBranch(git.DeleteBranchOpts{Branch: expectedBranch, Local: true})
			Expect(err).NotTo(HaveOccurred())

			_, err = afero.ReadFile(fs, ".git/refs/heads/test-branch")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("open .git/refs/heads/test-branch: no such file or directory"))
		})

		It("Should delete a remote branch", func() {
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

			// And a new branch and push it
			expectedBranch := "test-branch"
			_, err = git.CheckoutBranch(git.CheckoutBranchOpts{Create: true, Branch: expectedBranch})
			Expect(err).NotTo(HaveOccurred())

			command = exec.Command("git", "push", "--set-upstream", "origin", "test-branch")
			_, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			// When I delete the branch remotely
			_, err = git.DeleteBranch(git.DeleteBranchOpts{Branch: expectedBranch, Remote: true})
			Expect(err).NotTo(HaveOccurred())

			// Then the branch should not exist any more on the remote
			_, err = afero.ReadFile(fs, "remote/refs/heads/test-branch")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("open remote/refs/heads/test-branch: no such file or directory"))
		})
	})

	Describe("Determining branches", func() {
		It("Determine the current branch of a repository", func() {
			// Given a repository, with a new branch checked out
			expectedBranch := "test-branch"
			_, err := git.CheckoutBranch(git.CheckoutBranchOpts{Create: true, Branch: expectedBranch})
			Expect(err).NotTo(HaveOccurred())

			// When I check the branch of the repository
			actualBranch, err := git.GetCurrentBranch(fs, ".")
			Expect(err).NotTo(HaveOccurred())

			// Then it should be the branch checked out
			Expect(actualBranch).To(Equal(expectedBranch))
		})
	})

	Describe("Comparing branch heads", func() {
		It("Should identify branches that have the same heads", func() {
			// Given a repository, with a new branch checked out
			expectedBranch := "test-branch"
			_, err := git.CheckoutBranch(git.CheckoutBranchOpts{Create: true, Branch: expectedBranch})
			Expect(err).NotTo(HaveOccurred())

			// When I check the equality of the heads
			areEqual, err := git.HeadsAreEqual(fs, ".", "test-branch", "master")
			Expect(err).NotTo(HaveOccurred())

			// Then they should be equal
			Expect(areEqual).To(BeTrue())
		})
	})
})
