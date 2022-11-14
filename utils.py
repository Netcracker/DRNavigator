#!/usr/bin/env python3
import logging
import json
import ssl
from typing import Tuple, Dict

import requests.packages
import time
import os
import yaml
import base64
from kubernetes import client, config, watch
from requests.adapters import HTTPAdapter, Retry
from urllib3.exceptions import InsecureRequestWarning

SM_CONFIG_FILE = os.environ.get("SM_CONFIG_FILE", "")
if SM_CONFIG_FILE != "":
    try:
        SM_CONFIG = yaml.load(open(SM_CONFIG_FILE), Loader=yaml.FullLoader)
    except Exception as e:
        logging.fatal("Can not parse configuration file!: \n %s" % str(e))
        exit(1)
else:
    SM_CONFIG = {}

# Define kubernetes CR objects
SM_GROUP = os.environ.get("SM_GROUP", "netcracker.com")
SM_PLURAL = os.environ.get("SM_PLURAL", "sitemanagers")
SM_VERSION = os.environ.get("SM_VERSION", "v2")

# Define services default parameters
HTTP_SCHEME = os.environ.get("HTTP_SCHEME", "http://")

# site-manager WEB server parameters
SM_WEB_HOST = os.environ.get("SM_WEB_HOST", "0.0.0.0")
SM_WEB_PORT = os.environ.get("SM_WEB_PORT", 8443)

# define authentication mode
FRONT_HTTP_AUTH = os.environ.get("FRONT_HTTP_AUTH", False) in (1, True, "Yes", "yes", "True", "true")
BACK_HTTP_AUTH = os.environ.get("BACK_HTTP_AUTH", False) in (1, True, "Yes", "yes", "True", "true")

SM_DEBUG = os.environ.get("SM_DEBUG", False) in (1, True, "Yes", "yes", "True", "true")

SM_KUBECONFIG_FILE = os.environ.get("SM_KUBECONFIG_FILE", "")

SM_AUTH_TOKEN = ""

SM_CACERT = os.environ.get("SM_CACERT", True)
if SM_CACERT in ("Yes", "yes", "No", "no", "True", "true", "False", "false"):
    SM_CACERT = SM_CACERT in ("Yes", "yes", "True", "true")

SM_GET_REQUEST_TIMEOUT = int(os.environ.get("SM_GET_REQUEST_TIMEOUT", 5))
SM_POST_REQUEST_TIMEOUT = int(os.environ.get("SM_POST_REQUEST_TIMEOUT", 30))

def send_post(url, mode, no_wait):
    """ Method to send POST requests to services
    @param string url: the URL to service operator
    @param string mode: is the role of cluster part
    @param bool no_wait: special flag for microservice to show type of replication between of parts of database cluster
    """
    
    obj = {"mode": mode, "no-wait": no_wait}
    headers = {
        'Content-type': 'application/json',
        'Accept': 'application/json'
    }

    if SM_AUTH_TOKEN:
        headers["Authorization"] = f"Bearer {SM_AUTH_TOKEN}"

    logging.debug(f"REST url: {url}")
    logging.debug(f"REST data: {obj}")

    _, response, _ = io_make_http_json_request(url, http_body=obj,token=SM_AUTH_TOKEN,use_auth=BACK_HTTP_AUTH)
    if response:
        return response

    for _ in range(4):
        _, status, _ = io_make_http_json_request(url)
        if status.get("mode", "") != mode:
            _, response, _ = io_make_http_json_request(url, http_body=obj,token=SM_AUTH_TOKEN,use_auth=BACK_HTTP_AUTH)
            if response:
                return response
        time.sleep(2)

    logging.fatal(f"Can't successfully send post request to service endpoint {url}")
    return dict.fromkeys(['fatal'], True)


