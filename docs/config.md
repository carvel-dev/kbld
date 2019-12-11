## Config

You can configure kbld by adding configuration resources (they follow Kubernetes resource format, but are removed from kbld output). Configuration resources may be specified multiple times.

### Sources

Sources resource configures kbld to execute image building operation based on specified path.

Two builders are currently supported: [Docker](https://docs.docker.com/engine/reference/commandline/cli/) (default) and [pack](https://github.com/buildpack/pack).

```yaml
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Sources
sources:
- image: image1
  path: src/
```

For Docker (all options shown):

```yaml
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Sources
sources:
- image: image1
  path: src/
  docker:
    build:
      target: "some-target"
      pull: true
      noCache: true
      file: "hack/Dockefile.dev"
      rawOptions: ["--squash"]
```

For pack:

```yaml
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Sources
sources:
- image: image1
  path: src/
  pack:
    build:
      builder: cloudfoundry/cnb:bionic
```

#### Docker

- `docker.build.target` (string): Set the target build stage to build (no default)
- `docker.build.pull` (bool): Always attempt to pull a newer version of the image (default is false)
- `docker.build.noCache` (bool): Do not use cache when building the image (default is false)
- `docker.build.file` (string): Name of the Dockerfile (default is Dockerfile)
- `docker.build.rawOptions` ([]string): Refer to https://docs.docker.com/engine/reference/commandline/build/ for all available options

#### Pack

- `pack.build.builder` (string): Set builder image (required)
- `pack.build.buildpacks` ([]string): Set list of buildpacks to be used (no default)
- `pack.build.clearCache` (bool): Clear cache before building image (default is false)
- `pack.build.rawOptions` ([]string): Refer to `pack build -h` for all available flags

### ImageDestinations

ImageDestinations resource configures kbld to push built images to specified location.

Currently images are pushed via Docker daemon for both Docker and pack built images (since pack also uses Docker daemon).

```yaml
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageDestinations
destinations:
- image: adservice
  newImage: docker.io/dkalinin/microservices-demo-adservice
```

### ImageOverrides

ImageOverrides resource configures kbld to rewrite image location before trying to build it or resolve it.

```yaml
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageOverrides
overrides:
- image: unknown
  newImage: docker.io/library/nginx:1.14.2
```

It can also hold `preresolved` new image, so no building or resolution happens:

```yaml
---
apiVersion: kbld.k14s.io/v1alpha1
kind: ImageOverrides
overrides:
- image: unknown
  newImage: docker.io/library/nginx:1.14.2
  preresolved: true
```

For preresolved images, kbld will not connect to registry to obtain any metadata.
