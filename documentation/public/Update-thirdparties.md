# Thirdparties And External Packages Update

We have to update the thirdparties and external packages being used in our application frequently in order to keep our application up to date and get all the latest fetures and bug fixes in the thirdparties.

We have to check and updated thirdparties for below services -
1. site-manager
2. paas-geo-monitor
3. sm-check
4. sm-client
5. sm-ummy

## Thirdparties For site-manage

To address vulnerabilites with site-manager image, we need to make chnages in https://github.com/Netcracker/DRNavigator/tree/main/site-manager

* Update go version and golng image if applicable :
  1. To update go version, modify [go.mod](https://github.com/Netcracker/DRNavigator/blob/main/site-manager/go.mod) file and provide aplicable go version.
  2. To update latest version of modules being used, run command `go get -u` and then run `go mod tidy`
  3. To update source docker image for golang, we need to change image in [Dockerfile](https://github.com/Netcracker/DRNavigator/blob/main/site-manager/Dockerfile).

## Thirdparties For paas-geo-monitor

To address vulnerabilites with paas-geo-monitor image, we need to make chnages in https://github.com/Netcracker/DRNavigator/tree/main/paas-geo-monitor
* Update go version and golng image if applicable :
  1. To update go version, modify [go.mod](https://github.com/Netcracker/DRNavigator/blob/main/paas-geo-monitor/go.mod) file and provide aplicable go version.
  2. To update latest version of modules being used, run command `go get -u` and then run `go mod tidy`
  3. To update source docker image for golang, we need to change image in [Dockerfile](https://github.com/Netcracker/DRNavigator/blob/main/paas-geo-monitor/Dockerfile).

## Thirdparties For sm-check

To address vulnerabilites with sm-check image, we need to make chnages in https://github.com/Netcracker/DRNavigator/tree/main/sm-check  

* Update debian image as required and as applicable in [Dockerfile](https://github.com/Netcracker/DRNavigator/blob/main/paas-geo-monitor/Dockerfile).

## Thirdparties For sm-client

To address vulnerabilites with sm-client image, we need to make changes in [Dockerfile-sc](https://github.com/Netcracker/DRNavigator/blob/main/Dockerfile-sc)

1. Update python image as required and as applicable.
2. Update dependancies by modifying [requirements](https://github.com/Netcracker/DRNavigator/blob/main/requirements-sc.txt) file as applicable.

## Thirdparties For sm-dummy

To address vulnerabilites with sm-dummy image, we need to make changes in https://github.com/Netcracker/DRNavigator/tree/main/tests/sm-dummy
1. Update Python image as required and as applicable by modifying [Dockerfile](https://github.com/Netcracker/DRNavigator/blob/main/tests/sm-dummy/Dockerfile)
2. Update dependancies by modifying [requirements](https://github.com/Netcracker/DRNavigator/blob/main/tests/sm-dummy/requirements.txt) file as applicable.

## Finding laetst avialbe versions for dependacies

To find the latest avaialbe versions for python dependancies check the official repositories as follows and apply most suitable version in above mentioned modifications.
1. Flask - https://pypi.org/project/Flask/#history
2. Kubernetes - https://pypi.org/project/kubernetes/#history
3. Requests - https://pypi.org/project/requests/#history
4. PyYaml -  https://pypi.org/project/PyYAML/#history
5. Prettytable -  https://pypi.org/project/prettytable/#history