# How to work with sm-dummy

<!-- TOC -->
* [How to work with sm-dummy](#how-to-work-with-sm-dummy)
  * [Overview](#overview)
  * [Build](#build)
  * [Installation](#installation)
    * [Parameters](#parameters)
<!-- TOC -->

## Overview

Sm-dummy is a special service, that is uses to emulate dr-managed services during site-manager and
sm-client developing.

Sm-dummy follows dr-managed services [rest-api contract](../public/architecture.md#rest-api-contract),
but instead of real data replication during switchover and failover, it just waits some time
and updates its statuses.

sm-dummy source code is available [here](../../tests/sm-dummy)

## Build

Sm-dummy can be built as a docker image using following command:
```bash
docker build -t <sm-dummy-image-name> --no-cache tests/sm-dummy
```

## Installation

Sm-dummy has separate helm chart, that is available [here](../../tests/sm-dummy/charts).

The following command can be used to install sm-dummy using helm chart:
```bash
helm install sm-dummy tests/sm-dummy/charts/sm-dummy/ -n sm-dummy
```

Additionally, it's possible to install it as kubemarine plugin.
The examples, how it can be run, can be found [here](../../ci/cloud-tests/cluster.yaml).

### Parameters

| Parameter                           | Description                                                                             | Default value                                                          |
|-------------------------------------|-----------------------------------------------------------------------------------------|------------------------------------------------------------------------|
| name                                | sm-dummy entities name                                                                  | "sm-dummy"                                                             |
| PAAS_PLATFORM                       | define PAAS type. It can be "KUBERNETES" or "OPENSHIFT"                                 | "KUBERNETES"                                                           |
| smSecureAuth                        | Secure authorization mode in SM                                                         | false                                                                  |
| env.SMA_TIMEOUT                     | sleep timeout for DR procedure                                                          | 15                                                                     |
| env.SMA_REQUEST_TIMEOUT             | set timeout for /sitemanager requests                                                   | 0                                                                      |
| env.SMA_INIT_MODE                   | initial mode for service                                                                | "active"                                                               |
| env.SMA_DEBUG                       | debug mode for Flask                                                                    | True                                                                   |
| env.SMA_HEALTHZ                     | healthz output in dummy mode                                                            | "up"                                                                   |
| env.SMA_SEND_MESSAGE                | send messages with /sitemanager request                                                 | True                                                                   |
| env.SMA_HTTP_AUTH                   | check token for /sitemanager requests                                                   | True                                                                   |
| env.SMA_SM_NAMESPACE                | site-manager namespace (for authorization checking)                                     | "site-manager"                                                         |
| env.SMA_SM_SA                       | site-manager service account name (for authorization checking)                          | "site-manager-sa" if `smSecureAuth` is enabled. Otherwise `sm-auth-sa` |
| env.SMA_CUSTOM_AUDIENCE             | site-manager custom audience for rest API token. Used only is `smSecureAuth` is enabled | ""                                                                     |
| image                               | sm-dummy image                                                                          |                                                                        |
| ingressName                         | define URL for ingress                                                                  |                                                                        |
| serviceAccount.create               | enable/disable Service Account creation                                                 | true                                                                   |
| serviceAccount.name                 | name of Service Account for sm-dummy                                                    | "sm-dummy-sa"                                                          |
| siteManager.installCR               | enable sitemanager-cr                                                                   | True                                                                   |
| siteManager.module                  | cr module property                                                                      | "stateful"                                                             |
| siteManager.before                  | cr before property                                                                      | []                                                                     |
| siteManager.after                   | cr after property                                                                       | []                                                                     |
| siteManager.sequence                | cr sequence property                                                                    | ["standby", "active"]                                                  |
| siteManager.allowedStandbyStateList | cr allowedStandbyStateList property                                                     | ["up"]                                                                 |
| siteManager.timeout                 | cr timeout property                                                                     | 180                                                                    |
| siteManager.useDefaultParams        | add parameters for stateful module in cr (serviceEndpoint and healthzEndpoint)          | True                                                                   |
| siteManager.additionalParams        | you can customize parameters, adding needed properties inside this param                |                                                                        |
