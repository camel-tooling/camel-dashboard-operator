---
id: import
title: Import a Camel application
permalink: /configuration/import
---

The operator is instructed to watch `Deployment` and verify if they are marked as Camel application. You will likely need to update your deployment process and include automatically a `camel.apache.org/app` label for all the applications you want to monitor.

NOTE: you can configure the operator to watch for a different label setting the environment variable `LABEL_SELECTOR` in the operator Pod.

## Collect Camel metrics

The operator is designed to consume the services exposed by [Camel Observability Services component](https://camel.apache.org/components/next/others/observability-services.html).

It will works also when no services are exposed, but it won't be able to collect any meaningful metrics (likely only the status and the number of replicas).

## Run a new Camel application

Let's run some sample Camel application. We have prepared a few available to run some quick demo:

* A Camel main application available at `docker.io/squakez/db-app-main:1.0`
* A Camel Quarkus application available at `docker.io/squakez/db-app-quarkus:1.0`
* A Camel Spring Boot application available at `docker.io/squakez/db-app-sb:1.0`

These applications were created, exported and "containerized" via Camel JBang, which includes by default the aforementioned `camel-observability-services` dependency.

Let's run them in a Kubernetes cluster (it also works in a local cluster such as `Minikube`):

```
kubectl create deployment camel-app-main --image=docker.io/squakez/db-app-main:1.0
```

The application should start, but, since there is no label for the operator, this one cannot discover it.

NOTE: ideally your pipeline process should be the one in charge to include this and any other label to the applications.

Let's include the label via CLI:

```
kubectl label deployment camel-app-main camel.apache.org/app=camel-app-main
```

NOTE: you can test it straight away with any of your existing Camel application by adding the label as well.

The application is immediately imported by the operator. Its metrics are also scraped and available to be monitored:

```
kubectl get camelapps
NAME                PHASE     LAST EXCHANGE   EXCHANGE SLI   IMAGE                                  REPLICAS   INFO
camel-app-413       Running   8m32s           OK             squakez/cdb:4.13                       1          Main - 4.13.0-SNAPSHOT (4.13.0-SNAPSHOT)
```

NOTE: more information are available inspecting the custom resource (i.e. via `-o yaml`).

## Camel annotations synchronization

As you will discover in the configuraton chapter, you can provide specific configuration for each `CamelApp`. In order to keep the operator in synch with any deployment tool, you should therefore annotate the backing deployment object (ie, the `Deployment`) with such specific configuration. The operator will automatically synchronize any annotation prefixed with `camel.apache.org`.
