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
          svcPort: 80
          podPort: 8080
```
For more installation options see chart `values.yaml`. For `config` format see [Configuration](#configuration). 

# Configuration
TBD

# API
TBD