---
id: oc-console
title: Camel Openshift Console Plugin
permalink: /plugins/oc-console
carousels:
  - images: 
    - image: images/oc-console-list.png
    - image: images/oc-console-detail.png
    - image: images/oc-console-detail-resources.png
    - image: images/oc-console-detail-metrics.png
---

This operator can work standalone and you can use the data exposed in the `CamelApp` custom resource accordingly. However it has a great fit with the [Camel Openshift Console Plugin](https://github.com/camel-tooling/camel-openshift-console-plugin?tab=readme-ov-file#deployment-to-openshift), which is a visual representation of the services exposed by the operator.


## Camel Openshift Console Plugin dependencies matrix

The Camel Openshift Console Plugin is an extension of OpenShift Console exposing the data from the Camel Dashboard Operator.

Below you can find the compatibility list for its dependencies:

| Camel Openshift Console Plugin | Openshift          | Camel Dashboard Operator |
| ------------------------------ | ------------------ | ------------------------ |
| next (0.2.1)                   | Openshift 4.19     | 0.1.0                    |
| 0.2.0                          | Openshift 4.19     | 0.1.0                    |
| 0.1.0                          | Openshift 4.18     | 0.0.1                    |


NOTE: the old version 0.1.0 uses the old `App` custom resource.

## Installation


A [Helm](https://helm.sh) chart is available to deploy the plugin to an OpenShift environment.

The following Helm parameters are required:

`plugin.image:` The location of the image containing the plugin that was previously pushed

Additional parameters can be specified if desired. Consult the chart values file for the full set of supported parameters.

### Installing the Helm Chart

Install the chart using the name of the plugin as the Helm release name into a new namespace or an existing namespace as specified by the camel-openshift-console-plugin parameter and providing the location of the image within the `plugin.image` parameter by using the following command:

```
helm upgrade -i camel-openshift-console-plugin https://github.com/camel-tooling/camel-openshift-console-plugin/raw/refs/heads/release-1.0.x/docs/charts/camel-openshift-console-plugin-0.2.0.tgz --namespace camel-dashboard --set plugin.image=quay.io/camel-tooling/camel-openshift-console-plugin:0.2.0
```

NOTE: the installation procedure is still in alpha phase. Verify the latest release and change the version (ie, `0.2.0`) from the previous script accordingly.

## Overview

  {% include carousel.html height="50" unit="%" duration="7" number="1" %}

