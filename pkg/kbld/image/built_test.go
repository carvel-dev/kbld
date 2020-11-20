package image_test

import (
	"testing"

	"github.com/sclevine/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ctlbdk "github.com/k14s/kbld/pkg/kbld/builder/docker"
	ctlconf "github.com/k14s/kbld/pkg/kbld/config"
	"github.com/k14s/kbld/pkg/kbld/image"
	"github.com/k14s/kbld/pkg/kbld/image/imagefakes"
)

func TestBuiltImage(t *testing.T) {
	spec.Run(t, "testBuiltImage", testBuiltImage)
}

func testBuiltImage(t *testing.T, when spec.G, it spec.S) {
	var (
		fakeDocker    = &imagefakes.FakeDockerBuild{}
		fakePack      = &imagefakes.FakePackBuild{}
		fakeKBuildKit = &imagefakes.FakeKBuildKitBuild{}
	)

	when("when source contains `pack` build options", func() {
		var (
			builder    = "index.docker.io/some/builder"
			clearCache = true
			source     = ctlconf.Source{
				ImageRef: ctlconf.ImageRef{
					Image:     "some.place/image",
					ImageRepo: "some.repo",
				},
				Path: "/tmp/",
				Pack: &ctlconf.SourcePackOpts{
					Build: ctlconf.SourcePackBuildOpts{
						Builder:    &builder,
						Buildpacks: nil,
						ClearCache: &clearCache,
						RawOptions: nil,
					},
				},
			}
		)

		it("uses pack to build and pushes the resulting image, when output is an image", func() {
			fakeDocker.PushReturns(ctlbdk.NewDockerImageDigest("sha256:421c76d77563afa1914846b010bd164f395bd34c2102e5e99e0cb9cf173c1d87"), nil)
			fakePack.BuildReturns(ctlbdk.NewDockerTmpRef("localhost:5000/some-tmp-location"), nil)
			subject := image.NewBuiltImage(
				"testing.place.url",
				source,
				&ctlconf.ImageDestination{
					NewImage: "some.place/new-img",
				},
				fakeDocker,
				fakePack,
				fakeKBuildKit,
			)
			url, metas, err := subject.URL()
			require.NoError(t, err)
			assert.Equal(t, "some.place/new-img@sha256:421c76d77563afa1914846b010bd164f395bd34c2102e5e99e0cb9cf173c1d87", url)
			assert.Equal(t, []image.ImageMeta{image.BuiltImageSourceLocal{Type: "local", Path: "/tmp"}}, metas)
		})

		it("uses pack to build, when output is not an image", func() {
			fakePack.BuildReturns(ctlbdk.NewDockerTmpRef("localhost:5000/some-tmp-location"), nil)
			subject := image.NewBuiltImage(
				"testing.place.url",
				source,
				nil,
				fakeDocker,
				fakePack,
				fakeKBuildKit,
			)
			url, metas, err := subject.URL()
			require.NoError(t, err)
			assert.Equal(t, "localhost:5000/some-tmp-location", url)
			assert.Equal(t, []image.ImageMeta{image.BuiltImageSourceLocal{Type: "local", Path: "/tmp"}}, metas)
		})
	})
}
