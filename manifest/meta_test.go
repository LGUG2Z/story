package manifest_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"

	"github.com/LGUG2Z/story/manifest"
	"github.com/spf13/afero"
)

var _ = Describe("Meta", func() {
	Describe("Trying to load non-existent files", func() {
		It("Should throw an error", func() {
			fs := afero.NewMemMapFs()
			_, err := manifest.LoadMetaOnTrunk(fs)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Loading a meta manifest on the creation of a story", func() {
		It("Should unmarshal the .meta file into a Meta object", func() {
			// Given a meta file on the fs
			b := []byte(`{
  "artifacts": {
    "one": false
  },
  "organisation": "test-org",
  "projects": {
    "one": "git@github.com:test-org/one.git"
  }
}`)
			fs := afero.NewMemMapFs()
			Expect(afero.WriteFile(fs, ".meta", b, os.FileMode(0666))).To(Succeed())

			// When I load that .meta file while not on a branch
			_, err := manifest.LoadMetaOnTrunk(fs)

			// Then it gets unmarshaled successfully
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
