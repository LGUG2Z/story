package manifest_test

import (
	"os"

	"github.com/LGUG2Z/story/manifest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var _ = Describe("Meta", func() {
	Describe("Creating a new story", func() {
		It("Should create a story using the appropriate data from the meta file", func() {
			// Given a meta file
			m := NewMetaBuilder().
				Orgranisation("test-org").
				Deployables("one").
				Build()

			// When I create a story with that meta file as the base
			s := manifest.NewStory("test-story", m)

			// Then the story should inherit the organisation and the deployables from the meta file
			Expect(s.Name).To(Equal("test-story"))
			Expect(s.Orgranisation).To(Equal("test-org"))
			Expect(s.Deployables).To(HaveKeyWithValue("one", false))
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
				Deployables(true, "one").
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
  "blast-radius": {
    "one": null
  },
  "deployables": {
    "one": true
  },
  "story": "test-story",
  "organisation": "test-org",
  "projects": {
    "one": "git@github.com:test-org/one.git"
  }
}`
			Expect(actual).To(Equal(expected))
		})
	})

	Describe("Loading a story from a file", func() {
		It("Should load a valid story", func() {
			// Given a valid story file on an fs
			fs := afero.NewMemMapFs()
			b := []byte(`{
  "blast-radius": {
    "one": null
  },
  "deployables": {
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
  "blast-radius": {
    "one": null
  },
  "deployables": {
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
			allProjects["test-project"] = "test"
			s.AddToManifest(allProjects, "test-project")

			// It should update the story
			Expect(s.Projects).To(HaveKeyWithValue("test-project", "git@github.com:test-org/test-project.git"))
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

	Describe("Listing deployable artifacts", func() {
		It("Should provide a space separated string of artifacts", func() {
			// Given a story with a project with deployables
			s := NewStoryBuilder().
				Name("test-story").
				Deployables(true, "one", "three").
				Build()

			// When I get the deployables
			actual := s.GetDeployables()

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

		It("Should map the blast radius of a story to deployable artifacts", func() {
			// Given a story with a story, false deployables and a blast radius
			b := make(map[string][]string)
			b["four"] = []string{"one", "nine"}
			b["five"] = []string{"two", "ten"}

			s := NewStoryBuilder().
				Name("test-story").
				Projects("four", "five").
				Deployables(false, "one", "two", "three").
				BlastRadius(b).
				Build()

			// When I map the blast radius to deployable artifacts
			s.MapBlastRadiusToDeployables()

			// Then the expected deployable artifacts should be marked true
			Expect(s.Deployables).To(HaveKeyWithValue("one", true))
			Expect(s.Deployables).To(HaveKeyWithValue("two", true))

			// But those not within the blast radius should remain false
			Expect(s.Deployables).To(HaveKeyWithValue("three", false))

			// And no new projects should be added to the deployable set
			Expect(s.Deployables).ToNot(HaveKey("nine"))
			Expect(s.Deployables).ToNot(HaveKey("ten"))
		})
	})
})
