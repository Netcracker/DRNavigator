#!/usr/bin/env python3
import logging
import utils


def get_status(service, *args, **kwargs):
    """
    Method that collects complete information about the state of the service

    :param dict service: service's CR
    """

    output = dict()
    status = utils.send_get(service['parameters']["serviceEndpoint"])
    output["mode"] = status.get("mode", "--")
    output["status"] = status.get("status", "--")
    if service['parameters'].get("healthzEndpoint", "") != "":
        healthz = utils.send_get(service['parameters']["healthzEndpoint"])
        output["healthz"] = healthz.get("status", "--")
    else:
        output["healthz"] = "--"

    return output


def get_module_specific_cr(item):
    """
    Method preparing dictionary based on the service's CR

    :param dict item: service's CR
    """

    if item['spec']['sitemanager']['parameters'].get('serviceEndpoint', '') != '':

        if item["spec"]["sitemanager"]['parameters']["serviceEndpoint"].startswith("http://") or \
                item["spec"]["sitemanager"]['parameters']["serviceEndpoint"].startswith("https://"):
            service_endpoint = item["spec"]["sitemanager"]['parameters']["serviceEndpoint"]
        else:
            service_endpoint = utils.HTTP_SCHEME + item["spec"]["sitemanager"]['parameters']["serviceEndpoint"]
    else:
        service_endpoint = ''

    if item['spec']['sitemanager']['parameters'].get('healthzEndpoint', '') != '':

        if item['spec']['sitemanager']['parameters']['healthzEndpoint'].startswith("http://") or \
                item['spec']['sitemanager']['parameters']['healthzEndpoint'].startswith("https://"):
            healthz_endpoint = item['spec']['sitemanager']['parameters']['healthzEndpoint']
        else:
            healthz_endpoint = utils.HTTP_SCHEME + item['spec']['sitemanager']['parameters']['healthzEndpoint']
    else:
        healthz_endpoint = ''
    allowed_standby_state_list = [i.lower() for i in item['spec']['sitemanager'].get('allowedStandbyStateList', ["up"])]

    return {"namespace": item["metadata"]["namespace"],
            "module": item['spec']['sitemanager'].get('module', ''),
            "after": item['spec']['sitemanager'].get('after', []),
            "before": item['spec']['sitemanager'].get('before', []),
            "sequence": item['spec']['sitemanager'].get('sequence', []),
            "allowedStandbyStateList": allowed_standby_state_list,
            "timeout": item['spec']['sitemanager'].get('timeout', utils.SERVICE_DEFAULT_TIMEOUT),
            "parameters":
                {"serviceEndpoint": service_endpoint,
                 "healthzEndpoint": healthz_endpoint}}
