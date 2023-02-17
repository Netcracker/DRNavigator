#!/usr/bin/env python3
"""
Company:     NetCracker
Author:      Core PaaS Group
Version:     0.7
Date:        2021-11-19
Description: Service for management of microservices in active-standby scheme of kubernetes cluster
"""

import logging
import threading
import http
import copy

from kubernetes.client import ApiException

import utils
from flask import Flask, request, jsonify, make_response
from kubernetes import client, config
from prometheus_flask_exporter.multiprocess import GunicornInternalPrometheusMetrics
from prometheus_client import Gauge
from server_utils import *


# List of possible procedures
command_list = ["active", "standby", "disable", "list", "status"]

app = Flask(__name__)

app.config['DEBUG'] = utils.SM_DEBUG
app.config['JSONIFY_PRETTYPRINT_REGULAR'] = True
#metrics = GunicornInternalPrometheusMetrics(app)

site_manager_health = Gauge('site_manager_health', 'SM pod health')

if utils.SM_DEBUG:
    logging_level = logging.DEBUG
    logging_format = "[%(asctime)s] [%(process)d] [%(levelname)s] %(filename)s.%(funcName)s(%(lineno)d): %(message)s"
else:
    logging_level = logging.INFO
    logging_format = "[%(asctime)s] [%(process)d] [%(levelname)s] %(filename)s: %(message)s"

logging.basicConfig(format=logging_format, level=logging_level)

lock = threading.Lock()

if utils.FRONT_HTTP_AUTH or utils.BACK_HTTP_AUTH:
    utils.get_token(False)

    w_thread = threading.Thread(target=utils.get_token,
                                args=[True, ])
    w_thread.start()


def get_sitemanagers_dict(api_version=utils.SM_VERSION):
    """
    Method creates dictionary of donwloaded data from sitemanager CRs
    """

    # In testing mode return stab
    if utils.SM_CONFIG.get("testing", {}).get("enabled", False) and \
       utils.SM_CONFIG.get("testing", {}).get("sm_dict", {}) != {}:

       return utils.SM_CONFIG["testing"]["sm_dict"]

    if utils.SM_KUBECONFIG_FILE != "":
        k8s_api_client = config.load_kube_config(config_file=utils.SM_KUBECONFIG_FILE)

    else:
        k8s_api_client = config.load_incluster_config()

    try:
        response = client.CustomObjectsApi(api_client=k8s_api_client).list_cluster_custom_object(group=utils.SM_GROUP,
                                                                                                 version=api_version,
                                                                                                 plural=utils.SM_PLURAL,
                                                                                                 _request_timeout=10)

    except ApiException as e:
        if e.status == 404:
            logging.error(f"Can't get sitemanager objects: Desired CRD not found:\n"
                          f"\tplural={utils.SM_PLURAL}\n"
                          f"\tgroup={utils.SM_GROUP}\n"
                          f"\tversion={api_version}")
            raise utils.ProcedureException(output={"message": "Can't get sitemanager objects: Desired CRD not found",
                                                   "plural": utils.SM_PLURAL,
                                                   "group": utils.SM_GROUP,
                                                   "version": api_version})
        else:
            raise e

    output = dict()
    output['services'] = {}
    for item in response["items"]:
        serviceName = item['spec']['sitemanager'].get("alias",
                                           "%s.%s" % (item['metadata'].get('name'), item['metadata'].get('namespace')))
        output['services'][serviceName] = \
            get_module_specific_cr(item)

    return output


def json_response(code, body):
    """
    Method to perform json output
    :param int code: http code for operation
    :param dict body: dict with message or lists for http response
    """
    response = make_response(jsonify(body), code,)
    response.headers["Content-Type"] = "application/json"

    return response


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

    result = {"CRname": item["metadata"]["name"],
            "namespace": item["metadata"]["namespace"],
            "module": item['spec']['sitemanager'].get('module', ''),
            "after": item['spec']['sitemanager'].get('after', []),
            "before": item['spec']['sitemanager'].get('before', []),
            "sequence": item['spec']['sitemanager'].get('sequence', []),
            "allowedStandbyStateList": allowed_standby_state_list,
            "parameters":
                {"serviceEndpoint": service_endpoint,
                 "healthzEndpoint": healthz_endpoint}}
    if 'timeout' in item['spec']['sitemanager']:
        result['timeout'] = item['spec']['sitemanager']['timeout']
    if 'alias' in item['spec']['sitemanager']:
        result['alias'] = item['spec']['sitemanager']['alias']
    return result


@app.route('/', methods=['GET'])
def root_get():
    """
    Method for processing GET requests to /

    Right now is under construction and will be completely changed
    """

    return "Under construction"


