package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/afero"
	"github.com/LGUG2Z/story/manifest"
)

type PackageJSON struct {
	Name                 string              `json:"name,omitempty"`
	Version              string              `json:"version,omitempty"`
	Description          string              `json:"description,omitempty"`
	Keywords             []string            `json:"keywords,omitempty"`
	License              string              `json:"license,omitempty"`
	Homepage             string              `json:"homepage,omitempty"`
	Bugs                 map[string]string   `json:"bugs,omitempty"`
	Repository           map[string]string   `json:"repository,omitempty"`
	Author               string              `json:"author,omitempty"`
	Contributors         []string            `json:"contributors,omitempty"`
	Files                []string            `json:"files,omitempty"`
	Main                 string              `json:"main,omitempty"`
	Bin                  string              `json:"bin,omitempty"`
	Man                  string              `json:"man,omitempty"`
	Directories          map[string]string   `json:"directories,omitempty"`
	Scripts              map[string]string   `json:"scripts,omitempty"`
	Config               map[string]string   `json:"config,omitempty"`
	Dependencies         map[string]string   `json:"dependencies,omitempty"`
	DevDependencies      map[string]string   `json:"devDependencies,omitempty"`
	PeerDependencies     map[string]string   `json:"peerDependencies,omitempty"`
	OptionalDependencies map[string]string   `json:"optionalDependencies,omitempty"`
	BundledDependencies  map[string]string   `json:"bundledDependencies,omitempty"`
	Flat                 bool                `json:"flat,omitempty"`
	Resolutions          map[string]string   `json:"resolution,omitempty"`
	Engines              map[string]string   `json:"engines,omitempty"`
	OS                   map[string][]string `json:"os,omitempty"`
	CPU                  map[string][]string `json:"cpu,omitempty"`
	Private              bool                `json:"private,omitempty"`
	PublishConfig        map[string]string   `json:"private,omitempty"`
	Transform            map[string]string   `json:"transform,omitempty"`
	LintStaged           map[string][]string `json:"lint-staged,omitempty"`
	Jest                 *JestConfig         `json:"jest,omitempty"`
	Babel                *BabelConfig        `json:"babel,omitempty"`
	EslintConfig         *EslintConfig       `json:"eslintConfig,omitempty"`
}

type EslintConfig struct {
	Extends string `json:"extends"`
}

type BabelConfig struct {
	Presets []string `json:"presets"`
}

type JestConfig struct {
	Automock                     bool                              `json:"automock,omitempty"`
	Browser                      bool                              `json:"browser,omitempty"`
	Bail                         bool                              `json:"bail,omitempty"`
	CacheDirectory               string                            `json:"cacheDirectory,omitempty"`
	ClearMocks                   bool                              `json:"clearMocks,omitempty"`
	CollectCoverage              bool                              `json:"collectCoverage,omitempty"`
	CollectCoverageFrom          []string                          `json:"collectCoverageFrom,omitempty"`
	CoverageDirectory            string                            `json:"coverageDirectory,omitempty"`
	CoveragePathIgnorePatterns   []string                          `json:"coveragePathIgnorePatterns,omitempty"`
	CoverageReporters            []string                          `json:"coverageReporters,omitempty"`
	CoverageThreshold            map[string]map[string]interface{} `json:"coverageThreshold,omitempty"`
	ForceCoverageMatch           []string                          `json:"forceCoverageMatch,omitempty"`
	Globals                      map[string]interface{}            `json:"globals,omitempty"`
	GlobalSetup                  string                            `json:"globalSetup,omitempty"`
	GlobalTeardown               string                            `json:"globalTeardown,omitempty"`
	MocksPattern                 string                            `json:"mocksPattern,omitempty"`
	ModuleFileExtensions         []string                          `json:"moduleFileExtensions,omitempty"`
	ModuleDirectories            []string                          `json:"moduleDirectories,omitempty"`
	ModuleNameMapper             map[string]string                 `json:"moduleNameMapper,omitempty"`
	ModulePathIgnorePatterns     []string                          `json:"modulePathIgnorePatterns,omitempty"`
	ModulePaths                  []string                          `json:"modulePaths,omitempty"`
	Notify                       bool                              `json:"notify,omitempty"`
	NotifyMode                   string                            `json:"notifyMode,omitempty"`
	Preset                       string                            `json:"preset,omitempty"`
	Projects                     []string                          `json:"projects,omitempty"`
	Reporters                    []string                          `json:"reporters,omitempty"`
	ResetMocks                   bool                              `json:"resetMocks,omitempty"`
	ResetModules                 bool                              `json:"resetModules,omitempty"`
	Resolver                     string                            `json:"resolver,omitempty"`
	RestoreMocks                 bool                              `json:"restoreMocks,omitempty"`
	RootDir                      string                            `json:"rootDir,omitempty"`
	Roots                        []string                          `json:"roots,omitempty"`
	Runner                       string                            `json:"runner,omitempty"`
	SetupFiles                   []string                          `json:"setupFiles,omitempty"`
	SetupTestFrameworkScriptFile string                            `json:"setupTestFrameworkScriptFile,omitempty"`
	SnapshotSerializers          []string                          `json:"snapshotSerializers,omitempty"`
	TestEnvironment              string                            `json:"testEnvironment,omitempty"`
	TestEnvironmentOptions       map[string]string                 `json:"testEnvironmentOptions,omitempty"`
	TestMatch                    []string                          `json:"testMatch,omitempty"`
	TestPathDirs                 []string                          `json:"testPathDirs,omitempty"`
	TestPathIgnorePatterns       []string                          `json:"testPathIgnorePatterns,omitempty"`
	TestRegex                    string                            `json:"testRegex,omitempty"`
	TestResultsProcessor         string                            `json:"testResultsProcessor,omitempty"`
	TestRunner                   string                            `json:"testRunner,omitempty"`
	TestURL                      string                            `json:"testURL,omitempty"`
	Timers                       string                            `json:"timers,omitempty"`
	Transform                    map[string]string                 `json:"transform,omitempty"`
	TransformIgnorePatterns      []string                          `json:"transformIgnorePatterns,omitempty"`
	UnmockedModulePathPatterns   []string                          `json:"unmockedModulePathPatterns,omitempty"`
	Verbose                      bool                              `json:"verbose,omitempty"`
	WatchPathIgnorePatterns      []string                          `json:"watchPatchIgnorePatterns,omitempty"`
}

func (p *PackageJSON) Load(fs afero.Fs, project string) error {
	bytes, err := afero.ReadFile(fs, fmt.Sprintf("%s/package.json", project))
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, p)
}

func (p *PackageJSON) Write(fs afero.Fs, project string) error {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(&p); err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/package.json", project)
	return afero.WriteFile(fs, filename, buf.Bytes(), os.FileMode(0666))
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
