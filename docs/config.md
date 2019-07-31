## Config

You can configure kbld by adding configuration resources (they follow Kubernetes resource format, but are removed from kbld output). Configuration resources may be specified multiple times.

### Sources

Sources resource configures kbld to execute image building operation based on specified path.

```yaml
---
apiVersion: kbld.k14s.io/v1alpha1
kind: Sources
sources:
- image: image1
  path: src/
```

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

- `docker.build.target`: Set the target build stage to build (no default)
- `docker.build.pull`: Always attempt to pull a newer version of the image (default is false)
- `docker.build.noCache`: Do not use cache when building the image (default is false)
- `docker.build.file`: Name of the Dockerfile (default is Dockerfile)
- `docker.build.rawOptions`: Refer to https://docs.docker.com/engine/reference/commandline/build/ for all available options

### ImageDestinations

ImageDestinations resource configures kbld to push built images to specified location.

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
