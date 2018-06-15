package manifest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/spf13/afero"
	"os"
	"github.com/LGUG2Z/story/manifest"
)

var _ = Describe("Meta", func() {
	Describe("Moving a meta manifest on the creation of a story", func() {
		It("Should move the .meta file to a .meta.json file", func() {
			// Given a meta file on the fs
			b := []byte(`{
  "deployables": {
    "one": false
  },
  "organisation": "test-org",
  "projects": {
    "one": "git@github.com:test-org/one.git"
  }
}`)
			fs := afero.NewMemMapFs()
			Expect(afero.WriteFile(fs, ".meta", b, os.FileMode(0666))).To(Succeed())

			// And that meta file unmarshalled into an object
			m, err := manifest.LoadMetaOnTrunk(fs)
			Expect(err).NotTo(HaveOccurred())

			// When I move the meta as part of creating a story
			Expect(m.MoveForStory(fs)).To(Succeed())

			// Then then no file file exists on the fs at .meta
			exists, err := afero.Exists(fs, ".meta")
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeFalse())

			// But a file exists on the fs .meta.json
			exists, err = afero.Exists(fs, ".meta.json")
			Expect(err).NotTo(HaveOccurred())
			Expect(exists).To(BeTrue())

			// And it contains the same content as the previous .meta file
			bytes, err := afero.ReadFile(fs, ".meta.json")
			Expect(err).NotTo(HaveOccurred())
			actual := string(bytes)
			Expect(actual).To(Equal(string(b)))
		})
	})

	Describe("Loading a meta manifest in a story branch", func() {
		It("Should load the file from the proper location and unmarshal it into an object", func() {
			// Given a .meta.json file on the fs
			b := []byte(`{
  "deployables": {
    "one": false
  },
  "organisation": "test-org",
  "projects": {
    "one": "git@github.com:test-org/one.git"
  }
}`)
			fs := afero.NewMemMapFs()
			Expect(afero.WriteFile(fs, ".meta.json", b, os.FileMode(0666))).To(Succeed())

			// When I load the file in a story branch
			_, err := manifest.LoadMetaOnBranch(fs)

			// It should unmarshal into an object without error
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