@app.route('/validate', methods=['POST'])
def cr_validate():
    data = request.json["request"]
    logging.debug(f"Initial object from API for validating: {data}")
    allowed = True
    message = "All checks passed"
    uid = data["uid"]

    # Check name for unique
    sm_dict = get_sitemanagers_dict()
    service_name = data['object']['spec']['sitemanager'].get("alias",
                                                             "%s.%s" % (data.get('name'), data.get('namespace')))
    existed_cr = sm_dict['services'].get(service_name, None)
    if existed_cr is not None and existed_cr['namespace'] != data['namespace']:
        allowed = False
        message = f"Can't use alias {service_name}, this name is used for another service" \
            if "alias" in data['object']['spec']['sitemanager'] else \
            f"Can't use service with calculated name {service_name}, this name is used for another service"
        logging.debug(f"CR validation fails: {message}")

    return jsonify({"apiVersion": "admission.k8s.io/v1",
                    "kind": "AdmissionReview",
                    "response": {
                        "allowed": allowed,
                        "uid": uid,
                        "status": {"message": message}}})


@app.route('/convert', methods=['POST'])
def cr_convert():
    logging.debug(f"Initial object from API for converting: {request.json['request']}")

    spec = request.json["request"]["objects"]
    modified_spec = copy.deepcopy(spec)
    for i in range(len(modified_spec)):
        # v1 -> v2, v3 conversion
        if modified_spec[i]["apiVersion"] == "netcracker.com/v1":
            if "module" not in modified_spec[i]["spec"]["sitemanager"]:
                modified_spec[i]["spec"]["sitemanager"]["module"] = modified_spec[i]["spec"]["sitemanager"].get(
                    "module", "stateful")

            if "parameters" not in modified_spec[i]["spec"]["sitemanager"]:
                modified_spec[i]["spec"]["sitemanager"]["parameters"] = {}
                modified_spec[i]["spec"]["sitemanager"]["parameters"]["serviceEndpoint"] = modified_spec[i]["spec"][
                    "sitemanager"].pop("serviceEndpoint", "")
                modified_spec[i]["spec"]["sitemanager"]["parameters"]["ingressEndpoint"] = modified_spec[i]["spec"][
                    "sitemanager"].pop("ingressEndpoint", "")
                modified_spec[i]["spec"]["sitemanager"]["parameters"]["healthzEndpoint"] = modified_spec[i]["spec"][
                    "sitemanager"].pop("healthzEndpoint", "")

        # v2 -> v3 conversion
        if request.json["request"]["desiredAPIVersion"] == "netcracker.com/v3":
            # TODO: It's needed for automatic conversion not stateful modules
            if modified_spec[i]["spec"]["sitemanager"]["module"] != "stateful":
                modified_spec[i]["spec"]["sitemanager"]["alias"] = modified_spec[i]["metadata"]["name"]

            # skip CR, if it doesn't have any dependencies
            if modified_spec[i]["spec"]["sitemanager"]["before"] or modified_spec[i]["spec"]["sitemanager"]["after"]:
                sm_dict = get_sitemanagers_dict("v2")

                # Before services
                for j in range(len(modified_spec[i]["spec"]["sitemanager"]["before"])):
                    before_service_name = modified_spec[i]["spec"]["sitemanager"]["before"][j]
                    before_services = [key for key, value in sm_dict["services"].items()
                                       if value["CRname"] == before_service_name]
                    if not before_services:
                        logging.error("Found non-exist before dependency %s for CR %s" %
                                      (before_service_name, modified_spec[i]["metadata"]["name"]))
                    else:
                        modified_spec[i]["spec"]["sitemanager"]["before"][j] = before_services[0]

                # After services
                for j in range(len(modified_spec[i]["spec"]["sitemanager"]["after"])):
                    after_service_name = modified_spec[i]["spec"]["sitemanager"]["after"][j]
                    after_services = [key for key, value in sm_dict["services"].items()
                                      if value["CRname"] == after_service_name]
                    if not after_services:
                        logging.error("Found non-exist after dependency %s for CR %s" %
                                      (after_service_name, modified_spec[i]["metadata"]["name"]))
                    else:
                        modified_spec[i]["spec"]["sitemanager"]["after"][j] = after_services[0]

            if modified_spec[i]["spec"]["sitemanager"].get("parameters", {}).get("ingressEndpoint"):
                del modified_spec[i]["spec"]["sitemanager"]["parameters"]["ingressEndpoint"]
            modified_spec[i]["apiVersion"] = request.json["request"]["desiredAPIVersion"]

        # v3 -> v2 conversion
        if request.json["request"]["desiredAPIVersion"] == "netcracker.com/v2":
            modified_spec[i]["apiVersion"] = request.json["request"]["desiredAPIVersion"]

        # v3,v2->v1 conversion
        if request.json["request"]["desiredAPIVersion"] == "netcracker.com/v1":
            modified_spec[i]["apiVersion"] = request.json["request"]["desiredAPIVersion"]

            if "module" in modified_spec[i]["spec"]["sitemanager"]:
                del modified_spec[i]["spec"]["sitemanager"]["module"]

            if "parameters" in modified_spec[i]["spec"]["sitemanager"]:
                modified_spec[i]["spec"]["sitemanager"]["serviceEndpoint"] = modified_spec[i]["spec"]["sitemanager"]["parameters"].pop("serviceEndpoint", "")
                modified_spec[i]["spec"]["sitemanager"]["ingressEndpoint"] = modified_spec[i]["spec"]["sitemanager"]["parameters"].pop("ingressEndpoint", "")
                modified_spec[i]["spec"]["sitemanager"]["healthzEndpoint"] = modified_spec[i]["spec"]["sitemanager"]["parameters"].pop("healthzEndpoint", "")
                del modified_spec[i]["spec"]["sitemanager"]["parameters"]


    logging.debug("CR convertation is started.")
    logging.debug(f"Initial spec: {spec}")
    logging.debug(f"Modified spec: {modified_spec}")

    return jsonify(
        {
            "apiVersion": "apiextensions.k8s.io/v1",
            "kind": "ConversionReview",
            "response": {
                "uid": request.json["request"]["uid"],
                "result": {
                    "status": "Success"
                },
                "convertedObjects": modified_spec
            }
        }
    )


