# kbld

- Website: https://get-kbld.io
- Slack: [#k14s in Kubernetes slack](https://slack.kubernetes.io)

`kbld` can

- resolve images to their *immutable* image references (digests)
- optionally build Docker images (by delegating to tools such as Docker)
- export set of images as a single tarball, and import into a different registry
  - maintaining exactly same digests, hence guaranteeing exactly same content

![](docs/kbld-screenshot.png)

Example of using [ytt](https://github.com/k14s/ytt) + kbld + [kapp](https://github.com/k14s/kapp) to deploy an application:

```bash
ytt template -f kubernetes-manifests/ | kbld -f- | kapp -y deploy -a app1 -f-
```

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
```

`build.sh` depends on [ytt](https://github.com/k14s/ytt).
