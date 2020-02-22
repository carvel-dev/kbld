## Resolving image references to digests

kbld looks for `image` keys within YAML documents and tries to resolve image reference to its full digest form.

For example, following

```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kbld-test1
spec:
  selector:
    matchLabels:
      app: kbld-test1
  template:
    metadata:
      labels:
        app: kbld-test1
    spec:
      containers:
      - name: my-app
        image: nginx:1.14.2
        #!      ^-- image reference in its tag form
```

will be transformed to

```yaml
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kbld-test1
spec:
  selector:
    matchLabels:
      app: kbld-test1
  template:
    metadata:
      labels:
        app: kbld-test1
    spec:
      containers:
      - name: my-app
        image: index.docker.io/library/nginx@sha256:f7988fb6c02e0ce69257d9bd9cf37ae20a60f1df7563c3a2a6abe24160306b8d
        #!      ^-- resolved image reference to its digest form
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
