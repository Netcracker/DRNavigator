# Thirdparties And External Packages Update

- [Thirdparties And External Packages Update](#thirdparties-and-external-packages-update)
  - [Component list](#component-list)
    - [site-manager](#site-manager)
    - [paas-geo-monitor](#paas-geo-monitor)
    - [sm-check](#sm-check)
    - [sm-client](#sm-client)
    - [sm-dummy](#sm-dummy)
  - [Finding latest available versions for dependencies](#finding-latest-available-versions-for-dependencies)

We have to update the thirdparties and external packages being used in our application frequently in order to keep our application up to date and get all the latest fetures and bug fixes in the thirdparties.

## Component list

### [site-manager](../../site-manager/)

Update golang deps:

- [go.mod](../../site-manager/go.mod)
  - aplicable golang version
  - version of modules being used,  `go get -u` , `go mod tidy`
- source docker image for golang [Dockerfile](../../site-manager/Dockerfile)

### [paas-geo-monitor](../../paas-geo-monitor)

Update golang deps:

- [go.mod](../../paas-geo-monitor/go.mod)
  - aplicable golang version
  - version of modules being used,  `go get -u` , `go mod tidy`
- source docker image for golang [Dockerfile](../..//paas-geo-monitor/Dockerfile)

### [sm-check](../../sm-check/)

- Debian image [Dockerfile](../../paas-geo-monitor/Dockerfile#L1).

### [sm-client](../../sm-client)

- Python image [Dockerfile-sc](../..//Dockerfile-sc#L1)
- Python deps [requirements](../../requirements-sc.txt)

### [sm-dummy](../../tests/sm-dummy/)

- Python image [Dockerfile](../../tests/sm-dummy/Dockerfile)
- Python deps [requirements](../../tests/sm-dummy/requirements.txt)

## Finding latest available versions for dependencies

To find the latest avaialbe versions for python dependancies check the official repositories as follows and apply most suitable version in above mentioned modifications.

1. Flask - <https://pypi.org/project/Flask/#history>
2. Kubernetes - <https://pypi.org/project/kubernetes/#history>
3. Requests - <https://pypi.org/project/requests/#history>
4. PyYaml -  <https://pypi.org/project/PyYAML/#history>
5. Prettytable -  <https://pypi.org/project/prettytable/#history>
