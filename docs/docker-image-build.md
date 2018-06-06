# Build a `ks` docker image

The intended use case for building a docker images is to run `ks` from CI/CD. As a development tool it is recommended to install `ks` locally, though it is still possible to use the docker image for development as well.

Copy and paste the following commands:
```bash
# Clone the ksonnet repo into your GOPATH
go get github.com/ksonnet/ksonnet

# Build and install binary under shortname `ks` into $GOPATH/bin
cd "${GOPATH}/src/github.com/ksonnet/ksonnet"
make docker-image
```

# Running `ks` via docker

In order to run via docker, the `ks` process needs access to a kubernetes config file, certificates, and the current working directory. We can set these up using mounts. Here's what it looks like running on a local cluster with minikube:

```bash
docker run -e KUBECONFIG="~/.kube/config" --mount type=bind,source="~/.kube/conifg",target="~/.kube/config" --mount type=bind,source="~/.minikube",target="~/.minikube" --mount type=bind,source="$(pwd)",target="$(pwd)" -w "$(pwd)" ks --help
```

This sets the $KUBECONFIG environment variable inside the container, mounts the config and the directory holding certificates (which can be found inside the kubeconfig), and mounts the current working directory so that the `ks` binary knows where to work.

It is also possible to use aliases, though keep in mind that this sets the working directory statically to the directory of the app you are working on:

```bash
alias ks="docker run -e KUBECONFIG=/path/to/kube/config --mount type=bind,source=/path/to/kube/config,target=/path/to/kube/config --mount type=bind,source=/path/to/cert/dir,target=/path/to/cert/dir --mount type=bind,source=/path/to/app/dir,target=/path/to/app/dir -w /path/to/app/dir ks"
ks env list
```
