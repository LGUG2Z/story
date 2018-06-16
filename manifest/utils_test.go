package manifest_test

import (
	"fmt"

	"github.com/LGUG2Z/story/manifest"
	"github.com/LGUG2Z/story/node"
	"github.com/spf13/afero"
)

type MockRadius struct {
	Radius []string
}

type BlastRadiusBuilder struct {
	blastRadius *MockRadius
}

func NewBlastRadiusBuilder() *BlastRadiusBuilder {
	blastRadius := &MockRadius{}
	b := &BlastRadiusBuilder{blastRadius: blastRadius}
	return b
}

func (b *BlastRadiusBuilder) BlastRadius(projects ...string) *BlastRadiusBuilder {
	for _, project := range projects {
		b.blastRadius.Radius = append(b.blastRadius.Radius, project)
	}

	return b
}

func (b *BlastRadiusBuilder) Build() *MockRadius {
	return b.blastRadius
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

func (b *StoryBuilder) Artifacts(status bool, artifacts ...string) *StoryBuilder {
	b.story.Artifacts = make(map[string]bool)

	for _, artifact := range artifacts {
		b.story.Artifacts[artifact] = status
	}

	return b
}

func (b *StoryBuilder) Projects(projects ...string) *StoryBuilder {
	b.story.Projects = make(map[string]string)

	for _, project := range projects {
		b.story.Projects[project] = fmt.Sprintf("git@github.com:%s/%s.git", b.story.Orgranisation, project)
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

type MetaBuilder struct {
	meta *manifest.Meta
}

func NewMetaBuilder() *MetaBuilder {
	meta := &manifest.Meta{}
	b := &MetaBuilder{meta: meta}
	return b
}

func (b *MetaBuilder) Artifacts(artifacts ...string) *MetaBuilder {
	b.meta.Artifacts = make(map[string]bool)

	for _, artifact := range artifacts {
		b.meta.Artifacts[artifact] = false
	}

	return b
}

func (b *MetaBuilder) Orgranisation(orgranisation string) *MetaBuilder {
	b.meta.Orgranisation = orgranisation
	return b
}

func (b *MetaBuilder) Projects(projects ...string) *MetaBuilder {
	b.meta.Projects = make(map[string]string)

	for _, project := range projects {
		b.meta.Projects[project] = fmt.Sprintf("git@github.com:%s/%s.git", b.meta.Orgranisation, project)
	}

	return b
}

func (b *MetaBuilder) Build() *manifest.Meta {
	return b.meta
}
