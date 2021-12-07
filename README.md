
<!-- MarkdownTOC autolink="true" bracket="round" depth="3" -->
- [DRNavigator](#)
- [Prepare kubernetes cluster to work with site-manager](#prepare-kubernetes-cluster-to-work-with-site-manager)
- [Deployment with Restricted Rights](#deployment-with-restricted-rights)

<!-- /MarkdownTOC -->

# DRNavigator

DRNavigator is the project to manage applications in two or more kubernetes clusters.

# Prepare kubernetes cluster to work DRNavigator

To support ability of services be managed by `site-manager` you should prepare following steps:

1. Create CRD `sitemanagers.netcracker.com` from [YAML description file](manifests/crd-sitemanager.yaml)

2. Create namespace `site-manager`:

    ```bash
    $ kubectl create namespace site-manager
    ```

3. Install helm chart:

    ```bash
    $ helm install site-manager charts/site-manager/ -n site-manager
    ```

The `site-manager` helm chart can be customized with following parameters:

| Parameter | Description | Default value |
|-----------|-------------|---------------|
| env.SM_HTTP_AUTH | set athetication mode | "Yes" |
| env.SM_WEB_HOST | set IP address for WEB service | "0.0.0.0" |
| env.SM_WEB_PORT | set port for WEB service | 8080 |
| env.SM_DEBUG | set `debug` logging level | "False" |
| env.SM_GROUP | define API group for CRD | "netcracker.com" |
| env.SM_PLURAL | define object of API group  | "sitemanagers" |
| env.SM_VERSION | define API group version for CRD | "v2" |
| env.SERVICE_DEFAULT_TIMEOUT | set default timeout for every microservice DR procedure | 200 |
| env.HTTP_SCHEME | define HTTP scheme for connection to microservice operator | "http://" |
| serviceAccount.create | enable/disable Service Account creation | true |
| serviceAccount.name | name of Service Account for `site-manager` | "site-manager-sa" |
| image.repository | docker image repository name | ghcr.io/netcracker/site-manager |
| image.pullPolicy | docker image pull policy | Always |
| image.tag | docker image tag | v1.0 |
| ingress.create | enable/disable ingress creation | true |
| ingress.name | define URL for `site-manager` ingress | "" |

