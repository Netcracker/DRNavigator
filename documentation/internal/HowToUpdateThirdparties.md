# Third-parties and External Packages Update

<!-- TOC -->
* [Third-parties and External Packages Update](#third-parties-and-external-packages-update)
  * [Component List](#component-list)
    * [site-manager](#site-manager)
    * [paas-geo-monitor](#paas-geo-monitor)
    * [sm-check](#sm-check)
    * [sm-client](#sm-client)
    * [sm-dummy](#sm-dummy)
  * [Finding Latest Available Versions for Dependencies](#finding-latest-available-versions-for-dependencies)
<!-- TOC -->

We have to update the third parties and external packages being used in our application frequently in order to keep our application up to date and to get all the latest features and bug fixes in the third parties.

## Component List

The below subsections provide details on updates in the different components.

### [site-manager](../../site-manager)

Update golang deps:

- [go.mod](../../site-manager/go.mod)
  - applicable golang version
  - version of modules being used,  `go get -u` , `go mod tidy`
- source docker image for golang [Dockerfile](../../site-manager/Dockerfile)

### [paas-geo-monitor](../../paas-geo-monitor)

Update golang deps:

- [go.mod](../../paas-geo-monitor/go.mod)
  - applicable golang version
  - version of modules being used,  `go get -u` , `go mod tidy`
- source docker image for golang [Dockerfile](../../paas-geo-monitor/Dockerfile)

### [sm-check](../../sm-check)

- Debian image [Dockerfile](../../paas-geo-monitor/Dockerfile).

### [sm-client](../../sm-client)

- Python image [Dockerfile-sc](../../Dockerfile-sc)
- Python deps [requirements](../../requirements-sc.txt)

### [sm-dummy](../../tests/sm-dummy)

- Python image [Dockerfile](../../tests/sm-dummy/Dockerfile)
- Python deps [requirements](../../tests/sm-dummy/requirements.txt)

## Finding Latest Available Versions for Dependencies

To find the latest available versions for python dependencies, check the official repositories as follows and apply the most suitable version in the above-mentioned modifications.

- Flask - <https://pypi.org/project/Flask/#history>
- Kubernetes - <https://pypi.org/project/kubernetes/#history>
- Requests - <https://pypi.org/project/requests/#history>
- PyYaml -  <https://pypi.org/project/PyYAML/#history>
- Prettytable -  <https://pypi.org/project/prettytable/#history>
