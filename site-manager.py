"""
Company:     NetCracker
Author:      Core PaaS Group
Version:     0.7
Date:        2021-11-19
Description: Service for management of microservices in active-standby scheme of kubernetes cluster
"""
import json
import logging
import threading
import http
import copy

from kubernetes import client, config       # type: ignore
from kubernetes.client import ApiException  # type: ignore

from flask import Flask, request, jsonify, make_response, current_app
from prometheus_flask_exporter.multiprocess import GunicornInternalPrometheusMetrics  # type: ignore
from prometheus_client import Gauge

from common import utils
from site_manager import server_utils
from site_manager.swagger_config import get_apispec, swagger_ui_blueprint

# List of possible procedures
command_list = ["active", "standby", "disable", "list", "status"]

app = Flask(__name__)

app.config['DEBUG'] = server_utils.SM_DEBUG
app.config['JSONIFY_PRETTYPRINT_REGULAR'] = True
metrics = GunicornInternalPrometheusMetrics(app)

site_manager_health = Gauge('site_manager_health', 'SM pod health')

if server_utils.SM_DEBUG:
    logging_level = logging.DEBUG
    logging_format = "[%(asctime)s] [%(process)d] [%(levelname)s] %(filename)s.%(funcName)s(%(lineno)d): %(message)s"
else:
    logging_level = logging.INFO
    logging_format = "[%(asctime)s] [%(process)d] [%(levelname)s] %(filename)s: %(message)s"

logging.basicConfig(format=logging_format, level=logging_level)

lock = threading.Lock()

if server_utils.FRONT_HTTP_AUTH or server_utils.BACK_HTTP_AUTH:
    server_utils.get_token(False)

    w_thread = threading.Thread(target=server_utils.get_token,
                                args=[True, ])
    w_thread.start()


