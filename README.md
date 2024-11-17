# rh-ocp-examples

Red Hat OpenShift application deployment examples.

## Repository structure

```txt
- services        # Services
  - number        # A demo service that generates a random number
  - view          # A demo service that displays the random number
```

## How to deploy

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
