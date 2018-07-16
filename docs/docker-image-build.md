# Build a `ks` Docker image

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

In order to run via docker, the `ks` process needs access to a kubernetes config file, certificates, and the current working directory. Pass in the `--mount` flag to capture the path of these files. Here's what it looks like running on a local cluster with minikube:

```bash
docker run -e KUBECONFIG="$HOME/.kube/config" --mount type=bind,source="$HOME/.kube/config",target="$HOME/.kube/config" --mount type=bind,source="$HOME/.minikube",target="$HOME/.minikube" --mount type=bind,source="/tmp",target="/tmp" --mount type=bind,source="$(pwd)",target="$(pwd)" --network host -w "$(pwd)" ks:v0.12.0 --help
```

This sets the $KUBECONFIG environment variable inside the container, mounts the config and the directory holding certificates (which can be found inside the kubeconfig), and mounts the current working directory so that the `ks` binary knows where to work.

Optionally, set an alias to shorten the verbose command if using the docker image locally.

```bash
alias ks='docker run -e KUBECONFIG=/path/to/kube/config --mount type=bind,source=/path/to/kube/config,target=/path/to/kube/config --mount type=bind,source=/path/to/cert/dir,target=/path/to/cert/dir --mount type=bind,source="/tmp",target="/tmp" --mount type=bind,source="$(pwd)",target="$(pwd)" --network host -w "$(pwd)" ks:v0.12.0'
```
