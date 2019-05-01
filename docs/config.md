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
