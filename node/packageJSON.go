package node

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/afero"
)

type PackageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Description     string            `json:"description"`
	Main            string            `json:"main"`
	Bugs            map[string]string `json:"bugs,omitempty"`
	Scripts         map[string]string `json:"scripts,omitempty"`
	Dependencies    map[string]string `json:"dependencies,omitempty"`
	DevDependencies map[string]string `json:"devDependencies,omitempty"`
	Private         bool              `json:"private,omitempty"`
	License         string            `json:"license,omitempty"`
}

func (p *PackageJSON) Load(fs afero.Fs, project string) error {
	bytes, err := afero.ReadFile(fs, fmt.Sprintf("%s/package.json", project))
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, p)
}

func (p *PackageJSON) Write(fs afero.Fs, project string) error {
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/package.json", project)
	return afero.WriteFile(fs, filename, b, os.FileMode(0666))
}

func (p *PackageJSON) setPrivateDependencyBranchToStory(dependency, story string) {
	if strings.HasSuffix(p.Dependencies[dependency], ".git") {
		// Append #story-branch-name to the current git+ssh string
		p.Dependencies[dependency] = fmt.Sprintf("%s#%s", p.Dependencies[dependency], story)
	}
}

func (p *PackageJSON) ResetPrivateDependencyBranchesToMaster(story string) {
	storyBranch := fmt.Sprintf("#%s", story)
	for pkg, src := range p.Dependencies {
		if strings.HasSuffix(src, storyBranch) {
			p.Dependencies[pkg] = strings.TrimSuffix(src, storyBranch)
		}
	}
}

func (p *PackageJSON) resetPrivateDependencyBranch(dependency, story string) {
	storyBranch := fmt.Sprintf("#%s", story)
	if strings.HasSuffix(p.Dependencies[dependency], storyBranch) {
		p.Dependencies[dependency] = strings.TrimSuffix(p.Dependencies[dependency], storyBranch)
	}
}

func (p *PackageJSON) ResetPrivateDependencyBranches(toReset, story string, projects ...string) {
	for range projects {
		if _, exists := p.Dependencies[toReset]; exists {
			p.resetPrivateDependencyBranch(toReset, story)
		}
	}
}

func (p *PackageJSON) SetPrivateDependencyBranchesToStory(story string, projects ...string) {
	for _, project := range projects {
		if _, exists := p.Dependencies[project]; exists {
			p.setPrivateDependencyBranchToStory(project, story)
		}
	}
}
