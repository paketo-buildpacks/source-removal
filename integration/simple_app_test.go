package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/occam"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testSimpleApp(t *testing.T, context spec.G, it spec.S) {
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

	context("pushing a simple app where you want the working dir to be removed", func() {
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

		it("builds a working OCI image for a simple app that has an empty working dir", func() {
			var err error
			image, _, err = pack.Build.
				WithBuildpacks(goURI, goModURI, nosourceURI).
				Execute(name, filepath.Join("testdata", "simple_app"))
			Expect(err).NotTo(HaveOccurred())

			container, err = docker.Container.Run.WithCommand(`echo "hello world" && ls -a /workspace`).Execute(image.ID)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() string {
				logs, _ := docker.Container.Logs.Execute(container.ID)
				return logs.String()
			}, "5s").Should(ContainSubstring("hello world"))

			logs, err := docker.Container.Logs.Execute(container.ID)
			Expect(err).NotTo(HaveOccurred())
			Expect(logs.String()).NotTo(ContainSubstring("some-file"))
			Expect(logs.String()).To(ContainSubstring(".."))
		})
	})
}
