# How to develop site-manager

## Site-manager development

Site-manager is the server side tool, that is based on [controller-runtime library](https://github.com/kubernetes-sigs/controller-runtime) 
to manage CRs and also provides additional [main API](/documentation/public/architecture.md#sitemanager-api) to perform operations on services.  
The source code for it is located in separate [site-manager](/site-manager) directory. This directory should be used as root dir for development.

### Requirements for development

Required:
1. Golang v1.20 or above;
2. Docker with docker-compose v19.03 or above;
3. Kubernetes cluster, installed via [kubemarine](https://github.com/Netcracker/KubeMarine); 

Additionally, you should install following tools that are required for code-generation:
1. `swagg` is used for swagger API generation;
2. `controller-gen` is used for DeepCopy methods generation, that are required in CRs structures to make it matche the kubernetes 
`runtime.Object` interface;
They can be installed with `go install` command, e.g.:
```bash
$ go install github.com/swaggo/swag/cmd/swag
$ go install sigs.k8s.io/controller-tools/cmd/controller-gen
```

### How to build site-manager

Before building site-manager you should regenerate code, in case, if you've changed the following places:
1. CR structure (new or already existed ones);
2. SM main API;

For this you can use common `go generate` command to regenerate **all** places:
```bash
$ go generate ./...
```

**Note**: above operations are always done during `docker build` command automatically;

To build site-manager run the following command:
* For linux/mac: 
```bash
$ go build -a -o site-manager ./main.go
```
* For windows:
```bash
$ go build -a -o site-manager.exe ./main.go
```

### How to run site-manager for development

Site-manager has the scope of cli options and environment variables.
To see the full list of cli options, run site-manager help command after build, e.g. for linux:
```bash
$ ./site-manager -h
```
The full list of environment variables is available in [configuration file](/site-manager/config/config.go);

Site-manager can be run in following modes:
1. `without-cluster mode`;
2. `outside-the-cluster mode`;
3. `inside-the-cluster mode`;

#### without-cluster mode run

SM can run without access to any clusters, but in that case it can provide only main API for services, without cr webhooks 
or metrics API. This mode is good in case, if you develop contract between SM and sm-client or SM and services and do not 
affect CR changes.

This mode is used in [test docker-compose](/tests/README.md#site-manager-local-docker-compose) and 
[integration test](/tests/README.md#site-manager-integration-tests).

To run SM in following mode, you should run a couple of [sm-dummy](/tests/sm-dummy) services in docker (or docker-compose)
and then fill the special yam configuration file, that contains the list of SM objects, that SM should use. 
See the format of SM configuration in [file](/site-manager/pkg/model/sm_objects.go);
You can take [SM configuration from docker-compose](/tests/docker-compose/site-manager-config.yaml) as an example;

**Note**: `testing.enabled` should be `true` to enable without-cluster mode run;

After that you can run site-manager specifying the SM configuration `SM_CONFIG_FILE`for main API with command like:
```bash
$ SM_CONFIG_FILE=<path-to-sm-config-file> SM_DEBUG=true ./site-manager --bind=0.0.0.0:8443
```

#### outside-the-cluster mode run

SM can be integrated with cluster but run separately. In that mode webhook API and metrics API are available, for this case this mode is good, if you develop some CR or webhook changes.  

Restrictions:
1. CRD should be installed in the cloud;
2. Due to the fact that the SM is outside the cloud, it's not possible to contact with services in cluster except cases,
if ingress address is used in their CRs;
3. Provide the token to connect with services (in real SM it's token from `sm-auth-sa` service-account). For this create the special file
with token inside and provide the path to this token with `--tokenfile` option during SM run. As alternative, disable authorization with
services with `BACK_HTTP_AUTH` environment;
3. In following mode SM will use the token from `sm-auth-sa` secret in your cloud (as normal SM). Please disable 
authorization with `FRONT_HTTP_AUTH` environment if SM is not installed here yet; 

To run SM in following mode, please firstly generate test tls certificates for SM, because they are required for webhooks. 
Instructions, how to do it, are available in [p.2 of prerequisites](/documentation/public/installation.md#prerequisites);
Then enable bind webhook address (and bind metrics address if you wish) in SM run and provide the
kubeconfig to the cluster in `SM_KUBECONFIG_FILE` environment variable. Also use `--dev-mode` option to disable the leader
election for local SM:

```bash
$ SM_KUBECONFIG_FILE=<path-to-kubeconfig> SM_DEBUG=true ./site-manager \
--bind=0.0.0.0:8443 --bind-webhook=0.0.0.0:8442 --bind-metrics=0.0.0.0:9000 \
--certdir=<path to the dir with certificates> --certfile=<.crt file> --keyfile=<.key file> --tokenfile=<path to token> \
--dev-mode
```

#### inside-the-cluster mode run

SM can be installed in the cluster with helm chart. You can follow [installation guide](/documentation/public/installation.md) to deploy SM in your cloud;
Additionally, you can install/update SM as kubemarine plugin. This approach is used in [site-manager cloud tests](/tests/README.md#site-manager-cloud-tests);

Also it's possible to install the second SM in separate namespace with it's own CRD.
For it:
1. install the second CRD and webhooks manually with another kind and group; 
2. deploy site-manager with overriden `SM_KIND`, `SM_KIND_LIST` and `SM_GROUP`;

