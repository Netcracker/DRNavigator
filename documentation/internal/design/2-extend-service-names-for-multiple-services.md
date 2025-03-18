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
In our consept this name is unique, because we provide operations on these services by those names.

But this way is followed with the problem, because site can consist several identical services in different namespaces 
(e.g. some postgres DB, some kafka instances and so on). 
If user deploy such services, they will be deployed successfully, but site-manager will operate only with one of those services and another one will be invisible.
[Issue on it](https://github.com/Netcracker/DRNavigator/issues/49)

Now uniqueness of name is the concern of the user and requires to take account of the number of services, that should be deployed in site. 
If it can be more than one, user has to provide an opportunity to redefine CR name in helm chart or create automatic algorithm for it. 
And as practice shows, for most services there is no limit on their number in a cluster. Therefore, this issue is relevant.

Also, we have requirements to use dns name of operated service as service name in site-manager. 

## Considered Options

* **To require uniqueness of the CR name in the cluster**
This option means that we fully restrict creation CR with name, that is already exist in cluster. It requires to validate created/modified CRs and should be noticed in site-manager guides.
* **To modify the naming process in site-manager to ensure that it is unique across the cluster**  
This option involves algorithmically selecting the name of the service to work through the site-manager. Moreover, this option requires backward compatibility with the current approach.

## Proposal

It is proposed to make the following rules for service name:
```
<cr-name>.<namespace>
```

* **Namespace** is userful because CR name is unique per them (our CR is namespaced for this reason it's restricted by kubernetes).  
* **dot** makes the service name similar with the dns name, which partially fulfills the relevant requirement.

We have site-manager server as proxy server, for this reason we can override service names here during getting `sm-dict` with all services. 
In that case those names automatically will be used in interaction between sm-client and site-manager.   

Difficulties will only appear in places where service names are hardcoded:
* `after`/`before` sections in CRs - *most critical*;
* `run-service` in body of POST `/sitemanager` request to site-manager;
* `--run-services` and `--skip-services` options in sm-client;

In order for users not to urgently need to edit their CRs, it is required to maintain backward 
compatibility, i.e. it should be possible to use only the CR name (without namespace) like now.

For this it's proposed to implement new v3 CR api version to make a difference between current and new approaches.  
During conversion from v2 to v3 version, site-manager adds namespaces to service names in `after`/`before` sections.  
After that this CRs can be used for SM procedures.

If some service's already had namespace in CR names (e.g. it's dynamically calculated via helm), it should be excluded, CR should be 
created with new name and old CR should be removed.

Also, if service name is changed, it should be fixed manually in `run-service` and `--run-services`/`--skip-services` places.

As result:
1. Fresh installation process doesn't change;
2. After implementation in SM, dependencies are rebuilt for new approach, but namespace will be duplicated, if it's already used in CR name.
3. After excluding namespace from CR name, it's needed to redeploy not only CR for this service, but also CRs, where this service is dependent.
4. If service name is hardcoded outside cluster and if it's changed, it should be manually fixed there.

### Example

Suppose we have two services `serviceA` and `serviceB`, dependent from `serviceA`.
#### 0. Before changes

In services helms:
```yaml
apiVersion: qubership.org/v2
kind: SiteManager
metadata:
  name: {{ .Values.namespace }}-serviceA
spec:
  sitemanager:
    after: []
    before: []
---
apiVersion: qubership.org/v2
kind: SiteManager
metadata:
  name: {{ .Values.namespace }}-serviceB
spec:
  sitemanager:
    after: 
      {{ toYaml .Values.sm.after | indent 6 }}
    before: []
```

In cluster:
```yaml
apiVersion: qubership.org/v2
kind: SiteManager
metadata:
  name: ns1-serviceA
  namespace: ns1
spec:
  sitemanager:
    after: []
    before: []
---
apiVersion: qubership.org/v2
kind: SiteManager
metadata:
  name: ns2-serviceB
  namespace: ns2
spec:
  sitemanager:
    after: 
      - ns1-serviceA
    before: []
```

#### 1. After SM implementation
In services helms nothing changes;  
In cluster:
```yaml
apiVersion: qubership.org/v3
kind: SiteManager
metadata:
  name: ns1-serviceA
  namespace: ns1
spec:
  sitemanager:
    after: []
    before: []
---
apiVersion: qubership.org/v3
kind: SiteManager
metadata:
  name: ns2-serviceB
  namespace: ns2
spec:
  sitemanager:
    after: 
      - ns1-serviceA.ns1
    before: []
```

#### 2. After changing CR names in services helms and redeploy
In services helms:
```yaml
apiVersion: qubership.org/v3
kind: SiteManager
metadata:
  name: serviceA
spec:
  sitemanager:
    after: []
    before: []
---
apiVersion: qubership.org/v3
kind: SiteManager
metadata:
  name: serviceB
spec:
  sitemanager:
    after: 
      {{ toYaml .Values.sm.after | indent 6 }}
    before: []
```
In cluster:
```yaml
apiVersion: qubership.org/v3
kind: SiteManager
metadata:
  name: serviceA
  namespace: ns1
spec:
  sitemanager:
    after: []
    before: []
---
apiVersion: qubership.org/v3
kind: SiteManager
metadata:
  name: serviceB
  namespace: ns2
spec:
  sitemanager:
    after: 
      - serviceA.ns1
    before: []
```

### Positive Consequences
* It solves the problem with the multiplicity of services with the same name;
* It is intuitive to the user;
* Not need to change used data structures;
* Easy to implement in code;
* It doesn't break the concept of site-manager;
* Proposed rules for `after`/`before` don't change the work of already existing DR clusters;
* Changes in CRs are not required;

### Negative Consequences
* There is a restriction that the same service on different sites must have the same namespace. Otherwise, sm-client 
will recognise the as different services;
* Requires new CR version;
* Namespace will be duplicated, if it's already defined in cr name: to resolve this situation user has to recreate CR 
with new name;
* Requires manually steps outside the cluster;
* Additional test cases are required;

## Namespace Restriction Bypass

Unfortunately for some services it's important to have different namespaces on different sites of DR cluster. 
As result after changes in service name rules will be implemented, such service will be defined as two different services 
on different sites, e.g.:
```
+-----------------------------+--------------------------------------+--------------------------------------+
| Service                     |                site-1                |                site-2                |
+-----------------------------+--------------------------------------+--------------------------------------+
|                             | mode | DR status | healthz | message | mode | DR status | healthz | message |
| --------------------------  | ------------------------------------ | ------------------------------------ |
| sm-dummy.ns-in-site-1       |     standby / done / up / I'm OK     | -- / -- / -- / Service doesn't exist |
| sm-dummy.ns-in-site-2       | -- / -- / -- / Service doesn't exist |      active / done / up / I'm OK     |
+-----------------------------+--------------------------------------+--------------------------------------+
```
For this reason it's needed to increase functionality to support such situations.

### Considered Options

#### 1. Support aliases in sm-config

Because such situation appears in sm-client and as a result of the out of sync between sites, we can resolve this situation 
defining aliases for services, that will be used by sm-client:
```yaml
aliases:
  sm-dummy:
    site-1: sm-dummy.ns-in-site-1
    site-2: sm-dummy.ns-in-site-2
```

`sm-dummy` in that case is alias for services `sm-dummy.ns-in-site-1` and `sm-dummy.ns-in-site-2`. Because of this 
configuration sm-client can understand, that `sm-dummy.ns-in-site-1` and `sm-dummy.ns-in-site-2` represent one service 
on different sites:

```
+-----------------------------+--------------------------------------+--------------------------------------+
| Service                     |                site-1                |                site-2                |
+-----------------------------+--------------------------------------+--------------------------------------+
|                             | mode | DR status | healthz | message | mode | DR status | healthz | message |
| --------------------------  | ------------------------------------ | ------------------------------------ |
| sm-dummy                    |     standby / done / up / I'm OK     |      active / done / up / I'm OK     |
+-----------------------------+--------------------------------------+--------------------------------------+
```

### Positive Consequences
* It solves the problem with namespace restrictions;
* Don't touch behaviour of Site-Manager;
* Can be easily customized by user;
* Can be userful not only for namespaces but in case of mistakes in CR name for example.

### Negative Consequences
* Not easy to implement;
* Requires manually steps outside the cluster to reconfigure existed configuration;
* Additional test cases are required;

#### 2. Opponent name in CR

We can add special optional field to define service name for opponent site, e.g. for `site-1`:
```yaml
apiVersion: qubership.org/v3
kind: SiteManager
metadata:
  name: sm-dummy
  namespace: ns-in-site-1
spec:
  sitemanager:
    opponentService: sm-dummy.ns-in-site-2
    after: []
    before: []
```

In sm-client this field can be used to make links between services with different names.

### Positive Consequences
* It solves the problem with namespace restrictions;
* Can be userful not only for namespaces but in case of mistakes in CR name for example.

### Negative Consequences
* Requires implementation in CR and sm-client;
* It is not obvious what name should be used in sm-client: `--run-services`, `-skip-services`, status table, etc.;
* Unexpected behaviour, if `opponentService` options aren't consistent in DR sites` CRs.
* Automatic v2->v3 conversion for already existed services can be wrong;
* Additional test cases are required;

#### 3. Override service name

We can add special optional field to override service name, e.g. for `site-1`:
```yaml
apiVersion: qubership.org/v3
kind: SiteManager
metadata:
  name: sm-dummy
  namespace: ns-in-site-1
spec:
  sitemanager:
    serviceName: sm-dummy-true-name
    after: []
    before: []
```

`serviceName` value will be used instead of `<cr-name>.<namespace>` for all operations on this service:

```
+-----------------------------+--------------------------------------+--------------------------------------+
| Service                     |                site-1                |                site-2                |
+-----------------------------+--------------------------------------+--------------------------------------+
|                             | mode | DR status | healthz | message | mode | DR status | healthz | message |
| --------------------------  | ------------------------------------ | ------------------------------------ |
| sm-dummy-true-name          |        standby / done / up /         |        active / done / up /         |
+-----------------------------+--------------------------------------+--------------------------------------+
```

### Positive Consequences
* It solves the problem with namespace restrictions;
* Can be userful not only for namespaces but in case of mistakes in CR name for example.
* Easy to implement;

### Negative Consequences
* Requires to validate, if `serviceName` value is unique on cluster;
* Сan broken concept "service name like dns name";
* Automatic v2->v3 conversion for already existed services can be wrong;

#### 4. Turn on/off namespace concatenation

We can add special optional field to turn on/off adding namespace to service name (`isUniqueOnCluster`, default is false):
```yaml
apiVersion: qubership.org/v3
kind: SiteManager
metadata:
  name: sm-dummy
  namespace: ns-in-site-1
spec:
  sitemanager:
    isUniqueOnCluster: true
    after: []
    before: []
```

If `isUniqueOnCluster` is `true`, it means, that this service has only one exemplar in cluster and for it service name 
is `<cr-name>` like now:

```
+-----------------------------+--------------------------------------+--------------------------------------+
| Service                     |                site-1                |                site-2                |
+-----------------------------+--------------------------------------+--------------------------------------+
|                             | mode | DR status | healthz | message | mode | DR status | healthz | message |
| --------------------------  | ------------------------------------ | ------------------------------------ |
| sm-dummy                    |        standby / done / up /         |        active / done / up /         |
+-----------------------------+--------------------------------------+--------------------------------------+
```

### Positive Consequences
* It solves the problem with namespace restrictions;
* Easy to implement;

### Negative Consequences
* Requires to validate, if service names for unique if `isUniqueOnCluster` is true;
* Сan broken concept "service name like dns name";
* Automatic v2->v3 conversion for already existed services can be wrong;

### 5. Different module behaviour

We can add namespaces only for services with specific volumes (e.g. not stateful). As result, services with 
another module will have the same name:
```yaml
apiVersion: qubership.org/v3
kind: SiteManager
metadata:
  name: sm-dummy
  namespace: ns-in-site-1
spec:
  sitemanager:
    module: not-stateful
    after: []
    before: []
```

Result:

```
+-----------------------------+--------------------------------------+--------------------------------------+
| Service                     |                site-1                |                site-2                |
+-----------------------------+--------------------------------------+--------------------------------------+
|                             | mode | DR status | healthz | message | mode | DR status | healthz | message |
| --------------------------  | ------------------------------------ | ------------------------------------ |
| sm-dummy                    |        standby / done / up /         |        active / done / up /         |
+-----------------------------+--------------------------------------+--------------------------------------+
```

### Positive Consequences
* It solves the problem with namespace restrictions;
* Easy to implement;

### Negative Consequences
* Requires to validate unique of service names;
* Such situation doesn't relate to module and can appear for stateful services. Changing module for such
service follows additional problems with dependencies and flow;
* Сan broken concept "service name like dns name";
* Automatic v2->v3 conversion for already existed services can be wrong;

## Summary

As fast solution we can restrict CR names and make them unique on site (the first option), but the need to choose a unique 
CR name is transferred to the user.
But the second option can be implemented in the future after additional discussion with users and when we will have enough capacity.
