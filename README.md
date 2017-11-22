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

> You should have Go installed *(minimum version 1.8.1)*. If not, follow the instructions in the [official installation guide](https://golang.org/doc/install#install).

Copy and paste the following commands:
```bash
# Download ksonnet
go get github.com/ksonnet/ksonnet

# Build and install binary under shortname `ks`
cd $GOPATH/src/github.com/ksonnet/ksonnet
make install
```
If your ksonnet is properly installed, you should be able to run `ks --help` and see output describing the various `ks` commands.

#### Common issues
* **Ensure that your `$GOPATH` is set appropriately.** If `echo $GOPATH` results in empty output, you'll need to set it. If you're using OSX, trying adding the line `export GOPATH=$HOME/go` to the end of your `$HOME/.bash_profile`.

  Other systems may have different `$GOPATH` defaults (e.g. `/usr/local/go`), in which case you should use those instead. If you get stuck, [these instructions](https://github.com/golang/go/wiki/SettingGOPATH) may help).

* **You may need to specify your `$GOPATH` in the same command as `make install`.** For example, try `GOPATH=<your-go-path> make install` (making sure to replace `<your-go-path>`), instead of just `make install`.

* **If your error is "command not found", make sure that Go binaries are included in your $PATH**. You can do this by running `PATH=$PATH:$GOPATH/bin`.

## Example

Here we provide some commands that show some basic ksonnet features in action. You can run these commands to deploy and update a basic web app UI, via a Kubernetes Service and Deployment. This app is shown below:

<p align="center">
<img alt="guestbook screenshot" src="/docs/img/guestbook.png" style="width:60% !important;">
</p>

Note that we will not be implementing the entire app in this example, so the buttons will not work!

**Minimal explanation is provided here, and only basic ksonnet features are shown---this is intended to be a quick demonstration.** If you are interested in learning more, see [Additional Documentation](#additional-documentation).

### Prerequisites
* *You should have access to an up-and-running Kubernetes cluster (**maximum version 1.7**).* Support for Kubernetes 1.8 is under development.

  If you do not have a cluster, [choose a setup solution](https://kubernetes.io/docs/setup/) from the official Kubernetes docs.

* *You should have `kubectl` installed.* If not, follow the instructions for [installing via Homebrew (MacOS)](https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-with-homebrew-on-macos) or [building the binary (Linux)](https://kubernetes.io/docs/tasks/tools/install-kubectl/#tabset-1).

* *Your `$KUBECONFIG` should specify a valid `kubeconfig` file*, which points at the cluster you want to use for this demonstration.

### Example flow

You can copy and paste the commands below to deploy the web app UI:

```bash
# Start by creating your app directory (this is created at the current path)
# (This references your current cluster using $KUBECONFIG)
ks init ks-example
cd ks-example

# Autogenerate a basic manifest
ks generate deployed-service guestbook-ui \
  --name guestbook \
  --image alpinejay/dns-single-redis-guestbook:0.3 \
  --type ClusterIP

# Deploy your manifest to your cluster
ks apply default
```

Now there should be a Deployment and Service running on your cluster! Try accessing the `guestbook` service in your browser. (How you do this may depend on your cluster setup).

<details>
<summary><i>If you are unsure what to do, we suggest using <code>kubectl proxy</code>.</i></summary>
<pre>
# Set up an API proxy so that you can access the guestbook service locally
kubectl proxy > /dev/null &
PROXY_PID=$!
QUICKSTART_NAMESPACE=$(kubectl get svc guestbook -o jsonpath="{.metadata.namespace}")
GUESTBOOK_SERVICE_URL=http://localhost:8001/api/v1/proxy/namespaces/$QUICKSTART_NAMESPACE/services/guestbook
open $GUESTBOOK_SERVICE_URL
</pre>
</details>

<br>

*(Remember, the buttons won't work in this example.)*

Now let's try upgrading the container image to a new version:

```bash
# Bump the container image to a different version
ks param set guestbook-ui image alpinejay/dns-single-redis-guestbook:0.4

# See the differences between your local manifest and what's running on your cluster
# (ERROR is expected here since there are differences)
ks diff local:default remote:default

# Update your cluster with your latest changes
ks apply default

```

Check out the webpage again in your browser (force-refresh to update the javascript). Notice that it looks different! Clean up:

```bash
# Teardown
ks delete default

# There should be no guestbook service left running
kubectl get svc guestbook
```

*(If you ended up copying and pasting the `kubectl proxy` code above, make sure to clean up that process with `kill -9 $PROXY_PID`).*

Now, even though you've made modifications to the Guestbook app and removed it from your cluster, ksonnet still tracks all your manifests locally:

```bash
# View all expanded manifests (YAML)
ks show default
```

If you're still wondering how ksonnet differs from existing tools, the [tutorial](https://ksonnet-next-site.i.heptio.com/docs/tutorial) shows you how to use other ksonnet features to implement the rest of the Guestbook app (and yes, the buttons will work!).

## Additional documentation

ksonnet is a feature-rich framework. To learn more about how to integrate it into your workflow, check out the resources below:

* **[Tutorial](https://ksonnet-next-site.i.heptio.com/docs/tutorial)** - What can I build with ksonnet and why? This finishes the Guestbook app from the [Example](#example) above.

* **[Interactive tour of ksonnet](https://ksonnet-next-site.i.heptio.com/docs/tutorial/tour/welcome)** - How do `ks` commands work under the hood?

* **[CLI Reference](/docs/cli-reference#command-line-reference)** - What ksonnet commands are available, and how do I use them?

* **[Concept Reference](/docs/concepts.md)** - What do all these special ksonnet terms mean (e.g. *prototypes*) ?


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
