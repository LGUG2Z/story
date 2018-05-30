package meta_test

import (
	"encoding/json"
	"fmt"
	"github.com/AlexsJones/kepler/commands/node"
	"github.com/LGUG2Z/story/meta"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"io/ioutil"
	"os"
	"time"
)

var m meta.Manifest

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
	})

	AfterEach(func() {
		Expect(os.Unsetenv("ORGANISATION")).To(Succeed())

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

	Describe("Setting and resetting stories", func() {
		Context("With no story currently set", func() {
			Context("With a local story .meta for the requested story", func() {
				It("Should set the story using the existing story meta", func() {
					story := storyMetaWithProjects("test-story", []string{"one", "two"})
					afero.WriteFile(m.Fs, ".meta.test-story", story, os.FileMode(0666))

					Expect(m.Load(".meta")).To(Succeed())
					Expect(m.SetStory("test-story")).To(Succeed())

					s := meta.Manifest{Fs: m.Fs}
					Expect(s.Load(".meta")).To(Succeed())
					Expect(s.Projects).To(HaveKeyWithValue("one", "git@github.com:TestOrg/one.git"))
					Expect(s.Projects).To(HaveKeyWithValue("two", "git@github.com:TestOrg/two.git"))
				})
			})
		})

		Context("With a story currently set", func() {
			Context("Resetting to the global meta", func() {
				It("Should move the current story to a backup file", func() {
					Expect(m.Load(".meta")).To(Succeed())
					Expect(m.SetStory("some-story")).To(Succeed())
					s := meta.Manifest{Fs: m.Fs}
					Expect(s.Load(".meta")).To(Succeed())
					Expect(s.IsStory()).To(BeTrue())

					Expect(s.RestoreGlobal()).To(Succeed())
					Expect(afero.Exists(s.Fs, ".meta.some-story")).To(BeTrue())
				})

				It("Should set the meta file as the global meta", func() {
					Expect(m.Load(".meta")).To(Succeed())
					Expect(m.SetStory("some-story")).To(Succeed())
					s := meta.Manifest{Fs: m.Fs}
					Expect(s.Load(".meta")).To(Succeed())
					Expect(s.IsStory()).To(BeTrue())

					Expect(s.RestoreGlobal()).To(Succeed())
					g := meta.Manifest{Fs: m.Fs}
					Expect(s.Load(".meta")).To(Succeed())
					Expect(g.IsStory()).To(BeFalse())
				})
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
					fmt.Println(string(bytes))
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
						fmt.Println(string(bytes))
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
					fmt.Println(string(bytes))
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
			It("Should not prone those projects from the meta file", func() {
				// initial global meta
				Expect(ioutil.WriteFile(".meta", globalMeta, os.FileMode(0666))).To(Succeed())
				m.Fs = afero.NewOsFs()

				// initialise the git repos
				r1, err := git.PlainInit("one", false)
				Expect(err).NotTo(HaveOccurred())
				r2, err := git.PlainInit("two", false)
				Expect(err).NotTo(HaveOccurred())

				// write package.json files
				Expect(ioutil.WriteFile("one/package.json", one, os.FileMode(0666))).To(Succeed())
				Expect(ioutil.WriteFile("two/package.json", two, os.FileMode(0666))).To(Succeed())

				// commit the package.json files to master
				wt1, err := r1.Worktree()
				Expect(err).NotTo(HaveOccurred())
				_, err = wt1.Add("package.json")
				Expect(err).NotTo(HaveOccurred())

				_, err = wt1.Commit("package.json commit", &git.CommitOptions{
					Author: &object.Signature{
						Name:  "John Doe",
						Email: "john@doe.org",
						When:  time.Now(),
					},
				})
				Expect(err).NotTo(HaveOccurred())

				wt2, err := r2.Worktree()
				Expect(err).NotTo(HaveOccurred())
				_, err = wt2.Add("package.json")
				Expect(err).NotTo(HaveOccurred())

				_, err = wt2.Commit("package.json commit", &git.CommitOptions{
					Author: &object.Signature{
						Name:  "John Doe",
						Email: "john@doe.org",
						When:  time.Now(),
					},
				})
				Expect(err).NotTo(HaveOccurred())

				// checkout a new story branch on both projects
				ref := plumbing.ReferenceName("refs/heads/some-story")

				h1, err := r1.Head()
				Expect(err).NotTo(HaveOccurred())

				err = wt1.Checkout(&git.CheckoutOptions{
					Branch: ref,
					Hash:   h1.Hash(),
					Create: true,
				})
				Expect(err).NotTo(HaveOccurred())

				h2, err := r2.Head()
				Expect(err).NotTo(HaveOccurred())

				err = wt2.Checkout(&git.CheckoutOptions{
					Branch: ref,
					Hash:   h2.Hash(),
					Create: true,
				})
				Expect(err).NotTo(HaveOccurred())

				// set a new story and add project two
				Expect(m.Load(".meta")).To(Succeed())
				Expect(m.SetStory("some-story")).To(Succeed())
				s := meta.Manifest{Fs: m.Fs}
				Expect(s.Load(".meta")).To(Succeed())
				Expect(s.AddProjects([]string{"two"})).To(Succeed())

				// commit new package.json to story branch
				_, err = wt2.Add("package.json")
				Expect(err).NotTo(HaveOccurred())

				_, err = wt2.Commit("package.json story commit", &git.CommitOptions{
					Author: &object.Signature{
						Name:  "John Doe",
						Email: "john@doe.org",
						When:  time.Now(),
					},
				})
				Expect(err).NotTo(HaveOccurred())

				// run the pruner
				Expect(s.Prune()).To(Succeed())

				// expect that the projects are removed from the story
				Expect(s.Projects).NotTo(HaveKey("one"))
				Expect(s.Projects).To(HaveKey("two"))

				// expect that the package.json has been reverted
				bytes, err := ioutil.ReadFile("two/package.json")
				Expect(err).NotTo(HaveOccurred())

				p := &node.PackageJSON{}
				Expect(json.Unmarshal(bytes, p)).To(Succeed())
				Expect(p.Dependencies).To(HaveKeyWithValue("one", "git+ssh://git@github.com:TestOrg/one.git"))

				Expect(os.RemoveAll("one")).To(Succeed())
				Expect(os.RemoveAll("two")).To(Succeed())
				Expect(os.RemoveAll(".meta")).To(Succeed())
				Expect(os.RemoveAll(".meta.json")).To(Succeed())
			})
		})

		Context("With projects that point to the same commit on the story branch and master", func() {
			It("Should prune those projects from the meta file and reset any package.json changes", func() {
				// L O N G  A S S  S E T U P

				// TODO: Make this all work regardless of which directory the test command is called from
				// TODO: Needs to be called from within the meta directory for now

				// initial global meta
				Expect(ioutil.WriteFile(".meta", globalMeta, os.FileMode(0666))).To(Succeed())
				m.Fs = afero.NewOsFs()

				// initialise a git repo
				repository, err := git.PlainInit("two", false)
				Expect(err).NotTo(HaveOccurred())

				// write a package.json for project two
				Expect(ioutil.WriteFile("two/package.json", two, os.FileMode(0666))).To(Succeed())

				// commit the package.json to master
				wt, err := repository.Worktree()
				Expect(err).NotTo(HaveOccurred())
				_, err = wt.Add("package.json")
				Expect(err).NotTo(HaveOccurred())

				_, err = wt.Commit("package.json commit", &git.CommitOptions{
					Author: &object.Signature{
						Name:  "John Doe",
						Email: "john@doe.org",
						When:  time.Now(),
					},
				})
				Expect(err).NotTo(HaveOccurred())

				// checkout a new story branch
				head, err := repository.Head()
				Expect(err).NotTo(HaveOccurred())

				ref := plumbing.ReferenceName("refs/heads/some-story")

				err = wt.Checkout(&git.CheckoutOptions{
					Branch: ref,
					Hash:   head.Hash(),
					Create: true,
				})
				Expect(err).NotTo(HaveOccurred())

				// set a new story and add project two
				Expect(m.Load(".meta")).To(Succeed())
				Expect(m.SetStory("some-story")).To(Succeed())
				s := meta.Manifest{Fs: m.Fs}
				Expect(s.Load(".meta")).To(Succeed())
				Expect(s.AddProjects([]string{"two"})).To(Succeed())

				// run the pruner
				Expect(s.Prune()).To(Succeed())

				// expect that the projects are removed from the story
				Expect(s.Projects).NotTo(HaveKey("one"))
				Expect(s.Projects).NotTo(HaveKey("two"))

				// expect that the package.json has been reverted
				bytes, err := ioutil.ReadFile("two/package.json")
				Expect(err).NotTo(HaveOccurred())

				p := &node.PackageJSON{}
				Expect(json.Unmarshal(bytes, p)).To(Succeed())
				Expect(p.Dependencies).To(HaveKeyWithValue("one", "git+ssh://git@github.com:TestOrg/one.git"))

				Expect(os.RemoveAll("one")).To(Succeed())
				Expect(os.RemoveAll("two")).To(Succeed())
				Expect(os.RemoveAll(".meta")).To(Succeed())
				Expect(os.RemoveAll(".meta.json")).To(Succeed())
			})
		})
	})
})
