package integration_test

import (
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

			name string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
		})

		context("when you want everything in source to be removed", func() {
			it("builds a working OCI image for an app that has an empty working dir", func() {
				var err error
				image, _, err = pack.Build.
					WithBuildpacks(buildpack).
					Execute(name, filepath.Join("testdata", "remove_source"))
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

		context("when you want to perserve something in source", func() {
			it("builds a working OCI image for an app that has the files which were supposed to be perserved", func() {
				var err error
				image, _, err = pack.Build.
					WithBuildpacks(buildpack).
					Execute(name, filepath.Join("testdata", "perserve_source"))
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
	})
}
