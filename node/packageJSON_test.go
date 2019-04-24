package node_test

import (
	"os"

	"encoding/json"
	"fmt"

	"github.com/LGUG2Z/story/node"
	"github.com/iancoleman/orderedmap"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
)

var p node.PackageJSON
var fs afero.Fs

var invalidFile = []byte(`{
  "name": "test",
  "description": "test project",
  "version": "0.0.1",
  "devDependencies": {
    "mocha": "*"
  }`)

func packageJSONWithDependencies(dependencies ...string) []byte {
	pkg := node.PackageJSON{
		Dependencies: make(map[string]string),
		Raw:          orderedmap.New(),
	}

	deps := make(map[string]string)
	pkg.Raw.Set("dependencies", deps)

	for _, dep := range dependencies {
		deps[dep] = fmt.Sprintf("git+ssh://git@github.com:TestOrg/%s.git", dep)
		pkg.Dependencies[dep] = fmt.Sprintf("git+ssh://git@github.com:TestOrg/%s.git", dep)
	}

	bytes, _ := json.MarshalIndent(pkg.Raw, "", "  ")
	return bytes
}

var _ = Describe("PackageJSON", func() {
	BeforeEach(func() {
		fs = afero.NewMemMapFs()
		p = node.PackageJSON{}
	})

	Describe("Loading a file", func() {
		It("It should load a valid package.json file", func() {
			// Given a project with a valid package.json file
			validFile := packageJSONWithDependencies("one", "two", "three")
			if err := fs.MkdirAll("valid", os.FileMode(0700)); err != nil {
				Fail(err.Error())
			}

			if err := afero.WriteFile(fs, "valid/package.json", validFile, os.FileMode(0600)); err != nil {
				Fail(err.Error())
			}

			// When I load the file
			Expect(p.Load(fs, "valid")).To(Succeed())

			// Then I expect it to be unmarshalled into an object
			Expect(p.Dependencies["one"]).To(Equal("git+ssh://git@github.com:TestOrg/one.git"))
		})

		It("It should throw an error when trying to load an invalid package.json file", func() {
			// Given a project with an invalid package.json file
			if err := fs.MkdirAll("invalid", os.FileMode(0700)); err != nil {
				Fail(err.Error())
			}

			if err := afero.WriteFile(fs, "invalid/package.json", invalidFile, os.FileMode(0600)); err != nil {
				Fail(err.Error())
			}

			// When I load the file then an error is thrown
			Expect(p.Load(fs, "invalid")).NotTo(Succeed())
		})
	})

	Describe("Writing a file", func() {
		It("Should write out the object to a package.json file", func() {
			// Given a project with a valid package.json file
			validFile := packageJSONWithDependencies("one", "two", "three")
			if err := fs.MkdirAll("valid", os.FileMode(0700)); err != nil {
				Fail(err.Error())
			}

			if err := afero.WriteFile(fs, "valid/package.json", validFile, os.FileMode(0600)); err != nil {
				Fail(err.Error())
			}

			// When I load the file
			Expect(p.Load(fs, "valid")).To(Succeed())

			// And make a change to the file
			p.Dependencies["new"] = "test"

			// When I write the object for a project
			Expect(p.Write(fs, "valid")).To(Succeed())

			// Then the file is written
			content, err := afero.ReadFile(fs, "valid/package.json")
			Expect(err).NotTo(HaveOccurred())

			Expect(string(content)).To(ContainSubstring(`"new": "test"`))
		})
	})

	Describe("Updating dependency branches", func() {
		It("Should update only the dependencies of projects in allProjects", func() {
			// Given a map of all projects and story projects
			projects := []string{"one", "three"}

			// And an unmarshalled package.json file
			b := packageJSONWithDependencies("one", "two", "three")
			Expect(json.Unmarshal(b, &p)).To(Succeed())
			p.Dependencies["one"] = "git+ssh://git@github.com:TestOrg/one.git#somegithash123"

			// When I update the dependencies to the story branch
			p.SetPrivateDependencyBranchesToStory("test-story", projects...)

			// Then the project in allProjects should be updated
			Expect(p.Dependencies["one"]).To(Equal("git+ssh://git@github.com:TestOrg/one.git#test-story"))

			// If the project doesn't have a hash already, it should be split correctly
			Expect(p.Dependencies["three"]).To(Equal("git+ssh://git@github.com:TestOrg/three.git#test-story"))

			// But the project not in allProjects should not be updated
			Expect(p.Dependencies["two"]).To(Equal("git+ssh://git@github.com:TestOrg/two.git"))
		})

		It("Should reset references to removed dependencies to the master branch", func() {
			// Given a package.json file with a dependency pinned to a story branch
			b := packageJSONWithDependencies("one", "two")
			Expect(json.Unmarshal(b, &p)).To(Succeed())
			p.Dependencies["one"] = "git+ssh://git@github.com:TestOrg/one.git#test-story"

			// When I reset all the modified branches
			p.ResetPrivateDependencyBranches("one", "test-story")

			// Then that dependency should point to the master branch
			Expect(p.Dependencies["one"]).To(Equal("git+ssh://git@github.com:TestOrg/one.git"))
		})

		It("Should reset all modified dependencies to use the master branch", func() {
			// Given a package.json file with a dependency pinned to a story branch
			b := packageJSONWithDependencies("one", "two")
			Expect(json.Unmarshal(b, &p)).To(Succeed())
			p.Dependencies["one"] = "git+ssh://git@github.com:TestOrg/one.git#test-story"

			// When I reset all the modified branches
			p.ResetPrivateDependencyBranchesToMaster("test-story")

			// Then that dependency should point to the master branch
			Expect(p.Dependencies["one"]).To(Equal("git+ssh://git@github.com:TestOrg/one.git"))
		})
	})
})
