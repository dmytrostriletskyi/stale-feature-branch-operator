Delete stale feature branches in your `Kubernetes` cluster.

[![Release](https://img.shields.io/github/release/dmytrostriletskyi/stale-feature-branch-operator.svg)](https://github.com/dmytrostriletskyi/stale-feature-branch-operator/releases)
[![Build Status](https://travis-ci.com/dmytrostriletskyi/stale-feature-branch-operator.svg?branch=master)](https://travis-ci.com/dmytrostriletskyi/stale-feature-branch-operator)

* [Getting Started](#getting-started)
  * [Feature Branch](#feature-branch)
  * [Motivation](#motivation)
* [Installation](#installation)
* [Usage](#usage)
* [Guideline](#development)
  * [Requirements](#guideline-requirements)
  * [Running](#guideline-running)
* [API](#api)
  * [Version One](#version-one)
* [Development](#development)
  * [Requirements](#development-requirements)
  * [Cloning](#cloning)
  * [Running](#development-running)
  * [Docker Image](#docker-image)
  * [Contributing](#contributing)
    * [Code Style](#code-style)
    * [Testing](#testing)
    * [Custom Resource Definitions](#custom-resource-definitions)

## Getting Started

### Feature Branch

`Feature branch` (or `deploy preview`) means that a pull request is deployed as a separate instance of your application.
It allows preventing errors and bugs, responsible people can check a feature before it's merged to production.

One of the ways to create a feature branch in `Kubernetes` cluster is to use [namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/)
to separate production deployment from any other. Production configurations may look similar to:

```yaml
kind: Deployment
apiVersion: apps/v1
metadata:
  namespace: github-back-end
...
```

Otherwise, feature branches always have a different namespace. Such as `-pr-` prefix or postfix in its name. The example
is illustrated below:

```yaml
kind: Deployment
apiVersion: apps/v1
metadata:
  namespace: github-back-end-pr-17
...
```

More information about implementation of feature branches using namespaces is [here](https://itnext.io/feature-deployments-in-kubernetes-c74bdcff0d8e)
and [here](https://codefresh.io/kubernetes-tutorial/dynamically-creating-k8s-namespaces-every-branch-pull-request-2/).

### Motivation

To understand the motivation of the project, let's check common continuous integration lifecycle for a pull request:

1. A new commit is pushed to a branch.
2. Code style and tests are passed.
3. A feature branch's configurations are applied.
4. The feature branch's namespace and other resources are running in a cluster.
5. The branch is merged to a production branch.

One important thing is that good lifecycle will delete all existing feature branch resources for a particular commit
**before** applying configurations for the new commit. It's needed to ensure that each commit's deployment is done from
a clear state.
 
But after the branch is merged to the production branch, all feature branch's resources are still running in a cluster
and occupy its resources. **What are the ways to delete them?**

* On each master branch build, detect which branch was merged last (by fetching commits history).
* Write a service that receives branches' merging events.
* Create own `Cronjob`.

All of them are not ideal as have the following disadvantages: master branch builds may fail, own software takes time
for development and maintenance.

## Installation

Apply the latest release configurations with the command below, it will create the `StaleFeatureBranch` resource,
install the operator into `stale-feature-branch-operator` namespace, create a [service account](https://kubernetes.io/docs/reference/access-authn-authz/service-accounts-admin/)
and necessary [RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac) roles.

```bash
$ kubectl apply -f \
      https://raw.githubusercontent.com/dmytrostriletskyi/stale-feature-branch-operator/master/configs/production.yml
```

If you need any previous release, full list of versions is available [here](https://github.com/dmytrostriletskyi/stale-feature-branch-operator/releases).

## Usage

To delete stale feature branches, after applying installation instructions above, create a configuration file with
`feature-branch.dmytrostriletskyi.com/v1` as `apiVersion` and `StaleFeatureBranch` as kind:

```yaml
apiVersion: feature-branch.dmytrostriletskyi.com/v1
kind: StaleFeatureBranch
metadata:
  name: stale-feature-branch
spec:
  namespaceSubstring: -pr-
  afterDaysWithoutDeploy: 3
```

Choose any metadata's name for the resource and dive into specifications:

1. `namespaceSubstring` is needed to get all feature branches' namespaces. For instance, the example above will grab 
`github-back-end-pr-17` and `github-back-end-pr-33` if there are namespaces `github-back-end`, `github-front-end`,
`github-back-end-pr-17`, `github-back-end-pr-33` in a cluster as the `-pr-` substring occurs there.
2. `afterDaysWithoutDeploy` is needed to delete only old namespaces. If you set `3 days` there, namespaces created 
`1 day` or `2 days` ago will not be deleted, but created `3 days, 1 hour` or `4 days` will be deleted.

It processes feature branches' namespaces every `30 minutes` by default. The last available parameter in specifications
is `checkEveryMinutes`. You can configure a frequency of the processes in minutes if the default value doesn't fit you.

Check [guideline below](#guideline) if you want to know how it works under the hood.

## Guideline

This guideline shows how the deletion of stale feature branches works under the hood. **You should not reproduce the
instructions below for production cluster** as it's just a detailed example to understand the behavior of the operator.
For this chapter, testing `Kubernetes` cluster on your personal computer will be used.

<h3 id="guideline-requirements">Requirements</h3>

1. [Docker](https://docs.docker.com/get-docker). Virtualization to run the software in packages called containers.
2. [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube). Runs a single-node `Kubernetes` cluster in a 
virtual machine (or `Docker`) on your personal computer.
3. [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl). Command-line interface to access `Kubernetes`
cluster.

<h3 id="guideline-running">Running</h3>

Start `Kubernetes` cluster on your personal computer with the following command:

```bash
$ minikube start --vm-driver=docker
minikube v1.11.0 on Darwin 10.15.5
Using the docker driver based on existing profile.
Starting control plane node minikube in cluster minikube.
```

After, choose your cluster as the main one for `kubectl`. It's needed for cases you work with many clusters from the
single computer:

```bash
$ kubectl config use-context minikube
Switched to context "minikube".
```

Applied configurations in the same way you apply it to a production cluster. But as it's production configurations, they
will expect old namespaces present in your cluster. Our cluster is fresh, and no old resources are present there.
As you do not have them, the operator allows you to specify the debug parameter. If the debug is enabled, all
namespaces will be deleted without checking for an oldness:

Copy the production configurations to your personal computer:

```bash
$ curl https://raw.githubusercontent.com/dmytrostriletskyi/stale-feature-branch-operator/master/configs/production.yml > \
      stale-feature-branch-production-configs.yml
```

If you need any previous release, full list of versions is available [here](https://github.com/dmytrostriletskyi/stale-feature-branch-operator/releases).

Enable debug by changing the setting. For `Linux` it's:

```bash
$ sed -i 's|false|true|g' stale-feature-branch-production-configs.yml
```

For `macOS` it's:

```bash
$ sed -i "" 's|false|true|g' stale-feature-branch-production-configs.yml
```

Apply the changed production configurations:

```bash
$ kubectl apply -f stale-feature-branch-production-configs.yml
```

Fetch all resources in `Kubernetes` cluster, you will see `StaleFeatureBranch` resource is available to use:

```bash
$ kubectl api-resources | grep stalefeaturebranches
NAME                              SHORTNAMES   APIGROUP                               NAMESPACED   KIND
stalefeaturebranches              sfb          feature-branch.dmytrostriletskyi.com   true         StaleFeatureBranch
```

Fetch pods in `stale-feature-branch-operator` namespace, you will see an operator that listens for new `StaleFeatureBranch`
resources running there:

```bash
$ kubectl get pods --namespace stale-feature-branch-operator
NAME                                             READY   STATUS    RESTARTS   AGE
stale-feature-branch-operator-6bfbfd4df8-m7sch   1/1     Running   0          38s
```

Fetch the operator's logs to ensure it's running:

```bash
$ kubectl logs stale-feature-branch-operator-6bfbfd4df8-m7sch -n stale-feature-branch-operator
{"level":"info","ts":1592306900.8200202,"logger":"cmd","msg":"Operator Version: 0.0.1"}
...
{"level":"info","ts":1592306901.5672553,"logger":"controller-runtime.controller","msg":"Starting EventSource","controller":"stalefeaturebranch-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1592306901.6680624,"logger":"controller-runtime.controller","msg":"Starting Controller","controller":"stalefeaturebranch-controller"}
{"level":"info","ts":1592306901.6681142,"logger":"controller-runtime.controller","msg":"Starting workers","controller":"stalefeaturebranch-controller","worker count":1}
```

Create ready-to-use fixtures that contain two namespaces `project-pr-1` and `project-pr-2` with many other resources
as well (deployment, service, secrets, etc.).:

```bash
$ kubectl apply \
      -f https://raw.githubusercontent.com/dmytrostriletskyi/stale-feature-branch-operator/master/fixtures/first-feature-branch.yml \
      -f https://raw.githubusercontent.com/dmytrostriletskyi/stale-feature-branch-operator/master/fixtures/second-feature-branch.yml
namespace/project-pr-1 created
deployment.apps/project-pr-1 created
service/project-pr-1 created
horizontalpodautoscaler.autoscaling/project-pr-1 created
secret/project-pr-1 created
configmap/project-pr-1 created
ingress.extensions/project-pr-1 created
namespace/project-pr-2 created
deployment.apps/project-pr-2 created
service/project-pr-2 created
horizontalpodautoscaler.autoscaling/project-pr-2 created
secret/project-pr-2 created
configmap/project-pr-2 created
ingress.extensions/project-pr-2 created
```

You can check their existence by the following command:

```bash
$ kubectl get namespace,pods,deployment,service,horizontalpodautoscaler,configmap,ingress -n project-pr-1 && \
      kubectl get namespace,pods,deployment,service,horizontalpodautoscaler,configmap,ingress -n project-pr-2
...
NAME                                READY   STATUS    RESTARTS   AGE
pod/project-pr-1-848d5fdff6-rpmzw   1/1     Running   0          67s

NAME                           READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/project-pr-1   1/1     1            1           67s
...
```

As it's told above, when debug is enabled, all namespaces will be deleted without checking for an oldness. It means if
we create `StaleFeatureBranch` configurations, the namespaces will be deleted immediately. The fixture for
`StaleFeatureBranch` will check for namespaces that contain `-pr-` in their names **once a minute**.

```bash
$ kubectl apply -f \
      https://raw.githubusercontent.com/dmytrostriletskyi/stale-feature-branch-operator/master/fixtures/stale-feature-branch.yml
```

After, check the logs of the operator, and you will that namespaces are deleted:

```bash
{"level":"info","ts":1592322500.64014,"logger":"stale-feature-branch-controller","msg":"Stale feature branch is being processing.","namespaceSubstring":"-pr-","afterDaysWithoutDeploy":1,"checkEveryMinutes":1,"isDebug":"true"}
{"level":"info","ts":1592322500.7436411,"logger":"stale-feature-branch-controller","msg":"Namespace should be deleted due to debug mode is enabled.","namespaceName":"project-pr-1"}
{"level":"info","ts":1592322500.743676,"logger":"stale-feature-branch-controller","msg":"Namespace is being processing.","namespaceName":"project-pr-1","namespaceCreationTimestamp":"2020-06-16 18:43:58 +0300 EEST"}
{"level":"info","ts":1592322500.752212,"logger":"stale-feature-branch-controller","msg":"Namespace has been deleted.","namespaceName":"project-pr-1"}
{"level":"info","ts":1592322500.752239,"logger":"stale-feature-branch-controller","msg":"Namespace should be deleted due to debug mode is enabled.","namespaceName":"project-pr-2"}
{"level":"info","ts":1592322500.752244,"logger":"stale-feature-branch-controller","msg":"Namespace is being processing.","namespaceName":"project-pr-2","namespaceCreationTimestamp":"2020-06-16 18:43:58 +0300 EEST"}
{"level":"info","ts":1592322500.75804,"logger":"stale-feature-branch-controller","msg":"Namespace has been deleted.","namespaceName":"project-pr-2"}
```

If you check resources again, the output would be `Terminating` or empty.

```Bash
$ kubectl get namespace,pods,deployment,service,horizontalpodautoscaler,configmap,ingress -n project-pr-1 && \
      kubectl get namespace,pods,deployment,service,horizontalpodautoscaler,configmap,ingress -n project-pr-2
```

You can go through the process of the creation of resources again. At the end, in a minute or less, resources will be
deleted again.

## API

### Version One

Use `feature-branch.dmytrostriletskyi.com/v1` as `apiVersion`. Arguments for specification are the following:

| Arguments                | Type    | Required | Restrictions | Default  | Description                                                                   |
|:------------------------:|:-------:|:--------:|:------------:|:--------:|-------------------------------------------------------------------------------|
| `namespaceSubstring`     | String  | Yes      | -            | -        | Substring to grab feature branches' namespaces and not other once.            |
| `afterDaysWithoutDeploy` | Integer | Yes      | `>0`         | -        | Delete feature branches' namespaces if there is no deploy for number of days. |
| `checkEveryMinutes`      | Integer | No       | `>0`         | `30`     | Processes feature branches' namespaces each number of minutes.                |

## Development

<h3 id="development-requirements">Requirements</h3>

1. [Docker](https://docs.docker.com/get-docker). Virtualization to run software in packages called containers.
2. [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube). Runs a single-node `Kubernetes` cluster in a 
virtual machine (or `Docker`) on your personal computer.
3. [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl). Command line interface to access `Kubernetes`
cluster.

### Cloning

Clone the project with the following command:

```bash
$ mkdir -p $GOPATH/src/github.com/dmytrostriletskyi
$ cd $GOPATH/src/github.com/dmytrostriletskyi
$ git clone git@github.com:dmytrostriletskyi/stale-feature-branch-operator.git
$ cd stale-feature-branch-operator
```

<h3 id="development-running">Running</h3>

Start `Kubernetes` cluster on your personal computer with the following command:

```bash
$ minikube start --vm-driver=docker
minikube v1.11.0 on Darwin 10.15.5
Using the docker driver based on existing profile.
Starting control plane node minikube in cluster minikube.
```

After, choose your cluster as main one for `kubectl`. It's needed for cases you work with many clusters from the single
computer:

```bash
$ kubectl config use-context minikube
Switched to context "minikube".
```

Register `StaleFeatureBranch` resource by the following command:

```bash
$ kubectl create -f configs/development.yml
```

By fetching all resources in `Kubernetes` cluster, you will see `StaleFeatureBranch` resource is available to use there:

```bash
$ kubectl api-resources | grep stalefeaturebranches
NAME                              SHORTNAMES   APIGROUP                               NAMESPACED   KIND
stalefeaturebranches              sfb          feature-branch.dmytrostriletskyi.com   true         StaleFeatureBranch
```

Build the operator with the following command:

```bash
$ go build -a -o operator pkg/*.go
```

Run the operator with the following command:

```bash
$ ./operator
{"level":"info","ts":1592321007.8580391,"logger":"cmd","msg":"Operator Version: 0.0.1"}
...
{"level":"info","ts":1592321008.1686652,"logger":"controller-runtime.controller","msg":"Starting EventSource","controller":"stalefeaturebranch-controller","source":"kind source: /, Kind="}
{"level":"info","ts":1592321008.3716009,"logger":"controller-runtime.controller","msg":"Starting Controller","controller":"stalefeaturebranch-controller"}
{"level":"info","ts":1592321008.3717089,"logger":"controller-runtime.controller","msg":"Starting workers","controller":"stalefeaturebranch-controller","worker count":1}
```

The following environment variables are supported:

```bash
$ OPERATOR_NAME=stale-feature-branch-operator IS_DEBUG=true ./operator
```

| Arguments       | Type    | Required | Restrictions         | Default  | Description                                                                               |
|:---------------:|:-------:|:--------:|:--------------------:|:--------:|-------------------------------------------------------------------------------------------|
| `OPERATOR_NAME` | String  | Yes      | -                    | -        | Operator name.                                                                            |
| `IS_DEBUG`      | String  | No       | One of: true, false. | false    | If debug mode is enabled, all namespaces will be deleted without checking for an oldness. |

Create ready-to-use fixtures that container two namespaces `project-pr-1` and `project-pr-2` with many other resources
as well (deployment, service, secrets, etc.):

```bash
$ kubectl apply \
      -f fixtures/first-feature-branch.yml -f fixtures/second-feature-branch.yml
namespace/project-pr-1 created
deployment.apps/project-pr-1 created
service/project-pr-1 created
horizontalpodautoscaler.autoscaling/project-pr-1 created
secret/project-pr-1 created
configmap/project-pr-1 created
ingress.extensions/project-pr-1 created
namespace/project-pr-2 created
deployment.apps/project-pr-2 created
service/project-pr-2 created
horizontalpodautoscaler.autoscaling/project-pr-2 created
secret/project-pr-2 created
configmap/project-pr-2 created
ingress.extensions/project-pr-2 created
```

You can check their existence by the following command:

```bash
$ kubectl get namespace,pods,deployment,service,horizontalpodautoscaler,configmap,ingress -n project-pr-1 && \
      kubectl get namespace,pods,deployment,service,horizontalpodautoscaler,configmap,ingress -n project-pr-2
...
NAME                                READY   STATUS    RESTARTS   AGE
pod/project-pr-1-848d5fdff6-rpmzw   1/1     Running   0          67s

NAME                           READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/project-pr-1   1/1     1            1           67s
...
```

As it's told above, when debug is enabled, all namespaces will be deleted without checking for an oldness. It means if
we create `StaleFeatureBranch` configurations, the namespaces will be deleted immediately. The fixture for
`StaleFeatureBranch` will check for namespaces that contain `-pr-` in their names **once a minute**.

```bash
$ kubectl apply -f fixtures/stale-feature-branch.yml
```

After, check the logs of the operator, and you will that namespaces are deleted:

```bash
{"level":"info","ts":1592322500.64014,"logger":"stale-feature-branch-controller","msg":"Stale feature branch is being processing.","namespaceSubstring":"-pr-","afterDaysWithoutDeploy":1,"checkEveryMinutes":1,"isDebug":"true"}
{"level":"info","ts":1592322500.7436411,"logger":"stale-feature-branch-controller","msg":"Namespace should be deleted due to debug mode is enabled.","namespaceName":"project-pr-1"}
{"level":"info","ts":1592322500.743676,"logger":"stale-feature-branch-controller","msg":"Namespace is being processing.","namespaceName":"project-pr-1","namespaceCreationTimestamp":"2020-06-16 18:43:58 +0300 EEST"}
{"level":"info","ts":1592322500.752212,"logger":"stale-feature-branch-controller","msg":"Namespace has been deleted.","namespaceName":"project-pr-1"}
{"level":"info","ts":1592322500.752239,"logger":"stale-feature-branch-controller","msg":"Namespace should be deleted due to debug mode is enabled.","namespaceName":"project-pr-2"}
{"level":"info","ts":1592322500.752244,"logger":"stale-feature-branch-controller","msg":"Namespace is being processing.","namespaceName":"project-pr-2","namespaceCreationTimestamp":"2020-06-16 18:43:58 +0300 EEST"}
{"level":"info","ts":1592322500.75804,"logger":"stale-feature-branch-controller","msg":"Namespace has been deleted.","namespaceName":"project-pr-2"}
```

If you check resources again, the output would be `Terminating` or empty.

```Bash
$ kubectl get namespace,pods,deployment,service,horizontalpodautoscaler,configmap,ingress -n project-pr-1 && \
      kubectl get namespace,pods,deployment,service,horizontalpodautoscaler,configmap,ingress -n project-pr-2
```

You can go through the process of creation of resources again. At the end, in a minute or less, resources will be
deleted again.

### Docker Image

The operator is deployed to `Kubernetes` cluster as a pod in a deployment that can be found in `configs/production.yml`
file:

```yaml
kind: Deployment
apiVersion: apps/v1
...
      containers:
        - name: stale-feature-branch-operator
          image: dmytrostriletskyi/stale-feature-branch-operator:v0.0.1
...
```

To build, use the following command replacing registry, project name and version if needed:

```bash
$ docker build --tag dmytrostriletskyi/stale-feature-branch-operator:v$(cat .project-version) -f ops/Dockerfile .
```

To push,  use the following command replacing registry, project name and version if needed:

```bash
$ docker push dmytrostriletskyi/stale-feature-branch-operator:v$(cat .project-version)
```

If you want to run it locally, do the following command:

```bash
$ docker run dmytrostriletskyi/stale-feature-branch-operator:v$(cat .project-version) && \
      --name stale-feature-branch-operator
```

### Contributing

#### Code Style

Ensure, your code is formatted with the following command:

```bash
$ go fmt ./...
```

#### Testing

Ensure, you code is covered with tests using the following command:

```
$ go test ./... -v -count=1
```

#### Custom Resource Definitions

If you changed a custom resource definition schema such as `pkg/apis/featurebranch/v1/stale_feature_branch.go`,
you should:

1. Update corresponding `CustomResourceDefinition` resources in `configs/development.yml` and 
`configs/production.yml`. To generate `CustomResourceDefinition` resource based on your changes, use the following
command. It will output update configuration:

    ```bash
    $ make crds
    o: creating new go.mod: module tmp
    go: found sigs.k8s.io/controller-tools/cmd/controller-gen in sigs.k8s.io/controller-tools v0.2.5
    ../../go/bin/controller-gen crd:trivialVersions=true rbac:roleName=manager-role webhook output:stdout paths="./..."
    
    ---
    apiVersion: apiextensions.k8s.io/v1beta1
    kind: CustomResourceDefinition
    metadata:
      annotations:
        controller-gen.kubebuilder.io/version: v0.2.5
      creationTimestamp: null
      name: stalefeaturebranches.feature-branch.dmytrostriletskyi.com
    ...
    ```

2. Update deep copies for schema's structures. It will update file `pkg/apis/featurebranch/v1/zz_generated.deepcopy.go`
automatically:

    ```bash
    $ make deep-copy
    go: creating new go.mod: module tmp
    go: found sigs.k8s.io/controller-tools/cmd/controller-gen in sigs.k8s.io/controller-tools v0.2.5
    ../../go/bin/controller-gen object paths="./..."
    ```
