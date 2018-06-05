package meta_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/AlexsJones/kepler/commands/node"
	"github.com/LGUG2Z/story/meta"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"gopkg.in/src-d/go-billy.v4/memfs"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"strings"
)

var m meta.Manifest

func createMetaRepo(globalMeta []byte) error {
	r, err := git.PlainInit("test", false)
	if err := ioutil.WriteFile("test/.meta", globalMeta, os.FileMode(0666)); err != nil {
		return err
	}

	if err := ioutil.WriteFile("test/.gitignore", []byte("one\ntwo"), os.FileMode(0666)); err != nil {
		return err
	}

	wt, err := r.Worktree()
	if err != nil {
		return err
	}

	_, err = wt.Add(".meta")
	if err != nil {
		return err
	}

	_, err = wt.Add(".gitignore")
	if err != nil {
		return err
	}

	_, err = wt.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{Name: "John Doe", Email: "john@doe.org", When: time.Now()},
	})

	return err
}

func addProjectToMetaRepo(project string, packageJSON []byte) error {
	repo, err := git.PlainInit(project, false)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(fmt.Sprintf("%s/package.json", project), packageJSON, os.FileMode(0666)); err != nil {
		return err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return err
	}

	_, err = wt.Add("package.json")
	if err != nil {
		return err
	}

	_, err = wt.Commit("package.json commit", &git.CommitOptions{
		Author: &object.Signature{Name: "John Doe", Email: "john@doe.org", When: time.Now()},
	})

	return err
}

func commitFilesToProjectRepo(relativeProjectPath string, files []string) error {
	repo, err := git.PlainOpen(relativeProjectPath)
	Expect(err).NotTo(HaveOccurred())

	wt, err := repo.Worktree()
	Expect(err).NotTo(HaveOccurred())

	for _, file := range files {
		_, err = wt.Add(file)
		if err != nil {
			return err
		}
	}

	_, err = wt.Commit(fmt.Sprintf("updated: %s", strings.Join(files, ", ")), &git.CommitOptions{
		Author: &object.Signature{Name: "John Doe", Email: "john@doe.org", When: time.Now()},
	})

	return err
}

func getCurrentBranchForProjectRepo(relativeProjectPath string) (string, error) {
	repo, err := git.PlainOpen(relativeProjectPath)
	if err != nil {
		return "", err
	}
	head, err := repo.Head()
	if err != nil {
		return "", nil
	}

	return head.Name().String(), nil
}

func packageJSONWithDependencies(dependencies []string) []byte {
	pkg := node.PackageJSON{
		Dependencies: make(map[string]string),
	}

	for _, dep := range dependencies {
		pkg.Dependencies[dep] = fmt.Sprintf("git+ssh://git@github.com:TestOrg/%s.git", dep)
	}

	bytes, _ := json.MarshalIndent(pkg, "", "  ")
	return bytes
}

func globalMetaWithProjects(projects []string) []byte {
	global := meta.Manifest{Projects: make(map[string]string)}

	for _, project := range projects {
		global.Projects[project] = fmt.Sprintf("git@github.com:TestOrg/%s.git", project)
	}

	bytes, _ := json.MarshalIndent(global, "", "  ")
	return bytes
}

func storyMetaWithProjects(name string, projects []string) []byte {
	story := meta.Manifest{Name: name, Projects: make(map[string]string)}

	for _, project := range projects {
		story.Projects[project] = fmt.Sprintf("git@github.com:TestOrg/%s.git", project)
	}

	bytes, _ := json.MarshalIndent(story, "", "  ")
	return bytes
}

