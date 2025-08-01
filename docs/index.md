---
id: home
title: Camel Dashboard Operator
layout: default
---

The **Camel Dashboard Operator** is a project created to simplify the management of any Camel application on a Kubernetes cluster. The tool is in charge to **monitor any Camel application** and provide a set of basic information, useful to learn how your fleet of Camel (a caravan!?) is behaving.

The project is designed to be as simple and low resource consumption as possible. It only collects the most important Camel application KPI in order to quickly identify what's going on across your Camel applications.

NOTE: as the project is still in an experimental phase, the metrics collected can be changed at each development iteration.

## The Camel custom resource

The operator uses a simple custom resource known as `CamelApp` or `capp` which stores certain metrics around your running applications. The operator detects the Camel applications you're deploying to the cluster, identifying them in a given namespace or a given metadata label that need to be included when deploying your applications (all configurable on the operator side).

## How to see the dashboard

As the operator knows about `CamelApp` custom resource, you can use any Kubernetes tool (`kubectl`, `oc`) to visualize it. In order to render the resources graphically we are also developing an [Openshift plugin](https://github.com/camel-tooling/camel-openshift-console-plugin) which can be installed separately.