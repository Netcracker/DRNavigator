#!/usr/bin/env python3
import logging
import json
import requests
import time
import os
import yaml
import base64
from kubernetes import client, config, watch

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
SERVICE_DEFAULT_TIMEOUT = os.environ.get("SERVICE_DEFAULT_TIMEOUT", 200)
HTTP_SCHEME = os.environ.get("HTTP_SCHEME", "http://")

# site-manager WEB server parameters
SM_WEB_HOST = os.environ.get("SM_WEB_HOST", "0.0.0.0")
SM_WEB_PORT = os.environ.get("SM_WEB_PORT", 8443)

# define authentication mode
FRONT_HTTP_AUTH = os.environ.get("FRONT_HTTP_AUTH", False) in (1, True, "Yes", "yes", "True", "true")
BACK_HTTP_AUTH = os.environ.get("BACK_HTTP_AUTH", False) in (1, True, "Yes", "yes", "True", "true")

SM_DEBUG = os.environ.get("SM_DEBUG", False)

SM_KUBECONFIG_FILE = os.environ.get("SM_KUBECONFIG_FILE", "")

DEPLOYMENT_ADDITIONAL_DELAY = 30

SM_AUTH_TOKEN = ""


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

    if BACK_HTTP_AUTH:
        headers["Authorization"] = f"Bearer {SM_AUTH_TOKEN}"

    logging.debug(f"REST url: {url}")
    logging.debug(f"REST data: {obj}")

    response = _send_post(url, obj, headers)
    if response:
        return response

    for _ in range(4):
        status = send_get(url)
        if status.get("mode", "") != mode:
            response = _send_post(url, obj, headers)
            if response:
                return response
        time.sleep(2)

    logging.fatal(f"Can't successfully send post request to service endpoint {url}")
    return dict.fromkeys(['fatal'], True)


def _send_post(url, obj, headers):
    try:
        resp = requests.post(url, timeout=20, data=obj, headers=headers)
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
    except requests.exceptions.SSLError:
        logging.error("SSL certificate verify failed")
        raise # re-raise SSL exception to handle in the calling code TBD in more general manner
        # SSLEOFError(8, 'EOF occurred in violation of protocol (_ssl.c:1091)')
        # SSLCertVerificationError(1, '[SSL: CERTIFICATE_VERIFY_FAILED] certificate verify failed: unable to get local issuer certificate (_ssl.c:1125)'))
    except requests.exceptions.JSONDecodeError as e:
        logging.error("Wrong JSON data received: \n %s" % str(e))
    except requests.exceptions.RequestException as e:
        logging.error("General request error %s",e.__doc__)
    except:
        logging.error("General error")
    return None


def send_get(url):
    """
    Method to send GET requests to services

    :param string url: the URL to service operator
    """
    headers = dict()
    if BACK_HTTP_AUTH:
        headers["Authorization"] = f"Bearer {SM_AUTH_TOKEN}"

    logging.debug(f"REST url: {url}")

    for _ in range(5):
        try:
            resp = requests.get(url, timeout=10, headers=headers)
            return resp.json()
        except requests.exceptions.SSLError:
            logging.error("SSL certificate verify failed")
            raise # re-raise SSL exception to handle in the calling code TBD in more general manner
            # SSLEOFError(8, 'EOF occurred in violation of protocol (_ssl.c:1091)')
            # SSLCertVerificationError(1, '[SSL: CERTIFICATE_VERIFY_FAILED] certificate verify failed: unable to get local issuer certificate (_ssl.c:1125)'))
        except requests.exceptions.JSONDecodeError as e:
            logging.error("Wrong JSON data received: \n %s" % str(e))
        except requests.exceptions.RequestException as e:
            logging.error("General request error %s",e.__doc__)
        except:
            logging.error("General error")

    return {}


