package logic_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/source-removal/logic"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testInclude(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string
	)

	it.Before(func() {
		var err error
		workingDir, err = os.MkdirTemp("", "working-dir")
		Expect(err).NotTo(HaveOccurred())

		Expect(os.WriteFile(filepath.Join(workingDir, "some-file"), []byte(`some-contents`), os.ModePerm)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(workingDir, "some-dir", "some-other-dir", "another-dir"), os.ModePerm)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workingDir, "some-dir", "some-file"), []byte(`some-contents`), os.ModePerm)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workingDir, "some-dir", "some-other-dir", "some-file"), []byte(`some-contents`), os.ModePerm)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(workingDir, "some-dir", "some-other-dir", "another-dir", "some-file"), []byte(`some-contents`), os.ModePerm)).To(Succeed())
	})

	it.After(func() {
		Expect(os.RemoveAll(workingDir)).To(Succeed())
	})

	context("Include", func() {
		it("returns a result that deletes the contents of the working directroy except for the file that are meant to kept", func() {
			err := logic.Include(workingDir, "some-dir/some-other-dir/*:some-file")
			Expect(err).NotTo(HaveOccurred())

			Expect(workingDir).To(BeADirectory())
			Expect(filepath.Join(workingDir, "some-file")).To(BeAnExistingFile())
			Expect(filepath.Join(workingDir, "some-dir")).To(BeADirectory())
			Expect(filepath.Join(workingDir, "some-dir", "some-file")).NotTo(BeAnExistingFile())
			Expect(filepath.Join(workingDir, "some-dir", "some-other-dir", "some-file")).To(BeAnExistingFile())
			Expect(filepath.Join(workingDir, "some-dir", "some-other-dir", "another-dir", "some-file")).To(BeAnExistingFile())
		})
	})

	context("failure cases", func() {
		context("when the directory cannot be walked", func() {
			it.Before(func() {
				Expect(os.Chmod(workingDir, 0000)).To(Succeed())
			})

			it.After(func() {
				Expect(os.Chmod(workingDir, os.ModePerm)).To(Succeed())
			})

			it("returns an error", func() {
				err := logic.Include(workingDir, "some-dir/some-other-dir/*:some-file")
				Expect(err).To(MatchError(ContainSubstring("permission denied")))
			})
		})

		context("when there is a malformed glob in include", func() {
			it("returns an error", func() {
				err := logic.Include(workingDir, `\`)
				Expect(err).To(MatchError(ContainSubstring("syntax error in pattern")))
			})
		})
	})
}
