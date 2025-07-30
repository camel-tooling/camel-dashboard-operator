---
id: install
title: How to install the operator
permalink: /installation/
---

You can use [Helm](https://helm.sh) to install the operator resources. You can install it in any namespace (we conventionally use `camel-dashboard` namespace, which, has to be created previously). The default configuration is for a cluster scoped operator (use `--set operator.global=\"false\"` for a namespace scoped operator).

```
helm install camel-dashboard https://github.com/camel-tooling/camel-dashboard-operator/raw/refs/heads/main/docs/charts/camel-dashboard-0.0.1.tgz -n camel-dashboard
```

NOTE: the installation procedure is still in alpha phase. Verify the latest release and change the version (ie, `0.0.1) from the previous script accordingly.

You can check if the operator is running:

```
kubectl get pods -n camel-dashboard
NAME                                        READY   STATUS    RESTARTS   AGE
camel-dashboard-operator-7c6bcf5576-fwn7s   1/1     Running   0          4m18s
```