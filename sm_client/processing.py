"""Functions that are used for procedure processing"""
import copy
import logging
import threading
from queue import Queue
import time
from http import HTTPStatus
from typing import Tuple, Dict

from common import utils
from sm_client.data import settings
from sm_client.data.structures import TopologicalSorter2, ServiceDRStatus, SMClusterState

thread_pool: list = []
thread_result_queue: Queue = Queue(maxsize=-1)


def run_status_procedure(sm_dict: SMClusterState, service_dep_ordered: list):
    """ Runs status procedure for defined services"""

    for serv in service_dep_ordered:
        def run_in_thread(site, serv, sm_dict):  # to run each status service in parallel
            response, _, return_code = sm_process_service(site, serv, "status")
            if not sm_dict[site]['services'].get(serv):
                sm_dict[site]['services'][serv] = {}
            sm_dict[site]['services'][serv]['status'] = ServiceDRStatus(response) if return_code else False

        for site_i in sm_dict.get_available_sites():
            thread = threading.Thread(target=run_in_thread, args=(site_i, serv, sm_dict))
            thread.name = f"Thread: {serv}"
            thread.start()
            thread_pool.append(thread)
    for thread in thread_pool:
        thread.join()


def run_dr_or_site_procedure(sm_dict: SMClusterState, cmd: str, site: str):
    """ Runs dr or site procedure"""
    dr_status = True
    for elem in settings.module_flow:
        module, states = list(elem.items())[0]
        if not dr_status:  # fail rest of services in case failed before
            for i in sm_dict.get_module_services(sm_dict.get_available_sites()[0], module):
                skip_service_due_deps(i)
            continue
        if cmd in ['standby', 'disable', 'return'] and (states and states == ['active']):
            break
        if cmd in 'active' and (states and set(states) == {'standby', 'disable'}):
            continue
        process_module_services(module, states, cmd, site, sm_dict)
        if settings.failed_services:
            logging.debug(f"Module {module} failed. Failed services {settings.failed_services}")
            logging.error(f"Module {module} failed, skipping rest of services, exiting")
            dr_status = False


def skip_service_due_deps(service: str):
    """
    If service was skipped due dependencies or flow problems, it should be marked as not skipped due dependency
    """
    if service in settings.failed_services:
        return
    if service not in settings.skipped_due_deps_services:
        settings.skipped_due_deps_services.append(service)
    if service in settings.done_services:
        settings.done_services.remove(service)
    if service in settings.warned_services:
        settings.warned_services.remove(service)


def process_ts_services(ts: TopologicalSorter2, process_func, *run_args) -> None:
    """ Runs services in ts object one-by-one on both sites using process_func method.
    process_func have to put  ServiceDRStatus result in thread_result_queue  queue
    @param ts: TopologicalSorter2
    @param process_func: method with 1 mandatory param - service name from ts
    @param run_args: list of additional params  passed to process_func
    """
    failed_successors = []
    serv_thread_pool = []
    # global thread_result_queue

    # Make deep copy for ts to process modules by flow on different sites separately
    ts = copy.deepcopy(ts)

    while ts and ts.is_active():  # process all services one by one  in  sorted by dependency
        for serv in ts.get_ready():
            if serv in failed_successors:
                logging.info(f"Service {serv} marked as failed due to dependencies")
                skip_service_due_deps(serv)
                thread_result_queue.put(ServiceDRStatus({'services': {serv: {}}}))
                continue
            thread = threading.Thread(target=process_func,
                                      args=(serv,) + run_args)
            thread.name = f"Thread: {serv}"
            serv_thread_pool.append(thread)
            thread.start()
        service_response = thread_result_queue.get()

        if not service_response.is_ok():  # mark failed and skip successors of serv_done
            for s in ts.successors(service_response.service):
                logging.debug(f"Found successor {s} for failed {service_response.service} ")
                failed_successors.append(s)
        ts.done(service_response.service)
        if service_response.service not in settings.skipped_due_deps_services:
            service_response.sortout_service_results()
    for thread in serv_thread_pool:
        thread.join()


