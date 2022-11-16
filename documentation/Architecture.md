![Kubemarine_1280х640_3_JPEG](https://user-images.githubusercontent.com/5212888/162978291-63d55f19-7dc0-4126-ad39-cd69191e7e19.jpg)

<!-- TOC -->
* [Overview](#overview)
* [Managing Services](#managing-services)
  * [Service Sequence](#service-sequence)
  * [DR Procedures Flow](#dr-procedures-flow)
  * [Possible Schemes for sm-client and Site-Manager](#possible-schemes-for-sm-client-and-site-manager)
* [Security. Authorization](#security-authorization)
* [Monitoring](#monitoring)
* [SiteManager Contract](#sitemanager-contract)
  * [Custom Resource for Stateful](#custom-resource-for-stateful)
  * [REST API Definition](#rest-api-definition)
* [SiteManager API](#sitemanager-api)
* [sm-client](#sm-client)
* [Installation Procedure](#installation-procedure)
  * [SiteManager](#sitemanager)
  * [smclient](#smclient)

<!-- TOC -->
# Overview

The main purpose of DRNavigator is to manage services in a 2-cluster (Kubernetes or OpenShift) environment in the Active-StandBy scheme.  

The services can be any microservices that implement a special [contract](#sitemanager-contract).

There are several DR (Disaster Recovery) procedures that can be used. The basic are: 
* **Switchover** - swap Active and StandBy roles   
* **Failover** - move Active role to StandBy cluster 

The DRNavigator contains two components:
* `site-manager` - It is the management service to control the DR procedure flow on one cluster.
* `sm-client` - It is the client for management of DR procedures in two clusters. It can be launched as a container or as a cli util.

For more information, see [DR Operations](#sm-client).

**Common Scheme**

![](/documentation/images/site-manager-SM-common-scheme.png)

In a scheme with two clusters, there is `site-manager` that operates in each cluster and `sm-client` that is a client tool to manage the DR procedures.

# Managing Services

A modular system is implemented that allows different behavior to be applied for different microservices. 

The `stateful` module is supported by default. It is intended for services that are fully managed by the service's operators. This module expands the functionality of operators. An operator is an internal microservice that monitors the state of the service and listens to the HTTP port to receive a REST API for managing the service DR mode (active or standby). Most of the services have operators to manage their state, and it is proposed to expand the functionality for DR cases.

![](/documentation/images/site-manager-PG-service-with-CR-new.png)

Where:
  - `operator` is a service that manages the life-cycle of a service, such as `postgresql`, `rabbit`, or `kafka`.
  - `SiteManager CR` is the custom resource with description of the DR behavior for a service, specific for the stateful module.
  - `site-manager` is a dedicated service in a separate project. It can set a new state `active` or `standby` to other services and contains information about all services in the current Kubernetes cluster.

## Service Sequence

Service sequence is applied when a service should be reconfigured on one Kubernetes cluster before the second Kubernetes cluster (for 
example postgresql cluster), and when `standby` should be performed before `active`.

![](/documentation/images/site-manager-PG-service-with-CR-2clusters-new.png)

Some services need a different order. In case of a service does not need to follow the sequence of execution, the `sequence` parameter in the SiteManager CR can be omitted.

### Service Dependency

Often services depend on the sequence of starting other services. For example, `airflow` depends on running the `postgresql` cluster. To comply with dependencies of services, you can use `after` and `before` parameters. These parameters are lists and can contain the names of SiteManager CRs of other services. It is important to understand that we do not use names of services, we should use names of CRs.

When all dependencies of two or more services are fulfilled, `site-manager` can maintain all these services at the same time.

For one Kubernetes cluster:

![](/documentation/images/site-manager-SM-dependency-1.png)

For two Kubernetes clusters:

![](/documentation/images/site-manager-SM-dependency-2.png)

Where:
  - `1` is the first set of services.
  - `2` is the second set of services. It specifies that the services of that set depend on the services of set `1`.

### Service Endpoint

To start DR procedures, `site-manager` should send a REST request to the operator. You should use `serviceEndpoint` to define the URL for the operator.  

To check the health status of the current service, `site-manager` checks the URL from `healthzEndpoint`.

## DR Procedures Flow
### Switchover

Example for a service with the `sequence` parameter as ["standby", "active"]:

![](/documentation/images/site-manager_diagram_with_PG.png)

Postgres is the service with a defined sequence. It specifies that we have to send a new mode to the standby cluster and only when a procedure has successfully finished, we can send a new mode to the active site.

Example for a service without the `sequence` parameter:

![](/documentation/images/site-manager_diagram_with_Mongo.png)

For this example, Mongo is assumed as a service without a sequence. For this case, we send a new mode to both clusters and wait for the `done` status for both clusters.

### Failover

The following image is an example for two services, where the first service is with `sequence` and the second is without a sequence:

![](/documentation/images/site-manager_diagram_failover.png)

Working with services is similar to the switchover procedure. We have only one assumption - standby Kubernetes cluster can be unavailable. In this case, we should not wait for the correct status. We can omit statuses of all services of the standby Kubernetes cluster.

## Possible Schemes for sm-client and Site-Manager

1. Run `sm-client` as a cli util on any Linux host with access to both Kubernetes clusters:

    ![](/documentation/images/site-manager-SM-new-arch-1.png)

    `sm-client` starts as a cli util by an operator and prepares the DR procedure for Kubernetes clusters. All logs are in stdout and `sm-client` shows all operations in runtime. The `sm-client` exits after finishing the DR procedure.

2. Run `sm-client` as a service in the docker container on any container environment and send commands by REST:

    ![](/documentation/images/site-manager-SM-new-arch-2.png)

    `sm-client` runs as a service in the Docker container and can receive REST queries with commands. All procedures occur in runtime. `sm-client` does not exit after all DR procedures are finished and continues to listen for new REST queries. `sm-client` can be started on DVM or on the Operation portal VM. To achieve HA, the `sm-client` can be started on few nodes, but only one should launch DR procedures at a time.

`site-manager` is a service that is always started inside the Kubernetes cluster, and it has information only about services inside the same Kubernetes cluster.

`sm-client` is the util to manage DR procedures by sending REST requests to `site-manager` microservices in few Kubernetes clusters and respects dependencies and sequences between all services of Kubernetes clusters.

# Security. Authorization

To restrict access to `site-manager` from `sm-client`, there is a scheme for using authorization by Bearer Token:

1. In the Kubernetes cluster, the `sm-auth-sa` serviceaccount is available without any grants in the same namespace as `site-manager`.
2. `site-manager` is started with the `FRONT_HTTP_AUTH` env parameter  with value "True" or "Yes".
3. `site-manager` reads the secret created by Kubernetes for the `sm-auth-sa` serviceaccount and stores the token in memory. Also, `site-manager` uses the watch mode and waits for any updates of the secret. If the secret is updated, the `site-manager` also updates the token in the memory.
4. The operator fills config.yml for `sm-client` with the same token and sets the `FRONT_HTTP_AUTH` env parameter with value "True" or "Yes".
5. All REST operations between `sm-client` and `site-manager` contain the "Authorization: Bearer <TOKEN>" header, where, `TOKEN` is the token from serviceaccount `sm-auth-sa`.

To secure access to manageable services from `site-manager`, the same scheme is also added for using authorization by Bearer Token:

1. The value of the `BACK_HTTP_AUTH` env variable specifies whether the token from the `sm-auth-sa` serviceaccount is sent to manageable services in the header.
2. For more information about this scheme, see [API Security Model](#api-security-model).

**Note**: `site-manager` is installed by default with `FRONT_HTTP_AUTH` "Yes" and `BACK_HTTP_AUTH` "Yes", which means that authorization is enabled.

# Monitoring

To check metrics from running `site-manager` can be used `/metrics` endpoint. Output has prometheus specific format and intended for external monitoring system.

```
curl   --silent \
       --request GET \
       https://site-manager.example.com/metrics
```
# SiteManager Contract
The following API should be implemented by Cloud Service to be operated by SiteManager:
* Custom Resource 
* REST API

##  Custom Resource for stateful

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
  - `sequence` is the order of starting service sides. In case sequence is empty default `["standby","active"]` is used.
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
## REST API definition

This part describes which contract must implement the service so that the SiteManager can correctly interact with them.

- [`serviceEndpoint` GET](#check-service-mode-and-status-of-dr-procedure)
- [`serviceEndpoint` POST](#set-new-mode-for-service)
- [`healthzEndpoint` GET](#check-service-healthz-of-service)

### API Security model

**Note:** This functionality is enabled by default: `BACK_HTTP_AUTH: "Yes"` in the Helm chart, later this value can be changed in the SM deployment.

To implement secure access between SiteManager and services, the special header is added to POST and GET requests coming from SiteManager:

```
"Authorization": "Bearer <TOKEN>"
```

In order for the service to make sure that the request is secure, on the service side it is necessary to organize the verification of this token for authenticity and belonging to SiteManager. This is done as follows:

1) You need to make a request to verify the received token (`TokenReview`) through the `Kubernetes-client` or `Kubectl` of the following format:

```
apiVersion: authentication.k8s.io/v1
kind: TokenReview
spec:
  token: <TOKEN>
```

Where: `<TOKEN>` is a Bearer received from SiteManager `Authorization` request header.

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

Where `site-manager` is SiteManager's Namespace name and `sm-auth-sa` is SA name.

![](/documentation/images/site-manager-http-auth.png)

More information about token can be found at [SiteManager authorization](#security-authorization) part.

What might be required to implement this approach:

1. The operator which processing requests from SiteManager must have a `serviceaccount`.

2. The operator's `serviceaccount` must be wrapped in a `ClusterRoleBinding` with the `system:auth-delegator` cluster role:

```
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: <CRB-NAME>
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

### Check service mode and status of DR procedure:

**URL**: `<service>.<namespace>.svc.cluster.local/sitemanager`

**Method**: `/GET`

**Code**: 200

**Answer**: `{"mode": "active|standby|disable", "status": "running|done|failed", "message": "some string"}`

Where `mode` is the role of cluster part, `status` is the current state of DR procedure, `message` is optional field to describe process details.

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
{"mode":"active","status": "running", "message": "Activation process"}
```

### Set new mode for service:

**Important**: The contract between the SiteManager and manageable services for changing the mode implies that the Rest API of the service has the property of [idempotency](https://restfulapi.net/idempotent-rest-apis/), which means that changing the mode of the service to an already existing mode does not bring the service into an inconsistent state. If the service is already moving to the required mode, the response should contain the corresponding mode and the status equal to `running`.

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

### Check service healthz of service:

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

# SiteManager API

This part describes what requests a SiteManager can respond to on both kubernetes clusters.

- [SiteManager GET](#get-dict-with-all-services-and-settings)
- [SiteManager POST `{"procedure": "status"}`](#check-services-mode-and-status-of-dr-procedure)
- [SiteManager POST `{"procedure": "list"}`](#show-list-of-services)
- [SiteManager POST `{"procedure": "active|standby|disable"}`](#start-dr-procedure-active-standby-or-disable)
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
      "allowedStandbyStateList": [
        "up"
      ],
      "before": [], 
      "module": "stateful",
      "namespace": "kafka-service", 
      "parameters": {
        "healthzEndpoint": "http://kafka-disaster-recovery.kafka-service.svc.cluster.local:8068/healthz", 
        "serviceEndpoint": "http://kafka-disaster-recovery.kafka-service.svc.cluster.local:8068/sitemanager"
      },
      "sequence": [
        "active", 
        "standby"
      ],  
      "timeout": 360
    },
    "spark-operator-gcp-site-manager": {
      "after": [
        "paas"
      ], 
      "allowedStandbyStateList": [
        "up"
      ],
      "before": [],
      "module": "stateful", 
      "namespace": "spark-operator-gcp", 
      "parameters": {
        "healthzEndpoint": "http://spark-site-manager.spark-operator-gcp.svc.cluster.local:8080/health", 
        "serviceEndpoint": "http://spark-site-manager.spark-operator-gcp.svc.cluster.local:8080/sitemanager"
      },
      "sequence": [
        "standby", 
        "active"
      ],  
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
      "allowedStandbyStateList": [
        "up"
      ],
      "before": [], 
      "module": "stateful",
      "namespace": "kafka-service", 
      "parameters": {
        "healthzEndpoint": "http://kafka-disaster-recovery.kafka-service.svc.cluster.local:8068/healthz", 
        "serviceEndpoint": "http://kafka-disaster-recovery.kafka-service.svc.cluster.local:8068/sitemanager"
      },
      "sequence": [
        "active", 
        "standby"
      ],  
      "timeout": 360
    },
    "spark-operator-gcp-site-manager": {
      "after": [
        "paas"
      ], 
      "allowedStandbyStateList": [
        "up"
      ],
      "before": [],
      "module": "stateful", 
      "namespace": "spark-operator-gcp", 
      "parameters": {
        "healthzEndpoint": "http://spark-site-manager.spark-operator-gcp.svc.cluster.local:8080/health", 
        "serviceEndpoint": "http://spark-site-manager.spark-operator-gcp.svc.cluster.local:8080/sitemanager"
      },
      "sequence": [
        "standby", 
        "active"
      ],  
      "timeout": 120
    }
  }
}
```

### Check services mode and status of DR procedure:

**URL**: `site-manager.example.com/sitemanager`

**Method**: `/POST`

**Code**: 200

**Data Params**: `{"procedure": "status", "run-service": "service-name"}`

**Answer**: `"service": {[{"healthz": "up|down|degraded", "name": "service-name-1", "message": "some-message", "mode": "active|standby|disable", "status": "running|done|failed"}]}`

This command shows current status of DR procedures and results of health checks.

Example of `/GET` request with `curl` command shows output for service `paas`:

```
$ curl -XPOST \
       --header "Content-Type: application/json" \
       -d '{"procedure":"status", "run-service": "paas"}' \
       http://site-manager.example.com/sitemanager

```

Output:

```json
{
  "services": [
    { "paas": {
          "healthz": "--", 
          "message": "",
          "mode": "active", 
          "status": "done"
      }
    }
  ]
}
```

### Show list of services:

**URL**: `site-manager.example.com/sitemanager`

**Method**: `/POST`

**Code**: 200

**Data Params**: `{"procedure": "list"}`

**Answer**: `{"all_services": ["service-1", "service-2"]}`

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
  ]
}
```

### Start DR procedure `active`, `standby` or `disable`:

**URL**: `site-manager.example.com/sitemanager`

**Method**: `/POST`

**Code**: 200

**Data Params**: `{"procedure": "active|standby|disable", "run-service": "paas"}`

**Answer**: `{"message": "Procedure active is started", "procedure": "active", "service": "paas"}`

This command performs the specified procedure for the selected service.


Example:

```
$ curl -XPOST \
       --header "Content-Type: application/json" \
       -d '{"procedure":"active", "run-service": "paas"}' \
       http://site-manager.example.com/sitemanager

```

Output:

```json
{
  "message": "Procedure active is started", 
  "procedure": "active", 
  "service": "paas"
}
```

### Possible errors:

1. Wrong json:

```
$ curl -XPOST \
       --header "Content-Type: application/json" \
       -d '{"procedure":"status", "run-service": }' \
       http://site-manager.example.com/sitemanager
```

Output:

```json
{
  "message": "No valid JSON data was received"
}
```

HTTP Code: 400

2. Wrong procedure:

```
$ curl -XPOST \
       --header "Content-Type: application/json" \
       -d '{"procedure":"wrong", "run-service": "paas"}' \
       http://site-manager.example.com/sitemanager
```

Output:

```json
{
  "message": "You should define procedure from list: ['active', 'standby', 'disable', 'list', 'status']"
}
```

HTTP Code: 400

3. Don't define run-service or define it as non string type:

```
$ curl -XPOST \
       --header "Content-Type: application/json" \
       -d '{"procedure":"active"}' \
       http://site-manager.example.com/sitemanager
```

Output:

```json
{
  "message": "run-service value should be defined and have String type"
}
```

HTTP Code: 400

4.  Try to start procedure with incorrect service:

```
$ curl -XPOST \
       --header "Content-Type: application/json" \
       -d '{"procedure":"active", "run-service": "wrong-service"}' \
       http://site-manager.example.com/sitemanager
```

Output:

```json
{
    "message": "You defined service that does not exist in cluster",
    "wrong-service": "wrong-service"
}

```

HTTP Code: 400
# sm-client

`sm-client` is the cli tool to manage DR procedures between services that deployed in two kubernetes clusters. It can be started on any Linux host with installed python dependencies: on DVM host or operation portal. Another option is start `sm-client` in docker container. In this case `sm-client` can work even in kubernetes cluster if needed.

The main idea of `sm-client` is control of sequence of DR procedures for every  service for both kubernetes clusters. It reads CRs with type SiteManager in kubernetes clusters and creates tree with correct order of services. `sm-client` waits for successful finish of every DR procedures and starts next.

`sm-client` help section:

```
$ ./sm-client --help
usage: sm-client [-h] [-v] [-c CONFIG] [-f] [--run-services RUN_SERVICES] [--skip-services SKIP_SERVICES] {move,stop,return,disable,active,standby,list,status,daemon} ...

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
  | ACTIVE       | STANDBY       |  ===>  | ACTIVE       | stopped       |  =  | disable |
  +--------------+---------------+--------+--------------+---------------+-----+---------+

positional arguments:
  {move,stop,return,disable,active,standby,list,status}
    move                move Active functionality to Standby site
    stop                excludes site from Active-Standby scheme
    return              return stopped kubernetes cluster to Standby role
    disable             stop Standby kubernetes cluster for maintenance
    active              set kubernetes cluster services to active mode
    standby             set kubernetes cluster services to standby mode
    list                list all services from Active-Standby scheme managed by SiteManager with dependencies
    status              show current status of clusters and all services

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
  - `move site` is the action for **switchover**. Both sites are working, and you need to switch active site to new. Site in command will be `active` after apply.
  - `stop site` is the action for **failover**. This commands uses in case of `active` site is failed. Site in command will be `standby` after apply.
  - `return site` is the action for switching on `standby` site after failover. Site in command will be `standby` after apply. This action applied to only one site.
  - `disable site` is the action to switch microservices of site to mode `disable`. Site in command will be `disable` after apply. This action applied to only one site.
  - `active site` is the action to switch site to `active` mode. This action applied to only one site.
  - `standby site`  is the action to switch site to `standby` mode. This action applied to only one site.
  - `list` is the action to list all or part of microservices of sites.
  - `status` is the action to show all or part of microservices status and print them in the list ordered by dependencies.

## Configuration file

The main configuration file for `sm-client` in short format looks like this:

```yaml
---
sites:
  - name: k8s-1
    token: <TOKEN>
    site-manager: http://site-manager.k8s-1.example.com/sitemanager
    cacert: <path-to-ca-certificate>
  - name: k8s-2
    token: <TOKEN>
    site-manager: http://site-manager.k8s-2.example.com/sitemanager
    cacert: <path-to-ca-certificate>

sm-client:
  http_auth: True
```

where:
 - `sites` is the list of kubernetes clusters.
 - `name` is the short name of cluster
 - `token` is the token to have access to `site-manager` in kubernetes cluster
 - `cacert` is the path to CA certificate for `site-manager` with self-signed certificate
 - `http_auth` defines to use token for `site-manager` authorization

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
kubernetes services that will be processed: ['sm-test', 'postgres']
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
./sm-client disable k8s-1
```

Check status of services:

```
$ ./sm-client status

+---------------------------------+--------------------------------------+--------------------------------------+
| Service                         |        k8s-1                         |         k8s-2                        |
+---------------------------------+--------------------------------------+--------------------------------------+
|                                 | mode | DR status | healthz | message | mode | DR status | healthz | message |
|   --------------------------    |      --------------------------      |      --------------------------      |
| postgres-service-site-manager   | standby / done / up / --             |  active / done / up / --             |
| kafka                           | standby / done / up / some kafka msg |  active / done / up / some kafka msg |
| paas                            | standby / done / up / some paas msg  |  active / done / up / some paas msg  |
+---------------------------------+--------------------------------------+--------------------------------------+
```

Show list of services in kubernetes clusters with CR SiteManager:

```
./sm-client list

---------------------------------------------------------------------
Kubernetes services managed by site-manager: ['postgres', 'sm-test', 'mongo', 'paas']
kubernetes services that will be processed: ['postgres', 'sm-test', 'mongo', 'paas']
---------------------------------------------------------------------
```
## Custom modules support
It is possible to make custom DR flow (Switchover/Failover; Active/Standby/Disable) sequence based on [module of DR service](#managing-services)

```yaml
spec:
  sitemanager:
    module: "custom_module"
```
Optional section ```flow``` needs to be provided. 
It describes the sequence of modules with appropriate DR states which needs to be run during DR operation.  
For example:
```yaml
---
sites:
  - name: k8s-1
    token: <TOKEN>
    site-manager: http://site-manager.k8s-1.example.com/sitemanager
    cacert: <path-to-ca-certificate>
  - name: k8s-2
    token: <TOKEN>
    site-manager: http://site-manager.k8s-2.example.com/sitemanager
    cacert: <path-to-ca-certificate>
sm-client:
  http_auth: True
  
flow:
  - custom_module: [standby,disable]
  - stateful:
  - custom_module: [active]
```
The above example implies the following DR sequences:
* Switchover
1. Standby all `custom_module` services
2. Standby, Active for all `stateful` services , according to [DR sequence](#dr-procedures-flow)
3. Active all `custom_module` services
* Failover
1. Standby, Active`stateful` services 
2. Active `custom_module` services
* Active
1. Active `stateful` services 
2. Active `custom_module` services
* Standby
1. Standby `custom_module` services
2. Standby `stateful` services
* Disable
1. Disable `custom_module` services
2. Disable `stateful` services 

Note: the `stateful` module is default. It should not be specified in config in case no custom modules.
# Installation procedure

## SiteManager

### Requirements 
Before installation be aware, that you have enough resources for DRNavigator.
For one pod it's recommended to use 100m cpu and 80Mi memory limits per worker and by default DRNavigator uses 2 workers.
You can override worker count and cpu/memory limits using helm environments.

### Prerequisites 
Prepare kubernetes cluster to work with DRNavigator
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

### Installation  
   Install `site-manager` helm chart:

    ```bash
    $ helm install site-manager charts/site-manager/ -n site-manager
    ```

    The `site-manager` helm chart can be customized with following parameters:
    
| Parameter                                                     | Description                                                           | Default value                   |
|---------------------------------------------------------------|-----------------------------------------------------------------------|---------------------------------|
| env.FRONT_HTTP_AUTH                                           | set authentication mode between sm-client and Site-Manager            | "Yes"                           |
| env.BACK_HTTP_AUTH                                            | set authentication mode between Site-Manager and manageable services  | "Yes"                           |
| env.SM_DEBUG                                                  | set `debug` logging level                                             | "False"                         |
| env.SM_GROUP                                                  | define API group for CRD                                              | "netcracker.com"                |
| env.SM_PLURAL                                                 | define object of API group                                            | "sitemanagers"                  |
| env.SM_VERSION                                                | define API group version for CRD                                      | "v2"                            |
| env.SERVICE_DEFAULT_TIMEOUT                                   | set default timeout for every microservice DR procedure               | 200                             |
| env.HTTP_SCHEME                                               | define HTTP scheme for connection to microservice operator            | "http://"                       |
| env.SM_CACERT                                                 | TLS verification in operators (True, False or path to trusted CA file)| "True"                          |
| workerCount                                                   | count of parallel workers, that handle requests                       | 2                               |
| serviceAccount.create                                         | enable/disable Service Account creation                               | true                            |
| serviceAccount.name                                           | name of Service Account for `site-manager`                            | "site-manager-sa"               |
| image.repository                                              | docker image repository name                                          | ghcr.io/netcracker/site-manager |
| image.pullPolicy                                              | docker image pull policy                                              | Always                          |
| image.tag                                                     | docker image tag                                                      | v1.0                            |
| ingress.create                                                | enable/disable ingress creation                                       | true                            |
| ingress.name                                                  | define URL for `site-manager` ingress                                 | ""                              |
| limits.cpu                                                    | cpu limits per pod                                                    | 200m                            |
| limits.memory                                                 | memory limits per pod                                                 | 160Mi                           |
| paas_platform                                                 | define PAAS type. It can be "kubernetes" or "openshift"               | "kubernetes"                    |
| tls.generateCerts.enabled                                     | enable/disable certificates generation via cert-manager               | false                           |
| tls.generateCerts.clusterIssuerName                           | define name cluster issuer, if you wand to use it (if empty, will be created self-signed issuer )  | ""                              |
| tls.generateCerts.duration                                    | define duration (days) of created certificate via cert-manager                                     | 365                             |
| tls.generateCerts.subjectAlternativeName.additionalDnsNames   | additional trusted dns names in certificate                             | []                              |
| tls.generateCerts.subjectAlternativeName.additionalIpAddresses | additional trusted ips names in certificate                              | []                             |
| paasGeoMonitor                                                | see [paas-geo-monitor documentation](/paas-geo-monitor/docs) | |
   
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

## smclient
### Prepare environment
 - Create `config.yml` with at least with following content:
```yaml
    sites:
      - name: k8s-1
        token: <BEARER TOKEN>
        site-manager: http://site-manager.k8s-1.netcracker.com/sitemanager
        cacert: <path-to-ca-certificate>
      - name: k8s-2
        token: <BEARER TOKEN>
        site-manager: http://site-manager.k8s-2.netcracker.com/sitemanager
        cacert: <path-to-ca-certificate>
    
    sm-client:
      http_auth: True
```
where
- `<BEARER TOKEN>` should be taken from the `sm-auth-sa-token-*` secret. Its name is specified in `sm-auth-sa` Service Account and can be obtained as:
```
kubectl get sa sm-auth-sa -n site-manager -o yaml | grep sm-auth-sa-token | cut -d ' ' -f3
```
After that you need to decode this token using base64 decoding, for example:
```
kubectl get secret sm-auth-sa-token-pqkxj -n site-manager -o yaml | grep token: | cut -d ' ' -f4 | base64 --decode
```
- cacert is a content of `site-manager-tls.crt` which has been generated during SiteManager installation
### Running from cli
1. On the Linux host should be installed python with version 3.9 or newer
1. Copy or clone files from https://github.com/Netcracker/DRNavigator  to separate directory. For example `/opt/sm-client`
1. Install python libs:
    ```
    # pip3 install -r requirements-sc.txt
    ```
1. Start script `sm-client` with argument `status` to check status of clusters and services 
