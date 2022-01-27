<!-- MarkdownTOC autolink="true" bracket="round" depth="3" -->

- [Main concept](#main-concept)
    - [Stateful module concept](#stateful-module-concept)
    - [SiteManager Custom Resource for stateful](#sitemanager-custom-resource-for-stateful)
    - [Infra service sequence](#infra-service-sequence)
    - [Infra service dependencies](#infra-service-dependencies)
    - [Infra service endpoints](#infra-service-endpoints)
    - [Possible schemes for sm-client and site-manager](#possible-schemes-for-sm-client-and-site-manager)
    - [site-manager authorization](#site-manager-authorization)
- [REST API](#rest-api)
    - [Contract for communication with infra service](#contract-for-communication-with-infra-service)
        - [API Security model](#api-security-model)
    - [Contract for communication with site-manager](#contract-for-communication-with-site-manager)
    - [Metrics endpoint](#metrics-endpoint)
- [DR procedures flow](#dr-procedures-flow)
    - [Switchover](#switchover)
    - [Failover](#failover)
- [CLI tool sm-client](#cli-tool-sm-client)
    - [Configuration file](#configuration-file)
    - [Examples of using sm-client](#examples-of-using-sm-client)
    - [Daemon mode](#daemon-mode)

<!-- /MarkdownTOC -->

# Main concept

The main idea of the new approach is to manage operation modes of different microservices.

Common scheme:

![](/documentation/images/site-manager-SM-common-scheme.png)

In scheme with two kubernetes clusters there are `site-manager` services in each kubernetes cluster and `sm-client` - client tool to maintain the switching sequence between services in different clusters.

 - `site-manager` is the management service to control DR procedures flow in one kuberneets cluster.
 - `sm-client` - client for management of DR procedures in two or more kubernetes clusters. It can be launched as daemon or as cli util.

The modular system is introduced that will allow different behavior to be applied for different microservices. This system is more flexible and will allow us to add or exclude support for certain cases.

Currently the following modules have been added:

* `stateful` module - intended for services that will be fully managed through the service's operators.

## Stateful module concept

This module expands the functionality of operators. Operator is the internal microservice that monitors the state of infra service and listens HTTP port to receive REST API for managing infra service DR mode (active or standby). The most of infra services have operators to manage its state and we propose to expand functionality for DR cases.

![](/documentation/images/site-manager-PG-service-with-CR-new.png)

where:
  - `operator` - service that manages life-cycle of infra service, such as `postgresql`, `rabbit` or `kafka`
  - `SiteManager CR` - is the custom resource with description of DR behavior for infra service, specific for stateful module
  - `site-manager` - is the dedicated service in a separate project. It can set new state `active` or `standby` to other infra services and knows about all infra services in current kubernetes cluster.

### SiteManager Custom Resource for stateful

CR SiteManager description:

```yaml
apiVersion: netcracker.com/v2
kind: SiteManager
metadata:
  name: <SERVICE>
  namespace: <NAMESPACE>
  labels:
    app: <SERVICE>
spec:
  sitemanager:
    module: "stateful"
    after: ["<SERVICE-1>", "<SERVICE-3>"]
    before: ["<SERVICE-5>"]
    sequence: ["standby", "active"]
    timeout: <TIMEOUT-IN-SECONDS>
    allowedStandbyStateList: ["up", "down"]
    parameters:
      serviceEndpoint: "<SERVICENAME>.<NAMESPACE>.svc.cluster.local<:PORT>/sitemanager"
      healthzEndpoint: "<SERVICENAME>.<NAMESPACE>.svc.cluster.local<:PORT>/healthz"
```

where:
  - `module` is the name of the module through which the service should be controlled. 
  - `after` is the list of services, that should be done before service start. In case of `after` is empty or absent the service will start among the first services if no service with name of this service in section `before`.
  - `before` is the list of services, that should wait until service in running. May be empty or absent.
  - `sequence` is the order of starting service sides. In case of sequency is absent ot empty services on both clusters will start at the same time.
  - `timeout` is the timeout in seconds for polling operation. If `timeout` is empty or absent the default timeout 600 seconds will be used.
  - `allowedStandbyStateList` - is the list of possible healthz statuses for standby site. By default `["up"]`.
  - `serviceEndpoint` is the URL to access the operator inside kubernetes cluster.
  - `healthzEndpoint` is the URL to check cluster status.

Example for `postgres` service:

```yaml
apiVersion: netcracker.com/v2
kind: SiteManager
metadata:
  annotations:
    meta.helm.sh/release-name: postgres-service
    meta.helm.sh/release-namespace: postgres-service
  labels:
    app.kubernetes.io/managed-by: Helm
  name: postgres-service-site-manager
  namespace: postgres-service
spec:
  sitemanager:
    module: "stateful"
    after:
    - paas
    before: []
    sequence:
    - standby
    - active
    timeout: 360
    parameters:
      serviceEndpoint: postgres-operator.postgres-service.svc.cluster.local:8080/sitemanager
      healthzEndpoint: postgres-operator.postgres-service.svc.cluster.local:8080/health
```

**Important**: you must not read CRs as `v1` version, only `v2` read (default) is supported.  

## Infra service sequence

There is the situation then infra service should be reconfigured on one kubernetes cluster before the second kubernetes cluster. 

For example postgresql cluster. At first should be performed `standby` part of postgresql cluster and after `active` part:

![](/documentation/images/site-manager-PG-service-with-CR-2clusters-new.png)

Some services need different order. In case of infra service does not need to follow the sequence of execution parameter `sequence` in SiteManager CR can be omitted.

## Infra service dependencies

Often services depend on the sequence of starting another services. For example `airflow` depends of running `postgresql` cluster. To comply with dependencies of infra services you can use `after` and `before` parameters. These parameters are lists and can contain the names of CRs SiteManager of another infra services. It is impotent to understand that we dont use names of infra services, we should use names of CRs.

When all dependencies of two or more infra services are fulfilled `site-manager` can maintain all these services at the same time.

For one kubernetes cluster:

![](/documentation/images/site-manager-SM-dependency-1.png)

For two kubernetes clusters:

![](/documentation/images/site-manager-SM-dependency-2.png)

where:
  - `1` is the first set of services
  - `2` is the second set of services. It means that services of that set depend on services of set `1`

## Infra service endpoints

To start DR procedures `site-manager` should send REST request to operator. There are two parameters to define URL for operator: `serviceEndpoint` and `ingressEndpoint`. `serviceEndpoint` should be used for communication inside kubernetes cluster and define name of kuberneets service. `ingressEndpoint` should be used then operator has kubernetes ingress and can receive REST requests outside of kubernetes cluster.

To check current infra service health status `site-manager` checks URL from `healthzEndpoint`. 

## Possible schemes for sm-client and site-manager

1. Run `sm-client` as cli util on any Linux host with access to both kubernetes clusters:

    ![](/documentation/images/site-manager-SM-new-arch-1.png)

    `sm-client` starts as cli util by operator and prepare DR procedure for kubernetes clusters. All logs are in stdout and `sm-client` shows all operations in runtime. After `sm-client` finished DR procedure it exits.

2. Run `sm-client` as a service in docker container on DVM or Operation Portal and send commands by REST:

    ![](/documentation/images/site-manager-SM-new-arch-2.png)

    `sm-client` is running as service in Docker container and can receive REST queries with commands. All procedures occur in runtime. `sm-client` does not exit after all DR procedures finished and continues to listen for new REST queries. `sm-client` can be started on DVM or on Operation portal VM. To achieve HA the `sm-client` can be started on few nodes but only one should launch DR procedures at a time.


`site-manager` is the service always started inside kubernetes cluster and it has information only about infra services inside the same kubernetes cluster.

`sm-client` is the util to manage DR procedures by sending REST requests to `site-manager` microservices in few kubernetes clusters and respects dependencies and sequences between all infra services of kubernetes clusters.

## site-manager authorization

To restrict access to `site-manager` from `sm-client` there is the scheme with using authorization by Bearer:

1. In kubernetes cluster there is the serviceaccount `sm-auth-sa` without any grants in the same namespace as `site-manager`
2. `site-manager` is started with env parameter `FRONT_HTTP_AUTH` with value "True" or "Yes"
3. `site-manager` reads secret created by kubernetes for serviceaccount `sm-auth-sa` and store token in memory. Also `site-manager` uses watch mode and wait for any updates of secret. If secret was updated the `site-manager` also updates token in memory
4. Operator fills config.yml for `sm-client` with the same token and set env parameter `FRONT_HTTP_AUTH` with value "True" or "Yes"
5. All REST operations between `sm-client` and `site-manager` will be accompanied by a header "Authorization: Bearer <TOKEN>" where `TOKEN` is the token from serviceaccount `sm-auth-sa`

To secure access to manageable services from `Site-Manager` also added same scheme with using authorization by Bearer:

1. The value of env variable `BACK_HTTP_AUTH` means whether the token from serviceaccount `sm-auth-sa` will be sent to manageable services in header.
2. More about this scheme at [API Security model](#api-security-model) part.

**Note:** `site-manager` installed by default with `FRONT_HTTP_AUTH` "Yes" and `BACK_HTTP_AUTH` "Yes" which means that authorization enabled.

# REST API

## Contract for communication with infra service

This part describes which contract must implement the service so that the Site-manager can correctly interact with them.

- [`serviceEndpoint` GET](#check-infra-service-mode-and-status-of-dr-procedure)
- [`serviceEndpoint` POST](#set-new-mode-for-infra-service)
- [`healthzEndpoint` GET](#check-infra-service-healthz-of-service)

### API Security model

**Note:** This functionality is enabled by default: `BACK_HTTP_AUTH: "Yes"` in the Helm chart, later this value can be changed in the SM deployment.

To implement secure access between Site-Manager and services, the special header is added to POST and GET requests coming from Site-Manager:

```
"Authorization": "Bearer <TOKEN>"
```

In order for the service to make sure that the request is secure, on the service side it is necessary to organize the verification of this token for authenticity and belonging to Site-Manager. This is done as follows:

1) You need to make a request to verify the received token (`TokenReview`) through the `Kubernetes-client` or `Kubectl` of the following format:

```
apiVersion: authentication.k8s.io/v1
kind: TokenReview
spec:
  token: <TOKEN>
```

Where: `<TOKEN>` is a Bearer received from Site-Manager `Authorization` request header.

[Kubernetes-client TokenReview Api for Go](https://github.com/kubernetes-client/go/blob/master/kubernetes/docs/AuthenticationV1Api.md)

2) Kube-api for this request will return a response in the format:

```
apiVersion: authentication.k8s.io/v1
kind: TokenReview
spec:
  token: <TOKEN>
status:
  audiences:
  - <audience identifiers>
  authenticated: true
  user:
    groups:
    -system:serviceaccounts
    -system:serviceaccounts:site-manager
    -system:authenticated
    uid: c1a61275-608e-462e-89df-cf2a8ecc6d13
    username: system:serviceaccount:site-manager:sm-auth-sa
```

3) In this response, we are interested in the following fields:

```
  - status.authenticated = true
  - status.user.username = system:serviceaccount:site-manager:sm-auth-sa
```

Where `site-manager` is Site-Manager's Namespace name and `sm-auth-sa` is SA name.

![](/documentation/images/site-manager-http-auth.png)

More information about token can be found at [site-manager authorization](#site-manager-authorization) part.

What might be required to implement this approach:

1. The operator which processing requests from Site-Manager must have a `serviceaccount`.

2. The operator's `serviceaccount` must be wrapped in a `ClusterRoleBinding` with the `system:auth-delegator` cluster role:

```
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: replicator-auth
subjects:
- kind: ServiceAccount
  name: <SERVICE-ACCOUNT-NAME>
  namespace: <NAMESPACE>
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
```

3. Adding a handler in the operator that checks the token for GET and POST requests using `TokenReview` described earlier.

### Check infra service mode and status of DR procedure:

**URL**: `<service>.<namespace>.svc.cluster.local/sitemanager`

**Method**: `/GET`

**Code**: 200

**Answer**: `{"mode": "active|standby|disable", "status": "running|done|failed"}`

Where `mode` is the role of cluster part, `status` is the current state of DR procedure.

`mode` options:

  - `active` - set mode `active` for managed cluster. This mode is for setting part of cluster to be master or main.
  - `standby` - set mode `standby` for managed cluster. This mode is for setting part of cluster to be slave or backup.
  - `disable` - set mode `disable` for managed cluster. This mode is for maintenance procedure to stop part of cluster.

`status` options:

  - `running` means that DR procedure is in progress
  - `done` means that DR procedure is successfully done
  - `failed` means that DR procedure is failed. Operator should fix the problem by himself or restart DR procedure for specific service.

Example of `/GET` request with `curl` command:

```
$ curl --silent \
       --request GET \
       https://<service>.<namespace>.svc.cluster.local/sitemanager

```

Output:

```json
{"mode":"active","status": "running"}
```

### Set new mode for infra service:

**URL**: `<service>.<namespace>.svc.cluster.local/sitemanager`

**Method**: `/POST`

**Code**:    200

**Data Params**: `{"mode": "active|standby|disable", "no-wait": True|False}`

where:
  - `active` - set mode `active` for managed cluster. This mode is for setting part of cluster to be master or main.
  - `standby` - set mode `standby` for managed cluster. This mode is for setting part of cluster to be slave or backup.
  - `disable` - set mode `disable` for managed cluster. This mode is for maintenance procedure to stop part of cluster.
  - `no-wait` - special flag for microservice to show type of replication between of parts of database cluster. `no-wait=True` means master member of cluster should not wait replication with slave member. `no-wait=False` corresponds to only switchover DR operation. All other procedures occur with the flag `no-wait=True`.

Example of `/POST` request with `curl` command:

```
$ curl --silent \
       --request POST \
       --header "Content-Type: application/json" \
       --data '{"mode":"active"}' \
       http://<service>.<namespace>.svc.cluster.local/sitemanager

```

Output:

```json
{"mode":"active","status": "running"}
```

### Check infra service healthz of service:

**URL**: `<service>.<namespace>.svc.cluster.local/healthz`

**Method**: `/GET`

**Code**: 200

**Answer**: `{"status": "up|down|degraded"}`

where:
  - `status` - current status of the service.
      - `up` - means that the service correctly working.
      - `down` - means that the service is broken.
      - `degraded` - means that the service is partially broken.

Example of `/GET` request with `curl` command:

```
$ curl --silent \
       --request GET \
       http://<service>.<namespace>.svc.cluster.local/healthz

```

Output:

```json
{"status":"up"}
```

## Contract for communication with site-manager

This part describes what requests a Site-manager can respond to on both kubernetes clusters.

- [Site-manager GET](#get-dict-with-all-services-and-settings)
- [Site-manager POST `{"procedure": "status"}`](#check-infra-service-mode-and-status-of-dr-procedure)
- [Site-manager POST `{"procedure": "list"}`](#show-list-of-infra-services)
- [Site-manager POST `{"procedure": "active|standby|disable"}`](#start-dr-procedure-active-standby-or-disable)
- [Possible errors](#possible-errors)

### Get dict with all services and settings:

**URL**: `site-manager.example.com/sitemanager`

**Method**: `/GET`

**Code**: 200

**Answer**: 

```json
{
  "services": {
    "kafka": {
      "after": [], 
      "before": [], 
      "healthzEndpoint": "http://kafka-disaster-recovery.kafka-service.svc.cluster.local:8068/healthz", 
      "namespace": "kafka-service", 
      "sequence": [
        "active", 
        "standby"
      ], 
      "serviceEndpoint": "http://kafka-disaster-recovery.kafka-service.svc.cluster.local:8068/sitemanager", 
      "standbyState": "up", 
      "timeout": 360
    },
    "spark-operator-gcp-site-manager": {
      "after": [
        "paas"
      ], 
      "before": [], 
      "healthzEndpoint": "http://spark-site-manager.spark-operator-gcp.svc.cluster.local:8080/health", 
      "namespace": "spark-operator-gcp", 
      "sequence": [
        "standby", 
        "active"
      ], 
      "serviceEndpoint": "http://spark-site-manager.spark-operator-gcp.svc.cluster.local:8080/sitemanager", 
      "standbyState": "up", 
      "timeout": 120
    }
  }
}
```

This dict is compiled CRs of all services managed by `site-manager` and contains all necessary parameters and default values.


Example of `/GET` request with `curl` command:

```
$ curl --silent \
       --request GET \
       https://site-manager.example.com/sitemanager

```

Output:

```json
{
  "services": {
    "kafka": {
      "after": [], 
      "before": [], 
      "healthzEndpoint": "http://kafka-disaster-recovery.kafka-service.svc.cluster.local:8068/healthz", 
      "namespace": "kafka-service", 
      "sequence": [
        "active", 
        "standby"
      ], 
      "serviceEndpoint": "http://kafka-disaster-recovery.kafka-service.svc.cluster.local:8068/sitemanager", 
      "standbyState": "up", 
      "timeout": 360
    },
    "spark-operator-gcp-site-manager": {
      "after": [
        "paas"
      ], 
      "before": [], 
      "healthzEndpoint": "http://spark-site-manager.spark-operator-gcp.svc.cluster.local:8080/health", 
      "namespace": "spark-operator-gcp", 
      "sequence": [
        "standby", 
        "active"
      ], 
      "serviceEndpoint": "http://spark-site-manager.spark-operator-gcp.svc.cluster.local:8080/sitemanager", 
      "standbyState": "up", 
      "timeout": 120
    }
  }
}
```

### Check infra service mode and status of DR procedure:

**URL**: `site-manager.example.com/sitemanager`

**Method**: `/POST`

**Code**: 200

**Data Params**: `{"procedure": "status"}`

**Answer**: `[{"name": "service-name-1", "mode": "active|standby|disable", "status": "running|done|failed"}, {"name": "service-name-2", "mode": "active|standby|disable", "status": "running|done|failed"}]`

This command shows current status of DR procedures and results of health checks.

Example of `/GET` request with `curl` command shows output for services `paas` and `paas-1`:

```
$ curl -XPOST \
       --header "Content-Type: application/json" \
       -d '{"procedure":"status", "run-services": ["paas","paas-1"]}' \
       http://site-manager.example.com/sitemanager

```

Output:

```json
{
  "running-procedure": "", 
  "services": {
    "paas": {
      "healthz": "--", 
      "mode": "active", 
      "status": "done"
    }, 
    "paas-1": {
      "healthz": "up", 
      "mode": "active", 
      "status": "done"
    }
  }
}
```

### Show list of infra services:

**URL**: `site-manager.example.com/sitemanager`

**Method**: `/POST`

**Code**: 200

**Data Params**: `{"procedure": "list"}`

**Answer**: `{"all_services": ["service-1", "service-2"], "running-services": ["service-1", "service-2"]}`

Example:

```
$ curl -XPOST \
       --header "Content-Type: application/json" \
       -d '{"procedure":"list"}' \
       http://site-manager.k8s-1.openshift.sdntest.netcracker.com/sitemanager

```

Output:

```json
{
  "all-services": [
    "postgres", 
    "postgres-service-site-manager", 
    "paas", 
    "kafka", 
    "mongo", 
    "streaming-platform", 
    "paas-1", 
    "spark-operator-gcp-site-manager"
  ], 
  "running-services": [
    "postgres", 
    "postgres-service-site-manager", 
    "paas", 
    "kafka", 
    "mongo", 
    "streaming-platform", 
    "paas-1", 
    "spark-operator-gcp-site-manager"
  ]
}
```

### Start DR procedure `active`, `standby` or `disable`:

**URL**: `site-manager.example.com/sitemanager`

**Method**: `/POST`

**Code**: 200

**Data Params**: `{"procedure": "active|standby|disable"}`

**Answer**: `{"message": "Procedure active is started", "procedure": "active", "services": ["paas","paas-1"]}`

Example:

```
$ curl -XPOST \
       --header "Content-Type: application/json" \
       -d '{"procedure":"active", "run-services": "paas,paas-1"}' \
       http://site-manager.example.com/sitemanager

```

Output:

```json
{
  "message": "Procedure active is started", 
  "procedure": "active", 
  "services": [
    "paas", 
    "paas-1"
  ]
}
```

### Possible errors:

1. Wrong procedure:

```
$ curl -XPOST \
       --header "Content-Type: application/json" \
       -d '{"procedure":"wrong", "run-services": "paas,paas-1"}' \
       http://site-manager.example.com/sitemanager
```

Output:

```json
{
  "message": "You should define procedure from list: ['active', 'standby', 'disable', 'list', 'status']"
}
```

HTTP Code: 401

2. Start new procedure when previous is running:

```
$ curl -XPOST \
       --header "Content-Type: application/json" \
       -d '{"procedure":"active", "run-services": "paas,paas-1"}' \
       http://site-manager.example.com/sitemanager
```

Output:

```json
{
  "message": "Process active is still running. Procedure active rejected!"
}
```

HTTP Code: 409

3. Try to start procedure with incorrect services:

```
$ curl -XPOST \
       --header "Content-Type: application/json" \
       -d '{"procedure":"active", "run-services": "paas,paas-12"}' \
       http://site-manager.example.com/sitemanager
```

Output:

```json
{
  "message": "You defined service that does not exist in cluster", 
  "wrong-services": [
    "paas-12"
  ]
}

```

HTTP Code: 401

## Metrics endpoint

To check metricks from running `site-manager` can be used `/metrics` endpoint. Output has prometheus specific format and intended for external monitoring system.

```
curl   --silent \
       --request GET \
       https://site-manager.example.com/metrics
```

# DR procedures flow

## Switchover

Example for service with `sequence` parameter as ["standby", "active"]:

![](/documentation/images/site-manager_diagram_with_PG.png)

Postgres is the service with defined sequence. It means that we need to send new mode to the standby cluster and only in case of procedure successful finished we can send new mode to active site.

Example for service without `sequence` parameter:

![](/documentation/images/site-manager_diagram_with_Mongo.png)

For this example we consider Mongo as an example of service without sequence. For this case we send new mode to both clusters and wait for status `done` also for both clusters.

## Failover

Example for two services. The first service with defined `sequence` and the second without sequence:

![](/documentation/images/site-manager_diagram_failover.png)

Working with services looks like in switchover procedure. We have only one assumption - standby kubernetes cluster may be unavailable. In this case we should not wait for correct status. We can omit statuses of all infra services of standby kubernetes cluster.

# CLI tool sm-client

`sm-client` is the cli tool to manage DR procedures between infra services that deployed in two kubernetes clusters. It can be started on any Linux host with installed python dependencies: on DVM host or operation portal. Another option is start `sm-client` in docker container. In this case `sm-client` can work even in kubernetes cluster if needed.

The main idea of `sm-client` is control of sequence of DR procedures for every infra service for both kubernetes clusters. It reads CRs with type sitemanager in kubernetes clusters and creates tree with correct order of infra services. `sm-client` waits for successfull finish of every DR procedures and starts next.

`sm-client` help section:

```
$ ./sm-client --help
usage: sm-client [-h] [-v] [-c CONFIG] [-f] [--run-services RUN_SERVICES] [--skip-services SKIP_SERVICES] {move,stop,return,mntc,active,standby,list,status,daemon} ...

Script to manage DR cases in kubernetes Active-Standby scheme

How to use commands:

  +------------------------------+        +------------------------------+
  |      INITIAL CONDITION       |        |        FINAL CONDITION       |
  +--------------+---------------+--------+--------------+---------------+-----+---------+
  | ACTIVE SITE  | STANDBY SITE  |  ===>  | ACTIVE SITE  | STANDBY SITE  |     | COMMAND |
  +--------------+---------------+  ===>  +--------------+---------------+     +---------+
  | ACTIVE       | STANDBY       |  ===>  | STANDBY      | ACTIVE        |  =  | move    |
  | failed       | STANDBY       |  ===>  | stopped      | ACTIVE        |  =  | stop    |
  | stopped      | ACTIVE        |  ===>  | STANDBY      | ACTIVE        |  =  | return  |
  | ACTIVE       | stopped       |  ===>  | ACTIVE       | STANDBY       |  =  | return  |
  | ACTIVE       | STANDBY       |  ===>  | ACTIVE       | stopped       |  =  | mntc    |
  +--------------+---------------+--------+--------------+---------------+-----+---------+

positional arguments:
  {move,stop,return,mntc,active,standby,list,status,daemon}
    move                move Active functionality to Standby site
    stop                excludes site from Active-Standby scheme
    return              return stopped kuberneets cluster to Standby role
    mntc                stop Standby kubernetes cluster for maintenance
    active              set kubernetes cluster services to active mode
    standby             set kubernetes cluster services to standby mode
    list                list all services from Active-Standby scheme managed by site-manager with dependencies
    status              show current status of clusters and all services
    daemon              run site-manager as a service

optional arguments:
  -h, --help            show this help message and exit
  -v, --verbose         enable the verbosity mode
  -c CONFIG, --config CONFIG
                        define the path to configuration file
  -f, --force           force apply DR action and ignore healthz
  --run-services RUN_SERVICES
                        define the list of services to apply DR action, by default all services participate
  --skip-services SKIP_SERVICES
                        define the list of services what will not participate in DR action
  -k, --insecure        enable self-signed certificates
```

where:
  - `move site` is the action for switchoves. Both sites are working and you need to switch active site to new. Site in command will be `active` after apply.
  - `stop site` is the action for failover. This commands uses in case of `active` site is failed. Site in command will be `standby` after apply.
  - `return site` is the action for switching on `standby` site after failover. Site in command will be `standby` after apply. This action applied to only one site.
  - `mntc site` is the action to switch microservices of site to mode `disable`. Site in command will be `disable` after apply. This action applied to only one site.
  - `active site` is the action to switch site to `active` mode. This action applied to only one site.
  - `standby site`  is the action to switch site to `standby` mode. This action applied to only one site.
  - `list` is the action to list all or part of microservices of sites.
  - `status` is the action to show all or part of microservices status.

## Configuration file

The main confuration file for `sm-client` in short format looks like this:

```yaml
---
sites:
  - name: k8s-1
    token: <TOKEN>
    site-manager: http://site-manager.k8s-1.openshift.sdntest.netcracker.com/sitemanager
  - name: k8s-2
    token: <TOKEN>
    site-manager: http://site-manager.k8s-2.openshift.sdntest.netcracker.com/sitemanager

sm-client:
  logging-url: <URL to server to collect logs>
  logging-dir: <Directory to collect logs>
  logging-dir-size: 256
  bind-address: 0.0.0.0
  bind-port: 8080
```

where:
 - `sites` is the list of kuberneets clusters.
 - `sm-client` is the section for specific settings for `sm-client` as logging, listening port and ip in daemon mode and other.

## Examples of using sm-client

Failover of cluster k8s-1:

```
./sm-client stop k8s-1
```

Switchover to cluster k8s-1:

```
./sm-client move k8s-1

...
---------------------------------------------------------------------
Summary:
services that successfully done: ['sm-test', 'paas', 'postgres', 'mongo']
services that failed: []
services that ignored: []
---------------------------------------------------------------------
```

Switchover to cluster k8s-1 with skipping paas and mongo services:

```
./sm-client --skip-services paas,mongo move k8s-1

---------------------------------------------------------------------
Procedure: move

Active site is: k8s-1
Standby site is: k8s-2

Kubernetes services managed by site-manager: ['sm-test', 'paas', 'postgres', 'mongo']
Kuberneets services that will be processed: ['sm-test', 'postgres']
---------------------------------------------------------------------

......

---------------------------------------------------------------------
Summary:
services that successfully done: ['sm-test', 'postgres']
services that failed: []
services that ignored: ['paas', 'mongo']
---------------------------------------------------------------------
```

Stop standby cluster k8s-1 for maintenance:

```
./sm-client mntc k8s-1
```

Check status of services:

```
$ ./sm-client status
+---------------------------------+---------------------+-----------------------+
| Service                         |        k8s-1        |         k8s-2         |
+---------------------------------+---------------------+-----------------------+
| postgres-service-site-manager   | standby / done / up |   active / done / up  |
| kafka                           | standby / done / up |   active / done / up  |
| paas                            | standby / done / up |   active / done / up  |
+---------------------------------+---------------------+-----------------------+
```

Show list of services in kubernetes clusters with CR sitemanager:

```
./sm-client list

---------------------------------------------------------------------
Kubernetes services managed by site-manager: ['postgres', 'sm-test', 'mongo', 'paas']
Kuberneets services that will be processed: ['postgres', 'sm-test', 'mongo', 'paas']
---------------------------------------------------------------------
```

## Daemon mode

To start `sm-client` in daemon mode you should run:

```
$ ./sm-client daemon

 * Serving Flask app "sm-client" (lazy loading)
 * Environment: production
   WARNING: This is a development server. Do not use it in a production deployment.
   Use a production WSGI server instead.
 * Debug mode: off

2021-04-23 17:31:40,328 [INFO] _internal.py:  * Running on http://0.0.0.0:8080/ (Press CTRL+C to quit)
```

In daemon mode `sm-client` starts to listen TCP port 8080 and can receive HTTP REST requests. All output will be in logging yet.

To send REST API requests to `sm-client` you should comply with the following contract:

**URL**: `site-manager.example.com/sitemanager`

**Method**: `/POST`

**Code**: 200

**Data Params**: `{"procedure": "active|standby|disable", "site": NAME_OF_SITE_FROM_CONFIG, "run-services": LIST_OF_SERVICES, "skip-services": LIST_OF_SERVICES, "force": "True|False"}`

**Answer**: `{"message": "Procedure active is started", "procedure": "active", "services": ["SERVICE-1","SERVICE-2"]}`

where:
 - `procedure` is the command to execute by `sm-client`. It can be `move`, `stop`, `return`, `mntc`, `active`, `standby`, `list` or `status`. Mandatory parameter
 - `site` is the name of site. Should be the same as one of sites from `config.yml`. Mandatory parameter
 - `run-services` parameter in JSON object defines comma separated list of services for processing. Optional parameter.
 - `skip-services` parameter in JSON object defines comma separated list of services to be skipped for processing. Optional parameter.
 - `force` parameter defines ability to ignore `healthz` status different to `up`. Supported following values: "True", "true", 1. Optional parameter.

**Note:** `run-services` and `skip-services` can not be used in the same  command.

Example:

```
$ curl --silent \
       --request POST \
       --header "Content-Type: application/json" \
       --data '{"procedure": "standby", "site": "k8s-2", "run-services": "paas,postgres", "force": "true"}' \
       http://site-manager.example.com/sitemanager
```

In case of any of procedures is processing all new procedures will be rejected. New procedure will be applied only in waiting mode of `sm-client`.
