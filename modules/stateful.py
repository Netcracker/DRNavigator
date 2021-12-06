#!/usr/bin/env python3
import logging
import utils


def get_status(service, *args, **kwargs):
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


def run_service(service, options, procedure, force, no_wait):
    """
    Method to process one service on both kubernetes clusters

    :param string service: the name of service that will be processed
    :param dict options: the dictionary with parameters related to one service
    :param string procedure: the procedure that will be processed to services
    :param string force: flag to ignore healthz of service. Can be True, true, 1
    """
    mode = procedure

    logging.debug(f"Service: {service}. Check current mode")
    resp = utils.send_get(url=options['parameters']["serviceEndpoint"])
    service_status = dict()

    if "mode" not in resp or "status" not in resp:
        logging.warning(f"Service: {service}. Service is unavailable.")
        return "fatal", service_status

    if resp["mode"] == mode and resp["status"] == "done":
        logging.warning(f"Service: {service}. Service is already {mode} and has status done.")
        service_status = {"mode": resp["mode"], "status": resp["status"], "healthz": "--"}

        # Check current health status
        if options['parameters']["healthzEndpoint"] != "":
            logging.info(f"Service: {service}. Check current health status")
            service_status["healthz"] = utils.send_get(url=options['parameters']["healthzEndpoint"])["status"]
            if not is_healthy(service, procedure, options, service_status, force):
                return "unhealthy", service_status

    else:
        logging.info(f"Service: {service}. Current mode is {resp['mode']} and status is {resp['status']}")

        # Check health status before procedure
        if options['parameters']["healthzEndpoint"] != "":

            logging.info(f"Service: {service}. Check current health status")
            healthz_resp = utils.send_get(url=options['parameters']["healthzEndpoint"])

            if healthz_resp["status"].lower() != "up" and procedure == "standby":

                logging.critical(
                    f"Service: {service}. Current health status is {healthz_resp['status'].lower()}. Service failed")
                logging.warning(f"Service: {service}. Force mode enabled. Service healthz ignored")
                if not force:
                    return "unhealthy", service_status
            else:
                logging.info(f"Service: {service}. Current health status is {healthz_resp['status'].lower()}")

        logging.info(
            f"Service: {service}. Set mode {mode}. serviceEndpoint = {options['parameters']['serviceEndpoint']}. No-wait {no_wait}")
        resp = utils.send_post(url=options['parameters']["serviceEndpoint"], mode=mode, no_wait=no_wait)
        if resp.get("bad_response") or resp.get("fatal"):
            # TODO pushing back to sm-client bad response from service.
            return "fatal", service_status

        logging.info(f"Service: {service}. Start polling")
        service_status = utils.polling(service=service,
                                       procedure=procedure,
                                       mode=mode,
                                       url=options['parameters']["serviceEndpoint"],
                                       healthz_endpoint=options['parameters']["healthzEndpoint"],
                                       allowed_standby_state_list=options["allowedStandbyStateList"],
                                       timeout=options['timeout'])

        if service_status["status"] == "running":
            logging.error(f"Service: {service}. Procedure {procedure} is in running status still.")
            return "running", service_status

        elif service_status["status"] == "failed":
            logging.error(f"Service: {service}. Procedure {procedure} is in failed status.")
            return "failed", service_status

        elif service_status["status"] == "unknown":
            logging.error(f"Service: {service}. Procedure {procedure} is in unknown status.")
            return "unknown", service_status

        elif service_status["status"] == "done":
            logging.info(f"Service: {service}. Procedure {procedure} is in done status.")

        # Check current health status
        if options['parameters']["healthzEndpoint"] != "":
            if not is_healthy(service, procedure, options, service_status, force):
                return "unhealthy", service_status
            logging.info(f"Service: {service}. Procedure {procedure} is finished successfully.")

    return "success", service_status


def is_healthy(service, procedure, options, status, force):
    if (procedure == "active" and status["healthz"].lower() != "up") or \
            (procedure == "standby" and status["healthz"].lower() not in options[
                "allowedStandbyStateList"]):

        logging.critical(
            f"Service: {service}. Current health status is {status['healthz'].lower()}. Service failed")
        logging.warning(f"Service: {service}. Force mode enabled. Service healthz ignored")
        if not force:
            return False
    else:
        logging.info(f"Service: {service}. Current health status is {status['healthz'].lower()}")
    return True


def get_module_specific_cr(item):
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
