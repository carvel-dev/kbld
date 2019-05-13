## Resolving image references to digests

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

via

```bash
kbld -f file.yml
```

Few other variations

```bash
pbpaste | kbld -f-
kbld -f .
kbld -f file.yml -f config2.yml
```
