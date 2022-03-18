package sourceremoval_test

import (
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
		layersDir, err = os.MkdirTemp("", "layers")
		Expect(err).NotTo(HaveOccurred())

		cnbDir, err = os.MkdirTemp("", "cnb")
		Expect(err).NotTo(HaveOccurred())

		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.WriteFile(filepath.Join(workingDir, "some-file"), []byte(`some-contents`), os.ModePerm)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(workingDir, "some-dir", "some-other-dir", "another-dir"), os.ModePerm)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workingDir, "some-dir", "some-file"), []byte(`some-contents`), os.ModePerm)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workingDir, "some-dir", "some-other-dir", "some-file"), []byte(`some-contents`), os.ModePerm)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workingDir, "some-dir", "some-other-dir", "another-dir", "some-file"), []byte(`some-contents`), os.ModePerm)).To(Succeed())

		build = sourceremoval.Build()
	})

	it.After(func() {
		Expect(os.RemoveAll(layersDir)).To(Succeed())
		Expect(os.RemoveAll(cnbDir)).To(Succeed())
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("when there are no files to keep or exclude", func() {
		it("returns a result that keeps the contents of the working directroy", func() {
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
			Expect(filepath.Join(workingDir, "some-dir", "some-file")).To(BeAnExistingFile())
			Expect(filepath.Join(workingDir, "some-dir", "some-other-dir", "some-file")).To(BeAnExistingFile())
			Expect(filepath.Join(workingDir, "some-dir", "some-other-dir", "another-dir", "some-file")).To(BeAnExistingFile())
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

	context("when there are files to exclude", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_EXCLUDE_FILES", `some-dir/some-other-dir/*:some-file`)).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_EXCLUDE_FILES")).To(Succeed())
		})

		it("returns a result that deletes the contents of the working directroy that were specified", func() {
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
			Expect(filepath.Join(workingDir, "some-file")).NotTo(BeAnExistingFile())
			Expect(filepath.Join(workingDir, "some-dir")).To(BeADirectory())
			Expect(filepath.Join(workingDir, "some-dir", "some-file")).To(BeAnExistingFile())
			Expect(filepath.Join(workingDir, "some-dir", "some-other-dir", "some-file")).NotTo(BeAnExistingFile())
			Expect(filepath.Join(workingDir, "some-dir", "some-other-dir", "another-dir", "some-file")).NotTo(BeAnExistingFile())
		})
	})

	context("when both BP_INCLUDE_FILES and BP_EXCLUDE_FILES are set", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_INCLUDE_FILES", `some-dir/*:some-file`)).To(Succeed())
			Expect(os.Setenv("BP_EXCLUDE_FILES", `some-file`)).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_INCLUDE_FILES")).To(Succeed())
			Expect(os.Unsetenv("BP_EXCLUDE_FILES")).To(Succeed())
		})

		it("runs include logic followed by exclude logic", func() {
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
			Expect(filepath.Join(workingDir, "some-file")).NotTo(BeAnExistingFile())
			Expect(filepath.Join(workingDir, "some-dir")).To(BeADirectory())
			Expect(filepath.Join(workingDir, "some-dir", "some-file")).To(BeAnExistingFile())
			Expect(filepath.Join(workingDir, "some-dir", "some-other-dir", "some-file")).To(BeAnExistingFile())
			Expect(filepath.Join(workingDir, "some-dir", "some-other-dir", "another-dir", "some-file")).To(BeAnExistingFile())
		})
	})

	context("failure cases", func() {
		context("when there is a malformed glob in include", func() {
			it.Before(func() {
				Expect(os.Setenv("BP_INCLUDE_FILES", `\`)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BP_INCLUDE_FILES")).To(Succeed())
			})

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

		context("when there is a malformed glob in exclude", func() {
			it.Before(func() {
				Expect(os.Setenv("BP_EXCLUDE_FILES", `\`)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BP_EXCLUDE_FILES")).To(Succeed())
			})

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
	})
}
