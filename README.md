# Periodic Reconcile Controller Example

This project demonstrates how to build a simple Kubernetes controller that periodically fetches data from a third-party API and uses it to create or update a `Deployment` and `ConfigMap`.

## Overview

We define a custom resource (CRD) with two fields:

- `url`: Endpoint of the third-party API.
- `interval`: Sync interval (in seconds or duration string, e.g., `30s`, `1m`).

The controller watches these custom resources and periodically performs the following:

1. Connects to the third-party API using the URL from the CR.
2. Fetches and parses the response.
3. Creates or updates a `Deployment` and a `ConfigMap` based on the data.
4. Repeats this at the specified interval.

## CRD Example

```yaml
apiVersion: frontendpage.alex0m.io/v1alpha1
kind: FrontendPage
metadata:
  name: testpage
  namespace: default
spec:
  url: "https://k8s-controller.free.beeceptor.com"
  sync: 30