@app.route("/health", methods=["GET"])
def health():
    site_manager_health.set(1)
    return ("", http.HTTPStatus.NO_CONTENT)


@app.route('/sitemanager', methods=['GET'])
def sitemanager_get():
    """
    Method for processing GET requests to /sitemanager
    """
    result = check_authorization(request)
    if result:
        return result

    try:
        response = json_response(200, get_sitemanagers_dict())
    except Exception as e:
        logging.error("Can not get sitemanager structures \n %s" % str(e))
        response = json_response(500, {"message": "Can not get sitemanager structures"})

    return response


@app.route('/sitemanager', methods=['POST'])
def sitemanager_post():
    """
    Method for processing POST requests to /sitemanager
    """

    result = check_authorization(request)
    if result:
        return result

    try:
        data = request.get_json()

    except Exception as e:
        logging.error("Some problem occurred: \n %s" % str(e))
        return json_response(400, {"message": "No valid JSON data was received"})

    logging.info(f"Data was received: {data}")

    # Check all parameters in received data
    if data.get("procedure", "") not in command_list:

        return json_response(400, {"message": f"You should define procedure from list: {command_list}"})

    no_wait = True if data.get("no-wait", "") in (1, "1", True, "true", "True") else False

    try:
        sm_dict = get_sitemanagers_dict()
    except Exception as e:
        logging.error("Can not get sitemanager structures: \n %s" % str(e))
        return json_response(500, {"message": "Can not get sitemanager structures"})

    all_services = get_all_services(sm_dict)

    if data["procedure"] == "list":
        return json_response(200, {"all-services": all_services})

    # Check services for running
    if isinstance(data.get("run-service", None), str):
        run_service = data["run-service"]
    else:
        return json_response(400, {"message": f"run-service value should be defined and have String type"})

    # Check for service what does not exist in cluster
    if run_service not in all_services:

        return json_response(400, {"message": f"Service doesn't exist",
                                   "wrong-service": run_service})

    logging.info(f"Following service will be processed: {run_service}")

    if data["procedure"] == "status":
        try:
            output = dict()
            output["services"] = get_status_with_deps(run_service, sm_dict, **data)
            return json_response(200, output)
        except utils.ProcedureException as e:
            return json_response(400, e.output)

    mode = data["procedure"]
    url = sm_dict["services"][run_service]['parameters']['serviceEndpoint']
    logging.info(f"Service: {run_service}. Set mode {mode}. serviceEndpoint = {url}. No-wait {no_wait}")
    resp = utils.send_post(url=url, mode=mode, no_wait=no_wait)
    if resp.get("bad_response") or resp.get("fatal"):
        return json_response(500, {"message": f"Procedure {data['procedure']} failed",
                                   "service": run_service,
                                   "procedure": data['procedure']})
    else:
        return json_response(200, {"message": f"Procedure {data['procedure']} is started",
                                   "service": run_service,
                                   "procedure": data['procedure']})


def check_authorization(request):
    if utils.FRONT_HTTP_AUTH:
        if "Authorization" not in request.headers:
            return json_response(401, {"message": "You should use Bearer for authorization"})

        if len(request.headers["Authorization"].split(" ")) != 2 or \
           request.headers["Authorization"].split(" ")[1] != utils.SM_AUTH_TOKEN:
            return json_response(403, {"message": "Bearer is empty or wrong"})
