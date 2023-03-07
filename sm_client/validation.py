import logging
import ssl
from typing import Tuple

from sm_client.processing import sm_process_service
from sm_client.data.structures import *


def check_site_ssl_available(checked_site: str, sm_dict: SMClusterState):
    if not sm_dict[checked_site]['status']:
        logging.error(f"Site: {checked_site} is not available")
        if sm_dict[checked_site]['return_code'] == ssl.SSLErrorNumber.SSL_ERROR_SSL.__int__():
            logging.error(f"SSL certificate verify failed for site:{checked_site}. Please use key -k or --insecure")
        return False
    return True


def check_services_on_site(services: list, site: str, sm_dict: SMClusterState):
    ret = True
    for s in services:
        if s not in sm_dict[site]["services"].keys():  # need to rework to support modules, run-services as well
            logging.warning(f"Service '{s}' does not exist on '{site}' site")
            ret = False
    return ret


def check_dep_issue(sm_dict: SMClusterState, cmd: str, module: str):
    if sm_dict.globals[module]['deps_issue']:
        logging.warning(f"Module: {module}, found dependency issue")
        if cmd == "stop" and (sm_dict.globals[module]['ts'] is not None and
                              sm_dict.globals[module]['service_dep_ordered'] != []):  # not integrity_error
            logging.warning(f"Ignoring dependency issues for {cmd} command")
        else:
            return False
    return True


def check_state_restrictions(services: list, site: str, cmd: str):
    state_is_valid = True
    # Get services, that should be predicted
    services_to_predict = services if "*" in settings.state_restrictions else \
        [service for service in services if service in settings.state_restrictions.keys()]
    logging.debug(f"Services to predict {services_to_predict}")
    # Predict state for services and compare with restrictions
    for service in services_to_predict:
        # Convert restricted states
        restricted_states_list = []
        for restricted_state_str in settings.state_restrictions.get(service, []) + settings.state_restrictions.get("*", []):
            restricted_states_list.append(dict(zip(settings.sm_conf.keys(), restricted_state_str.split("-"))))
        logging.debug(f"Restricted states for service {service}: {restricted_states_list}")

        # Get state on opposite site
        opposite_site = settings.sm_conf.get_opposite_site(site)
        response, ret, _ = sm_process_service(opposite_site, service, "status")
        if not ret:
            logging.error(f"Can't get service {service} on site {opposite_site}")
            state_is_valid = False
            continue
        serviceDRstatus = ServiceDRStatus(response)
        if serviceDRstatus.mode == '---':
            logging.error(f"Can't recognize current mode for service {service} on site {opposite_site}")
            state_is_valid = False
            continue

        # Predicted state
        state = {site: settings.sm_conf.convert_sitecmd_to_dr_mode(cmd), opposite_site: serviceDRstatus.mode}
        logging.debug(f"Predicted state for service {service}: {state}")
        if state in restricted_states_list:
            logging.error(f"Final state {state} for service {service} is restricted")
            state_is_valid = False
    if not state_is_valid:
        logging.error(f"State restrictions validation fail. To skip it use --ignore-restrictions option")
    return state_is_valid


def validate_stop_operation(sm_dict: SMClusterState, cmd, site=None, service_dep_ordered=None,
                            module=settings.default_module):
    """ Validate command compliance to current site state for stop procedure
    @param sm_dict: populated sm_dict
    @returns: Allowed or not to proceed operation <cmd> on <site>
    @todo warn/fail in case deps are different
    """
    if not check_site_ssl_available(settings.sm_conf.get_opposite_site(site), sm_dict) or \
            not check_dep_issue(sm_dict, cmd, module):
        raise NotValid
    check_services_on_site(service_dep_ordered, settings.sm_conf.get_opposite_site(site), sm_dict)



def validate_move_operation(sm_dict: SMClusterState, cmd, site=None, service_dep_ordered=None,
                            module=settings.default_module):
    """ Validate command compliance to current site state for move procedure
    @param sm_dict: populated sm_dict
    @returns: Allowed or not to proceed operation <cmd> on <site>
    @todo warn/fail in case deps are different
    """
    if not all(check_site_ssl_available(site_i, sm_dict) for site_i in settings.sm_conf.keys()) or \
            not all(
                check_services_on_site(service_dep_ordered, site_i, sm_dict) for site_i in settings.sm_conf.keys()) or \
            not check_dep_issue(sm_dict, cmd, module):
        raise NotValid


def validate_readonly_operation(sm_dict: SMClusterState, cmd, site=None, service_dep_ordered=None,
                            module=settings.default_module):
    """ Validate command compliance to current site state for status and lists procedure
    @param sm_dict: populated sm_dict
    @returns: Allowed or not to proceed operation <cmd> on <site>
    @todo warn/fail in case deps are different
    """
    if not any([check_site_ssl_available(site, sm_dict)] if site else
              [check_site_ssl_available(s, sm_dict) for s in settings.sm_conf.keys()]):
        raise NotValid

    # Fail if required site (or site merging) has deps issues:
    if not check_dep_issue(sm_dict, cmd, module):
        raise NotValid


def validate_sites_operation(sm_dict: SMClusterState, cmd, site=None, service_dep_ordered=None,
                            module=settings.default_module):
    """ Validate command compliance to current site state for active, standby, disable, return procedures
    @param sm_dict: populated sm_dict
    @returns: Allowed or not to proceed operation <cmd> on <site>
    @todo warn/fail in case deps are different
    """

    if not check_site_ssl_available(site, sm_dict) or \
            not check_services_on_site(service_dep_ordered, site, sm_dict) or \
            not check_dep_issue(sm_dict, cmd, module) or \
            not check_state_restrictions(service_dep_ordered, site, cmd):
        raise NotValid


validation_func = {
    'stop': validate_stop_operation,
    'move': validate_move_operation,
    'status': validate_readonly_operation,
    'list': validate_readonly_operation,
    'active': validate_sites_operation,
    'standby': validate_sites_operation,
    'disable': validate_sites_operation,
    'return': validate_sites_operation,
}


def validate_operation(sm_dict: SMClusterState, cmd, site=None, services_to_run=None, module=settings.default_module) \
        -> list:
    """ Validate command compliance to current site state
    @param sm_dict: populated sm_dict
    @returns: Allowed or not to proceed operation <cmd> on <site>
    @todo warn/fail in case deps are different
    """

    service_dep_ordered = [s for s in services_to_run
                           if sm_dict[sm_dict.get_available_sites()[0]]['services']
                           .get(s, {})
                           .get("module", settings.default_module) == module] \
        if services_to_run else sm_dict.globals[module]['service_dep_ordered']

    if cmd in validation_func:
        validation_func[cmd](sm_dict, cmd, site, service_dep_ordered, module)
        return service_dep_ordered
    else:
        raise NotValid
