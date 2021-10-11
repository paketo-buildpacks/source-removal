package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDefault(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("pushing an app where you want source removal", func() {
		var (
			image     occam.Image
			container occam.Container

			name   string
			source string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())

			source, err = occam.Source(filepath.Join("testdata", "default"))
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		context("when you want everything in source to be removed", func() {
			it("builds a working OCI image for an app that has an empty working dir", func() {
				var err error
				image, _, err = pack.Build.
					WithBuildpacks(buildpack).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				container, err = docker.Container.Run.WithCommand(`ls -a /workspace && echo "hello world"`).Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() string {
					logs, _ := docker.Container.Logs.Execute(container.ID)
					return logs.String()
				}, "10s").Should(ContainSubstring("hello world"))

				logs, err := docker.Container.Logs.Execute(container.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(logs.String()).NotTo(ContainSubstring("some-file"))
				Expect(logs.String()).NotTo(ContainSubstring("other-file"))
				Expect(logs.String()).To(ContainSubstring(".."))
			})
		})

		context("when you want to include something in source", func() {
			it("builds a working OCI image for an app that contains the files which were supposed to be included", func() {
				// The .occam-key needs to be include in order to ensure that a unique
				// image is made to make the test thread safe
				var err error
				image, _, err = pack.Build.
					WithEnv(map[string]string{
						"BP_INCLUDE_FILES": "some-file:.occam-key",
					}).
					WithBuildpacks(buildpack).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				container, err = docker.Container.Run.WithCommand(`ls -a /workspace && echo "hello world"`).Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() string {
					logs, _ := docker.Container.Logs.Execute(container.ID)
					return logs.String()
				}, "10s").Should(ContainSubstring("hello world"))

				logs, err := docker.Container.Logs.Execute(container.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(logs.String()).To(ContainSubstring("some-file"))
				Expect(logs.String()).NotTo(ContainSubstring("other-file"))
				Expect(logs.String()).To(ContainSubstring(".."))
			})
		})

		context("when you want to exclude something in source", func() {
			it("builds a working OCI image for an app that does not contain the files which were supposed to be exclude", func() {
				var err error
				image, _, err = pack.Build.
					WithEnv(map[string]string{
						"BP_EXCLUDE_FILES": "some-file",
					}).
					WithBuildpacks(buildpack).
					Execute(name, source)
				Expect(err).NotTo(HaveOccurred())

				container, err = docker.Container.Run.WithCommand(`ls -a /workspace && echo "hello world"`).Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(func() string {
					logs, _ := docker.Container.Logs.Execute(container.ID)
					return logs.String()
				}, "10s").Should(ContainSubstring("hello world"))

				logs, err := docker.Container.Logs.Execute(container.ID)
				Expect(err).NotTo(HaveOccurred())
				Expect(logs.String()).NotTo(ContainSubstring("some-file"))
				Expect(logs.String()).To(ContainSubstring("other-file"))
				Expect(logs.String()).To(ContainSubstring(".."))
			})
		})
	})
}
