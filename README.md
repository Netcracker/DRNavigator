
<!-- MarkdownTOC autolink="true" bracket="round" depth="3" -->
- [DRNavigator](#)
- [Prepare kubernetes cluster to work with site-manager](#prepare-kubernetes-cluster-to-work-with-site-manager)
- [Deployment with Restricted Rights](#deployment-with-restricted-rights)

<!-- /MarkdownTOC -->

# DRNavigator

DRNavigator is the project to manage applications in two or more kubernetes clusters.

# Prepare kubernetes cluster to work with DRNavigator

To support ability of services be managed by `site-manager` you should prepare following steps:

1. Generate self-signed certificates for `site-manager` service

    Create configuration file for generation SSL certificate:

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

    Create CA certificate:

    ```bash
    $ openssl req -nodes -new -x509 -keyout ca.key -out ca.crt -subj "/CN=SM service"
    ```

    Create KEY for `site-manager` service:

    ```bash
    $ openssl genrsa -out site-manager-tls.key 2048
    ```

    Create CRT file for `site-manager`:

    ```bash
    $ openssl req -new -key site-manager-tls.key -subj "/CN=site-manager.site-manager.svc" -config server.conf | \
      openssl x509 -req -CA ca.crt -CAkey ca.key -CAcreateserial -out site-manager-tls.crt -extensions v3_req -extfile server.conf
    ```

2. Create CRD `sitemanagers.netcracker.com`

    Generate base64 string from ca.crt certificate:

    ```
    cat ca.crt | base64 - | tr -d '\n'
    ```

    Open file `manifests/crd-sitemanager.yaml` end insert output of previous command to `caBundle` parameter:

    ```
    ...
      conversion:
        strategy: Webhook
        webhook:
        conversionReviewVersions: ["v1"]
        clientConfig:
            service:
            namespace: site-manager
            name: site-manager
            path: /convert
            caBundle: <INSERT-BASE64-OUTPUT-HERE>
    ```

    Create CRD `sitemanagers.netcracker.com`:

    ```bash
    $ kubectl apply -f manifests/crd-sitemanager.yaml
    ```

3. Create namespace `site-manager`:

    ```bash
    $ kubectl create namespace site-manager
    ```

4. Create secret with SSL certificates for `site-manager`

    ```
    $ kubectl -n site-manager create secret tls sm-certs --cert site-manager-tls.crt --key site-manager-tls.key
    ```

5. Install `site-manager` helm chart:

    ```bash
    $ helm install site-manager charts/site-manager/ -n site-manager
    ```

The `site-manager` helm chart can be customized with following parameters:

| Parameter | Description                                                  | Default value |
|-----------|--------------------------------------------------------------|---------------|
| env.FRONT_HTTP_AUTH | set authentication mode between sm-client and Site-Manager | "Yes" |
| env.BACK_HTTP_AUTH | set authentication mode between Site-Manager and manageable services | "Yes" |
| env.SM_DEBUG | set `debug` logging level                                    | "False" |
| env.SM_GROUP | define API group for CRD                                     | "netcracker.com" |
| env.SM_PLURAL | define object of API group                                   | "sitemanagers" |
| env.SM_VERSION | define API group version for CRD                             | "v2" |
| env.SERVICE_DEFAULT_TIMEOUT | set default timeout for every microservice DR procedure      | 200 |
| env.HTTP_SCHEME | define HTTP scheme for connection to microservice operator   | "http://" |
| serviceAccount.create | enable/disable Service Account creation                      | true |
| serviceAccount.name | name of Service Account for `site-manager`                   | "site-manager-sa" |
| image.repository | docker image repository name                                 | ghcr.io/netcracker/site-manager |
| image.pullPolicy | docker image pull policy                                     | Always |
| image.tag | docker image tag                                             | v1.0 |
| ingress.create | enable/disable ingress creation                              | true |
| ingress.name | define URL for `site-manager` ingress                        | "" |
