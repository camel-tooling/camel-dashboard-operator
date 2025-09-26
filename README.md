<h1 align="center">
  <a href="https://camel-tooling.github.io/camel-dashboard/docs/operator/">Camel Dashboard Operator</a>
</h1>

<p align=center>
  <a href="https://github.com/camel-tooling/camel-dashboard-operator/blob/main/LICENSE"><img src="https://img.shields.io/github/license/camel-tooling/camel-dashboard-operator?color=104d92&style=for-the-badge" alt="License"/></a>
  <a href="https://camel-tooling.github.io/camel-dashboard/docs/operator/"><img src="https://img.shields.io/badge/Documentation-Camel_Dashboard_Operator-white?color=cf7428&style=for-the-badge" alt="Visit"/></a>
</p><br/>

<h2 align="center">The <a href="https://github.com/camel-tooling/camel-dashboard">Camel Dashboard</a> Monitoring for Kubernetes</h2>

The Camel Dashboard Operator is a project created to simplify the management of any Camel application on a Kubernetes cluster. The tool is in charge to monitor any Camel application and provide a set of basic information, useful to learn how your fleet of Camel (a caravan!?) is behaving.

The project is designed to be as simple and low resource consumption as possible. It only collects the most important Camel application KPI in order to quickly identify what's going on across your Camel applications.

> **_NOTE:_** as the project is still in an experimental phase, the metrics collected can be changed at each development iteration.

## The Camel custom resource

The operator uses a simple custom resource known as `CamelApp` or `capp` which stores certain metrics around your running applications. The operator detects the Camel applications you're deploying to the cluster, identifying them in a given namespace or a given metadata label that need to be included when deploying your applications (all configurable on the operator side).

## Install the operator

To install the Camel Dashboard Operator please see the [installation documentation](https://camel-tooling.github.io/camel-dashboard/docs/operator/).

## Configure a Camel application

To create a new Camel Application or modify an existing Camel Application to be monitored by the Camel Dashboard Operator please see the [Camel Application configuration documentation](https://camel-tooling.github.io/camel-dashboard/docs/operator/configuration/import/)

## Tuning configuration

To review the several configuration you can apply separately to each of your Camel application please see the [tuning documentation](https://camel-tooling.github.io/camel-dashboard/docs/operator/configuration/tuning/)

## Openshift plugin

This operator can work standalone and you can use the data exposed in the `CamelApp` custom resource accordingly. However it has a great fit with the [Camel Dashboard Console](https://camel-tooling.github.io/camel-dashboard/docs/console/), which is a visual representation of the services exposed by the operator.

## Development

In order to build the project, you need Go (version 1.24+) needed. Refer to the [Go](https://go.dev/) website for the installation.

In order to deploy the project to an actual kubernetes cluster you need:
* [helm](https://helm.sh)
* [docker](https://www.docker.com/) and the [buildx plugin](https://github.com/docker/buildx)
* a Kubernetes/Openshift cluster

To build the whole project you now need to run:
```
NOTEST=1 make
```

After a successful build, if you’re connected to a Docker daemon, you can build the operator Docker image by running:
```
NOTEST=1 make images
```

You might need to produce `camel-dashboard-operator` images that need to be pushed to the custom repository e.g. docker.io/myrepo/camel-dashboard-operator, to do that you can pass a parameter CUSTOM_IMAGE to make as shown below:
```
NOTEST=1 make CUSTOM_IMAGE='docker.io/myrepo/camel-dashboard-operator' images
```

> **_NOTE:_** The image `quay.io/camel-tooling/camel-dashboard-operator:latest-amd64` is published so it can be pulled instead.

Deploy using helm:
```sh
helm upgrade -i camel-dashboard-operator helm/camel-dashboard --namespace camel-dashboard --set operator.image=quay.io/camel-tooling/camel-dashboard-operator:latest-amd64 
```