def polling(service, procedure, mode, url, allowed_standby_state_list, healthz_endpoint=None, timeout=SERVICE_DEFAULT_TIMEOUT):
    """
    Method to poll GET|POST requests to services

    :param string service: the name of service that will be processed
    :param string procedure: the procedure that will be processed to services
    :param string url: the URL to service operator
    :param list allowed_standby_state_list: list of allowed statuses after changing service mode to standby
    :param string mode: the mode to be set for the service
    :param string healthz_endpoint: the URL to service operator healthz entrypoint
    :param int timeout: the timeout for processing service in seconds
    """

    service_status = {"mode": "unknown", "status": "unknown", "healthz": "--"}
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
            service_status["status"] = data["status"]
            time.sleep(5)
            continue

        if data["mode"] == mode:
            service_status["mode"] = data["mode"]
            if data["status"] == "failed":
                service_status["status"] = "failed"
                return service_status
            if data["status"] == "done":
                service_status["status"] = "done"
                break

    if healthz_endpoint:
        count = 0
        service_status["healthz"] = "unknown"
        while int(time.time()) < init_time + int(timeout):

            count += 1

            logging.info(f"Service: {service}. Polling service status. Iteration {count}")
            logging.info(f"Service: {service}. "
                         f"{int(timeout) - (int(time.time()) - init_time)} seconds left until timeout")

            data = send_get(healthz_endpoint)

            logging.debug(f"Service: {service}. Received status: {data}")
            if "status" not in data:
                time.sleep(5)
                continue

            if (procedure == "active" and data["status"].lower() == "up") or \
               (procedure == "standby" and data["status"].lower() in allowed_standby_state_list):
                service_status["healthz"] = data["status"]
                break

            if data["status"] in ("degraded", "down"):
                service_status["healthz"] = data["status"]
                time.sleep(5)
                continue

            time.sleep(5)

    return service_status


def poll_deployment(name, namespace, mode, options, k8s_api_client, session_data):
    """
    Method for polling deployments during the procedure

    :param string name: microservice name
    :param string namespace: microservice's namespace
    :param string mode: target value for drModeEnv variable
    :param dict options: service's CR
    :param ApiClient k8s_api_client: kube-api client
    :param dict session_data: dictionary for storing microservices statuses
    """
    dr_mode_env = options["parameters"]["drModeEnv"]
    logging.info("starting deployment check: namespace=%s name=%s " % (namespace, name))
    deployment_name = namespace + "/" + name
    session_data[deployment_name] = dict()
    deployment_info = _read_deployment_info(name, namespace, k8s_api_client)
    if not deployment_info:
        session_data[deployment_name]["mode"] = "Unknown"
        session_data[deployment_name]["status"] = "Unknown"
        return
    if not _check_env(deployment_info, options, mode):
        termination_grace_period_seconds = \
            deployment_info["spec"]["template"]["spec"].get("terminationGracePeriodSeconds", 30)
        init_time = int(time.time())
        env_updated = False
        while int(time.time()) < init_time + int(termination_grace_period_seconds):
            logging.info("trying to recheck drModeEnv value namespace=%s name=%s " % (namespace, name))
            time.sleep(5)
            deployment_info = _read_deployment_info(name, namespace, k8s_api_client)
            if not deployment_info:
                session_data[deployment_name]["mode"] = "Unknown"
                session_data[deployment_name]["status"] = "Unknown"
                return
            if _check_env(deployment_info, options, mode):
                env_updated = True
                continue
        if not env_updated and _is_env_exists(deployment_info, dr_mode_env):
            session_data[deployment_name]["mode"] = "Unchanged"
            session_data[deployment_name]["status"] = "Unhealthy"
            return
    if _check_deployment_status(deployment_info):
        if _is_env_exists(deployment_info, dr_mode_env):
            session_data[deployment_name]["mode"] = mode
        else:
            session_data[deployment_name]["mode"] = "active"
        session_data[deployment_name]["status"] = "Ready"
        return
    timeout = _get_timeout_for_deployment(deployment_info)
    init_time = int(time.time())
    while int(time.time()) < init_time + int(timeout):
        logging.info("trying to recheck deployment status namespace=%s name=%s " % (namespace, name))
        time.sleep(5)
        deployment_info = _read_deployment_info(name, namespace, k8s_api_client)
        if not deployment_info:
            session_data[deployment_name]["mode"] = "Unknown"
            session_data[deployment_name]["status"] = "Unknown"
            return
        if _check_deployment_status(deployment_info):
            if _is_env_exists(deployment_info, dr_mode_env):
                session_data[deployment_name]["mode"] = mode
            else:
                session_data[deployment_name]["mode"] = "active"
            session_data[deployment_name]["status"] = "Ready"
            return
    if _is_env_exists(deployment_info, dr_mode_env):
        session_data[deployment_name]["mode"] = mode
    else:
        session_data[deployment_name]["mode"] = "active"
    session_data[deployment_name]["status"] = "Unhealthy"
    return