def process_module_services(module, states, cmd, site, sm_dict):
    """ Process services for specific module and states"""
    def get_cmd():
        """get site cmd for module"""
        if states:  #  [standby,disable] or ['active']
            if cmd in ['move', 'stop']:
                return states[0]
            return settings.sm_conf.convert_sitecmd_to_dr_mode(cmd)
        return cmd

    def get_site():
        """ get target site for module"""
        if states and \
                (get_cmd() in 'active' and cmd in 'stop' or
                 get_cmd() not in 'active' and cmd in 'move'):
            return settings.sm_conf.get_opposite_site(site)
        return site

    logging.info(f"Processing {module} module by cmd: {get_cmd()} on site: {get_site()}")

    process_ts_services(sm_dict.globals[module]["ts"],
                        sm_process_service_with_polling,
                        get_site(), get_cmd(), sm_dict, cmd == "stop")


def sm_process_service_with_polling(service, site, cmd, sm_dict, is_failover=False) -> None:
    """ Processes the service with specific site cmd with polling """
    #global thread_result_queue

    service_response = ServiceDRStatus({'services': {service: {}}})

    logging.info(f"Processing {service} in thread start...")
    if cmd in settings.site_cmds:
        if service in sm_dict[site]['services']:
            mode = settings.sm_conf.convert_sitecmd_to_dr_mode(cmd)
            data, ok, status_code = sm_process_service(site, service, mode)
            if is_failover and mode == "standby":
                force = True
                allow_failure = True
                logging.info(f"Force key enabled for procedure 'stop' for service {service} on passivated site")
            else:
                allow_failure = False
                force = settings.force
            if not ok:
                logging.error(f"Failed changing status for service {service}, returned status code {status_code}")
                service_response = ServiceDRStatus(data, sm_dict, site, mode, force, allow_failure)
            else:
                service_response = sm_poll_service_required_status(site, service, mode, sm_dict, force, allow_failure)
                if service_response.service not in settings.skipped_due_deps_services:
                    service_response.sortout_service_results()
        else:
            logging.warning(f"Skip procedure {cmd} for service {service} on site {site}")
            service_response = ServiceDRStatus({'services': {service: {"message": "Service doesn't exist"}}})
    elif cmd in settings.dr_procedures:
        for site_to_process, mode in sm_dict.get_dr_operation_sequence(service, cmd, site):
            if cmd == "stop" and mode == "standby":
                force = True
                allow_failure = True
                logging.info(f"Force key enabled for procedure 'stop' for service {service} on passivated site")
            else:
                allow_failure = False
                force = settings.force

            if service not in sm_dict[site_to_process].get("services", []):
                logging.warning(f"Skip procedure {cmd} for service {service} on site {site_to_process}")
                service_response = ServiceDRStatus({'services': {service: {"message": "Service doesn't exist"}}})
            elif sm_dict[site_to_process]['status']:  # to process only available sites
                data, ok, status_code = sm_process_service(site_to_process, service, mode, 'move' not in cmd)
                if not ok:
                    logging.error(f"Failed changing status for service {service}, returned status code {status_code}")
                    service_response = ServiceDRStatus(data, sm_dict, site, mode, force, allow_failure)
                else:
                    service_response = sm_poll_service_required_status(site_to_process, service, mode, sm_dict, force, allow_failure)
                    if service_response.service not in settings.skipped_due_deps_services:
                        service_response.sortout_service_results()
                if not service_response.is_ok():
                    logging.info(f"Service {service} failed on {site_to_process}, skipping it on another site...")
                    break
    else:
        logging.error(f"Invalid command '{cmd}' for service '{service}'. No processing performed.")

    thread_result_queue.put(service_response)

    logging.info(f"Processing {service} in thread finished")


