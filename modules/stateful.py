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


def is_healthy(service, procedure, options, status, force):
    """
    Method for checking the status of a service during the execution of a procedure

    :param string service: the name of service that will be processed
    :param string procedure: the procedure that will be processed to services
    :param dict options: service's CR
    :param dict status: the dictionary containing the state of the service
    :param bool force: flag to ignore healthz of service.
    """

    if (procedure == "active" and status["healthz"].lower() != "up") or \
            (procedure == "standby" and status["healthz"].lower() not in options[
                "allowedStandbyStateList"]):

        logging.critical(
            f"Service: {service}. Current health status is {status['healthz'].lower()}. Service failed")
        if not force:
            return False
        logging.warning(f"Service: {service}. Force mode enabled. Service healthz ignored")
    else:
        logging.info(f"Service: {service}. Current health status is {status['healthz'].lower()}")
    return True


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
