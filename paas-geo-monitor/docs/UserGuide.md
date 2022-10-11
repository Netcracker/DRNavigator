# Abstract
Service for monitoring connectivity between geographically distributed clusters.
The primary focus is to check following aspects:
1. DNS resolving for service names from another cluster.
2. Pod-to-service connectivity between clusters.
3. Pod-to-pod connectivity between clusters.

The service should be deployed in each cluster in a schema.
Each instance should be able to discover other instances (neighbors) 
in other clusters, so each instance should be provided with a 
special configuration, containing service names of the neighbors.

Each instance reports connectivity status for the current cluster only.
So, to gather information about all clusters, 
you should contact each instance.

# Installation

The installation is done using a [helm chart](/charts/site-manager). By default, this service is not installed.
To install the service, it is required to enable it and provide information about neighbors, for example:
```yaml
paas_geo_monitor:
  install: true
  config:
    neighbors:
      - name: neighbor1
        clusterIp:
          name: paas-geo-monitor.site-manager-3.svc.cluster.local
```
For more installation options see chart `values.yaml`. For `config` format see [Configuration](#configuration). 

# Configuration
The service configuration allows you to configure instance-wide options and neighbors.

| Field                                           | Description                                            |
|-------------------------------------------------|--------------------------------------------------------|
| **port**<br/>_int_                              | The port that the service should listen. Default 8080. |
| **neighbors**<br/>_array [Neighbor](#neighbor)_ | Array of neighbor service instances in other clusters. |

## Neighbor
Neighbor represents a service instance from another cluster.

| Field                                       | Description                                  |
|---------------------------------------------|----------------------------------------------|
| **name**<br/>_string_                       | The name of the neighbor. Mandatory.         |
| **clusterIp**<br/>_[ClusterIp](#clusterip)_ | ClusterIp service representing the neighbor. |

## ClusterIp
ClusterIp represents neighbor k8s ClusterIP service.

| Field                     | Description                                 |
|---------------------------|---------------------------------------------|
| **name**<br/>_string_     | The hostname of k8s service. Mandatory      |
| **svcPort**<br/>_int_     | Service port of the neighbor. Default 8080. |
| **podPort**<br/>_int_     | Pod port of the neighbor. Default 8080.     |
| **protocol**<br/>_string_ | The protocol to use. Default `http`.        |

# API
TBD