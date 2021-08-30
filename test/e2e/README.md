# End to End Tests

This test suite verifies `kbld`'s ability to build, resolve, and push images using real components (e.g. an actual OCI registry rather than a simulated one).

Given that `kbld` integrates with image-building CLIs (e.g. Docker, `ko`) and Kubernetes Cluster-hosted image-building services (e.g. `kubectl-build`), both environments need to be configured for the full suite to run successfully.

## Pre-requisites

To run the test suite, install the following:
- Docker; https://docs.docker.com/get-docker/
- `kubectl`; https://kubernetes.io/docs/tasks/tools/
- `minikube`; https://minikube.sigs.k8s.io/docs/start/

Download (the very specific version of) the following build tools...

- `ko` (v0.8.0); https://github.com/google/ko/releases/tag/v0.8.0
- `kubectl-buildkit` (v0.1.0); https://github.com/vmware-tanzu/buildkit-cli-for-kubectl#installing-the-tarball
- `pack` (v0.8.1); https://github.com/buildpacks/pack/releases/tag/v0.8.1 (follow the "Download the .tgz..." instructions)

... and place a copy of each binary in `/usr/local/bin/`

## Start local Kubernetes Cluster with minikube

Some `kbld` builders use a Kubernetes cluster.

Startup a local cluster with minikube:

```bash
$ minikube start
```

## Running End-to-End Tests with DockerHub

1. Setup the E2E to use your DockerHub username in the suite:
    ```bash
    $ export KBLD_E2E_DOCKERHUB_USERNAME=(your docker username)
    ```

2. Ensure Docker is using same credentials
    ```bash
    $ docker logout; docker login
    Removing login credentials for https://index.docker.io/v1/
    Login with your Docker ID to push and pull images from Docker Hub. If you don't have a Docker ID, head over to https://hub.docker.com to create one.
    Username: (your docker username)
    Password: (your docker password)

    Login Succeeded
   ```

3. Configure buildkit to use the same credentials:
    ```bash
    $ kubectl create secret docker-registry buildkit --docker-server=https://index.docker.io/v1/ --docker-username=$KBLD_E2E_DOCKERHUB_USERNAME --docker-password="(your docker password)"
    ```

4. Run the test suite
    ```bash
    $ ./hack/test-e2e.sh
    ```

   DockerHub rate limits traffic (see https://docs.docker.com/docker-hub/download-rate-limit/). This test suite includes
   a number of test cases that involve many downloads. You can skip such tests by setting the KBLD_E2E_SKIP_STRESS_TESTS
   flag:

    ```bash
    $ KBLD_E2E_SKIP_STRESS_TESTS=true ./hack/test-e2e.sh
    ```

   Any additional arguments are passed along to the `go test ` invocation. For example, to run a specific (set of)
   test(s):

    ```bash
    $ ./hack/test-e2e.sh -run TestVersion
    ```

## Customizing End-to-End Tests

See the `Env` struct in `./test/e2e/env.go` for additional recognized environment variables.