def _check_env(deployment_info, options, mode):
    """
    Method for finding and checking the deployment env variable

    :param dict deployment_info: content of deployment config
    :param dict options: service's CR
    :param string mode: target value for drModeEnv variable
    """

    dr_mode_env = options["parameters"]["drModeEnv"]
    for container in deployment_info["spec"]["template"]["spec"]["containers"]:
        env_updated = False
        for env in container.get("env", {}):
            if env["name"] == dr_mode_env and env["value"] == mode:
                env_updated = True
                break
        if not env_updated:
            return False
    return True


def _is_env_exists(deployment_info, dr_mode_env):
    counter = 0
    for container in deployment_info["spec"]["template"]["spec"]["containers"]:
        for env in container.get("env", {}):
            if env["name"] == dr_mode_env:
                counter += 1
                break
    if counter == 0:
        return False
    return True


def _check_deployment_status(deployment_info):
    """
    Method to check the readiness of the microservice

    :param dict deployment_info: content of deployment config
    """

    for condition in deployment_info["status"]["conditions"]:
        if condition["type"] == "Available" and \
           condition["status"] == "True" and \
           deployment_info["status"]["replicas"] == deployment_info["status"]["ready_replicas"]:
            return True


def _read_deployment_info(name, namespace, k8s_api_client):
    """
    Method for loading from kubernetes content of deployment config by name and namespace

    :param string name: microservice name
    :param string name: microservice namespace
    :param ApiClient k8s_api_client: kube-api client
    """

    deployment_info = dict()
    try:
        deployment_info = client.AppsV1Api(api_client=k8s_api_client).read_namespaced_deployment(name, namespace).to_dict()
    except Exception as e:
        logging.error("Can't find deployment: \n %s" % str(e))
    return deployment_info


def _get_timeout_for_deployment(deployment_info):
    timeouts = list()
    for i, container in enumerate(deployment_info["spec"]["template"]["spec"]["containers"]):
        if container.get("readinessProbe", {}):
            rp = container["readinessProbe"]
            timeouts.append(rp["initialDelaySeconds"] + (rp["timeoutSeconds"] + rp["periodSeconds"]) \
                      * rp["failureThreshold"] + DEPLOYMENT_ADDITIONAL_DELAY)
        else:
            timeouts.append(DEPLOYMENT_ADDITIONAL_DELAY)
    timeouts.sort(reverse=True)
    logging.error(f"timeouts: {timeouts}")
    return timeouts[0]


def collect_deployments_statuses(dr_marker, dr_mode_env):
    """
    Method for loading current microservices status

    :param string dr_marker: the label for finding deployments for processing
    :param string dr_mode_env: deployment's env variable name
    """

    results = dict()
    if SM_KUBECONFIG_FILE != "":
        k8s_api_client = config.load_kube_config(config_file=SM_KUBECONFIG_FILE)
    else:
        k8s_api_client = config.load_incluster_config()
    try:
        deployments = client.AppsV1Api(api_client=k8s_api_client).list_deployment_for_all_namespaces(
            label_selector=dr_marker).to_dict()
    except Exception as e:
        logging.error("Can't get Deployments list: \n %s" % str(e))
        return {}
    for item in deployments["items"]:
        deployment_name = item["metadata"]["namespace"] + "/" + item["metadata"]["name"]
        dr_mode_env_value = "Unknown"
        status = "Unhealthy"
        for container in item["spec"]["template"]["spec"]["containers"]:
            for env in container.get("env", {}):
                if env["name"] == dr_mode_env:
                    dr_mode_env_value = env["value"]
                    break
        for condition in item["status"]["conditions"]:
            if condition["type"] == "Available" and condition["status"] == "True":
                status = "Not Ready"
                if item["status"]["replicas"] == item["status"]["ready_replicas"]:
                    status = "Ready"
        results[deployment_name] = dict()
        if _is_env_exists(item, dr_mode_env):
            results[deployment_name]["mode"] = dr_mode_env_value
        else:
            results[deployment_name]["mode"] = "active"
        results[deployment_name]["status"] = status
    return results


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
