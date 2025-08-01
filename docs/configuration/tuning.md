---
id: tuning
title: Configure Camel application
permalink: /configuration/tuning
---

There are several configuration you can apply separately to each of your Camel application. They mostly work at general level (setting an environment variable on the operator) or at application level (setting an annotation on the Deployment resource).

## Configure the metrics polling

You can watch the metrics evolving as long as the application is running, for example via `-w` parameter:

```
kubectl get camelapps -w

NAME                PHASE     LAST EXCHANGE   EXCHANGE SLI   IMAGE                                  REPLICAS   INFO
...
camel-app-413       Running   8m32s           OK             squakez/cdb:4.13                       1          Main - 4.13.0-SNAPSHOT (4.13.0-SNAPSHOT)
camel-app-main      Running                   OK             docker.io/squakez/db-app-main:1.0      1          Main - 4.11.0 (4.11.0)
camel-app-quarkus   Running                   Warning        docker.io/squakez/db-app-quarkus:1.0   1
camel-app-sb        Running                   Error          docker.io/squakez/db-app-sb:1.0        1          Spring-Boot - 3.4.3 (4.11.0)
```

The `CamelApp` are polled every minute by default. It should be enough in most cases, as the project is really a dashboard and not a proper monitoring tool. However, you can change this configuration if you want a more or less reactive polling. You can configure this value both at operator level (which would affect all the applications) or at single application level.

### Operator level

You can setup the environment variables `POLL_INTERVAL_SECONDS` with the number of seconds between each metrics polling.

NOTE: this will affect all your applications. Setting it a low value can reduce the performances of both the operator and the same Camel applications which will need to use compute resources to read from the HTTP service.

### Application level

You can add an annotation to the `Deployment` resource, `camel.apache.org/polling-interval-seconds` with the value you want.

NOTE: although this configuration will only affect the single application, consider the right balance to avoid affecting the application performances.

## Configure the SLI Exchange error and warning percentage

The operator is in charge to automatically calculate the success rate percentage of exchanges in the last polling interval time. It has some default configuration and will return a `Success`, `Warning` or `Error` status if it detects that the failure of exchanges during the interval exceeds the thresholds. It returns an `Error` when the failure exceed the 5% of exchanges failed, `Warning` if the failure is above 10%, `Success`. However, these values can be configured.

### Operator level

You can setup the environment variables `SLI_ERR_PERCENTAGE` and `SLI_WARN_PERCENTAGE`. It requires an `int` value.

### Application level

You can add an annotation to the `Deployment` resource, `camel.apache.org/sli-exchange-error-percentage` and `camel.apache.org/sli-exchange-warning-percentage` with the value expected for that specific application only.

## Configure the observability services port

The operator is able to discover applications thanks to the presence of the `camel-observability-services` component. By default this component exposes the metrics on port `9876` (which is also the operator default if you don't configure it). However this value can be changed by the user to any other port (including the regular business service port). You can configure is both at Operator or Application level.

### Operator level

You can setup the environment variables `OBSERVABILITY_PORT` with the number of the port where the operator has to get the metrics.

### Application level

You can add an annotation to the `Deployment` resource, `camel.apache.org/observability-services-port` with the value expected for that specific application only.