# Status

In progress.

# Context

To support geo distributed schemas, for example two-clusters DR, it is currently required that k8s clusters 
meet some specific requirements, that is:
1. pod-to-pod connectivity, i.e. pod should be able to reach pod from another cluster by IP address;
2. pod-to-service connectivity, i.e. pod should be able to reach service from another cluster by IP address;
3. cross-cluster service names resolving, i.e. pod should be able to resolve service name from another cluster.

These requirements are provided by IaaS and PaaS layers.
To verify that connectivity and resolving actually work, it is required to 
run some workload in one cluster and try to reach workload in another cluster by name, service IP and pod IP.
These checks are currently not automated, but they are very important to understand clusters geo health.

# Decision

We are going to provide a special `dr-monitor` application which will be responsible for verifying clusters geo health 
from inside each cluster. Particularly it is going to verify dns resolving, pod-to-pod and pod-to-service connectivity
by trying to resolve and reach `dr-monitor` service in another cluster. 
The information about clusters geo health could be pulled by external tools, like monitoring or KubeMarine `check_paas`.

### Peers configuration format
The `dr-monitor` instance in one cluster should know about `dr-monitor` instances in other clusters, 
this information should be provided through configuration, for example:
```yaml
peers:
  - name: cluster-b
    clusterIp:
      name: dr-monitor.ns.svc.cluster-b.local
  - name: cluster-c
    clusterIp:
      name: dr-monitor.ns.svc.cluster-c.local
    port: 8081
```
Here, we specify that our `dr-monitor` instance has two peers, 
which should be reached by the name of the ClusterIP service.

### Getting peers status
To get the status of DNS resolving, pod-to-pod and pod-to-service connectivity for peers, external tool should
contact `dr-monitor` instance in one of the clusters. This instance verifies resolving and connectivity from the
perspective of its own cluster. Thus, to get the full picture it is required to aggregate information
from all `dr-monitor` instances.

Below is the diagram showing the process of getting peers status.

![](/documentation/images/dr-monitor.png)

The information about peers is collected on demand when external tools 
send HTTP GET requests to `/peers/status` path. 
The response contains information about all peers DNS resolving, pod-to-service and pod-to-pod connectivity statuses,
for example:
```yaml
peers:
  - name: cluster-b
    clusterIp:
      name: dr-monitor.ns.svc.cluster-b.local
      dnsStatus:
        resolved: true
      svcStatus:
        available: true
        address: 1.1.1.1
      podStatus:
        available: true
        address: 2.2.2.2
  - name: cluster-c
    clusterIp:
      name: dr-monitor.ns.svc.cluster-c.local
      dnsStatus:
        resolved: false
        error: "no such host"
```
The statuses are collected in order - first DNS resolving is checked, then pod-to-service and then pod-to-pod
connectivity. If a certain status fails, an error message is attached and further statuses are not collected.

Internally, `dr-monitor` instance works as following for each peer:
1. It tries to resolve peer DNS name and get ClusterIP address.
2. It sends HTTP GET request to peer `/ping` path using ClusterIP address. 
The response contains peer pod IP address.
3. It sends HTTP GET request to peer `/ping` path using pod IP address.

### Integration with `check_paas`
The `check_paas` procedure should be modified to support DR checks using `dr-monitor` instance. 
Since `dr-monitor` is an optional service, `check_paas` does not use it by default.
To let `check_paas` know that it should perform DR checks, a `procedure.yaml` should be passed
containing information about `dr-monitor` instance:
```yaml
drMonitor:
  name: dr-monitor.ns.svc.cluster-a.local
```
The information includes at least service name, but it could also contain port, protocol, cert path, etc.
To send the request, `check_paas` uses `curl` from one of the cluster nodes. 
All the checks are performed as single check since they are ordered as dependencies.

### Integration with Prometheus
The information about clusters geo health could be sent to Prometheus. For this, a separate `/metrics` path is used,
which serves DNS, pod-to-service and pod-to-pod statuses metrics for each peer.

# Consequences

* New `dr-monitor` application should be implemented. It should be part of DRNavigator repository and deployed using site-manager chart.
* KubeMarine `check_paas` procedure should support DR checks using `dr-monitor` API.