def get_sitemanagers_dict(api_version=server_utils.SM_VERSION):
    """
    Method creates dictionary of donwloaded data from sitemanager CRs
    """

    # In testing mode return stab
    if server_utils.SM_CONFIG.get("testing", {}).get("enabled", False) and \
            server_utils.SM_CONFIG.get("testing", {}).get("sm_dict", {}) != {}:
        return server_utils.SM_CONFIG["testing"]["sm_dict"]

    if server_utils.SM_KUBECONFIG_FILE != "":
        config.load_kube_config(config_file=server_utils.SM_KUBECONFIG_FILE)

    else:
        config.load_incluster_config()

    try:
        response = client.CustomObjectsApi().list_cluster_custom_object(group=server_utils.SM_GROUP,
                                                                        version=api_version,
                                                                        plural=server_utils.SM_PLURAL,
                                                                        _request_timeout=10)

    except ApiException as e:
        if e.status == 404:
            logging.error(f"Can't get sitemanager objects: Desired CRD not found:\n"
                          f"\tplural={server_utils.SM_PLURAL}\n"
                          f"\tgroup={server_utils.SM_GROUP}\n"
                          f"\tversion={api_version}")
            raise utils.ProcedureException(output={"message": "Can't get sitemanager objects: Desired CRD not found",
                                                   "plural": server_utils.SM_PLURAL,
                                                   "group": server_utils.SM_GROUP,
                                                   "version": api_version})
        raise e

    output = {'services': {}}
    for item in response["items"]:
        service_name = item['spec']['sitemanager'].get("alias",
                                           "%s.%s" % (item['metadata'].get('name'), item['metadata'].get('namespace')))
        output['services'][service_name] = \
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
            service_endpoint = server_utils.HTTP_SCHEME + item["spec"]["sitemanager"]['parameters']["serviceEndpoint"]
    else:
        service_endpoint = ''

    if item['spec']['sitemanager']['parameters'].get('healthzEndpoint', '') != '':

        if item['spec']['sitemanager']['parameters']['healthzEndpoint'].startswith("http://") or \
                item['spec']['sitemanager']['parameters']['healthzEndpoint'].startswith("https://"):
            healthz_endpoint = item['spec']['sitemanager']['parameters']['healthzEndpoint']
        else:
            healthz_endpoint = server_utils.HTTP_SCHEME + item['spec']['sitemanager']['parameters']['healthzEndpoint']
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
    ---
    get:
        summary: Root request to check SM availability
        responses:
            '200':
                description: Always return 'Under construction'
        tags:
        - site-manager
    """

    return "Under construction"


@app.route("/health", methods=["GET"])
def health():
    """
    Health endpoint
    ---
    get:
        summary: Health check
        responses:
            '204':
                description: site-manager health up
        tags:
        - site-manager
    """
    site_manager_health.set(1)
    return "", http.HTTPStatus.NO_CONTENT


@app.route('/sitemanager', methods=['GET'])
def sitemanager_get():
    """
    Method for processing GET requests to /sitemanager
    ---
    get:
        summary: Get the dict of CRs for all services managed by site-manager
        security:
            - token: ['secret']
        responses:
            '200':
                description: CRs dictionary
            '401':
                description: Unauthorized user
            '403':
                description: Invalid token
            '500':
                description: Server error
        tags:
        - site-manager
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
    ---
    post:
        summary: Process service
        security:
            - token: ['secret']
        requestBody:
            content:
                application/json:
                    schema: ProcessingBodySchema
        responses:
            '200':
                description: Procedure runs
            '400':
                description: Wrong data
            '401':
                description: Unauthorized user
            '403':
                description: Invalid token
            '500':
                description: Server error
        tags:
        - site-manager
    """

    result = check_authorization(request)
    if result:
        return result

    try:
        data = request.get_json()

    except Exception as e:
        logging.error(f"Some problem occurred: \n {str(e)}")
        return json_response(400, {"message": "No valid JSON data was received"})

    logging.info(f"Data was received: {data}")

    # Check all parameters in received data
    if data.get("procedure", "") not in command_list:

        return json_response(400, {"message": f"You should define procedure from list: {command_list}"})

    no_wait = data.get("no-wait", "") in (1, "1", True, "true", "True")

    try:
        sm_dict = get_sitemanagers_dict()
    except Exception as e:
        logging.error(f"Can not get sitemanager structures: \n {str(e)}")
        return json_response(500, {"message": "Can not get sitemanager structures"})

    all_services = server_utils.get_all_services(sm_dict)

    if data["procedure"] == "list":
        return json_response(200, {"all-services": all_services})

    # Check services for running
    if isinstance(data.get("run-service", None), str):
        run_service = data["run-service"]
    else:
        return json_response(400, {"message": "run-service value should be defined and have String type"})

    # Check for service what does not exist in cluster
    if run_service not in all_services:

        return json_response(400, {"message": "Service doesn't exist",
                                   "wrong-service": run_service})

    logging.info(f"Following service will be processed: {run_service}")

    if data["procedure"] == "status":
        try:
            output = {
                "services": server_utils.get_status_with_deps(run_service, sm_dict, **data)
            }
            return json_response(200, output)
        except utils.ProcedureException as e:
            return json_response(400, e.output)

    mode = data["procedure"]
    url = sm_dict["services"][run_service]['parameters']['serviceEndpoint']
    logging.info(f"Service: {run_service}. Set mode {mode}. serviceEndpoint = {url}. No-wait {no_wait}")
    resp = server_utils.send_post(url=url, mode=mode, no_wait=no_wait)
    if resp.get("bad_response") or resp.get("fatal"):
        return json_response(500, {"message": f"Procedure {data['procedure']} failed",
                                   "service": run_service,
                                   "procedure": data['procedure']})
    return json_response(200, {"message": f"Procedure {data['procedure']} is started",
                                   "service": run_service,
                                   "procedure": data['procedure']})


@app.route('/swagger')
def create_swagger_spec():
   return json.dumps(get_apispec(app).to_dict())


app.register_blueprint(swagger_ui_blueprint)

def check_authorization(req):
    """Checks authorization token"""
    if server_utils.FRONT_HTTP_AUTH:
        if "Authorization" not in req.headers:
            return json_response(401, {"message": "You should use Bearer for authorization"})

        if len(req.headers["Authorization"].split(" ")) != 2 or \
           req.headers["Authorization"].split(" ")[1] != server_utils.SM_AUTH_TOKEN:
            return json_response(403, {"message": "Bearer is empty or wrong"})
    return None
