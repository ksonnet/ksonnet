# ksonnet

[![Build Status](https://travis-ci.org/ksonnet/ksonnet.svg?branch=master)](https://travis-ci.org/ksonnet/ksonnet)
[![Go Report Card](https://goreportcard.com/badge/github.com/ksonnet/ksonnet)](https://goreportcard.com/report/github.com/ksonnet/ksonnet)

*ksonnet* is a framework for writing, sharing, and deploying Kubernetes application manifests. With its CLI, you can generate a complete application from scratch in only a few commands, or manage a complex system at scale.

Specifically, ksonnet allows you to:
* **Reuse** common manifest patterns (within your app or from external libraries)
* **Customize** manifests directly with powerful object concatenation syntax
* **Deploy** app manifests to multiple environments
* **Diff** across environments to compare two running versions of your app
* **Track** the entire state of your app configuration in version controllable files

All of this results in a more iterative process for developing manifests, one that can be supplemented by continuous integration (CI).

*STATUS: Development is ongoing—this tool is pre-alpha.*

## Install

> You should have Go installed, and an appropriately set `$GOPATH`. (For most systems, the minimum Go version is 1.7. However, recent OSX may [require golang >=1.8.1](https://github.com/golang/go/issues/19734) to
avoid an immediate `Killed: 9`.) If you need additional help, see the [official Go installation guide](https://golang.org/doc/install#install).

ksonnet is installed like any other Go binary. Run the following shell commands to download ksonnet and ensure that it is runnable in any directory:

```bash
go get github.com/ksonnet/ksonnet
PATH=$PATH:$GOPATH/bin
```

If your ksonnet is properly installed, you should be able to run the following `--help` command and see similar output:

```
$ ksonnet --help

Synchronise Kubernetes resources with config files

Usage:
  ks [command]

Available Commands:
  apply       Apply local configuration to remote cluster
  delete      Delete Kubernetes resources described in local config
  diff        Display differences between server and local config, or server and server config
  env         Manage ksonnet environments
  generate    Expand prototype, place in components/ directory of ksonnet app
  ...
```

## Quickstart

Here we provide a shell script that shows some basic ksonnet features in action. You can run this script to deploy and update a basic web app UI, via a Kubernetes Service and Deployment. This app is shown below:

<p align="center">
<img alt="guestbook screenshot" src="/docs/img/guestbook.png" style="width:60% !important;">
</p>

Note that we will not be implementing the entire app in this quickstart, so the buttons will not work!

**Minimal explanation is provided here, and only basic ksonnet features are shown---this is intended to be a quick demonstration.** If you are interested in learning more, see [Additional Documentation](#additional-documentation).

### Prerequisites
* *You should have access to an up-and-running Kubernetes cluster (**maximum version 1.7**).* Support for Kubernetes 1.8 is under development.

  If you do not have a cluster, [choose a setup solution](https://kubernetes.io/docs/setup/) from the official Kubernetes docs.

* *You should have `kubectl` installed.* If not, follow the instructions for [installing via Homebrew (MacOS)](https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-with-homebrew-on-macos) or [building the binary (Linux)](https://kubernetes.io/docs/tasks/tools/install-kubectl/#tabset-1).

* *Your `$KUBECONFIG` should specify a valid `kubeconfig` file*, which points at the cluster you want to use for this demonstration.

### Script

Copy and paste the script below to deploy the container image for a basic web app UI:

```
# Start by creating your app directory
# (This references your current cluster using $KUBECONFIG)
ksonnet init quickstart
cd quickstart

# Autogenerate a basic manifest
ksonnet generate deployment-exposed-with-service guestbook-ui \
--name guestbook \
--image alpinejay/dns-single-redis-guestbook:0.3 \
--type ClusterIP

# Deploy your manifest to your cluster
ksonnet apply default

# Set up an API proxy so that you can access the guestbook service locally
kc proxy > /dev/null &
PROXY_PID=$!
QUICKSTART_NAMESPACE=$(kubectl get svc guestbook -o jsonpath="{.metadata.namespace}")
GUESTBOOK_SERVICE_URL=http://localhost:8001/api/v1/proxy/namespaces/$QUICKSTART_NAMESPACE/services/guestbook

# Check out the guestbook app in your browser (NOTE: the buttons don't work!)
open $GUESTBOOK_SERVICE_URL

```

The rest of this script upgrades the container image to a new version:

```
# Bump the container image to a different version
ksonnet param set guestbook-ui image alpinejay/dns-single-redis-guestbook:0.4

# See the differences between your local manifest and what's running on your cluster
# (ERROR is expected here since there are differences)
ksonnet diff local:default remote:default

# Update your cluster with your latest changes
ksonnet apply default

# (Wait a bit) and open another tab to see newly added javascript
open $GUESTBOOK_SERVICE_URL

```

Notice that the webpage looks different! Now clean up:

```
# Teardown
kill -9 $PROXY_PID
ksonnet delete default

# There should be no guestbook service left running
kubectl get svc guestbook
```

Even though you've made modifications to the Guestbook app and removed it from your cluster, ksonnet still tracks all your manifests locally:

```
# View all expanded manifests (YAML)
ks show default
```

If you're wondering how ksonnet differs from existing tools, the full-length tutorial (WIP) shows you how to use other ksonnet features to implement the rest of the Guestbook app (so that the buttons work!).

## Additional documentation

ksonnet is a feature-rich framework. To learn more about how to integrate it into your workflow, check out the resources below:

* **Tutorial (WIP)** - How do I use ksonnet and why? This finishes the Guestbook app from the [Quickstart](#quickstart) above.

* **Interactive tour of ksonnet (WIP)** - Where does the ksonnet magic come from?

* **[CLI Reference](/docs/cli-reference#command-line-reference)** - What ksonnet commands are available, and how do I use them?

* **[Concept Reference (WIP)](/docs/concepts.md)** - What do all these special ksonnet terms mean (e.g. *prototypes*) ?


## Troubleshooting

If you encounter any problems that the documentation does not address, [file an issue](https://github.com/ksonnet/ksonnet/issues).

## Contributing

Thanks for taking the time to join our community and start contributing!

#### Before you start

* Please familiarize yourself with the [Code of
Conduct](CODE_OF_CONDUCT.md) before contributing.
* Read the contribution guidelines in [CONTRIBUTING.md](CONTRIBUTING.md).
* There is a [mailing list](https://groups.google.com/forum/#!forum/ksonnet) and [Slack channel](https://ksonnet.slack.com/) if you want to interact with
other members of the community.

#### Pull requests

* We welcome pull requests. Feel free to dig through the [issues](https://github.com/ksonnet/ksonnet/issues) and jump in.

## Changelog

See [the list of releases](https://github.com/ksonnet/ksonnet/releases) to find out about feature changes.
