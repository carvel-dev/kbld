# How to Contribute

We’d love to accept your patches and contributions to this project. Please review the following guidelines you'll need to follow in order to make a contribution.

# Communication

We prefer communicating asynchronously through GitHub issues and the [#carvel Slack channel](https://kubernetes.slack.com/archives/CH8KCCKA5). In order to be inclusive to the community, if a conversation related to an issue happens outside of these channels, we appreciate summarizing the conversation's context and adding it to an issue.

# Propose a Change

Pull requests are welcome for all changes. When adding new functionality, we encourage including test coverage. If significant effort will be involved, we suggest beginning by submitting an issue so any high level feedback can be addressed early.

Please submit feature requests and bug reports by using GitHub issues.

Before submitting an issue, please search through open ones to ensure others have not submitted something similar. If a similar issue exists, please add any additional information as a comment.

## Issues Lifecycle

Once an issue is labeled with `in-progress`, a team member has begun investigating it. We keep `in-progress` issues open until they have been resolved and released. Once released, a comment containing release information will be posted in the issue's thread.

# Contributor License Agreement

All contributors to this project must have a signed Contributor License Agreement (**"CLA"**) on file with us. The CLA grants us the permissions we need to use and redistribute your contributions as part of the project; you or your employer retain the copyright to your contribution. Before a PR can pass all required checks, our CLA action will prompt you to accept the agreement. Head over to https://cla.pivotal.io/ to see your current agreement(s) on file or to sign a new one.

We generally only need you (or your employer) to sign our CLA once and once signed, you should be able to submit contributions to any VMware project.

Note: if you would like to submit an "_obvious fix_" for something like a typo, formatting issue or spelling mistake, you may not need to sign the CLA. Please see our information on [obvious fixes](https://cla.pivotal.io/about#obvious-fix) for more details.

# Development

## Prerequisites

- [minikube](https://minikube.sigs.k8s.io/docs/)
- [ytt](https://github.com/k14s/ytt)
- [pack 0.8.1](https://github.com/buildpacks/pack)

## Run Unit tests
```bash
# Run all tests
./hack/test.sh
# or run single test
./hack/test.sh -run TestLogger
```

## Run E2E tests against minikube registry
```bash
# Bootstrap k8s cluster and enable docker registry
# X.X.X.X must be replaced with your subnetmask of "minikube ip"
minikube start --driver=docker --insecure-registry=X.X.X.X/16
# Build kbld binary for testing
./hack/build.sh
# Make your env aware of the docker registry
eval $(minikube docker-env)
# Run all tests
./hack/test-all-minikube-local-registry.sh
# or run single test
./hack/test-all-minikube-local-registry.sh -run TestDockerBuildSuccessful
```

## Run E2E tests against private docker registry
```bash
# Bootstrap k8s cluster
minikube start
# Build kbld binary for testing
./hack/build.sh
# Make your env aware of the docker registry
eval $(minikube docker-env)
docker login ...
export KBLD_E2E_DOCKERHUB_USERNAME=...
# Run all tests
./hack/test-all.sh
```

## Website build
```bash
# include goog analytics in 'kbld website' command for https://get-kbld.io
# (goog analytics is _not_ included in release binaries)
BUILD_VALUES=./hack/build-values-get-kbld-io.yml ./hack/build.sh
```