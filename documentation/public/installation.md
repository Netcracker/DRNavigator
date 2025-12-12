# DR Navigator Installation Procedure

<!-- TOC -->
* [DR Navigator Installation Procedure](#dr-navigator-installation-procedure)
  * [Site-manager](#site-manager)
    * [Requirements](#requirements)
    * [Prerequisites](#prerequisites)
    * [Installation](#installation)
    * [Certificate Renewal Procedure](#certificate-renewal-procedure)
  * [Paas-Geo-Monitor](#paas-geo-monitor)
    * [Installation Procedure](#installation-procedure)
    * [Configuration](#configuration)
  * [sm-client](#sm-client)
    * [Prepare Environment](#prepare-environment)
    * [Running From CLI](#running-from-cli)
    * [Running From Docker](#running-from-docker)
<!-- TOC -->

## Site-manager

### Requirements

Before the installation, ensure that you have enough resources for DRNavigator.
For one pod by default DRNavigator uses 20m CPU and 100Mi memory.
You can override cpu/memory limits using helm environments.

Needed quotas depends on the number of services, for this reason, if you have problems with services processing (typically, with timeouts),
you can do one of following solutions:

- Increase timeouts for site-manager and sm-client;
- Increase quotas for site-manager;

### Prerequisites

1. Generate self-signed certificates for the `site-manager` service if you do not want to integrate with cert-manager or OpenShift service serving certificates mechanism.

    2.1. Create a configuration file for generating the SSL certificate:

    ```bash
    cat <<EOF > server.conf 
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
    IP.1 = 127.0.0.1
    DNS.1 = site-manager
    DNS.2 = site-manager.site-manager
    DNS.3 = site-manager.site-manager.svc
    DNS.4 = <specify there ingress name of site-manager>
    EOF
    ```

    **Important**: Do not forget to specify any other IP addresses and DNS names that you plan to use to connect to the site-manager.
    For example, it is required to specify the ingress name (`ingress.name` value from helm installation) to use this certificate
    to connect outside the cloud (from sm-client).  
    For this, specify the additional `DNS.#` and `IP.#` fields.

    2.2. Create the CA certificate:

    ```bash
    openssl req -days 730 -nodes -new -x509 -keyout ca.key -out ca.crt -subj "/CN=SM service"
    ```

    2.3. Create KEY for the `site-manager` service:

    ```bash
    openssl genrsa -out site-manager-tls.key 2048
    ```

    2.4. Create CRT file for `site-manager`:

    ```bash
    openssl req -new -key site-manager-tls.key -subj "/CN=site-manager.site-manager.svc" -config server.conf | \
    openssl x509 -req -days 730 -CA ca.crt -CAkey ca.key -CAcreateserial -out site-manager-tls.crt -extensions v3_req -extfile server.conf
    ```
    2.5. Specify data from generated files in site-manager chart under `tls.crt`, `tls.key` and `tls.ca` sections respectively. 

### Installation

   Install `site-manager` helm chart:

   ```bash
   helm install site-manager charts/site-manager/ -n site-manager
   ```

   The `site-manager` helm chart can be customized with the following parameters:

| Parameter                                                      | Description                                                                                                                                                              | Default value                   |
|----------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------|
| crd.install                                                    | Enable/disable site-manager CRD installation                                                                                                                             | true                           |
| env.FRONT_HTTP_AUTH                                            | Set the authentication mode between sm-client and Site-Manager.                                                                                                          | "Yes"                           |
| env.BACK_HTTP_AUTH                                             | Set the authentication mode between Site-Manager and manageable services.                                                                                                | "Yes"                           |
| env.SM_DEBUG                                                   | Set `debug` logging level.                                                                                                                                               | "False"                         |
| env.SM_GROUP                                                   | Define API group for CRD.                                                                                                                                                | "legacy.qubership.org"                |
| env.SM_KIND                                                    | Define the kind of API group.                                                                                                                                            | "SiteManager"                   |
| env.SM_KIND_LIST                                               | Define the kind  for list API group.                                                                                                                                     | "SiteManagerList"               |
| env.HTTP_SCHEME                                                | Define the HTTP scheme for connection to microservice operator.                                                                                                          | "http://"                       |
| env.SM_CACERT                                                  | TLS verification in operators (True, False or path to trusted CA file).                                                                                                  | "True"                          |
| env.SM_GET_REQUEST_TIMEOUT                                     | Timeout for GET requests: service status and health.                                                                                                                     | 10                              |
| env.SM_POST_REQUEST_TIMEOUT                                    | Timeout for POST requests: service procedures.                                                                                                                           | 30                              |
| workerCount                                                    | The count of parallel workers that handle requests.                                                                                                                      | 2                               |
| serviceAccount.create                                          | Enable/disable Service Account creation.                                                                                                                                 | true                            |
| serviceAccount.name                                            | The name of Service Account for `site-manager`.                                                                                                                          | "site-manager-sa"               |
| image.repository                                               | The docker image repository name.                                                                                                                                        | ghcr.io/netcracker/site-manager |
| image.pullPolicy                                               | The docker image pull policy.                                                                                                                                            | Always                          |
| image.tag                                                      | The docker image tag.                                                                                                                                                    | v1.0                            |
| ingress.create                                                 | Enable/disable ingress creation.                                                                                                                                         | true                            |
| ingress.name                                                   | Define URL for `site-manager` ingress.                                                                                                                                   | ""                              |
| ingress.className                                              | Define class name for ingress.                                                                                                                                           | ""                              |
| limits.cpu                                                     | CPU limits per pod.                                                                                                                                                      | 200m                            |
| limits.memory                                                  | Memory limits per pod.                                                                                                                                                   | 160Mi                           |
| requests.cpu                                                   | CPU requests per pod.                                                                                                                                                    | 100m                            |
| requests.memory                                                | Memory requests per pod.                                                                                                                                                 | 75Mi                            |
| affinity                                                       | Define affinity rules.                                                                                                                                                   | {}                              |
| PAAS_PLATFORM                                                  | Define PAAS type. It can be "KUBERNETES" or "OPENSHIFT".                                                                                                                 | "KUBERNETES"                    |
| paasGeoMonitor                                                 | Refer to [paas-geo-monitor documentation](#paas-geo-monitor).                                                                                                            |                                 |
| priorityClassName                                              | The Priority Class Name for site-manager and paas-geo-monitor deployments                                                                                                | ""                              |
| smSecureAuth                                                   | The mode for SM authorization with dr-services. See [API Security Model](architecture.md#api-security-model) for details                                                 | false                           |
| customAudience                                                 | Custom audience for rest api token, that is used to connect with services. Worked only if `smSecureAuth=true`                                                            | "sm-services"                   |
| tls.enabled                                                    | Enable https in ingress/route                                                                                                                                            | true                            |
| tls.defaultIngressTls                                          | Use default tls certificate instead of generated one for ingress/route                                                                                                   | false                           |
| tls.ca                                                         | CA tls certificate (content of `ca.crt` file after [prerequisites](#prerequisites) step 2). Required, if integration with cert-manager is disabled                       | ""                              |
| tls.crt                                                        | SM public tls certificate (content of `site-manager-tls.crt` file after [prerequisites](#prerequisites) step 2). Required, if integration with cert-manager is disabled  | ""                              |
| tls.key                                                        | SM private tls certificate (content of `site-manager-tls.key` file after [prerequisites](#prerequisites) step 2). Required, if integration with cert-manager is disabled | ""                              |
| tls.generateCerts.executor                                     | Choose executor for certificates generation. Certificate generation is enabled automatically if custom crt/key/ca are not provided. Possible values: "cert-manager" and "openshift"                                                                             | cert-manager                    |
| tls.generateCerts.clusterIssuerName                            | In case of cert-manager integration, define the cluster name issuer if required (if empty, it is created by a self-signed issuer).                                       | ""                              |
| tls.generateCerts.duration                                     | In case of cert-manager integration, define the duration (days) of created certificate using cert-manager.                                                               | 365                             |
| tls.generateCerts.subjectAlternativeName.additionalDnsNames    | In case of cert-manager integration, additional trusted DNS names in the certificate.                                                                                    | []                              |
| tls.generateCerts.subjectAlternativeName.additionalIpAddresses | In case of cert-manager integration, additional trusted IP names in the certificate.                                                                                     | []                              |

   **Warning**: Some parameters (e.g. `tls.ca`, `tls.crt` and `tls.key`) have multiline values in common cases. To override them, you
   can use `--set-file` helm option or separate values yaml file with multiline yaml strings, like:

   ```yaml
   tls:
     crt: |
       -----BEGIN CERTIFICATE-----
       ...
       -----END CERTIFICATE-----
     crt: |
       -----BEGIN CERTIFICATE-----
       ...
       -----END CERTIFICATE-----
     key: |
       -----BEGIN RSA PRIVATE KEY-----
       ...
       -----END RSA PRIVATE KEY-----
   ```

   Or you may install `site-manager` by specifying additional arguments:

   ```bash
   $ helm install site-manager charts/site-manager/
                 -n site-manager \
                 --set image.repository=ghcr.io/netcracker/site-manager \
                 --set image.tag=<image tag> \
                 --set PAAS_PLATFORM=OPENSHIFT \
                 --set-file tls.ca=<path to ca.crt> \
                 --set-file tls.crt=<path to site-manager-tls.crt> \
                 --set-file tls.key=<path to site-manager-tls.key> \
                 --set ingress.name=site-manager.apps.example.com
   ```

   Where:

- `ingress.name` parameter is mandatory for OpenShift.
- `PAAS_PLATFORM` should be set to "OPENSHIFT".
- `<image tag>` is the image tag reference.
- `<path to ca.crt>`, `<path to site-manager-tls.crt>`, `<path to site-manager-tls.key>` is paths to tls files, that you generated using `openssl` command.

### Certificate Renewal Procedure

To renew a certificate:

1. Execute the instructions in the **Prerequisites** section from the step 2.2 to 3.2;
2. Redeploy SM with new certificate parameters: `tls.ca`, `tls.crt` and `tls.key`;
3. Restart pod `site-manager`

    ```bash
    kubectl rollout restart deployment site-manager -n site-manager
    ```

4. Update `config.yml` (Token and path cacert).

## Paas-Geo-Monitor

### Installation Procedure

The installation is done using a [helm chart](/charts/site-manager). By default, this service is not installed.
To install the service, it is required to enable it and provide information
about peers, for example:

```yaml
paasGeoMonitor:
  install: true
  config:
    peers:
      - name: cluster-2
        clusterIp:
          name: paas-geo-monitor.site-manager.svc.cluster-2.local
```

By default, paas-geo-monitor pings peers status in background process and print warnings if any problems occurs.
Additionally, it save peers status in following metrics:

- `peer_dns_status` is responsible for DNS resolution;  
- `peer_svs_status` is responsible for pod-to-service connection;  
- `peer_pod_status` is responsible for pod-to-pod connection;

You can disable this using `PAAS_PING_PEERS` parameter or increase time between ping attempts using `PAAS_PING_TIME`.

Full parameters list:

| Parameter                                             | Description                                                        | Default value                             |
|-------------------------------------------------------|--------------------------------------------------------------------|-------------------------------------------|
| paasGeoMonitor.install                                | Enable/disable paas-geo-monitor installation                       | false                                     |
| paasGeoMonitor.image                                  | Paas-geo-monitor image                                             | ghcr.io/netcracker/paas-geo-monitor:0.9.0 |
| paasGeoMonitor.config                                 | Peers configuration. See format on [Configuration](#configuration) | peers: []                                 |
| paasGeoMonitor.env.PAAS_PING_PEERS                    | Enable/disable ping peers status. Required for monitoring          | true                                      |
| paasGeoMonitor.env.PAAS_PING_TIME                     | Time between ping peers status attempts (in sec)                   | 5                                         |
| paasGeoMonitor.env.PAAS_DEBUG                         | Enable/disable debug logs                                          | false                                     |

### Configuration

The service configuration allows you to configure instance-wide options and
peers.

**Note**: `PING_IP` env variable should be set, otherwise service will not work. See [ping endpoint](architecture.md#ping).

| Field                                                      | Description                                             |
|------------------------------------------------------------|---------------------------------------------------------|
| **port**<br/>*int*                                         | The port that the service should listen. Default 8080.  |
| **peers**<br/>*array [Peer](architecture.md#peers-status)* | Array of peer service <br/>instances in other clusters. |

## sm-client

### Prepare Environment

- Create `config.yml` using the following commands:

```yaml
    sites:
      - name: k8s-1
        token: <BEARER TOKEN>
        site-manager: http://site-manager.example.com/sitemanager
        cacert: <path-to-ca-certificate>
      - name: k8s-2
        token: <BEARER TOKEN>
        site-manager: http://site-manager.example.com/sitemanager
        cacert: <path-to-ca-certificate>
    
    sm-client:
      http_auth: True
```

Where,

- `<BEARER TOKEN>` should be taken from the `sm-auth-sa` service-account. It might be specified in two ways
  - endless token from `sm-auth-sa-token` secret, that can be obtained with following command:
  ```shell
  kubectl get secret sm-auth-sa-token -n site-manager -o yaml | grep token: | cut -d ' ' -f4 | base64 --decode
  ```
  - temporary token generated with specified duration using command:
  ```shell
  kubectl create token -n site-manager sm-auth-sa --duration=1h
  ```

- cacert is a content of `ca.crt` which has been generated during the SiteManager installation.

### Running From CLI

1. On the Linux host, install python with version 3.9 or newer.
2. Copy or clone files from <https://github.com/Netcracker/DRNavigator> to a separate directory. For example, `/opt/sm-client`
3. Install python libs:

    ```shell
    pip3 install -r requirements-sc.txt
    ```

4. Start script `sm-client` with the `status` argument to check the status of clusters and services.

### Running From Docker

1. Pull sm-client docker image, e.g.:

    ```bash
    docker pull ghcr.io/netcracker/sm-client
    ```

2. Run SM client from docker with mounted config.yaml and site-manager certificate files (in that case paths to files
from config.yaml and sm-client cli command should be specified for docker container file system), e.g. for status:

```bash
docker run \
--mount type=bind,source="$(pwd)"/config.yaml,target=/data/config.yaml \
--mount type=bind,source="$(pwd)"/CA-site1.cer,target=/secrets/CA-site1.cer \
--mount type=bind,source="$(pwd)"/CA-site2.cer,target=/secrets/CA-site2.cer \
sm-client ./sm-client -c /data/config.yaml status
```
