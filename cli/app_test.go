package cli_test

import (
	"os"

	"encoding/json"
	"os/exec"

	"github.com/LGUG2Z/story/cli"
	"github.com/LGUG2Z/story/git"
	"github.com/LGUG2Z/story/manifest"
	"github.com/LGUG2Z/story/node"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
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

	Describe("Reset", func() {
		It("Should reset if currently working on a story", func() {
			// Given an initialised metarepo with a story
			Expect(cli.App().Run([]string{"story", "create", "test-story"})).To(Succeed())

			// When I reset the story then it resets without error
			Expect(cli.App().Run([]string{"story", "reset"})).To(Succeed())
		})

		It("Should return an error if extra arguments are given", func() {
			// Given an initialised metarepo with a story
			Expect(cli.App().Run([]string{"story", "create", "test-story"})).To(Succeed())

			// When I reset the story
			err := cli.App().Run([]string{"story", "reset", "test-story"})

			// Then it returns an error
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(cli.ErrCommandTakesNoArguments))
		})

		It("Should return an error if not working on a story", func() {
			// Given an initialised metarepo not on a story

			// When I reset the story
			err := cli.App().Run([]string{"story", "reset"})

			// Then it returns an error
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(cli.ErrNotWorkingOnAStory))
		})
	})

	Describe("List", func() {
		It("Should print a list of projects in the story", func() {
			// Given an initialised metarepo with a story and a project added
			Expect(cli.App().Run([]string{"story", "create", "test-story"})).To(Succeed())
			s, err := manifest.LoadStory(fs)
			Expect(err).NotTo(HaveOccurred())
			s.Projects = map[string]string{"one": "git@github.com:test-org/one.git"}
			Expect(s.Write(fs)).To(Succeed())

			// When I run the list command, Then it should succeed
			Expect(cli.App().Run([]string{"story", "list"})).To(Succeed())

			// TODO: Maybe check the stdout and assert on that too
		})

		It("Should return an error if extra arguments are given", func() {
			// Given an initialised metarepo with a story and a project added
			Expect(cli.App().Run([]string{"story", "create", "test-story"})).To(Succeed())

			// When I try to list projects,
			err := cli.App().Run([]string{"story", "list", "test-story"})

			// Then it returns an error
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(cli.ErrCommandTakesNoArguments))
		})

		It("Should return an error if not working on a story", func() {
			// Given an initialised metarepo not on a story

			// When I try to list projects,
			err := cli.App().Run([]string{"story", "list"})

			// Then it returns an error
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(cli.ErrNotWorkingOnAStory))
		})
	})

	Describe("Artifacts", func() {
		It("Should print a list of the story artifacts", func() {
			// Given an initialised metarepo with a story and a project added
			Expect(cli.App().Run([]string{"story", "create", "test-story"})).To(Succeed())
			s, err := manifest.LoadStory(fs)
			Expect(err).NotTo(HaveOccurred())
			s.Artifacts = map[string]bool{"one": true}
			Expect(s.Write(fs)).To(Succeed())

			// When I run the list command, Then it succeeds
			Expect(cli.App().Run([]string{"story", "artifacts"})).To(Succeed())

			// TODO: Maybe check the stdout and assert on that too
		})

		It("Should return an error if extra arguments are given", func() {
			// Given an initialised metarepo with a story and a project added
			Expect(cli.App().Run([]string{"story", "create", "test-story"})).To(Succeed())

			// When I try to list projects,
			err := cli.App().Run([]string{"story", "artifacts", "test-story"})

			// Then it returns an error
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(cli.ErrCommandTakesNoArguments))
		})

		It("Should return an error if not working on a story", func() {
			// Given an initialised metarepo not on a story

			// When I try to list projects,
			err := cli.App().Run([]string{"story", "artifacts"})

			// Then it returns an error
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(cli.ErrNotWorkingOnAStory))
		})
	})

	Describe("Add", func() {
		It("Should add a project to a story", func() {
			// Given an initialised metarepo with projects and a story
			Expect(fs.MkdirAll("one", os.FileMode(0700))).To(Succeed())
			p := node.PackageJSON{}
			b, err := json.Marshal(p)
			Expect(err).NotTo(HaveOccurred())
			Expect(afero.WriteFile(fs, "one/package.json", b, os.FileMode(0666))).To(Succeed())

			command := exec.Command("git", "init")
			command.Dir = "one"
			_, err = command.CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			_, err = git.Add(git.AddOpts{Project: "one", Files: []string{"package.json"}})
			Expect(err).NotTo(HaveOccurred())

			_, err = git.Commit(git.CommitOpts{Project: "one", Messages: []string{"initial commit"}})
			Expect(err).NotTo(HaveOccurred())

			Expect(cli.App().Run([]string{"story", "create", "test-story"})).To(Succeed())

			// When I add a project to the story
			Expect(cli.App().Run([]string{"story", "add", "one"})).To(Succeed())

			// Then the project is in the story meta file
			s, err := manifest.LoadStory(fs)
			Expect(err).NotTo(HaveOccurred())
			Expect(s.Projects).To(HaveKeyWithValue("one", "git@github.com:test-org/one.git"))

			// And the project is on the story branch
			branch, err := git.GetCurrentBranch(fs, "one")
			Expect(err).NotTo(HaveOccurred())
			Expect(branch).To(Equal("test-story"))

			// And artifacts are updated
			Expect(s.Artifacts).To(HaveKeyWithValue("one", true))
		})

		It("Should return an error if no arguments are given", func() {
			// Given an initialised metarepo with a story
			Expect(cli.App().Run([]string{"story", "create", "test-story"})).To(Succeed())

			// When I try to add a project
			err := cli.App().Run([]string{"story", "add"})

			// Then it returns an error
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(cli.ErrCommandRequiresAnArgument))
		})

		It("Should return an error if not working on a story", func() {
			// Given an initialised metarepo not on a story

			// When I try to list projects,
			err := cli.App().Run([]string{"story", "add", "one"})

			// Then it returns an error
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(cli.ErrNotWorkingOnAStory))
		})

	})
})
