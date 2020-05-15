package main_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	main "github.com/ForestEckhardt/clear-source"
	"github.com/cloudfoundry/packit"
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
		Expect(os.MkdirAll(filepath.Join(workingDir, "some-dir"), os.ModePerm)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(workingDir, "some-dir", "some-file"), []byte(`some-contents`), os.ModePerm)).To(Succeed())

		build = main.Build()
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

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

		Expect(filepath.Join(workingDir, "some-file")).ToNot(BeAnExistingFile())
		Expect(filepath.Join(workingDir, "some-dir")).ToNot(BeADirectory())
	})

	context("failure cases", func() {
		context("the working dir cannot be read", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
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

		context("the subdir cannot be read", func() {
			it.Before(func() {
				Expect(os.Chmod(filepath.Join(workingDir, "some-dir"), 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(filepath.Join(workingDir, "some-dir"), os.ModePerm)).To(Succeed())
			})

			it("returns a result that installs node", func() {
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
