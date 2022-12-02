# Extend service names in Site-Manager

Contents:
* [Issue](#issue)
  * [Issue Examples](#issue-examples)
  * [Requirements and Reflections](#requirements-and-reflections)
* [Proposal](#proposal)
  * [Configuration options](#configuration-options)
  * [Positive Consequences](#positive-consequences)
  * [Negative Consequences](#negative-consequences)
* [Summary](#summary)

## Issue

At the moment we can use atomic commands like `active`, `standby` in sm-client, that change state of one site and don't 
touch another one.  
But practice shows, that some cluster states can be incorrect for some services or some data can be lost.  

As result, we should restrict some cluster final states to forbid user to do it.

### Issue Examples
* **OpenSearch case**  
issue state: stadnby-standby   
severity: high  
For OpenSearch moving to standby state cleans all data. So, using command "standby left" and "standby right" all data 
will be lost. 

* **Kafka case**  
issue state: standby-standby  
severity: medium  
For Kafka moving to standby state runs replication from opposite side. So, using command "standby left" and 
"standby right" the data replication will be cycled.

### Requirements and Reflections
1. **What services are restricted?**
  * *For all* - it easier to implement, but services, for which restricted states can be used, can exist;
  * *Defined services* - can be customized;
2. **What states are restricted?**
  * *Only standby-standby* - we know only about cases with them, but theoretically can be others (e.g. active-active or disable-disable);
  * *Defined states* - can be customized, but special configuration should be;
3. **Should intermediate states should be restricted?**
  * DR operations (`move`, `stop`) are described with sequence;
  * Atomic procedures don't have intermediate states;
  * Runtime problems can't be fully predicted;
4. **What operations should be checked for restrictions?**
  * Cases appear after atomic operations: `active`, `standby`, `disable`, `return`;
5. **Are restricted states are similar for all operations?**
  * Problems are related only with *final* sites' state;
6. **Where restrictions rules should be described?**
  * *Hardcode as constants* - it is only appropriate if the restrictions apply to all services on the predefined states;
  * *In services' CRs* - it requires new CR version, and at the same time there are described per-site rules, for this 
reason we have to resolve descriptions differences from different sites.
  * *In sm-client configuration* - it's the most logical, but configuration format should be designed;
7. **What restrictions should be default?**
  * *All services in standby-standby* - this is the most frequent case;
  * *No restrictions* - writing new restrictions is easier than clearing predefined;
8. **Should health and dr status be taken into account?**
  * No, because those are runtime problems;
9. **Where and when restrictions should be checked?**
  * Operation result should be predicted and checked *before* running procedure;
  * We have already had *validate_operation* function in sm-client: where we check dependencies, availability of 
services on sites, etc. 
10. **Should we really always restrict those states?**
  * We can add special option in sm-client like `--ignore-restrictions` to skip state restrictions check;
11. **What should we do if service status is unknown (e.g. no connection to service from site-manager)?**
  * sm-client should fail (operation can be applied with force, if option from point #10 will be implemented);

## Proposal

It is proposed to configure state restrictions per service in sm-client configuration.
Those restrictions describe, which services shouldn't go into certain problem states.  
If user runs some of atomic procedure: `active`, `standby`, `return`, `disable`, final states 
will be predicted before running procedure and if some of those states are problem for some 
service, this operation should be forbidden.  
User should be able to skip those restrictions check if he wants. For this it's proposed to add additional option 
in sm-client like `--ignore-restrictions`, that will let him do operation regardless of the final state.  
If for some reason the final state can't be predicted (for example, the status of one of the services is unknown),
the operation should be forbidden.  
If operation was forbidden, logs should contain information about the reason: there are should be current and predicted 
cluster states and problem services, for which this state is restricted;  

### Configuration options

##### 1. Forbid standby-standby state for all services
Pros:
  * Not need additional configuration;  

Cons:
  * Used always and can't be overriden;

##### 2. Get services list for every restricted state
```yaml
restrictions:
  standby-standby:
    - kafka-service
    - opensearch
  active-active:
    - * 
```
`*` means applying for all services;

Procs:
* Services and states can be overriden;
* Intuitive to the user;

Cons:
* Unexpected behaviour if sites number is more or less than 2;
* More difficult to get a list of restricted states for concrete service (check applies per service);
* States like `standby-disable` should be listed twice or should be additionally calculated with any order; 

Notes:
* Using `*` is optional, it complicates the algorithm but simplifies the configuration;

##### 3. Get states list for every restricted service
```yaml
restrictions:
  kafka-service:
  - standby-standby
  opensearch: 
  - standby-standby
  *:
  - active-active
```
`*` means applying for all services;

Procs:
* Services and states can be overriden;
* Intuitive to the user;
* Easy to get restricted states per service;

Cons:
* Unexpected behaviour if sites number is more or less than 2;
* A bit redundant because there are fewer restricted states than services;
* States like `standby-disable` should be listed twice or should be additionally calculated with any order; 

Notes:
* Using `*` is optional, it complicates the algorithm but simplifies the configuration;

##### 4. Specify the exact number of sites that can be in a particular state at the same time
```yaml
restrictions:
  kafka-service:
    standby: 2
  opensearch: 
    standby: 2
  *:
    active: 0
```
`*` means applying for all services;

Procs:
* Services and states can be overriden;
* Not depend on states order;
* Easy to get restricted states per service;
* Can be used with any counts of sites;

Cons:
* Not as obvious as previous options;
* Can't configure only xor-like states, e.g. `standby-disable`;
* Redundant because there are fewer restricted states than services;

Notes:
* Using `*` is optional, it complicates the algorithm but simplifies the configuration;

### Positive Consequences
* It solves the problem with restricted states;
* It is intuitive to the user;
* Not need to change used data structures;
* Not need to update CRs;
* We have already had architecture possibilities to validate operations before running them;
* Can be turned off if needed;

### Negative Consequences
* Atomic operations requires services status in not requested sites to be known;
* Additional test cases are required;
* Runtime issues are not taken into account;

## Summary

The verification algorithm is quite simple, if the sm-client will predict the final states.
The third configuration option is the most optimal, because used format is quite obvious and fully covers 
work with DR clusters. 
