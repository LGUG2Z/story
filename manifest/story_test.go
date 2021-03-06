package manifest_test

import (
	"os"

	"github.com/LGUG2Z/story/manifest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Story", func() {
	Describe("Trying to load non-existent files", func() {
		It("Should throw an error", func() {
			fs := afero.NewMemMapFs()
			_, err := manifest.LoadStory(fs)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Creating a new story", func() {
		It("Should create a story using the appropriate data from the meta file", func() {
			// Given a meta file
			m := NewMetaBuilder().
				Organisation("test-org").
				Projects("one", "two").
				Artifacts("one").
				Build()

			// When I create a story with that meta file as the base
			s := manifest.NewStory("test-story", m)

			// Then the story should inherit the organisation and the artifacts from the meta file
			Expect(s.Name).To(Equal("test-story"))
			Expect(s.Orgranisation).To(Equal("test-org"))
			Expect(s.Artifacts).To(HaveKeyWithValue("one", false))
			Expect(s.AllProjects).To(HaveKeyWithValue("one", "git@github.com:test-org/one.git"))
			Expect(s.AllProjects).To(HaveKeyWithValue("two", "git@github.com:test-org/two.git"))
		})
	})

	Describe("Writing a story to a file", func() {
		It("Should marshal the object with indentation and write to a .meta file", func() {
			// Given a story object
			var br []string

			s := NewStoryBuilder().
				Name("test-story").
				Organisation("test-org").
				Projects("one").
				Artifacts(true, "one").
				BlastRadius(map[string][]string{"one": br}).
				Build()

			fs := afero.NewMemMapFs()

			// When I write the story to a file
			Expect(s.Write(fs)).To(Succeed())

			// Then then the file exists on the fs
			exists, err := afero.Exists(fs, ".meta")
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())

			// And it is indented with two spaces
			bytes, err := afero.ReadFile(fs, ".meta")
			Expect(err).NotTo(HaveOccurred())
			actual := string(bytes)
			expected := `{
  "story": "test-story",
  "organisation": "test-org",
  "projects": {
    "one": "git@github.com:test-org/one.git"
  },
  "blastRadius": {
    "one": null
  },
  "artifacts": {
    "one": true
  },
  "allProjects": null
}`
			Expect(actual).To(Equal(expected))
		})
	})

	Describe("Loading a story from a file", func() {
		It("Should load a valid story", func() {
			// Given a valid story file on an fs
			fs := afero.NewMemMapFs()
			b := []byte(`{
  "allProjects": {
    "one": "git@github.com:test-org/one.git"
  },
  "blastRadius": {
    "one": null
  },
  "artifacts": {
    "one": true
  },
  "story": "test-story",
  "organisation": "test-org",
  "projects": {
    "one": "git@github.com:test-org/one.git"
  }
}`)
			Expect(afero.WriteFile(fs, ".meta", b, os.FileMode(0666))).To(Succeed())

			// When I try to load that story
			_, err := manifest.LoadStory(fs)

			// Then the file should be unmarshalled into an object without error
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should not load an invalid story", func() {
			// Given a valid story file on an fs
			fs := afero.NewMemMapFs()
			b := []byte(`{
  "blastRadius": {
    "one": null
  },
  "artifacts": {
    "one": true
  },
  "story": "test-story",
  "organisation": "test-org",
  "projects": {
    "one": "git@github.com:test-org/one.git"
  },
}`)
			Expect(afero.WriteFile(fs, ".meta", b, os.FileMode(0666))).To(Succeed())

			// When I try to load that story
			_, err := manifest.LoadStory(fs)

			// Then the file should be unmarshalled into an object without error
			Expect(err).To(HaveOccurred())

		})
	})

	Describe("Adding a project", func() {
		It("Should add the project to the manifest", func() {
			// Given a story
			s := NewStoryBuilder().Name("test-story").Organisation("test-org").Build()

			// When I add a project to that story
			allProjects := make(map[string]string)
			allProjects["test-project"] = "git@github.com:test-org/test-project.git"
			Expect(s.AddToManifest(allProjects, "test-project")).To(Succeed())

			// It should update the story
			Expect(s.Projects).To(HaveKeyWithValue("test-project", "git@github.com:test-org/test-project.git"))
		})

		It("Should not add the project that is not in the metarepo", func() {
			// Given a story
			s := NewStoryBuilder().Name("test-story").Organisation("test-org").Build()

			// When I add a project to that story that is not in the metarepo
			allProjects := make(map[string]string)
			allProjects["real-project"] = "test"
			err := s.AddToManifest(allProjects, "test-project")

			// It should not succeed and an informative error should be returned
			Expect(err).To(HaveOccurred())

			Expect(err.Error()).To(Equal("this project is not in the metarepo"))
		})
	})

	Describe("Removing a project", func() {
		It("Should return early if there are no projects", func() {
			// Given a story with a project
			s := NewStoryBuilder().
				Name("test-story").
				Organisation("test-org").
				Build()

			// When I remove the project from that story
			s.RemoveFromManifest("test-project")

			// It should update the story
			Expect(s.Projects).To(BeNil())
		})

		It("Should remove the project from the manifest", func() {
			// Given a story with a project
			s := NewStoryBuilder().
				Name("test-story").
				Organisation("test-org").
				Projects("test-project").
				Build()

			// When I remove the project from that story
			s.RemoveFromManifest("test-project")

			// It should update the story
			Expect(s.Projects).ToNot(HaveKeyWithValue("test-project", "git@github.com:test-org/test-project.git"))
		})
	})

	Describe("Listing artifacts", func() {
		It("Should provide a space separated string of artifacts", func() {
			// Given a story with a project with artifacts
			s := NewStoryBuilder().
				Name("test-story").
				Artifacts(true, "one", "three").
				Build()

			// When I get the artifacts
			actual := s.GetArtifacts()

			// It should produce a space separated string
			Expect(actual).To(Equal("one three"))
		})
	})

	Describe("Calculating the Blast Radius of a story", func() {
		It("Should add the blast radius of each project to the manifest", func() {
			// Given a story
			s := NewStoryBuilder().
				Name("test-story").
				Projects("one").
				Build()

			// And a Blast Radius calculator
			b := NewBlastRadiusBuilder().
				BlastRadius("three", "four", "five").
				Build()

			// WHen I calculate the blast radius for a story
			Expect(s.CalculateBlastRadiusForProject(afero.NewMemMapFs(), b, "one")).To(Succeed())

			// It should be reflected in the manifest
			Expect(s.BlastRadius).To(HaveKeyWithValue("one", []string{"three", "four", "five"}))
		})

		It("Should map the blast radius of a story to artifacts", func() {
			// Given a story with a story, false artifacts and a blast radius
			b := make(map[string][]string)
			b["four"] = []string{"one", "nine"}
			b["five"] = []string{"two", "ten"}

			s := NewStoryBuilder().
				Name("test-story").
				Projects("four", "five", "nine").
				Artifacts(false, "one", "two", "three", "nine").
				BlastRadius(b).
				Build()

			// When I map the blast radius to artifacts
			s.MapBlastRadiusToArtifacts()

			// Then the expected artifacts should be marked true
			Expect(s.Artifacts).To(HaveKeyWithValue("one", true))
			Expect(s.Artifacts).To(HaveKeyWithValue("two", true))
			Expect(s.Artifacts).To(HaveKeyWithValue("nine", true))

			// But those not within the blast radius should remain false
			Expect(s.Artifacts).To(HaveKeyWithValue("three", false))

			// And no new projects should be added to the artifacts map
			Expect(s.Artifacts).ToNot(HaveKey("ten"))
		})
	})

	Describe("Getting commit hashes for all projects in the story", func() {
		It("Should return a project-hash map", func() {
			// Given a story with n projects
			s := NewStoryBuilder().Name("test-story").Projects("one", "two").Build()

			// And the appropriate git files
			fs := afero.NewMemMapFs()
			Expect(fs.MkdirAll("one/.git/refs/heads", os.FileMode(0700))).To(Succeed())
			Expect(fs.MkdirAll("two/.git/refs/heads", os.FileMode(0700))).To(Succeed())
			Expect(afero.WriteFile(fs, "one/.git/refs/heads/test-story", []byte("hash-one"), os.FileMode(0666))).To(Succeed())
			Expect(afero.WriteFile(fs, "two/.git/refs/heads/test-story", []byte("hash-two"), os.FileMode(0666))).To(Succeed())

			// When I get the commit hashes for all projects
			hashes, err := s.GetCommitHashes(fs)
			Expect(err).NotTo(HaveOccurred())

			// Then I expect all projects with their latest commit messages to be in the project-hash map
			Expect(hashes).To(HaveKeyWithValue("one", "hash-one"))
			Expect(hashes).To(HaveKeyWithValue("two", "hash-two"))
		})
	})
})
