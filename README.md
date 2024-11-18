# rh-ocp-examples

Red Hat OpenShift application deployment examples.

## Repository structure

```txt
- services        # Services
  - number        # A demo service that generates a random number
  - view          # A demo service that displays the random number
```

## Prerequisites

- OpenShift >= 4.17 cluster
- Enabled GitOps subscription in the OpenShift cluster (ArgoCD)
- `oc` CLI installed and configured to access the OpenShift cluster

## How to deploy

Before we can deploy any of the services, we need to modify our central ArgoCD management application to allow sourcing `Application` resources from other than the `openshift-gitops` namespace. This is necessary because the services are deployed in a separate project, namespace and ArgoCD application.

To do this, execute the following patch command:

```bash
oc -n openshift-gitops patch argocds.argoproj.io openshift-gitops \
  --type merge \
  --patch '{"spec":{"sourceNamespaces":["openshift-gitops","example-application-basic","example-application-progressive"]}}'
```

The `openshift-gitops` namespace is the `default` namespace and might not be necessary to add to the `sourceNamespaces` list. However, it is added here for completeness.

### Deploy the services

1. Clone the repository

```bash
git clone git@github.com:leonsteinhaeuser/rh-ocp-examples.git
cd rh-ocp-examples
```

2. Deploy the management app and project

```bash
oc apply --server-side -f ./
```

3. Wait for ArgoCD to sync the other project setup manifests stored in the `ocp/deployments` directory

4. Wait for the ArgoCD services and applications to be created
