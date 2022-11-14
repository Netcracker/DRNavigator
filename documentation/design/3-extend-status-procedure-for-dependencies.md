# Extend status procedure for getting full info about cluster

Contents:
* [Issue](#issue)
* [Considered Options](#considered-options)
* [Proposal](#proposal)
  * [Example](#example)
  * [Positive Consequences](#positive-consequences)
  * [Negative Consequences](#negative-consequences)
* [Summary](#summary)

## Issue

Now we have special status procedure in site-manager, that returns status information about concrete service. 
Example:
```commandline
$ curl -X POST -H "Content-Type: application/json" \
    -d '{"procedure": "status", "run-service": "some-service"}' \
    https://site-manager:8443/sitemanager
{
    "services": {
        "some-service": {
            "healthz": "up",
            "message": "",
            "mode": "active",
            "status": "done"
        }
    }
}
```

But in real this information is not enough to understand, if site active or not and if service really ready for 
other procedures because it is not clear what happens to dependent services;
If user wants to get such information, he has to get current sm-dict, calculate dependent services from there and 
call status procedure for all of them (especially since their dependent services may have own dependencies).

## Proposal

It is proposed to extend status procedure and get information about dependencies for all depth.
For this, response will return additional `deps` field with information about dependent services.   
Template looks like:
```json
{
  "services": {
    "some-service": {
      "healthz": "up",
      "message": "",
      "mode": "active",
      "status": "done",
      "deps": {
        "before": { <before services statuses> },
        "after": { <after services statuses> }
      }
    }
  }
}
```

Statuses for dependent services also contains information about dependencies, descending recursively to final services.

### Example

Suppose we have 4 services on the site: `serviceA`, `serviceB`, `serviceC` and `serviceD` with following dependencies:
```yaml
serviceA:
  before: []
  after: []
serviceB: 
  before: []
  after: [serviceA]
serviceC:
  before: []
  after: []
serviceD:
  before: [serviceB]
  after: [serviceC]
```
Dependencies in status procedure for `serviceA` and `serviceC` will be empty, because they don't have ane dependencies in their CRs:
* For `serviceA`:
```json
{
  "services": {
    "serviceA": {
      "healthz": "<healthz>",
      "message": "<message>",
      "mode": "<mode>",
      "status": "<status>",
      "deps": {
        "before": {},
        "after": {}
      }
    }
  }
}
```
* For `serviceC`:
```json
{
  "services": {
    "serviceC": {
      "healthz": "<healthz>",
      "message": "<message>",
      "mode": "<mode>",
      "status": "<status>",
      "deps": {
        "before": {},
        "after": {}
      }
    }
  }
}
```
`ServiceB` will contain information about `serviceA` status:
```json
{
  "services": {
    "serviceB": {
      "healthz": "<healthz>",
      "message": "<message>",
      "mode": "<mode>",
      "status": "<status>",
      "deps": {
        "before": {},
        "after": {
          "serviceA": {
            "healthz": "<healthz>",
            "message": "<message>",
            "mode": "<mode>",
            "status": "<status>",
            "deps": {
              "before": {},
              "after": {}
            }
          }
        }
      }
    }
  }
}
```

`ServiceD` will contain information about all other services because it's recursively dependent from them:
```json
{
  "services": {
    "serviceD": {
      "healthz": "<healthz>",
      "message": "<message>",
      "mode": "<mode>",
      "status": "<status>",
      "deps": {
        "before": {
          "serviceB": {
            "healthz": "<healthz>", 
            "message": "<message>", 
            "mode": "<mode>", 
            "status": "<status>", 
            "deps": {
              "before": {}, 
              "after": {
                "serviceA": {
                  "healthz": "<healthz>",
                  "message": "<message>", 
                  "mode": "<mode>", 
                  "status": "<status>", 
                  "deps": {
                    "before": {}, 
                    "after": {}
                  }
                }
              }
            }
          }
        },
        "after": {
          "serviceC": {
            "healthz": "<healthz>",
            "message": "<message>",
            "mode": "<mode>",
            "status": "<status>",
            "deps": {
              "before": {},
              "after": {}
            }
          }
        }
      }
    }
  }
}
```

### Positive Consequences
* It returns information about statuses for all dependent services;
* It is intuitive to the user;
* It can be turned on/off with special flag in post request;
* It can be used in sm-client;
* Easy to implement using recursion;

### Negative Consequences
* Increased procedure's working time (can be controlled, if we calculate dependencies only if necessary);
* Added exception with cyclic dependencies in status procedure, with is needed to be handled;
* The work of site-manager goes a little beyond the proxy service;
* Requires additional test cases;

## Summary

This is a fairly simple way to supplement information about the status of the service and its real readiness for other procedures.
However, for a full-fledged implementation, you should use this logic in the cm-client, since in most cases the user interacts with the site manager through it.

