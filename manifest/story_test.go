package manifest_test

import (
	"fmt"

	"github.com/LGUG2Z/story/manifest"
	"github.com/LGUG2Z/story/node"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

type MockRadius struct {
	Radius []string
}

type BlastRadiusBuilder struct {
	mockRadius *MockRadius
}

func NewBlastRadiusBuilder() *BlastRadiusBuilder {
	blastRadius := &MockRadius{}
	b := &BlastRadiusBuilder{mockRadius: blastRadius}
	return b
}

func (b *BlastRadiusBuilder) BlastRadius(projects ...string) *BlastRadiusBuilder {
	for _, project := range projects {
		b.mockRadius.Radius = append(b.mockRadius.Radius, project)
	}

	return b
}

func (b *BlastRadiusBuilder) Build() *MockRadius {
	return b.mockRadius
}

func (r *MockRadius) Calculate(fs afero.Fs, metarepo, project string) ([]string, error) {
	return r.Radius, nil
}

type StoryBuilder struct {
	story *manifest.Story
}

func NewStoryBuilder() *StoryBuilder {
	story := &manifest.Story{}
	b := &StoryBuilder{story: story}
	return b
}

func (b *StoryBuilder) Name(name string) *StoryBuilder {
	b.story.Name = name
	return b
}

func (b *StoryBuilder) Organisation(organisation string) *StoryBuilder {
	b.story.Orgranisation = organisation
	return b
}

func (b *StoryBuilder) Deployables(status bool, deployables ...string) *StoryBuilder {
	b.story.Deployables = make(map[string]bool)

	for _, deployable := range deployables {
		b.story.Deployables[deployable] = status
	}

	return b
}

func (b *StoryBuilder) Projects(projects ...string) *StoryBuilder {
	b.story.Projects = make(map[string]string)

	for _, project := range projects {
		b.story.Projects[project] = fmt.Sprintf("git+ssh://git@github.com:%s/%s.git", b.story.Orgranisation, project)
	}

	return b
}

func (b *StoryBuilder) PackageJSONs(packageJSONs map[string]*node.PackageJSON) *StoryBuilder {
	b.story.PackageJSONs = packageJSONs
	return b
}

func (b *StoryBuilder) BlastRadius(blastRadius map[string][]string) *StoryBuilder {
	b.story.BlastRadius = blastRadius
	return b
}

func (b *StoryBuilder) Build() *manifest.Story {
	return b.story
}

var _ = Describe("Meta", func() {
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
			Expect(s.Projects).ToNot(HaveKeyWithValue("test-project", "git+ssh://git@github.com:test-org/test-project.git"))
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
