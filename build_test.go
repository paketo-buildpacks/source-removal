package sourceremoval_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	sourceremoval "github.com/ForestEckhardt/source-removal"
	"github.com/paketo-buildpacks/packit"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir  string
		cnbDir     string
		workingDir string

		build packit.BuildFunc
	)

	it.Before(func() {
		var err error
		layersDir, err = ioutil.TempDir("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = ioutil.TempDir("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = ioutil.TempDir("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-file"), []byte(`some-contents`), os.ModePerm)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(workingDir, "some-dir", "some-other-dir", "another-dir"), os.ModePerm)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-dir", "some-file"), []byte(`some-contents`), os.ModePerm)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-dir", "some-other-dir", "some-file"), []byte(`some-contents`), os.ModePerm)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-dir", "some-other-dir", "another-dir", "some-file"), []byte(`some-contents`), os.ModePerm)).To(Succeed())

		build = sourceremoval.Build()
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when there are no files to keep", func() {
		it("returns a result that deletes the contents of the working directroy", func() {
			result, err := build(packit.BuildContext{
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				WorkingDir: workingDir,
				Plan:       packit.BuildpackPlan{},
				Layers:     packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.BuildResult{}))

			Expect(workingDir).To(BeADirectory())
			Expect(filepath.Join(workingDir, "some-file")).ToNot(BeAnExistingFile())
			Expect(filepath.Join(workingDir, "some-dir")).ToNot(BeADirectory())
		})
	})

	context("when there are files to keep", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_INCLUDE_FILES", `some-dir/some-other-dir/*:some-file`)).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_INCLUDE_FILES")).To(Succeed())
		})

		it("returns a result that deletes the contents of the working directroy except for the file that are meant to kept", func() {
			result, err := build(packit.BuildContext{
				CNBPath:    cnbDir,
				Stack:      "some-stack",
				WorkingDir: workingDir,
				Plan:       packit.BuildpackPlan{},
				Layers:     packit.Layers{Path: layersDir},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(packit.BuildResult{}))

			Expect(workingDir).To(BeADirectory())
			Expect(filepath.Join(workingDir, "some-file")).To(BeAnExistingFile())
			Expect(filepath.Join(workingDir, "some-dir")).To(BeADirectory())
			Expect(filepath.Join(workingDir, "some-dir", "some-file")).NotTo(BeAnExistingFile())
			Expect(filepath.Join(workingDir, "some-dir", "some-other-dir", "some-file")).To(BeAnExistingFile())
			Expect(filepath.Join(workingDir, "some-dir", "some-other-dir", "another-dir", "some-file")).To(BeAnExistingFile())
		})
	})

	context("failure cases", func() {

		it.Before(func() {
			Expect(os.Setenv("BP_INCLUDE_FILES", `\`)).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_INCLUDE_FILES")).To(Succeed())
		})

		context("when there is a malformed glob in keep", func() {
			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					WorkingDir: workingDir,
					Plan:       packit.BuildpackPlan{},
					Layers:     packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError(ContainSubstring("syntax error in pattern")))
			})
		})

		context("when the directory cannot be removed", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0666)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				_, err := build(packit.BuildContext{
					CNBPath:    cnbDir,
					Stack:      "some-stack",
					WorkingDir: workingDir,
					Plan:       packit.BuildpackPlan{},
					Layers:     packit.Layers{Path: layersDir},
				})
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})
	})
}
