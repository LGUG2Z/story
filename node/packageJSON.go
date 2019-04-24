package node

import (
	"encoding/json"

	"github.com/iancoleman/orderedmap"

	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/LGUG2Z/story/manifest"
	"github.com/spf13/afero"
)

type PackageJSON struct {
	Raw             *orderedmap.OrderedMap
	Dependencies    map[string]string
	DevDependencies map[string]string
}

func (p *PackageJSON) Load(fs afero.Fs, project string) error {
	b, err := afero.ReadFile(fs, fmt.Sprintf("%s/package.json", project))
	if err != nil {
		return err
	}

	p.Raw = orderedmap.New()
	p.Dependencies = make(map[string]string)
	p.DevDependencies = make(map[string]string)

	if err = p.Raw.UnmarshalJSON(b); err != nil {
		return err
	}

	if dependencies, ok := p.Raw.Get("dependencies"); ok {
		d := dependencies.(orderedmap.OrderedMap)
		for _, k := range d.Keys() {
			if v, ok := d.Get(k); ok {
				p.Dependencies[k] = v.(string)
			}
		}
	}

	if devDependencies, ok := p.Raw.Get("devDependencies"); ok {
		d := devDependencies.(orderedmap.OrderedMap)
		for _, k := range d.Keys() {
			if v, ok := d.Get(k); ok {
				p.DevDependencies[k] = v.(string)
			}
		}
	}

	return nil
}

func (p *PackageJSON) Write(fs afero.Fs, project string) error {
	dependencies, err := json.Marshal(p.Dependencies)
	if err != nil {
		return err
	}

	devDependencies, err := json.Marshal(p.DevDependencies)
	if err != nil {
		return err
	}

	p.Raw.Set("dependencies", json.RawMessage(dependencies))
	p.Raw.Set("devDependencies", json.RawMessage(devDependencies))

	b, err := json.MarshalIndent(&p.Raw, "", "  ")
	if err != nil {
		return err
	}

	b = bytes.Replace(b, []byte("\\u003c"), []byte("<"), -1)
	b = bytes.Replace(b, []byte("\\u003e"), []byte(">"), -1)
	b = bytes.Replace(b, []byte("\\u0026"), []byte("&"), -1)

	filename := fmt.Sprintf("%s/package.json", project)
	return afero.WriteFile(fs, filename, b, os.FileMode(0666))
}

func (p *PackageJSON) setPrivateDependencyBranchToStory(dependency, story string) {
	// TODO: Update this to strip out any commit hashes
	if strings.Contains(p.Dependencies[dependency], ".git") {
		// Append #story-branch-name to the current git+ssh string
		s := strings.Split(p.Dependencies[dependency], "#")
		p.Dependencies[dependency] = fmt.Sprintf("%s#%s", s[0], story)
	}
}

func (p *PackageJSON) setPrivateDependencyBranchToCommitHash(dependency, commitHash string) {
	// TODO: Update this to strip out any commit hashes
	if strings.Contains(p.Dependencies[dependency], ".git") {
		// Append #story-branch-name to the current git+ssh string
		s := strings.Split(p.Dependencies[dependency], "#")
		p.Dependencies[dependency] = fmt.Sprintf("%s#%s", s[0], commitHash)
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

func (p *PackageJSON) ResetPrivateDependencyBranchesToCommitHash(story *manifest.Story) {
	storyBranch := fmt.Sprintf("#%s", story.Name)
	for pkg, src := range p.Dependencies {
		if strings.HasSuffix(src, storyBranch) {
			p.Dependencies[pkg] = strings.TrimSuffix(src, storyBranch)
			p.Dependencies[pkg] = fmt.Sprintf("%s#%s", p.Dependencies[pkg], story.Hashes[pkg])
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

func (p *PackageJSON) SetPrivateDependencyBranchesToCommitHashes(story *manifest.Story, projects ...string) {
	for _, project := range projects {
		if _, exists := p.Dependencies[project]; exists {
			p.setPrivateDependencyBranchToStory(project, story.Hashes[project])
		}
	}
}
