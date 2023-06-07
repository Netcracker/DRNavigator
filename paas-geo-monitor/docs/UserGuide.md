# Abstract
Service for monitoring connectivity between geographically distributed clusters.
The primary focus is to check following aspects:
1. DNS resolving for service names from another cluster.
2. Pod-to-service connectivity between clusters.
3. Pod-to-pod connectivity between clusters.

The service should be deployed in each cluster in a schema.
Each instance should be able to discover other instances (peers) 
in other clusters, so each instance should be provided with a 
special configuration, containing service names of the peers.

Each instance reports connectivity status for the current cluster only.
So, to gather information about all clusters, 
you should contact each instance.

# Installation

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

Full parameters list:

| Parameter                                             | Description                                                        | Default value                             |
|-------------------------------------------------------|--------------------------------------------------------------------|-------------------------------------------|
| paasGeoMonitor.install                                | Enable/disable paas-geo-monitor installation                       | false                                     |
| paasGeoMonitor.image                                  | Paas-geo-monitor image                                             | ghcr.io/netcracker/paas-geo-monitor:0.9.0 |
| paasGeoMonitor.config                                 | Peers configuration. See format on [Configuration](#configuration) | peers: []                                 |
| paasGeoMonitor.env.PAAS_PING_PEERS                    | Enable/disable ping peers status (for monitoring)                  | true                                      |
| paasGeoMonitor.env.PAAS_PING_TIME                     | Time between ping peers status attempts (in sec)                   | 5                                         |
| paasGeoMonitor.env.PAAS_DEBUG                         | Enable/disable debug logs                                          | false                                     |

# Configuration
The service configuration allows you to configure instance-wide options and 
peers.

**Note**: `PING_IP` env variable should be set, otherwise service will not work. See [ping endpoint](#ping).

| Field                               | Description                                             |
|-------------------------------------|---------------------------------------------------------|
| **port**<br/>_int_                  | The port that the service should listen. Default 8080.  |
| **peers**<br/>_array [Peer](#peer)_ | Array of peer service <br/>instances in other clusters. |

## Peer
Peer represents a service instance from another cluster.

| Field                                       | Description                                           |
|---------------------------------------------|-------------------------------------------------------|
| **name**<br/>_string_                       | The name of the peer. Mandatory.                  |
| **clusterIp**<br/>_[ClusterIp](#clusterip)_ | K8s ClusterIP service <br/>representing the peer. |

## ClusterIp
ClusterIp represents peer k8s ClusterIP service.

| Field                     | Description                                 |
|---------------------------|---------------------------------------------|
| **name**<br/>_string_     | The hostname of k8s service. Mandatory      |
| **svcPort**<br/>_int_     | Service port of the peer. Default 8080. |
| **podPort**<br/>_int_     | Pod port of the peer. Default 8080.     |
| **protocol**<br/>_string_ | The protocol to use. Default `http`.        |

# API
The service provides following HTTP APIs. 

## peers status
peers status endpoint:
* Path: `/peers/status`
* Method: `GET`

The endpoint returns information about DNS, pod-to-pod and pod-to-service statuses of all peers.
An example response:
```yaml
- name: cluster-2
  clusterIpStatus:
    name: paas-geo-monitor.ns.svc.cluster-2.local
    svcPort: 8080
    podPort: 8080
    protocol: http
    dnsStatus:
      resolved: false
      error: 'failed to resolve peer: unable to resolve name dr-monitor.ns.svc.cluster-2.local: lookup dr-monitor.ns.svc.cluster-2.local: no such host'
- name: cluster-3
  clusterIpStatus:
    name: dr-monitor.ns.svc.cluster-3.local
    svcPort: 8080
    podPort: 8080
    protocol: http
    dnsStatus:
      resolved: true
    svcStatus:
      available: true
      address: 1.1.1.1
    podStatus:
      available: true
      address: 2.2.2.2
```

## Ping
Ping endpoint:
* Path: `/ping`
* Method: `GET`

The endpoint returns IP address of the instance, which is then used to verify pod-to-pod connectivity.
The IP address is taken from `PING_IP` env variable.