def sm_poll_service_required_status(site, service, mode, sm_dict, force: bool = settings.force, allow_failure=False) -> ServiceDRStatus:
    """ Polls service status command till desired mode is reached
        @param force: True/False --force mode to ignore healthz
    """

    def service_status_polling(site, service, expected_state: dict, error_states: list, delay=5) -> Dict:
        """  Polls "site-manager status" <service> command till <expected_state> dict or
        error_state dict returns during <timeout> period with <delay>
        @param expected_state:  expected dict state from site-manager service status command. {"status": ["up"]}
        @param error_states:  error states in case return immediately. [{"status": ["failed"]}]
        """
        timeout = sm_dict[site]["services"][service]["timeout"] if sm_dict[site]["services"].get(service, {}) \
                                                                       .get("timeout", None) is not None else \
            settings.SERVICE_DEFAULT_TIMEOUT

        init_time = int(time.time())
        count = 0
        data: dict = {'services': {service: {}}}
        while int(time.time()) <= init_time + int(timeout):
            count += 1

            logging.info(f"Service: {service}. Site: {site}. Polling procedure {expected_state} Iteration {count}")
            logging.info(
                f"Service: {service}. Site: {site}. {int(timeout) - (int(time.time()) - init_time)} seconds left until timeout")

            data, ret, _ = sm_process_service(site, service, "status")
            data = {'services': {service: {}}} if not data else data

            logging.info(f"Service: {service}. Site: {site}. Received data: {data}. Return code: {ret}")

            def check_state():
                for error_state in error_states:
                    if all(data["services"][service][key] in val for key, val in expected_state.items()):
                        logging.info(f"Service: {service}. Site: {site}. Expected state {expected_state} occurred.")
                        return True
                    if all(data["services"][service][key] in val for key, val in error_state.items()):
                        logging.info(f"Service: {service}. Site: {site}. Error state {error_state} occurred.")
                        return True
                return False

            if ret and check_state():
                return data
            time.sleep(delay)

        logging.info(f"Service: {service}. Site: {site}. Timeout expired.")
        # healthz = "--" is needed to understand, that status is not ok
        if 'services' in data and service in data['services']:
            data['services'][service]['healthz'] = "--"
        return data

    if mode == "standby":
        data = service_status_polling(site, service,
                                      {"status": ["done"], "mode": [mode], "healthz":
                                          sm_dict[site]["services"][service]["allowedStandbyStateList"]},
                                      [{'status': ['done'], 'healthz': ['down', 'degraded']},
                                       {'status': ['failed']}])
    else:  # active and rest commands
        data = service_status_polling(site, service,
                                      {"status": ["done"], "mode": [mode], "healthz": ["up"]},
                                      [{'status': ['done'], 'healthz': ['down', 'degraded']},
                                       {'status': ['failed']}])

    if force:
        logging.warning(f"Service: {service}. Force mode enabled. Service healthz ignored")

    return ServiceDRStatus(data, sm_dict, site, mode, force, allow_failure)


def sm_process_service(site, service, site_cmd: str, no_wait=True, force=False) -> Tuple[Dict, bool, int]:
    """ Processes the service with specific site cmd """
    if site_cmd in ["status", "list"]:  # RO operations
        body = {} if service == "site-manager" else {"procedure": "status", "run-service": service}
    else:
        body = {"procedure": settings.sm_conf.convert_sitecmd_to_dr_mode(site_cmd), "run-service": service,
                "no-wait": no_wait, "force": force}

    _, response, return_code = utils.io_make_http_json_request(settings.sm_conf[site]["url"],
                                                                 settings.sm_conf[site]["token"],
                                                                 settings.sm_conf[site]["cacert"],
                                                                 body,
                                                                 use_auth=settings.FRONT_HTTP_AUTH)
    return response, return_code == HTTPStatus.OK, return_code