var _ = Describe("Meta", func() {
	globalMeta := globalMetaWithProjects([]string{"one", "two", "three"})
	one := packageJSONWithDependencies([]string{})
	two := packageJSONWithDependencies([]string{"one"})
	three := packageJSONWithDependencies([]string{"one", "three"})

	BeforeEach(func() {
		m = meta.Manifest{Fs: afero.NewMemMapFs()}
		Expect(afero.WriteFile(m.Fs, ".meta", globalMeta, os.FileMode(0666))).To(Succeed())

		Expect(m.Fs.Mkdir("one", os.FileMode(0666))).To(Succeed())
		Expect(afero.WriteFile(m.Fs, "one/package.json", one, os.FileMode(0666))).To(Succeed())

		Expect(m.Fs.Mkdir("two", os.FileMode(0666))).To(Succeed())
		Expect(afero.WriteFile(m.Fs, "two/package.json", two, os.FileMode(0666))).To(Succeed())

		Expect(m.Fs.Mkdir("three", os.FileMode(0666))).To(Succeed())
		Expect(afero.WriteFile(m.Fs, "three/package.json", three, os.FileMode(0666))).To(Succeed())
		Expect(os.Setenv("ORGANISATION", "TestOrg")).To(Succeed())
		Expect(os.Setenv("TEST", "1")).To(Succeed())
	})

	AfterEach(func() {
		Expect(os.Unsetenv("ORGANISATION")).To(Succeed())
		Expect(os.Unsetenv("TEST")).To(Succeed())

	})

	Describe("Reading .meta files", func() {
		Context("With an invalid story meta file", func() {
			It("Should return an error", func() {
				_, err := m.Fs.Create(".meta")
				Expect(err).NotTo(HaveOccurred())
				Expect(m.Load(".meta")).NotTo(Succeed())
			})
		})

		Context("With a valid story meta file", func() {
			It("Should be identifiable as a story meta file", func() {
				story := storyMetaWithProjects("story-1", []string{})
				Expect(afero.WriteFile(m.Fs, ".meta", story, os.FileMode(0666))).To(Succeed())

				s := meta.Manifest{Fs: m.Fs}
				Expect(s.Load(".meta")).To(Succeed())
				Expect(s.IsStory()).To(BeTrue())
			})
		})

		Context("With an invalid global meta file", func() {
			It("Should return an error", func() {
				_, err := m.Fs.Create(".meta")
				Expect(err).NotTo(HaveOccurred())
				Expect(m.Load(".meta")).NotTo(Succeed())
			})
		})

		Context("With a valid global meta file", func() {
			It("Should be identifiable as a global meta file", func() {
				Expect(m.Load(".meta")).To(Succeed())
			})
		})

		Context("With a missing global metal file", func() {
			It("Should return an error", func() {
				Expect(m.Fs.Remove(".meta")).To(Succeed())
				Expect(m.Load(".meta")).NotTo(Succeed())
			})
		})
	})

	Describe("Setting a story", func() {
		Context("With an existing story on a checked out branch", func() {
			It("Should check out the branches for each project in the story", func() {
				Expect(os.Unsetenv("TEST")).To(Succeed())
				Expect(os.RemoveAll("test")).To(Succeed())

				// GIVEN a meta repo
				Expect(createMetaRepo(globalMeta)).To(Succeed())
				Expect(os.Chdir("test")).To(Succeed())
				m.Fs = afero.NewOsFs()

				// AND one repo within the meta repo
				Expect(addProjectToMetaRepo("one", one)).To(Succeed())

				// AND a story set
				Expect(m.Load(".meta")).To(Succeed())
				Expect(m.SetStory("some-story")).To(Succeed())

				// WITH project one added to the story and updated
				s := meta.Manifest{Fs: m.Fs}
				Expect(s.Load(".meta")).To(Succeed())
				Expect(s.AddProjects([]string{"one"})).To(Succeed())
				Expect(commitFilesToProjectRepo(".", []string{".meta", ".meta.json"})).To(Succeed())

				// AND then being reset back to the master branch
				Expect(s.Reset()).To(Succeed())

				// WHEN I run set story again
				s = meta.Manifest{Fs: m.Fs}
				Expect(s.SetStory("some-story")).To(Succeed())

				// THEN the meta repo and project one should be on the some-story branch
				metaRepoBranch, err := getCurrentBranchForProjectRepo(".")
				Expect(err).NotTo(HaveOccurred())

				projectOneBranch, err := getCurrentBranchForProjectRepo("one")
				Expect(err).NotTo(HaveOccurred())

				Expect(metaRepoBranch).To(Equal("refs/heads/some-story"))
				Expect(projectOneBranch).To(Equal("refs/heads/some-story"))

				Expect(os.Chdir("..")).To(Succeed())
				Expect(os.RemoveAll("test")).To(Succeed())
			})
		})
	})

	Describe("Adding projects", func() {
		Context("With no story currently set", func() {
			It("Should return an error", func() {
				Expect(m.Load(".meta")).To(Succeed())
				Expect(m.AddProjects([]string{"one"})).NotTo(Succeed())
			})
		})

		Context("With a story set", func() {
			Context("Adding a project that isn't in the global meta", func() {
				It("Should skip that project and continue to add valid projects", func() {
					Expect(m.Load(".meta")).To(Succeed())
					Expect(m.SetStory("some-story")).To(Succeed())
					s := meta.Manifest{Fs: m.Fs}
					Expect(s.Load(".meta")).To(Succeed())

					Expect(s.AddProjects([]string{"one", "not-a-project"})).To(Succeed())
					Expect(s.Projects).To(HaveKeyWithValue("one", "git@github.com:TestOrg/one.git"))
					Expect(s.Projects).ToNot(HaveKey("not-a-project"))
				})
			})

			Context("Adding a project with no dependencies", func() {
				It("Should add just the project specified", func() {
					Expect(m.Load(".meta")).To(Succeed())
					Expect(m.SetStory("some-story")).To(Succeed())
					s := meta.Manifest{Fs: m.Fs}
					Expect(s.Load(".meta")).To(Succeed())

					Expect(s.AddProjects([]string{"one"})).To(Succeed())
					Expect(s.Projects).To(HaveKeyWithValue("one", "git@github.com:TestOrg/one.git"))
				})

			})

			Context("Adding a project with dependencies", func() {
				It("Should log an error and continue if a project has an invalid package.json", func() {
					Expect(m.Load(".meta")).To(Succeed())
					Expect(m.SetStory("some-story")).To(Succeed())
					s := meta.Manifest{Fs: m.Fs}
					Expect(s.Load(".meta")).To(Succeed())

					Expect(afero.WriteFile(m.Fs, "two/package.json", []byte{}, os.FileMode(0666))).To(Succeed())
					Expect(s.AddProjects([]string{"two"})).To(Succeed())

					Expect(s.Projects).To(HaveKeyWithValue("two", "git@github.com:TestOrg/two.git"))
				})

				It("Should add the dependencies of the project", func() {
					Expect(m.Load(".meta")).To(Succeed())
					Expect(m.SetStory("some-story")).To(Succeed())
					s := meta.Manifest{Fs: m.Fs}
					Expect(s.Load(".meta")).To(Succeed())

					Expect(s.AddProjects([]string{"two"})).To(Succeed())

					Expect(s.Projects).To(HaveKeyWithValue("one", "git@github.com:TestOrg/one.git"))
					Expect(s.Projects).To(HaveKeyWithValue("two", "git@github.com:TestOrg/two.git"))
				})

				It("Should only add the given project as a primary project", func() {
					Expect(m.Load(".meta")).To(Succeed())
					Expect(m.SetStory("some-story")).To(Succeed())
					s := meta.Manifest{Fs: m.Fs}
					Expect(s.Load(".meta")).To(Succeed())

					Expect(s.AddProjects([]string{"two"})).To(Succeed())

					Expect(s.Primaries).To(HaveKeyWithValue("two", true))
					Expect(s.Primaries).ToNot(HaveKeyWithValue("one", true))
				})

				It("Should update dependencies in the package.json by appending #story", func() {
					Expect(m.Load(".meta")).To(Succeed())
					Expect(m.SetStory("some-story")).To(Succeed())
					s := meta.Manifest{Fs: m.Fs}
					Expect(s.Load(".meta")).To(Succeed())

					Expect(s.AddProjects([]string{"two"})).To(Succeed())

					Expect(s.Projects).To(HaveKeyWithValue("one", "git@github.com:TestOrg/one.git"))
					Expect(s.Projects).To(HaveKeyWithValue("two", "git@github.com:TestOrg/two.git"))

					bytes, err := afero.ReadFile(m.Fs, "two/package.json")
					Expect(err).NotTo(HaveOccurred())

					p := &node.PackageJSON{}
					Expect(json.Unmarshal(bytes, p)).To(Succeed())
					Expect(p.Dependencies).To(HaveKeyWithValue("one", "git+ssh://git@github.com:TestOrg/one.git#some-story"))
				})
			})
		})
	})

	Describe("Removing projects", func() {
		Context("With no story currently set", func() {
			It("Should return an error", func() {
				Expect(m.Load(".meta")).To(Succeed())
				Expect(m.RemoveProjects([]string{"one"})).NotTo(Succeed())
			})
		})

		Context("With a story set", func() {
			Context("Removing an added project with no dependencies", func() {
				It("Should remove the given project", func() {
					Expect(m.Load(".meta")).To(Succeed())
					Expect(m.SetStory("some-story")).To(Succeed())
					s := meta.Manifest{Fs: m.Fs}
					Expect(s.Load(".meta")).To(Succeed())

					Expect(s.AddProjects([]string{"one"})).To(Succeed())
					Expect(s.Projects).To(HaveKeyWithValue("one", "git@github.com:TestOrg/one.git"))

					Expect(s.RemoveProjects([]string{"one"})).To(Succeed())
					Expect(s.Projects).ToNot(HaveKeyWithValue("one", "git@github.com:TestOrg/one.git"))
				})

				Context("That is a dependency of a primary project", func() {
					It("Should remove the given project and update the package.json in the primary project", func() {
						Expect(m.Load(".meta")).To(Succeed())
						Expect(m.SetStory("some-story")).To(Succeed())
						s := meta.Manifest{Fs: m.Fs}
						Expect(s.Load(".meta")).To(Succeed())
						Expect(s.AddProjects([]string{"two"})).To(Succeed())

						Expect(s.RemoveProjects([]string{"one"})).To(Succeed())

						bytes, err := afero.ReadFile(m.Fs, "two/package.json")
						Expect(err).NotTo(HaveOccurred())

						p := &node.PackageJSON{}
						Expect(json.Unmarshal(bytes, p)).To(Succeed())
						Expect(p.Dependencies).To(HaveKeyWithValue("one", "git+ssh://git@github.com:TestOrg/one.git"))
					})
				})
			})

			Context("Removing a project with dependencies", func() {
				It("Should update dependencies in the package.json by removing the #story suffix", func() {
					Expect(m.Load(".meta")).To(Succeed())
					Expect(m.SetStory("some-story")).To(Succeed())
					s := meta.Manifest{Fs: m.Fs}
					Expect(s.Load(".meta")).To(Succeed())
					Expect(s.AddProjects([]string{"two"})).To(Succeed())

					Expect(s.RemoveProjects([]string{"two"})).To(Succeed())

					bytes, err := afero.ReadFile(m.Fs, "two/package.json")
					Expect(err).NotTo(HaveOccurred())

					p := &node.PackageJSON{}
					Expect(json.Unmarshal(bytes, p)).To(Succeed())
					Expect(p.Dependencies).To(HaveKeyWithValue("one", "git+ssh://git@github.com:TestOrg/one.git"))
				})

				Context("That are not shared by other primary projects", func() {
					It("Should remove the project and its dependencies", func() {
						Expect(m.Load(".meta")).To(Succeed())
						Expect(m.SetStory("some-story")).To(Succeed())
						s := meta.Manifest{Fs: m.Fs}
						Expect(s.Load(".meta")).To(Succeed())

						Expect(s.AddProjects([]string{"two"})).To(Succeed())
						Expect(s.Projects).To(HaveKeyWithValue("one", "git@github.com:TestOrg/one.git"))
						Expect(s.Projects).To(HaveKeyWithValue("two", "git@github.com:TestOrg/two.git"))

						Expect(s.RemoveProjects([]string{"two"})).To(Succeed())
						Expect(s.Projects).ToNot(HaveKeyWithValue("one", "git@github.com:TestOrg/one.git"))
						Expect(s.Projects).ToNot(HaveKeyWithValue("two", "git@github.com:TestOrg/two.git"))
					})
				})

				Context("That are shared by other primary projects", func() {
					It("Should remove the project and leave dependencies which are also in other primary projects", func() {
						Expect(m.Load(".meta")).To(Succeed())
						Expect(m.SetStory("some-story")).To(Succeed())
						s := meta.Manifest{Fs: m.Fs}
						Expect(s.Load(".meta")).To(Succeed())

						Expect(s.AddProjects([]string{"two"})).To(Succeed())
						Expect(s.AddProjects([]string{"three"})).To(Succeed())
						Expect(s.Projects).To(HaveKeyWithValue("one", "git@github.com:TestOrg/one.git"))
						Expect(s.Projects).To(HaveKeyWithValue("two", "git@github.com:TestOrg/two.git"))
						Expect(s.Projects).To(HaveKeyWithValue("three", "git@github.com:TestOrg/three.git"))

						Expect(s.RemoveProjects([]string{"three"})).To(Succeed())
						Expect(s.Projects).To(HaveKeyWithValue("one", "git@github.com:TestOrg/one.git"))
						Expect(s.Projects).To(HaveKeyWithValue("two", "git@github.com:TestOrg/two.git"))
						Expect(s.Projects).ToNot(HaveKeyWithValue("three", "git@github.com:TestOrg/three.git"))
					})
				})
			})
		})
	})

	Describe("Pruning unchanged projects", func() {
		Context("With projects that point to different commits on the story branch and master", func() {
			It("Should not prune those projects from the meta file", func() {
				// Setup
				Expect(os.Unsetenv("TEST")).To(Succeed())
				Expect(os.RemoveAll("test")).To(Succeed())

				// GIVEN a meta repo
				Expect(createMetaRepo(globalMeta)).To(Succeed())
				Expect(os.Chdir("test")).To(Succeed())
				m.Fs = afero.NewOsFs()

				// AND two repos within the meta repo
				Expect(addProjectToMetaRepo("one", one)).To(Succeed())
				Expect(addProjectToMetaRepo("two", two)).To(Succeed())

				// AND a story set
				Expect(m.Load(".meta")).To(Succeed())
				Expect(m.SetStory("some-story")).To(Succeed())

				// WITH project two added to the story and updated
				s := meta.Manifest{Fs: m.Fs}
				Expect(s.Load(".meta")).To(Succeed())
				Expect(s.AddProjects([]string{"two"})).To(Succeed())
				Expect(commitFilesToProjectRepo("two", []string{"package.json"})).To(Succeed())

				// WHEN I run the pruner
				Expect(s.Prune()).To(Succeed())

				// THEN project one is removed from the story but project two remains
				Expect(s.Projects).NotTo(HaveKey("one"))
				Expect(s.Projects).To(HaveKey("two"))

				// AND the package.json of project two has been reverted
				bytes, err := ioutil.ReadFile("two/package.json")
				Expect(err).NotTo(HaveOccurred())

				p := &node.PackageJSON{}
				Expect(json.Unmarshal(bytes, p)).To(Succeed())
				Expect(p.Dependencies).To(HaveKeyWithValue("one", "git+ssh://git@github.com:TestOrg/one.git"))

				// Cleanup
				Expect(os.Chdir("..")).To(Succeed())
				Expect(os.RemoveAll("test")).To(Succeed())
			})
		})

		Context("With projects that point to the same commit on the story branch and master", func() {
			It("Should prune those projects from the meta file and reset any package.json changes", func() {
				// Setup
				Expect(os.Unsetenv("TEST")).To(Succeed())
				Expect(os.RemoveAll("test")).To(Succeed())

				// GIVEN a meta repo
				Expect(createMetaRepo(globalMeta)).To(Succeed())
				Expect(os.Chdir("test")).To(Succeed())
				m.Fs = afero.NewOsFs()

				// AND two projects within the meta repo
				Expect(addProjectToMetaRepo("one", one)).To(Succeed())
				Expect(addProjectToMetaRepo("two", two)).To(Succeed())

				// AND a new story with project two added
				Expect(m.Load(".meta")).To(Succeed())
				Expect(m.SetStory("some-story")).To(Succeed())
				s := meta.Manifest{Fs: m.Fs}
				Expect(s.Load(".meta")).To(Succeed())
				Expect(s.AddProjects([]string{"two"})).To(Succeed())

				// AND changes to project two reverted
				Expect(ioutil.WriteFile("two/package.json", two, os.FileMode(0666))).To(Succeed())

				// WHEN I run the pruner
				Expect(s.Prune()).To(Succeed())

				// THEN both projects are removed from the story
				Expect(s.Projects).NotTo(HaveKey("one"))
				Expect(s.Projects).NotTo(HaveKey("two"))

				// AND both projects are on the master branch
				projectOneBranch, err := getCurrentBranchForProjectRepo("one")
				Expect(err).NotTo(HaveOccurred())
				projectTwoBranch, err := getCurrentBranchForProjectRepo("two")
				Expect(err).NotTo(HaveOccurred())

				Expect(projectOneBranch).To(Equal("refs/heads/master"))
				Expect(projectTwoBranch).To(Equal("refs/heads/master"))

				// Cleanup
				Expect(os.Chdir("..")).To(Succeed())
				Expect(os.RemoveAll("test")).To(Succeed())
			})
		})
	})

	Describe("Adding additional projects within the blast radius of a story", func() {
		Context("With a project within the blast radius", func() {
			It("Should add it to the story's .meta file", func() {
				Expect(m.Load(".meta")).To(Succeed())
				Expect(m.SetStory("some-story")).To(Succeed())
				s := meta.Manifest{Fs: m.Fs}
				Expect(s.Load(".meta")).To(Succeed())

				Expect(s.AddProjects([]string{"two"})).To(Succeed())
				Expect(s.Blast()).To(Succeed())

				Expect(s.Projects).To(HaveKeyWithValue("three", "git@github.com:TestOrg/three.git"))
			})
		})
	})

	Describe("Deleting branches", func() {
		Context("In a project repo that has multiple branches", func() {
			It("Should delete the specified branch", func() {
				Expect(os.Unsetenv("TEST")).To(Succeed())
				// given a repo with a commit on master
				s := memory.NewStorage()
				wt := memfs.New()

				repository, err := git.Init(s, wt)
				Expect(err).NotTo(HaveOccurred())

				_, err = wt.Create("README.md")
				Expect(err).NotTo(HaveOccurred())
				workTree, err := repository.Worktree()

				Expect(err).NotTo(HaveOccurred())
				workTree.Add("README.md")

				_, err = workTree.Commit("adding readme", &git.CommitOptions{
					Author: &object.Signature{Name: "some-author", Email: "some@author.com"},
				})
				Expect(err).NotTo(HaveOccurred())

				// AND another branch
				Expect(meta.CheckoutBranch("test-story", repository)).To(Succeed())

				// WHEN I delete that branch
				Expect(meta.DeleteBranch("test-story", repository)).To(Succeed())

				// THEN I should not see that branch referenced in the repo anymore
				var branch *plumbing.Reference
				branches, err := repository.Branches()
				branches.ForEach(func(r *plumbing.Reference) error {
					if r.Name().String() == "refs/heads/test-story" {
						branch = r
					}
					return nil
				})

				Expect(branch).To(BeNil())
			})
		})
	})

	Describe("Checking out branches", func() {
		Context("In a project repo that doesn't have the branch to be created", func() {
			It("Should create the branch", func() {
				Expect(os.Unsetenv("TEST")).To(Succeed())
				// given a repo with a commit on master
				s := memory.NewStorage()
				wt := memfs.New()

				repository, err := git.Init(s, wt)
				Expect(err).NotTo(HaveOccurred())

				_, err = wt.Create("README.md")
				Expect(err).NotTo(HaveOccurred())
				workTree, err := repository.Worktree()

				Expect(err).NotTo(HaveOccurred())
				workTree.Add("README.md")

				_, err = workTree.Commit("adding readme", &git.CommitOptions{
					Author: &object.Signature{Name: "some-author", Email: "some@author.com"},
				})
				Expect(err).NotTo(HaveOccurred())

				// when I run CheckoutBranch
				Expect(meta.CheckoutBranch("test-story", repository)).To(Succeed())
				branches, err := repository.Branches()

				var branch *plumbing.Reference

				// then I should have that branch in my repo
				branches.ForEach(func(r *plumbing.Reference) error {
					if r.Name().String() == "refs/heads/test-story" {
						branch = r
					}
					return nil
				})

				Expect(branch).NotTo(BeNil())
			})
		})
	})

	Describe("Completing a story", func() {
		It("Should remove all references to the story branch in all package.json files", func() {
			Expect(m.Load(".meta")).To(Succeed())
			Expect(m.SetStory("some-story")).To(Succeed())
			s := meta.Manifest{Fs: m.Fs}
			Expect(s.Load(".meta")).To(Succeed())
			Expect(s.AddProjects([]string{"one", "two", "three"})).To(Succeed())

			Expect(s.Complete()).To(Succeed())

			for project := range s.Projects {
				bytes, err := afero.ReadFile(m.Fs, fmt.Sprintf("%s/package.json", project))
				Expect(err).NotTo(HaveOccurred())
				p := &node.PackageJSON{}

				Expect(json.Unmarshal(bytes, p)).To(Succeed())

				for _, repo := range p.Dependencies {
					Expect(repo).NotTo(HaveSuffix("#some-story"))
				}
			}
		})
	})
})
