## Hipster Shop

Repo: [https://github.com/GoogleCloudPlatform/microservices-demo](https://github.com/GoogleCloudPlatform/microservices-demo)

```bash
$ git clone https://github.com/GoogleCloudPlatform/microservices-demo
$ kbld apply -R -f microservices-demo/kubernetes-manifests -f ../examples/gcp-microservices-demo/prebuilt.yml | kapp deploy -a md -f - -y
```

With your own application one would expect that `prebuilt.yml` configuration would be part of that repository (instead of similar to this example, where `prebuilt.yml` is in a separate repo).

Alternatively if you want to build images, use `build-local.yml` instead of `prebuilt.yml`.
