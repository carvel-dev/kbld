## End-to-end test

### Prerequisites

To run the end to end tests, you must have the following utilities installed on your test runner. The test script checks for the specific version referenced. While a package installer will install the utility, it may install a newer, or different version than the one specified. Be sure to install these versions.

- [pack 0.8.1](https://github.com/buildpacks/pack)
- [kubectl-buildkit 0.1.0](https://github.com/vmware-tanzu/buildkit-cli-for-kubectl/releases/tag/v0.1.0)
- [ko 0.8.0](https://github.com/google/ko/releases/tag/v0.8.0)
- [bazel 4.2.0](https://github.com/bazelbuild/bazel/releases/tag/4.2.0)

### Run End to End Tests

Set environment variables

```bash
export KBLD_E2E_DOCKERHUB_USERNAME=joeexample
```

To use a registry other than DockerHub, set the KBLD_E2E_DOCKERHUB_HOSTNAME variable. You can specify a local registry or use an anonymous service, such as ttl.sh. You will need to authenticate with non-anonymous registeries for the tests to pass.

```bash
# OPTIONAL
export KBLD_E2E_DOCKERHUB_HOSTNAME=ttl.sh
```

It is possible to specify the absolute path to the kbld binary that you wish to test with. This could be useful for testing a version installed by a package manager, or a previously built custom binary in non-standard location. This value defaults to the local binary.

```bash
# OPTIONAL
export KBLD_BINARY_PATH=/usr/local/bin/kbld
```

Run the end to end tests

```bash
$ ./hack/test-e2e.sh
$ ./hack/test-e2e.sh -run TestVersion
```

See `./test/e2e/env.go` for required environment variables for some tests.

