## Packaging images

kbld can be used to package all referenced container images into a single tarball so that they can be easily imported into a same or different registry. Use cases:

- packaging applications for fully offline environments with a private registry
- moving images from one registry to another
- backing up images

For example, to package referenced images into a single tarball:

1. First make sure all image references are in their digest form

    ```bash
    $ cat /tmp/manifest
    images:
    - image: nginx
    - image: haproxy

    $ kbld -f /tmp/manifest > /tmp/resolved-manifest
    resolve | final: haproxy -> index.docker.io/library/haproxy@sha256:6dae9c8674e2e5f418c3dd040041a05f6b490597315139c0bcacadf65a46cfd5
    resolve | final: nginx -> index.docker.io/library/nginx@sha256:e71b1bf4281f25533cf15e6e5f9be4dac74d2328152edf7ecde23abc54e16c1c

    $ cat /tmp/resolved-manifest
    images:
    - image: index.docker.io/library/nginx@sha256:e71b1bf4281f25533cf15e6e5f9be4dac74d2328152edf7ecde23abc54e16c1c
    - image: index.docker.io/library/haproxy@sha256:6dae9c8674e2e5f418c3dd040041a05f6b490597315139c0bcacadf65a46cfd5
    ```

1. Feed `/tmp/resolved-manifest` to `kbld pkg` command to download images and pack them into a single tarball `/tmp/packaged-images.tar`:

    ```bash
    $ kbld pkg -f /tmp/resolved-manifest --output /tmp/packaged-images.tar
    package | exporting 2 images...
    package | will export index.docker.io/library/nginx@sha256:e71b1bf4281f25533cf15e6e5f9be4dac74d2328152edf7ecde23abc54e16c1c
    package | will export index.docker.io/library/haproxy@sha256:6dae9c8674e2e5f418c3dd040041a05f6b490597315139c0bcacadf65a46cfd5
    package | exported 2 images

    $ ls -lah /tmp/packaged-images.tar
    -rw-r--r-- 1 root root 314M May  3 18:59 /tmp/packaged-images.tar
    ```

    Note: Depending on your internet connection this may be slow.

To import packaged images from a tarball:

1. Specify new repository location `docker.io/dkalinin/app1` and provide tarball:

    ```bash
    $ kbld unpkg -f /tmp/resolved-manifest --input /tmp/packaged-images.tar --repository docker.io/dkalinin/app1
    unpackage | importing 2 images...
    unpackage | importing index.docker.io/library/nginx@sha256:e71b1bf4281f25533cf15e6e5f9be4dac74d2328152edf7ecde23abc54e16c1c -> docker.io/dkalinin/app1@sha256:e71b1bf4281f25533cf15e6e5f9be4dac74d2328152edf7ecde23abc54e16c1c...
    unpackage | importing index.docker.io/library/haproxy@sha256:6dae9c8674e2e5f418c3dd040041a05f6b490597315139c0bcacadf65a46cfd5 -> docker.io/dkalinin/app1@sha256:6dae9c8674e2e5f418c3dd040041a05f6b490597315139c0bcacadf65a46cfd5...
    unpackage | imported 2 images
    images:
    - image: docker.io/dkalinin/app1@sha256:e71b1bf4281f25533cf15e6e5f9be4dac74d2328152edf7ecde23abc54e16c1c
    - image: docker.io/dkalinin/app1@sha256:6dae9c8674e2e5f418c3dd040041a05f6b490597315139c0bcacadf65a46cfd5
    ```

    Images will be imported under a single new repository `docker.io/dkalinin/app1`. **You are guaranteed that images are exactly same as they are referenced by the same digests in produced YAML configuration (though under a different repository name)**.

### Authentication

Even though `kbld pkg/unpkg` commands use registry APIs directly, by default they rely on credentials stored in `~/.docker/config.json` which are typically generated via `docker login` command.

### Authenticating to gcr.io

- Create service account with "Storage Admin" for push access
  - See [Permissions and Roles](https://cloud.google.com/container-registry/docs/access-control#permissions_and_roles)
- Download JSON service account key and place it somewhere on filesystem (e.g. `/tmp/key`)
  - See [Advanced authentication](https://cloud.google.com/container-registry/docs/advanced-authentication#json_key_file)
- Run `cat /tmp/key | docker login -u _json_key --password-stdin https://gcr.io` to authenticate
- Run `kbld unpkg -f /tmp/resolved-manifest --input /tmp/packaged-images.tar --repository gcr.io/{project-id}/app1` to import images (e.g. project id is `dkalinin`)

### Authenticating to AWS ECR

**Note**: AWS ECR does not support manifest list media types from Docker Registry v2 API. Manifest lists are used for images that are built against multiple architectures and platforms and are referenced through a single digest or tag. Most user-built images do not use manifest lists (as it's a single image); however, common Docker Hub library images are represented by manifest lists and will fail upon import into AWS ECR. You will see following error: `Writing image index: Retried 5 times: uploading manifest2: UNSUPPORTED: Invalid parameter at 'imageTag' failed to satisfy constraint: 'must satisfy regular expression '[a-zA-Z0-9-_.]+'`. Related [AWS feature request](https://forums.aws.amazon.com/thread.jspa?threadID=292294)).

- Create ECR repository
- Create IAM user with ECR policy that allows to read/write
  - See [Amazon ECR Policies](https://docs.aws.amazon.com/AmazonECR/latest/userguide/ecr_managed_policies.html)
- Run `aws configure` and specify access key ID, secret access key and region
  - To install on Ubuntu, run `apt-get install pip3` and `pip3 install awscli`
    - See [Installing the AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html)
- Run `eval $(aws ecr get-login --no-include-email)` to authenticate
  - See [get-login command](https://docs.aws.amazon.com/cli/latest/reference/ecr/get-login.html)
- Run `kbld unpkg -f /tmp/resolved-manifest --input /tmp/packaged-images.tar --repository {uri}` to import images (e.g. uri is `823869848626.dkr.ecr.us-east-1.amazonaws.com/k14s/kbld-test`)

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ecr:GetAuthorizationToken",
                "ecr:BatchCheckLayerAvailability",
                "ecr:GetDownloadUrlForLayer",
                "ecr:GetRepositoryPolicy",
                "ecr:DescribeRepositories",
                "ecr:ListImages",
                "ecr:DescribeImages",
                "ecr:BatchGetImage",
                "ecr:InitiateLayerUpload",
                "ecr:UploadLayerPart",
                "ecr:CompleteLayerUpload",
                "ecr:PutImage"
            ],
            "Resource": "*"
        }
    ]
}
```

### Notes

- Produced tarball does not have duplicate image layers, as they are named by their digest (see `tar tvf /tmp/packaged-images.tar`).
- If digest reference points to an image index, all children (images and other image indexes) will be included in the export. Saving only a portion of contents would of course change the digest.
- Only Docker v2 and OCI images and indexes are supported. Docker v1 format is not supported, hence, not all images out there could be exported and only registries supporting v2 format can be used for imports.
- Images that once were in different repositories are imported into the same repository to make it easier to manage them in bulk.
