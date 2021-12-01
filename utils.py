#!/usr/bin/env python3
import logging
import json
import requests
import time
import os
import yaml

# Define kubernetes CR objects
SM_GROUP = os.environ.get("SM_GROUP", "netcracker.com")
SM_PLURAL = os.environ.get("SM_PLURAL", "sitemanagers")
SM_VERSION = os.environ.get("SM_VERSION", "v2")

# Define services default parameters
SERVICE_DEFAULT_TIMEOUT = os.environ.get("SERVICE_DEFAULT_TIMEOUT", 200)
HTTP_SCHEME = os.environ.get("HTTP_SCHEME", "http://")

SM_HTTP_AUTH = os.environ.get("SM_HTTP_AUTH", False) in (1, True, "Yes", "yes", "True", "true")
SM_CLIENT_TOKEN = ""

# site-manager WEB server parameters
SM_WEB_HOST = os.environ.get("SM_WEB_HOST", "0.0.0.0")
SM_WEB_PORT = os.environ.get("SM_WEB_PORT", 8080)

SM_DEBUG = os.environ.get("SM_DEBUG", False)

# Set main parameters from env
SM_KUBECONFIG_FILE = os.environ.get("SM_KUBECONFIG_FILE", "")

SM_CONFIG_FILE = os.environ.get("SM_CONFIG_FILE", "")
if SM_CONFIG_FILE != "":
    try:
        SM_CONFIG = yaml.load(open(SM_CONFIG_FILE), Loader=yaml.FullLoader)
    except Exception as e:
        logging.fatal("Can not parse configuration file!: \n %s" % str(e))
        exit(1)
else:
    SM_CONFIG = {}


def send_post(url, mode, no_wait):
    """
    Method to send POST requests to services

    :param string url: the URL to service operator
    :param string mode: is the role of cluster part
    :param bool no_wait: special flag for microservice to show type of replication between of parts of database cluster
    """

    obj = json.dumps({"mode": mode, "no-wait": no_wait})
    headers = {
        'Content-type': 'application/json',
        'Accept': 'application/json'
    }
    logging.debug(f"REST url: {url}")
    logging.debug(f"REST data: {obj}")

    for _ in range(5):
        try:
            resp = requests.post(url, timeout=10, data=obj, headers=headers)
            logging.debug(f"REST response: {resp} and return code: {resp.status_code}")
            response = resp.json()
            ret_code = resp.status_code
            if ret_code == 200:
                if response.get("message", ""):
                    logging.info(f"Code: {ret_code}. Message: {response['message']}")
            if ret_code != 200:
                if response.get("message", ""):
                    logging.error(f"Code: {ret_code}. Message: {response['message']}")
                response["bad_response"] = ret_code
            return response
        except Exception as e:
            logging.error("Wrong JSON data received: \n %s" % str(e))

        time.sleep(2)
    logging.fatal(f"Can't successfully send post request to service endpoint {url}")
    return dict.fromkeys(['fatal'], True)


def send_get(url):
    """
    Method to send GET requests to services

    :param string url: the URL to service operator
    """

    logging.debug(f"REST url: {url}")

    for _ in range(5):
        try:
            resp = requests.get(url, timeout=10)
            return resp.json()
        except Exception as e:
            logging.error("Wrong JSON data received: \n %s" % str(e))

    return {}


def polling(service, url, mode, healthz_endpoint=None,  timeout=SERVICE_DEFAULT_TIMEOUT):
    """
    Method to poll GET|POST requests to services

    :param string service: the name of service that will be processed
    :param string url: the URL to service operator
    :param string mode: the mode to be set for the service
    :param string healthz_endpoint: the URL to service operator healthz entrypoint
    :param int timeout: the timeout for processing service in seconds
    """

    result = {"procedure_status": "unknown",
              "service_status": "unknown"}

    init_time = int(time.time())

    count = 0
    while int(time.time()) < init_time + int(timeout):
        count += 1

        logging.info(f"Service: {service}. Polling procedure status. Iteration {count}")
        logging.info(f"Service: {service}. {int(timeout) - (int(time.time()) - init_time)} seconds left until timeout")

        data = send_get(url)

        logging.debug(f"Service: {service}. Received data: {data}")
        if "status" not in data or "mode" not in data:
            time.sleep(5)
            continue

        if data["status"] == "running":
            result["procedure_status"] = "running"
            time.sleep(5)
            continue

        if data["mode"] == mode:
            if data["status"] == "failed":
                result["procedure_status"] = "failed"
                return result
            if data["status"] == "done":
                result["procedure_status"] = "done"
                break

    count = 0
    while int(time.time()) < init_time + int(timeout):

        if healthz_endpoint is None or healthz_endpoint == "":
            result["service_status"] = "Unknown"
            break

        count += 1

        logging.info(f"Service: {service}. Polling service status. Iteration {count}")
        logging.info(f"Service: {service}. {int(timeout) - (int(time.time()) - init_time)} seconds left until timeout")

        data = send_get(healthz_endpoint)

        logging.debug(f"Service: {service}. Received status: {data}")
        if "status" not in data:
            continue

        if data["status"] in ("degraded", "down"):
            result["service_status"] = data["status"]
            continue

        if data["status"] in ("up", "disable"):
            result["service_status"] = data["status"]
            break

        time.sleep(5)

    return result
