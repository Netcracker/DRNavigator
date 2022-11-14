# Extend service names in Site-Manager

Contents:
* [Issue](#issue)
* [Considered Options](#considered-options)
* [Proposal](#proposal)
  * [Example](#example)
  * [Positive Consequences](#positive-consequences)
  * [Negative Consequences](#negative-consequences)
* [Summary](#summary)

## Issue

At the moment we get the name of the service in the site-manager as the name of CR that is deployed along with this service. 
In our consept this name is unique, because we provide operations on this services by those names.

But this way is followed with the problem, because site can consist several identical services in different namespaces 
(e.g. some postgres DB, some kafka instances and so on). 
If user deploy such services, they will be deployed successfully, but site-manager will operate only with one of this services and another one will be invisible.
[Issue on it](https://github.com/Netcracker/DRNavigator/issues/49)

Now uniqueness of name is the concern of the user and requires to take account of the number of services, that should be deployed in site. 
If it can be more than one, user has to provide an opportunity to redefine CR name in helm chart or create automatic algorithm for it. 
And as practice shows, for most services there is no limit on their number in a cluster. Therefore, this issue is relevant.

## Considered Options

* **To require uniqueness of the CR name in the cluster**
This option means that we fully restrict creation CR with name, that is already exist in cluster. It requires to validate created/modified CRs and should be noticed in site-manager guides.
* **To modify the naming process in site-manager to ensure that it is unique across the cluster**  
This option involves algorithmically selecting the name of the service to work through the site-manager. Moreover, this option requires backward compatibility with the current approach.

## Proposal

It is proposed to make the following rules for service name:
```
<namespace>/<cr-name>
```

* **Namespace** is userful because CR name is unique per them (our CR is namespaced for this reason it's restricted by kubernetes).  
* **Slash symbol** can't be used in kubernetes names but is not special symbol in yaml and json format, that we use inside site-manager. 
As result, it can't appear like part of cr name or namespace and the same time we don't have to additionally escape it.

We have site-manager server as proxy server, for this reason we can override service names here during getting `sm-dict` with all services. 
In that case those names automatically will be used in interaction between sm-client and site-manager.   

Difficulties will only appear in places where service names are hardcoded:
* `after`/`before` sections in CRs - *most critical*;
* `run-service` in body of POST /sitemanager request to site-manager;
* `--run-services` and `--skip-services` options in sm-client;

In order for users not to urgently need to edit their CRs, it is required to maintain backward 
compatibility, i.e. it should be possible to use only the CR name (without namespace).  

Therefore, it is suggested to use the following rule:  
* If only the CR name is specified, this means that this rule/request applies to all services with the given name on the cluster.
For example, if CR of service contain line like ```after: [cassandra]```, it means, that this service is dependent from all cassandras on site.
If you want to specify concrete cassandra you should specify namespace, like ```after: [devops-tools-ns/cassandra]```

This way can be simply implemented in `sm-dict` calculation for `after`/`before` because this place is the most critical. 
Also, it can be extended to an additional feature if we additionally implement it in `run-service` and `--run-services`/`--skip-services`.

### Example

Suppose we have 3 services on the site: `serviceA` on namespaces `ns1` and `ns2` and `serviceB` on `ns3`, that is dependent from `serviceA`. 
Their CRs look like:
```yaml
apiVersion: netcracker.com/v2
kind: SiteManager
name: serviceA
namespace: ns1
spec:
  sitemanager:
    after: []
    allowedStandbyStateList:
    - up
    before: []
    module: stateful
    parameters:
    healthzEndpoint: http://serviceA.ns1:8080/healthz
    serviceEndpoint: http://serviceA.ns1:8080/sitemanager
    sequence:
    - standby
    - active
    timeout: 360
---
apiVersion: netcracker.com/v2
kind: SiteManager
name: serviceA
namespace: ns2
spec:
  sitemanager:
    after: []
    allowedStandbyStateList:
    - up
    before: []
    module: stateful
    parameters:
    healthzEndpoint: http://serviceA.ns2:8080/healthz
    serviceEndpoint: http://serviceA.ns2:8080/sitemanager
    sequence:
    - standby
    - active
    timeout: 360
---
apiVersion: netcracker.com/v2
kind: SiteManager
name: serviceB
namespace: ns3
spec:
  sitemanager:
    after: [serviceA]
    allowedStandbyStateList:
    - up
    before: []
    module: stateful
    parameters:
    healthzEndpoint: http://serviceB.ns3:8080/healthz
    serviceEndpoint: http://serviceB.ns3:8080/sitemanager
    sequence:
    - standby
    - active
    timeout: 360
```
`sm-dict` for them will look like this (pay attention to the names of the services and to the `after` section at the `serviceB`):
```json
{
    "services": {
        "ns1/serviceA": {
            "after": [],
            "allowedStandbyStateList": ["up"],
            "before": [],
            "module": "stateful",
            "name": "serviceA",
            "namespace": "ns1",
            "parameters": {
                "healthzEndpoint": "http://serviceA.ns1:8080/healthz",
                "serviceEndpoint": "http://serviceA.ns1:8080/sitemanager"
            },
            "sequence": [
                "standby",
                "active"
            ],
            "timeout": 360
        },
        "ns2/serviceA": {
            "after": [],
            "allowedStandbyStateList": ["up"],
            "before": [],
            "module": "stateful",
            "name": "serviceA",
            "namespace": "ns2",
            "parameters": {
                "healthzEndpoint": "http://serviceA.ns2:8080/healthz",
                "serviceEndpoint": "http://serviceA.ns2:8080/sitemanager"
            },
            "sequence": [
                "standby",
                "active"
            ],
            "timeout": 360
        },
        "ns3/serviceB": {
            "after": ["ns1/serviceA", "ns2/serviceA"],
            "allowedStandbyStateList": ["up"],
            "before": [],
            "module": "stateful",
            "name": "serviceB",
            "namespace": "ns3",
            "parameters": {
                "healthzEndpoint": "http://serviceB.ns3:8080/healthz",
                "serviceEndpoint": "http://serviceB.ns3:8080/sitemanager"
            },
            "sequence": [
                "standby",
                "active"
            ],
            "timeout": 360
        }
    }
}
```

sm-client status table will look like:
```
+----------------------------+--------------------------------------+--------------------------------------+
| Service                    |                site-1                |                site-2                |
+----------------------------+--------------------------------------+--------------------------------------+
|                            | mode | DR status | healthz | message | mode | DR status | healthz | message |
| -------------------------- | ------------------------------------ | ------------------------------------ |
| ns1/serviceA               |        active / done / up /          |        standby / done / up /         |
| ns2/serviceA               |        active / done / up /          |        standby / done / up /         |
| ns3/serviceB               |        active / done / up /          |        standby / done / up /         |
+----------------------------+--------------------------------------+--------------------------------------+
```

### Positive Consequences
* It solves the problem with the multiplicity of services with the same name;
* It is intuitive to the user;
* Not need to change used data structures;
* CR changes aren't needed;
* Easy to implement in code;
* It doesn't break the concept of site-manager;
* Proposed rules for `after`/`before` don't change the work of already existing DR clusters;
* Can be implemented like additional site-manager mode (and can be switched on/off);

### Negative Consequences
* There is a restriction that the same service on different sites must have the same namespace. Otherwise, sm-client will recognise the as different services;
* It is necessary to notify users of this changes so that new output of the sm-client does not come as a surprise to them.
* If another service with the same name is deployed on an existing cluster, the user must keep in mind that it will be taken 
into account in all dependencies where it does not have a namespace specified.
* Users feedback is needed, because maybe not all cases are taken into account.
* Additional test cases are required;

## Summary

As fast solution we can restrict CR names and make them unique on site (the second option), but the need to choose a unique CR name is transferred to the user.
But the first future can be implemented in the future after additional discussion with users and when we will have enough capacity.
