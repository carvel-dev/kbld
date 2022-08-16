![logo](docs/CarvelLogo.png)

# kbld

- Website: https://carvel.dev/kbld
- Slack: [#carvel in Kubernetes slack](https://slack.kubernetes.io)
- [Docs](docs/README.md) with topics about building, packaging images, blog posts, etc.
- Install: Grab prebuilt binaries from the [Releases page](https://github.com/vmware-tanzu/carvel-kbld/releases) or [Homebrew Carvel tap](https://github.com/vmware-tanzu/homebrew-carvel)
- Backlog: [See what we're up to](https://github.com/orgs/vmware-tanzu/projects/16/views/1?filterQuery=repo%3A%22vmware-tanzu%2Fcarvel-kbld%22).

`kbld` (pronounced: `kei·bild`) seamlessly incorporates image building and image pushing into your development and deployment workflows.

Features:

- Orchestrates image builds (delegates to tools like Docker, pack, kubectl-buildkit) and registry pushes
- Works with local Docker daemon and remote registries, for development and production cases
- Records metadata about image sources in annotation on Kubernetes resources (see examples below)
- Resolves image references to their digest form (*immutable*) ([details](https://carvel.dev/kbld/docs/latest/#why-digest-references))
- Provides a way to transport set of images in a single tarball between registries
  - maintaining exactly same digests, hence guaranteeing exactly same content
- Not specific to Kubernetes, but works really well with Kubernetes configuration files  

![](docs/kbld-screenshot.png)

See [building and deploying simple Go application to Kubernetes example](https://github.com/vmware-tanzu/carvel-simple-app-on-kubernetes#step-3-building-container-images-locally) that uses kbld.

### Join the Community and Make Carvel Better
Carvel is better because of our contributors and maintainers. It is because of you that we can bring great software to the community.
Please join us during our online community meetings. Details can be found on our [Carvel website](https://carvel.dev/community/).

You can chat with us on Kubernetes Slack in the #carvel channel and follow us on Twitter at @carvel_dev.

Check out which organizations are using and contributing to Carvel: [Adopter's list](https://github.com/vmware-tanzu/carvel/blob/master/ADOPTERS.md)

# Development

Consult [docs/dev.md](docs/dev.md) for build instructions, code structure details.
