## Docs

kbld looks for `image` keys within YAML documents and tries to resolve image reference to its full digest form.

For example, following

```yaml
kind: Object
spec:
- image: nginx:1.14.2
```

will be transformed to

```yaml
kind: Object
spec:
- image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
```

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
