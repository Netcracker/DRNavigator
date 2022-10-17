![DR Navigator_1280х640_3](https://user-images.githubusercontent.com/5212888/162979551-96cf5038-2015-4b45-9e02-94ed1a8cbd3f.png)

<!-- MarkdownTOC autolink="true" bracket="round" depth="3" -->
- [DRNavigator](#)
- [Prepare kubernetes cluster to work with site-manager](#prepare-kubernetes-cluster-to-work-with-site-manager)
- [Deployment with Restricted Rights](#deployment-with-restricted-rights)

<!-- /MarkdownTOC -->

# DRNavigator

DRNavigator is the project to manage applications in two or more kubernetes clusters.

# Prepare kubernetes cluster to work with DRNavigator

To support ability of services be managed by `site-manager` you should prepare following steps:

1. Create namespace `site-manager`:

    ```bash
    $ kubectl create namespace site-manager
    ```

2. Generate self-signed certificates for `site-manager` service if you don't integrate with cert-manager

    2.1. Create configuration file for generation SSL certificate:

    ```bash
    $ cat <<EOF > server.conf 
    [req]
    req_extensions = v3_req
    distinguished_name = req_distinguished_name
    prompt = no

    [req_distinguished_name]
    CN = site-manager.site-manager.svc

    [ v3_req ]
    basicConstraints = CA:FALSE
    keyUsage = nonRepudiation, digitalSignature, keyEncipherment
    extendedKeyUsage = clientAuth, serverAuth
    subjectAltName = @alt_names
    [alt_names]
    DNS.1 = site-manager
    DNS.2 = site-manager.site-manager
    DNS.3 = site-manager.site-manager.svc
    EOF
    ```

    2.2. Create CA certificate:

    ```bash
    $ openssl req -days 730 -nodes -new -x509 -keyout ca.key -out ca.crt -subj "/CN=SM service"
    ```

    2.3. Create KEY for `site-manager` service:

    ```bash
    $ openssl genrsa -out site-manager-tls.key 2048
    ```

    2.4. Create CRT file for `site-manager`:

    ```bash
    $ openssl req -new -key site-manager-tls.key -subj "/CN=site-manager.site-manager.svc" -config server.conf | \
      openssl x509 -req -days 730 -CA ca.crt -CAkey ca.key -CAcreateserial -out site-manager-tls.crt -extensions v3_req -extfile server.conf
    ```

3. Create CRD `sitemanagers.netcracker.com`

    3.1. In case of integration with cert-manager you should add following annotation in crd, that helps update caBundle in it's webhook:

    ```
    apiVersion: apiextensions.k8s.io/v1
        kind: CustomResourceDefinition
        metadata:
            name: sitemanagers.netcracker.com
            annotations:
                cert-manager.io/inject-ca-from: <NAMESPACE>/site-manager-tls-certificate
    ```

     Create CRD `sitemanagers.netcracker.com` without caBundle field:

    ```
    $ cat manifests/crd-sitemanager.yaml | sed "/caBundle/d" | kubectl apply -f -
    ```

    If you've already had site-manager crd in your cloud and want to migrate to cert-manager integration, it's enough to annotate it:

    ```
    $ kubectl annotate crds sitemanagers.netcracker.com cert-manager.io/inject-ca-from=<NAMESPACE>/site-manager-tls-certificate
    ```
    
    3.2. In other case generate base64 string from ca.crt certificate:

    ```
    $ CA_BUNDLE=$(cat ca.crt | base64 - | tr -d '\n')
    ```

    Create CRD `sitemanagers.netcracker.com`:

    ```
    $ cat manifests/crd-sitemanager.yaml | sed "s/<base-64-encoded-ca-bundle>/${CA_BUNDLE}/" | kubectl apply -f -
    ```

4. Create secret with SSL certificates for `site-manager` if you don't integrate with cert-manager

    ```
    $ kubectl -n site-manager create secret tls sm-certs --cert site-manager-tls.crt --key site-manager-tls.key
    ```

    In case of cert-manager integration it will be created automatically during helm chart installation

5. Install `site-manager` helm chart:

    ```bash
    $ helm install site-manager charts/site-manager/ -n site-manager
    ```

    The `site-manager` helm chart can be customized with following parameters:

    | Parameter                   | Description                                                                                       | Default value                   |
    |---------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------|---------------------------------|
    | env.FRONT_HTTP_AUTH         | set authentication mode between sm-client and Site-Manager                                        | "Yes"                           |
    | env.BACK_HTTP_AUTH          | set authentication mode between Site-Manager and manageable services                              | "Yes"                           |
    | env.SM_DEBUG                | set `debug` logging level                                                                         | "False"                         |
    | env.SM_GROUP                | define API group for CRD                                                                          | "netcracker.com"                |
    | env.SM_PLURAL               | define object of API group                                                                        | "sitemanagers"                  |
    | env.SM_VERSION              | define API group version for CRD                                                                  | "v2"                            |
    | env.SERVICE_DEFAULT_TIMEOUT | set default timeout for every microservice DR procedure                                           | 200                             |
    | env.HTTP_SCHEME             | define HTTP scheme for connection to microservice operator                                        | "http://"                       |
    | env.SM_CACERT               | TLS verification in operators (True, False or path to trusted CA file)                            | "True"                          |
    | workerCount                 | count of parallel workers, that handle requests                                                   | 2                               |
    | serviceAccount.create       | enable/disable Service Account creation                                                           | true                            |
    | serviceAccount.name         | name of Service Account for `site-manager`                                                        | "site-manager-sa"               |
    | image.repository            | docker image repository name                                                                      | ghcr.io/netcracker/site-manager |
    | image.pullPolicy            | docker image pull policy                                                                          | Always                          |
    | image.tag                   | docker image tag                                                                                  | v1.0                            |
    | ingress.create              | enable/disable ingress creation                                                                   | true                            |
    | ingress.name                | define URL for `site-manager` ingress                                                             | ""                              |
    | paas_platform               | define PAAS type. It can be "kubernetes" or "openshift"                                           | "kubernetes"                    |
    | tls.generateCerts.enabled   | enable/disable certificates generation via cert-manager                                           | false                           |
    | tls.generateCerts.clusterIssuerName | define name cluster issuer, if you wand to use it (if empty, will be created self-signed issuer ) | ""                              |
    | tls.generateCerts.duration          | define duration (days) of created certificate via cert-manager                                    | 365                             |
    | tls.generateCerts.subjectAlternativeName.additionalDnsNames    | additional trusted dns names in certificate                                                       | []                              |
    | tls.generateCerts.subjectAlternativeName.additionalIpAddresses | additional trusted ips names in certificate                                                       | []                             |
    | paasGeoMonitor | see [paas-geo-monitor documentation](/paas-geo-monitor/docs) | |

6. Install `site-manager` to OpenShift

    ```
    $ helm install site-manager charts/site-manager/
                  -n site-manager \
                  --set image.repository=ghcr.io/netcracker/site-manager \
                  --set image.tag=<image tag> \
                  --set paas_platform=openshift \
                  --set ingress.name=site-manager.apps.example.com
    ```

    where:
      - `ingress.name` parameter is mandatory for OpenShift
      - `paas_platform` should be set to "openshift"
      - `<image tag>` image tag reference
