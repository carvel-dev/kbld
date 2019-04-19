# kbld

`kbld`

1. optionally builds Docker images (by delegating to other tools such as Docker, kaniko, etc.)
1. resolves images to their *immutable* image references (digests)
1. and, outputs YAML configuration with *immutable* image references

... so that output can be used with other Kubernetes deployment tools.

For example, one could use [ytt](https://github.com/k14s/ytt) + kbld + [kapp](https://github.com/k14s/kapp) to deploy an application:

```bash
ytt template -R -f kubernetes-manifests/ | kbld apply -f- | kapp -y deploy -a app1 -f-
```

## Docs

- [Docs](docs/README.md)

## Install

Grab prebuilt binaries from the [Releases page](https://github.com/k14s/kbld/releases).

## Development

```bash
./hack/build.sh
./hack/test-all.sh
```

`build.sh` depends on [ytt](https://github.com/k14s/ytt).
