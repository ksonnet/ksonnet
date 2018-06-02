# Build a `ks` docker image

Copy and paste the following commands:
```bash
# Clone the ksonnet repo into your GOPATH
go get github.com/ksonnet/ksonnet

# Build and install binary under shortname `ks` into $GOPATH/bin
cd $GOPATH/src/github.com/ksonnet/ksonnet
dep ensure
make docker-image
```

# Running `ks` via docker

In order to run via docker, the `ks` process needs access to a kubernetes config file, certificates, and the current working directory. We set these up using bind mounts:

```bash
docker run -e KUBECONFIG="${KUBECONFIG}" --mount type=bind,source="${KUBECONFIG}",target="${KUBECONFIG}" --mount type=bind,source="${HOME}/.minikube",target="${HOME}/.minikube" --mount type=bind,source="$(pwd)",target="$(pwd)" -w "$(pwd)" ks --help
```
