import copy
import logging
import threading
from queue import Queue
import time
from http import HTTPStatus
from typing import Tuple, Dict

from common import utils
from sm_client.data.structures import *

thread_pool = []
thread_result_queue = Queue(maxsize=-1)


def process_ts_services(ts: TopologicalSorter2, process_func, *run_args: ()) -> None:
    """ Runs services in ts object one-by-one on both sites using process_func method.
    process_func have to put  ServiceDRStatus result in thread_result_queue  queue
    @param ts: TopologicalSorter2
    @param process_func: method with 1 mandatory param - service name from ts
    @param run_args: list of additional params  passed to process_func
    """
    failed_successors = []
    serv_thread_pool = []
    global thread_result_queue

    # Make deep copy for ts to process modules by flow on different sites separately
    ts = copy.deepcopy(ts)

    while ts and ts.is_active():  # process all services one by one  in  sorted by dependency
        for serv in ts.get_ready():
            if serv in failed_successors:
                logging.info(f"Service {serv} marked as failed due to dependencies")
                thread_result_queue.put(ServiceDRStatus({'services': {serv: {}}}))
                break
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
            else:
                return settings.sm_conf.convert_sitecmd_to_dr_mode(cmd)
        return cmd

    def get_site():
        """ get target site for module"""
        if states and \
                (get_cmd() in 'active' and cmd in 'stop' or
                 get_cmd() not in 'active' and cmd in 'move'):
            return settings.sm_conf.get_opposite_site(site)
        else:
            return site

    logging.info(f"Processing {module} module by cmd: {get_cmd()} on site: {get_site()}")

    process_ts_services(sm_dict.globals[module]["ts"],
                        sm_process_service_with_polling,
                        get_site(), get_cmd(), sm_dict)


def sm_process_service_with_polling(service, site, cmd, sm_dict) -> None:
    global thread_result_queue

    service_response = ServiceDRStatus({'services': {service: {}}})

    logging.info(f"Processing {service} in thread start...")
    if cmd in settings.site_cmds:
        mode = settings.sm_conf.convert_sitecmd_to_dr_mode(cmd)
        sm_process_service(site, service, mode)
        service_response = sm_poll_service_required_status(site, service, mode, sm_dict)
        thread_result_queue.put(service_response)
    elif cmd in settings.dr_procedures:
        # todo to handle False(no service on this site)
        for site_to_process, mode in sm_dict.get_dr_operation_sequence(service, cmd, site):
            if cmd == "stop" and mode == "standby":
                force = True
                logging.info(f"Force key enabled for procedure 'stop' for service {service} on passivated site")
            else:
                force = settings.force

            if service not in sm_dict[site_to_process].get("services", []):
                logging.warning(f"Service {service} doesn't exist on site {site_to_process}, skip it")
            elif sm_dict[site_to_process]['status']:  # to process only available sites
                sm_process_service(site_to_process, service, mode, False if 'move' in cmd else True)
                service_response = sm_poll_service_required_status(site_to_process, service, mode, sm_dict, force)
                if cmd in 'move' and not service_response.is_ok():
                    logging.info(f"Service {service} failed on {site_to_process}, skipping it on another site...")
                    break

        thread_result_queue.put(service_response)

    logging.info(f"Processing {service} in thread finished")


def sm_poll_service_required_status(site, service, mode, sm_dict, force: bool = False) -> ServiceDRStatus:
    """ Polls service status command till desired mode is reached
        @param force: True/False --force mode to ignore healthz
    """

    def service_status_polling(site, service, expected_state: {}, error_states: [], delay=5) -> Dict:
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
        data = {'services': {service: {}}}
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
                    elif all(data["services"][service][key] in val for key, val in error_state.items()):
                        logging.info(f"Service: {service}. Site: {site}. Error state {error_state} occurred.")
                        return True
                return False

            if ret and check_state():
                return data
            else:
                time.sleep(delay)
                continue

        logging.info(f"Service: {service}. Site: {site}. Timeout expired.")
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
        stat = ServiceDRStatus(data)
        stat.service_status = True
        return stat

    return ServiceDRStatus(data, sm_dict, site, mode)


def sm_process_service(site, service, site_cmd: str, no_wait=True, force=False) -> Tuple[Dict, bool, int]:
    """ Processes the service with specific site cmd """
    if site_cmd in ["status", "list"]:  # RO operations
        body = {} if service == "site-manager" else {"procedure": "status", "run-service": service}
    else:
        body = {"procedure": settings.sm_conf.convert_sitecmd_to_dr_mode(site_cmd), "run-service": service,
                "no-wait": no_wait, "force": force}

    ret, response, return_code = utils.io_make_http_json_request(settings.sm_conf[site]["url"],
                                                                 settings.sm_conf[site]["token"],
                                                                 settings.sm_conf[site]["cacert"],
                                                                 body,
                                                                 use_auth=settings.FRONT_HTTP_AUTH)
    return response, True if return_code == HTTPStatus.OK else False, return_code
