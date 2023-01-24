// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestRelocateSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, env.KbldBinaryPath, Logger{}}

	// redis:5.0.4
	input := `
kind: Object
spec:
# references image index
- image: index.docker.io/library/redis@sha256:000339fb57e0ddf2d48d72f3341e47a8ca3b1beae9bdcb25a96323095b72a79b
# references plain image
- image: gcr.io/cloud-builders/gcs-fetcher@sha256:055519529bf1ba12bf916fa42d6d3f68bdc581413621c269425bb0fee2467a93
`

	out, _ := kbld.RunWithOpts([]string{"relocate", "-f", "-", "--repository", env.WithRegistries("docker.io/*username*/kbld-test-relocate")}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := env.WithRegistries(`---
kind: Object
spec:
- image: index.docker.io/*username*/kbld-test-relocate@sha256:000339fb57e0ddf2d48d72f3341e47a8ca3b1beae9bdcb25a96323095b72a79b
- image: index.docker.io/*username*/kbld-test-relocate@sha256:055519529bf1ba12bf916fa42d6d3f68bdc581413621c269425bb0fee2467a93
`)

	if out != expectedOut {
		t.Fatalf("Expected unpackage output >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}

func TestRelocateSuccessfulWithManyImages(t *testing.T) {
	env := BuildEnv(t)

	if env.SkipStressTests {
		fmt.Printf("This is a stress test; skipping.")
		return
	}

	kbld := Kbld{t, env.Namespace, env.KbldBinaryPath, Logger{}}

	input := `
kind: Object
spec:
- image: index.docker.io/cfidentity/uaa@sha256:9f1e7e399c96309935145624d1824b2c2bf93656fd9c4dcf1c593b55f98aa6a8
- image: index.docker.io/cloudfoundry/capi-kpack-watcher@sha256:67125e0d3a4026a23342d80e09aad9284c08ab4f7b3d9a993ae66e403d5d0796
- image: index.docker.io/cloudfoundry/capi@sha256:51e4e48c457d5cb922cf0f569e145054e557e214afa78fb2b312a39bb2f938b6
- image: index.docker.io/cloudfoundry/cloud-controller-ng@sha256:374f967edd7db4d7efc2f38cb849988aa36a8248dd240d56f49484b8159fd800
- image: index.docker.io/cloudfoundry/cnb@sha256:5b03a853e636b78c44e475bbc514e2b7b140cc41cca8ab907e9753431ae8c0b0
- image: index.docker.io/istio/citadel@sha256:420a331a528886aca47bed5b8c549c78d594e52d771f876f3137d3557207712f
- image: index.docker.io/istio/galley@sha256:26e744bdfd3db289d4cfc9be63e38e7c7a424a9f76d1224cbdbbe58374229b68
- image: index.docker.io/istio/mixer@sha256:ff6f39732c31999911790b00b484a471b6fe87192223d3266b0b6e752a374287
- image: index.docker.io/istio/node-agent-k8s@sha256:7e17ab509777a54f3c0dfb4518692a9ca179d1e8c41df87dc81a734339b37152
- image: index.docker.io/istio/pilot@sha256:2bca5900d6bf20d5f0bf2b6673bc4d2885bab8cca2a9060336a0024930665b59
- image: index.docker.io/istio/proxyv2@sha256:fc09ea0f969147a4843a564c5b677fbf3a6f94b56627d00b313b4c30d5fef094
- image: index.docker.io/istio/sidecar_injector@sha256:ba446f8cf98bafdad4514fd492432dd180243cbc55a0b9c6bebfe31cb169033d
- image: index.docker.io/eirini/opi@sha256:2e0b84c5fcb1e6e5cdb07a70210f2e462aa52119f7a330660a7444a938deefbb
- image: gcr.io/cf-build-service-public/kpack/controller@sha256:1d7d80257e2019a474417ba0c7dcfff5612aeec55e24d91ef7b2e4bd0a521a40
- image: gcr.io/cf-build-service-public/kpack/webhook@sha256:c2461ef9634c771f2a06bc0371040b43c9a78dd0e4ac1c9fde3f4525e0ae21f2
- image: index.docker.io/bitnami/postgresql@sha256:9762d9a80b90a5efe299d4848057ac5c45fb384570b36f60aad38fe2b1704bd6
- image: index.docker.io/metacontroller/metacontroller@sha256:ad85cb5f5ad9a61a3f38277fed371df43ea0fc55d9073dfa8f4fc2e27c127603
- image: index.docker.io/minio/minio@sha256:5e96d539583afd9a7da14e0d9bf2360d316e4e8219659d82b8ef106a9d75b16c
- image: index.docker.io/cloudfoundry/cf-k8s-logging@sha256:d8c73e6c87b2a71c8b6798205761bb7870fb2080a4329c4eefb0b4620656eeaa
- image: index.docker.io/logcache/log-cache-cf-auth-proxy@sha256:6a436c864e5e6d2e153da4776f08fa064021eb365407f5f435a4a9f47afdef3d
- image: index.docker.io/logcache/log-cache-gateway@sha256:65b34fb624b40a263b6d1be9410ba61d55d515ac340226860d8fd7ef4ac0dbf1
- image: index.docker.io/logcache/log-cache@sha256:20ffd743bd6b52ff217918b2df3df4886969877feb5565b47248bfefe7b2b210
- image: index.docker.io/cloudfoundry/syslog-server@sha256:39a386521f94c70071eab4a7fb67cc7e28adba2e2dd8113d6df155c17b19f5a5
- image: index.docker.io/oratos/metric-proxy@sha256:a2a0d2201d1a57602a3db337bfa256d6e042dfc09a63ba1b6f39c952847e00dc
- image: index.docker.io/prom/statsd-exporter@sha256:e3174186628b401e4a441b78513ba06e957644267332436be0c77dd7af9bdddc
`

	out, _ := kbld.RunWithOpts([]string{"relocate", "-f", "-", "--repository", env.WithRegistries("docker.io/*username*/kbld-test-relocate-successful-with-many-images"), "--concurrency=8"}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := env.WithRegistries(`---
kind: Object
spec:
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:9f1e7e399c96309935145624d1824b2c2bf93656fd9c4dcf1c593b55f98aa6a8
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:67125e0d3a4026a23342d80e09aad9284c08ab4f7b3d9a993ae66e403d5d0796
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:51e4e48c457d5cb922cf0f569e145054e557e214afa78fb2b312a39bb2f938b6
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:374f967edd7db4d7efc2f38cb849988aa36a8248dd240d56f49484b8159fd800
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:5b03a853e636b78c44e475bbc514e2b7b140cc41cca8ab907e9753431ae8c0b0
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:420a331a528886aca47bed5b8c549c78d594e52d771f876f3137d3557207712f
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:26e744bdfd3db289d4cfc9be63e38e7c7a424a9f76d1224cbdbbe58374229b68
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:ff6f39732c31999911790b00b484a471b6fe87192223d3266b0b6e752a374287
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:7e17ab509777a54f3c0dfb4518692a9ca179d1e8c41df87dc81a734339b37152
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:2bca5900d6bf20d5f0bf2b6673bc4d2885bab8cca2a9060336a0024930665b59
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:fc09ea0f969147a4843a564c5b677fbf3a6f94b56627d00b313b4c30d5fef094
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:ba446f8cf98bafdad4514fd492432dd180243cbc55a0b9c6bebfe31cb169033d
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:2e0b84c5fcb1e6e5cdb07a70210f2e462aa52119f7a330660a7444a938deefbb
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:1d7d80257e2019a474417ba0c7dcfff5612aeec55e24d91ef7b2e4bd0a521a40
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:c2461ef9634c771f2a06bc0371040b43c9a78dd0e4ac1c9fde3f4525e0ae21f2
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:de15fb9d960ecd36082550ec7a24dcad12a63f7ca53d110fbb4c9c59eca9b544
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:9762d9a80b90a5efe299d4848057ac5c45fb384570b36f60aad38fe2b1704bd6
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:ad85cb5f5ad9a61a3f38277fed371df43ea0fc55d9073dfa8f4fc2e27c127603
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:5e96d539583afd9a7da14e0d9bf2360d316e4e8219659d82b8ef106a9d75b16c
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:7ee4b08db98627c76091b8b82210eaa9e3a0646db56b534feba7dd0a35c35948
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:6a436c864e5e6d2e153da4776f08fa064021eb365407f5f435a4a9f47afdef3d
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:65b34fb624b40a263b6d1be9410ba61d55d515ac340226860d8fd7ef4ac0dbf1
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:20ffd743bd6b52ff217918b2df3df4886969877feb5565b47248bfefe7b2b210
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:623ade911f957d38b36944d9f069feff3d4139ece24557a94d09b28f2efbe3d8
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:a2a0d2201d1a57602a3db337bfa256d6e042dfc09a63ba1b6f39c952847e00dc
- image: index.docker.io/*username*/kbld-test-relocate-successful-with-many-images@sha256:e3174186628b401e4a441b78513ba06e957644267332436be0c77dd7af9bdddc
`)
	if out != expectedOut {
		t.Fatalf("Expected >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}
}

func TestRelocateLockSuccessful(t *testing.T) {
	env := BuildEnv(t)
	kbld := Kbld{t, env.Namespace, env.KbldBinaryPath, Logger{}}

	input := `
apiVersion: kbld.k14s.io/v1alpha1
kind: Config
overrides:
# ignored because it's not preresolved
- image: gcs-fetcher
  newImage: gcr.io/cloud-builders/gcs-fetcher@sha256:055519529bf1ba12bf916fa42d6d3f68bdc581413621c269425bb0fee2467a93
- image: redis
  newImage: index.docker.io/library/redis@sha256:000339fb57e0ddf2d48d72f3341e47a8ca3b1beae9bdcb25a96323095b72a79b
  preresolved: true
`

	path := "/tmp/kbld-test-pkg-unpkg-with-lock-successful"
	defer os.RemoveAll(path)

	relocatedLockPath := path + "-relocated"
	defer os.RemoveAll(relocatedLockPath)

	out, _ := kbld.RunWithOpts([]string{"relocate", "-f", "-", "--repository", env.WithRegistries("docker.io/*username*/kbld-test-relocate-lock"), "--lock-output", relocatedLockPath}, RunOpts{
		StdinReader: strings.NewReader(input),
	})

	expectedOut := ""

	if out != expectedOut {
		t.Fatalf("Expected package output >>>%s<<< to match >>>%s<<<", out, expectedOut)
	}

	lockOutBs, err := ioutil.ReadFile(relocatedLockPath)
	if err != nil {
		t.Fatalf("Expected to find relocated lock file")
	}

	kbldVersionOutput, _ := kbld.RunWithOpts([]string{"version"}, RunOpts{})
	kbldVersion := strings.SplitAfter(kbldVersionOutput, " ")[2]

	expectedLockOut := strings.ReplaceAll(env.WithRegistries(`apiVersion: kbld.k14s.io/v1alpha1
kind: Config
minimumRequiredVersion: __ver__
overrides:
- image: redis
  newImage: index.docker.io/*username*/kbld-test-relocate-lock@sha256:000339fb57e0ddf2d48d72f3341e47a8ca3b1beae9bdcb25a96323095b72a79b
  preresolved: true
- image: index.docker.io/library/redis@sha256:000339fb57e0ddf2d48d72f3341e47a8ca3b1beae9bdcb25a96323095b72a79b
  newImage: index.docker.io/*username*/kbld-test-relocate-lock@sha256:000339fb57e0ddf2d48d72f3341e47a8ca3b1beae9bdcb25a96323095b72a79b
  preresolved: true
`), "__ver__", strings.TrimSuffix(kbldVersion, "\n"))

	if string(lockOutBs) != expectedLockOut {
		t.Fatalf("Expected unpackage lock output >>>%s<<< to match >>>%s<<<", lockOutBs, expectedLockOut)
	}
}
