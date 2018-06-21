package node

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/afero"
)

type PackageJSON struct {
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	Repository      map[string]string   `json:"repository"`
	Version         string              `json:"version"`
	Author          string              `json:"author"`
	Dependencies    map[string]string   `json:"dependencies,omitempty"`
	DevDependencies map[string]string   `json:"devDependencies,omitempty"`
	Scripts         map[string]string   `json:"scripts,omitempty"`
	LintStaged      map[string][]string `json:"lint-staged,omitempty"`
	Engines         map[string]string   `json:"engines,omitempty"`
	Jest            struct {
		CollectCoverageFrom     []string          `json:"collectCoverageFrom,omitempty"`
		SetupFiles              []string          `json:"setupFiles,omitempty"`
		TestMatch               []string          `json:"testMatch,omitempty"`
		TestEnvironment         string            `json:"testEnvironment,omitempty"`
		TestURL                 string            `json:"testURL,omitempty"`
		Transform               map[string]string `json:"transform,omitempty"`
		TransformIgnorePatterns []string          `json:"transformIgnorePatterns,omitempty"`
		ModuleNameMapper        map[string]string `json:"moduleNameMapper,omitempty"`
		ModuleFileExtensions    []string          `json:"moduleFileExtensions,omitempty"`
	} `json:"jest,omitempty"`
	Babel        map[string][]string `json:"babel,omitempty"`
	EslintConfig map[string]string   `json: eslintConfig,omitempty`
	Main         string              `json:"main,omitempty"`
	Bugs         map[string]string   `json:"bugs,omitempty"`
	Private      bool                `json:"private,omitempty"`
	License      string              `json:"license,omitempty"`
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

func (p *PackageJSON) ResetPrivateDependencyBranches(toReset, story string) {
	if _, exists := p.Dependencies[toReset]; exists {
		p.resetPrivateDependencyBranch(toReset, story)
	}
}

func (p *PackageJSON) SetPrivateDependencyBranchesToStory(story string, projects ...string) {
	for _, project := range projects {
		if _, exists := p.Dependencies[project]; exists {
			p.setPrivateDependencyBranchToStory(project, story)
		}
	}
}
