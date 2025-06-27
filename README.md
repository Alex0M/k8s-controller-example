# Kubernetes Controller for Syncing Third-Party API State

This repository demonstrates how to build a Kubernetes controller that integrates with a third-party API, synchronizes its state periodically, and stores it in a custom Kubernetes Custom Resource Definition (CRD). It also shows how another controller can act upon the synced data to manage Kubernetes-native resources like `ConfigMap` and `Deployment`.

---

## Overview

The system is composed of two main controllers:

1. **FrontendSync Controller** 
   - `internal/controller/frontendsync_controller.go`
   - Periodically queries a third-party API using parameters from a user-defined custom resource `FrontendSync`.
   - Creates or updates a `FrontendPage` custom resource to store the state retrieved from the API.

2. **FrontendPage Controller**
   - `internal/controller/frontendpage_controller.go` 
   - Watches for changes to the `FrontendPage` CRD.
   - Generates a `ConfigMap` and a `Deployment` based on the data in the `FrontendPage` resource.

---

## Components

### 1. Third-Party API (Example)

An external API server returns structured data that the Sync Controller fetches.
`./cmd/fe-config-api/main.go`

### 2. FrontendSync Controller

- **Custom Resource:** `FrontendSync`
- **Fields:**
  - `url`: API endpoint to fetch data from.
  - `syncInterval`: Interval for syncing with the third-party API.
- **Behavior:**
  - Fetches data from API on the specified interval.
  - Writes data into a `FrontendPage` CR.

### 3. FrontendPage Controller

- **Custom Resource:** `FrontendPage`
- **Fields:**
  - `image`: Container image to be used in the generated Deployment
  - `replicas`: Number of replicas for the Deployment
  - `contents`: Contents to be written into the ConfigMap
- **Behavior:**
  - Creates, updates, or deletes a `ConfigMap` with the fetched content.
  - Creates, updates, or deletes a Kubernetes `Deployment` using the fetched container image and specified number of replicas.