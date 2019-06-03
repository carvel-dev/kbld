# kbld

- Website: https://get-kbld.io
- Slack: [#k14s in Kubernetes slack](https://slack.kubernetes.io)

`kbld` seamlessly incorporates image building and image pushing into your development and deployment workflows.

Features:

- Orchestrates image builds (delegates to tools like Docker) and registry pushes
- Works with local Docker daemon and remote registries, for development and production cases
- Records metadata about image sources in annotation on Kubernetes resources (see examples below)
- Resolves image references to their digest form (*immutable*) ([details](https://get-kbld.io/#why))
- Provides a way to transport set of images in a single tarball between registries
  - maintaining exactly same digests, hence guaranteeing exactly same content
- Not specific to Kubernetes, but works really well with Kubernetes configuration files  

![](docs/kbld-screenshot.png)

See [building and deploying simple Go application to Kubernetes example](https://github.com/k14s/k8s-simple-app-example#step-3-building-container-images-locally) that uses kbld.

## Docs

- [Resolving image references to digests](docs/resolving.md)
- [Building images from source](docs/building.md)
- [Packaging images for distribution](docs/packaging.md)
- [Configuration](docs/config.md)

## Install

Grab prebuilt binaries from the [Releases page](https://github.com/k14s/kbld/releases).

## Development

```bash
./hack/build.sh
./hack/test-all.sh

BUILD_VALUES=./hack/build-values-get-kbld-io.yml ./hack/build.sh # includes goog analytics
```

`build.sh` depends on [ytt](https://github.com/k14s/ytt).
