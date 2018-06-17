package cli_test

import (
	"os"

	"github.com/LGUG2Z/story/cli"
	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("App", func() {
	BeforeEach(func() {
		if err := fs.MkdirAll("test", os.FileMode(0700)); err != nil {
			Fail(err.Error())
		}

		if err := os.Chdir("test"); err != nil {
			Fail(err.Error())
		}

		if err := initialiseMetarepo("."); err != nil {
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

	Describe("Create", func() {
		It("Should create a new story if on trunk", func() {
			// Given an initialised metarepo

			// When I run the create command
			Expect(cli.App().Run([]string{"story", "create", "test-story"})).To(Succeed())

			// I should be on a new branch
			branch, err := git.GetCurrentBranch(fs, ".")
			Expect(err).NotTo(HaveOccurred())
			Expect(branch).To(Equal("test-story"))

			// And I should have a story .meta file
			_, err = manifest.LoadStory(fs)
			Expect(err).NotTo(HaveOccurred())

			// And I should have a global .meta.json file
			_, err = manifest.LoadMetaOnTrunk(fs)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should return an error if on trunk and a story name isn't supplied", func() {
			// Given an initialised metarepo

			// When I run the create command
			err := cli.App().Run([]string{"story", "create"})

			// Then an error is returned
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(cli.ErrCommandRequiresAnArgument))
		})

		It("Should return an error if already working on a story", func() {
			// Given an initialised metarepo with a story
			Expect(cli.App().Run([]string{"story", "create", "test-story"})).To(Succeed())

			// When I try to create another story
			err := cli.App().Run([]string{"story", "create", "test-story"})
			Expect(err).To(HaveOccurred())

			// Then an error is returned
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(cli.ErrAlreadyWorkingOnAStory))
		})
	})

	Describe("Load", func() {
		It("Should load the story branches", func() {
			// Given an initialised metarepo with a story which is then reset
			Expect(cli.App().Run([]string{"story", "create", "test-story"})).To(Succeed())
			_, err := git.Add(git.AddOpts{Files: []string{".meta", ".meta.json"}})
			Expect(err).NotTo(HaveOccurred())

			_, err = git.Commit(git.CommitOpts{Messages: []string{"story start"}})
			Expect(err).NotTo(HaveOccurred())

			_, err = git.CheckoutBranch(git.CheckoutBranchOpts{Branch: "master"})
			Expect(err).NotTo(HaveOccurred())

			// When I load the story
			err = cli.App().Run([]string{"story", "load", "test-story"})

			// Then it loads without error
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should return an error if on trunk and a story name isn't supplied", func() {
			// Given an initialised metarepo

			// When I run the load command
			err := cli.App().Run([]string{"story", "load"})

			// Then an error is returned
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(cli.ErrCommandRequiresAnArgument))
		})

		It("Should return an error if already working on a story", func() {
			// Given an initialised metarepo with a story
			Expect(cli.App().Run([]string{"story", "create", "test-story"})).To(Succeed())

			// When I try to load another story
			err := cli.App().Run([]string{"story", "load", "test-story"})
			Expect(err).To(HaveOccurred())

			// Then an error is returned
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(cli.ErrAlreadyWorkingOnAStory))
		})
	})
})