def io_make_http_json_request(url="", token="", verify=True, http_body:dict=None, retry=3, use_auth=FRONT_HTTP_AUTH) -> Tuple[bool, Dict, int]:
    """ Sends GET/POST request to service
    @param string url: the URL to service operator
    @param token: Bearer token
    @param verify: Server side SSL verification
    @param retry: the number of retries
    @param http_body: the dictionary with procedure and list of services
    @returns: True/False, Dict with not empty json body in case Ok/{}, HTTP_CODE/
    IO SSL codes: ssl.SSLErrorNumber.SSL_ERROR_SSL/SSLErrorNumber.SSL_ERROR_EOF
    """
    if not os.getenv("DEBUG"):
        # Disable warnings about self-signed certificates from requests library
        requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

    if token != "" and use_auth:
        headers = {"Authorization": f"Bearer {token}"}
    else:
        headers = {}
    if not http_body:
        http_body = {}
    logging.debug(f"REST url: {url}")
    logging.debug(f"REST data: {http_body}")

    session = requests.Session()
    retries = Retry(total=retry)
    session.mount('https://', HTTPAdapter(max_retries=retries))
    session.mount('http://', HTTPAdapter(max_retries=retries))

    logging.getLogger("urllib3").setLevel(logging.CRITICAL)

    try:
        if any(http_body):
            resp = session.post(url, json=http_body, timeout=SM_POST_REQUEST_TIMEOUT, headers=headers, verify=verify)
        else:
            resp = session.get(url, timeout=SM_GET_REQUEST_TIMEOUT, headers=headers, verify=verify)
        logging.debug(f"REST response: {resp.json()}")
        return True, resp.json() if resp.json() else {}, resp.status_code # return ANY content with HTTP code

    except requests.exceptions.SSLError as e:
        logging.error("SSL certificate verify failed")
        #TODO in more accurate manner error handling manner
        if "SSLCertVerificationError" in str(e.args): ## SSL Verification fails ; SSLCertVerificationError(1, '[SSL: CERTIFICATE_VERIFY_FAILED] certificate verify failed: unable to get local issuer certificate (_ssl.c:1125)')
            return False, {}, ssl.SSLErrorNumber.SSL_ERROR_SSL.__int__() # - 1
        elif "SSLEOFError" in str(e.args): # SSL connect error, SSL resource is not accessible vi ha-proxy  ; SSLEOFError(8, 'EOF occurred in violation of protocol (_ssl.c:1091)')
            #TODO need a test for this case
            return False, {}, ssl.SSLErrorNumber.SSL_ERROR_EOF.__int__() # - 8
    except requests.exceptions.JSONDecodeError as e:
        logging.error("Wrong JSON data received %s", e)
    except requests.exceptions.RequestException as e:
        logging.error("General request error %s", e)
    except Exception as e:
        logging.error("General error %s",e)

    return False,{},False


def get_token(api_watch=False):
    """
    Method to get token of sm-auth-sa from kubernetes. Method rewrites global var SM_CLIENT_TOKEN with actual token value

    :param bool api_watch: special flag to define method mode: get token once or follow the token changes.
    """
    global SM_AUTH_TOKEN

    # In testing mode return stab
    if SM_CONFIG.get("testing", {}).get("enabled", False) and \
            SM_CONFIG.get("testing", {}).get("sm_dict", {}) != {}:

        SM_AUTH_TOKEN = SM_CONFIG["testing"].get("token", "123")

        return

    if SM_KUBECONFIG_FILE != "":
        k8s_api_client = config.load_kube_config(config_file=SM_KUBECONFIG_FILE)

        _, current_context = config.list_kube_config_contexts(config_file=SM_KUBECONFIG_FILE)
        namespace = current_context['context'].get('namespace', 'default')

    else:
        k8s_api_client = config.load_incluster_config()
        namespace = open("/var/run/secrets/kubernetes.io/serviceaccount/namespace").read()

    logging.info(f"Current namespace: {namespace}")

    if not api_watch:

        try:
            service_account = client.CoreV1Api(api_client=k8s_api_client).read_namespaced_service_account("sm-auth-sa", namespace)
            secret_name = [s for s in service_account.secrets if 'token' in s.name][0].name
            btoken = client.CoreV1Api(api_client=k8s_api_client).read_namespaced_secret(
                name=secret_name, namespace=namespace).data['token']
            token = base64.b64decode(btoken).decode()

        except Exception as e:
            logging.error("Can not get sm-auth-sa token: \n %s" % str(e))
            os._exit(1)

        SM_AUTH_TOKEN = token

    else:
        counter = 1
        w = watch.Watch()

        while True:
            logging.debug(f"Start watching serviceaccount sm-auth-sa. Iteration {counter}")
            counter += 1

            for event in w.stream(client.CoreV1Api(api_client=k8s_api_client).list_namespaced_service_account,
                                  namespace,
                                  timeout_seconds=30):
                if event['object'].metadata.name == "sm-auth-sa":
                    if event['type'] in ["ADDED", "MODIFIED"]:
                        try:
                            secret_name = [s for s in event['object'].secrets][0].name
                        except: # hit here when secret for appropriate  SA is not ready yet
                            continue

                        btoken = client.CoreV1Api(api_client=k8s_api_client).read_namespaced_secret(
                            name=secret_name, namespace=namespace).data['token']
                        token = base64.b64decode(btoken).decode()

                        logging.info(f"Serviceaccount sm-auth-sa was {event['type']}. Token was updated.")

                        SM_AUTH_TOKEN = token

                    if event['type'] == "DELETED":
                        logging.fatal("Serviceaccount sm-auth-sa was deleted. Exit")
                        os._exit(1)
            time.sleep(15)
