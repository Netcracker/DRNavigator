"""Functions that are used for validation"""
import logging
import ssl

from sm_client.data import settings
from sm_client.data.structures import SMClusterState, ServiceDRStatus, NotValid
from sm_client.processing import sm_process_service


def check_site_ssl_available(checked_site: str, sm_dict: SMClusterState):
    """ Check if required site is available
    @param checked_site: site to check
    @param sm_dict: populated sm_dict
    @returns: Returns true, if site is available
    """

    if not sm_dict[checked_site]['status']:
        logging.error(f"Site: {checked_site} is not available")
        if sm_dict[checked_site]['return_code'] == ssl.SSLErrorNumber.SSL_ERROR_SSL.__int__():
            logging.error(f"SSL certificate verify failed for site:{checked_site}. Please use key -k or --insecure")
        return False
    return True


def check_services_on_sites(services: list, sites: list, sm_dict: SMClusterState):
    """ Check if required services are exists on needed sites
    @param services: services list to process
    @param sites: list of sites to process
    @param sm_dict: populated sm_dict
    @returns: Returns true, if all required services exist on site
    """

    ret = True
    for site in sites:
        for s in services:
            if s not in sm_dict[site]["services"].keys():  # need to rework to support modules, run-services as well
                logging.warning(f"Service '{s}' does not exist on '{site}' site")
                ret = False
    return ret


def check_dep_issue(sm_dict: SMClusterState, cmd: str, module: str):
    """ Check if calculated service order doesn't have undesirable dependency issues
    @param sm_dict: populated sm_dict
    @param cmd:  called cmd command
    @param module: services module
    @returns: Returns true, if cluster doesn't have global problems with dependencies
    """

    if sm_dict.globals[module]['deps_issue']:
        logging.warning(f"Module: {module}, found dependency issue")
        if (cmd == "stop" or cmd in settings.readonly_cmd) and (sm_dict.globals[module]['ts'] is not None and
                              sm_dict.globals[module]['service_dep_ordered'] != []):  # not integrity_error
            logging.warning(f"Ignoring dependency issues for {cmd} command")
        else:
            return False
    return True


def check_state_restrictions(services: list, site: str, cmd: str):
    """ Check if services final state is not restricted in config
    @param services: services list to process
    @param site: site to process
    @param  cmd:  called cmd command
    @returns: Returns true, if final state is permitted for all defined services
    """

    state_is_valid = True
    # Get services, that should be predicted
    services_to_predict = services if "*" in settings.state_restrictions else \
        [service for service in services if service in settings.state_restrictions.keys()]
    logging.debug(f"Services to predict {services_to_predict}")
    # Predict state for services and compare with restrictions
    for service in services_to_predict:
        # Convert restricted states
        restricted_states_list = []
        for restricted_state_str in settings.state_restrictions.get(service, []) + settings.state_restrictions.get("*",
                                                                                                                   []):
            restricted_states_list.append(dict(zip(settings.sm_conf.keys(), restricted_state_str.split("-"))))
        logging.debug(f"Restricted states for service {service}: {restricted_states_list}")

        # Get state on opposite site
        opposite_site = settings.sm_conf.get_opposite_site(site)
        response, ret, _ = sm_process_service(opposite_site, service, "status")
        if not ret:
            logging.error(f"Can't get service {service} on site {opposite_site}")
            state_is_valid = False
            continue
        serviceDRstatus = ServiceDRStatus(service, response)
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
        logging.error("State restrictions validation fail. To skip it use --ignore-restrictions option")
    return state_is_valid


def check_deps_consistency(sm_dict: SMClusterState, services: list, sites: list):
    """ Check services dependencies between sites
    @param sm_dict: populated sm_dict
    @param services: services list to process
    @param sites: sites to check
    @returns: Returns true, if sequence equal on all sites
    """

    is_consistent = True

    for service in services:
        site_to_compare = None
        before_issue = False
        after_issue = False

        for site in sites:
            # Skip service, if it doesn't exist on site
            if service not in sm_dict[site]['services']:
                continue
            # Mark first site as compared with
            if site_to_compare is None:
                site_to_compare = site
                continue
            # Check dependencies
            if set(sm_dict[site_to_compare]['services'][service].get('before', [])) != \
                    set(sm_dict[site]['services'][service].get('before', [])):
                before_issue = True
            if set(sm_dict[site_to_compare]['services'][service].get('after', [])) != \
                    set(sm_dict[site]['services'][service].get('after', [])):
                after_issue = True

        if before_issue:
            message = "\n".join(f"\t{site}: {sm_dict[site]['services'][service].get('before', [])}" if \
                                    service in sm_dict[site]['services'] else f"\t{site}: Service doesn't exist" for
                                site in sites)
            logging.warning(f"Found inconsistent before dependencies for service {service}: \n{message}")
            is_consistent = False
        if after_issue:
            message = "\n".join(f"\t{site}: {sm_dict[site]['services'][service].get('after', [])}" if \
                                    service in sm_dict[site]['services'] else f"\t{site}: Service doesn't exist" for
                                site in sites)
            logging.warning(f"Found inconsistent after dependencies for service {service}: \n{message}")
            is_consistent = False

    return is_consistent


def check_sequence_consistency(sm_dict: SMClusterState, services: list, sites: list):
    """ Check services sequence between sites
    @param sm_dict: populated sm_dict
    @param services: services list to process
    @param sites: sites to check
    @returns: Returns true, if sequence equal on all sites
    """

    is_consistent = True

    for service in services:
        site_to_compare = None

        for site in sites:
            # Skip service, if it doesn't exist on site
            if service not in sm_dict[site]['services']:
                continue
            # Mark first site as compared with
            if site_to_compare is None:
                site_to_compare = site
                continue
            # Check sequence
            if sm_dict[site_to_compare]['services'][service].get('sequence', []) != \
                    sm_dict[site]['services'][service].get('sequence', []):
                message = "\n".join(f"\t{site}: {sm_dict[site]['services'][service].get('sequence', [])}" if \
                                        service in sm_dict[site]['services'] else f"\t{site}: Service doesn't exist"
                                    for site in sites)
                logging.warning(f"Found inconsistent sequence for service {service}: \n{message}")
                is_consistent = False
                break

    return is_consistent


def validate_stop_operation(sm_dict: SMClusterState, cmd, site=None, service_dep_ordered=None,
                            module=settings.default_module):
    """ Validate command compliance to current site state for stop procedure
    @param sm_dict: populated sm_dict
    @param cmd: called cmd command
    @param site: specified site
    @param service_dep_ordered: services list to process
    @param module: services module
    @returns: Allowed or not to proceed operation <cmd> on <site>
    """
    # Only warn if inconsistent
    check_deps_consistency(sm_dict, service_dep_ordered, sm_dict.get_available_sites())
    check_sequence_consistency(sm_dict, service_dep_ordered, sm_dict.get_available_sites())

    if not check_site_ssl_available(settings.sm_conf.get_opposite_site(site), sm_dict) or \
            not check_dep_issue(sm_dict, cmd, module):
        raise NotValid
    check_services_on_sites(service_dep_ordered, list(settings.sm_conf.keys()), sm_dict)


def validate_move_operation(sm_dict: SMClusterState, cmd, site=None, service_dep_ordered=None,
                            module=settings.default_module):
    """ Validate command compliance to current site state for move procedure
    @param sm_dict: populated sm_dict
    @param cmd: called cmd command
    @param site: specified site
    @param service_dep_ordered: services list to process
    @param module: services module
    @returns: Allowed or not to proceed operation <cmd> on <site>
    """
    if not all(check_site_ssl_available(site_i, sm_dict) for site_i in settings.sm_conf.keys()) or \
            not check_services_on_sites(service_dep_ordered,  list(settings.sm_conf.keys()), sm_dict) or \
            not check_dep_issue(sm_dict, cmd, module) or \
            not check_deps_consistency(sm_dict, service_dep_ordered, sm_dict.get_available_sites()) or \
            not check_sequence_consistency(sm_dict, service_dep_ordered, sm_dict.get_available_sites()):
        raise NotValid


def validate_readonly_operation(sm_dict: SMClusterState, cmd, site=None, service_dep_ordered=None,
                                module=settings.default_module):
    """ Validate command compliance to current site state for status and lists procedure
    @param sm_dict: populated sm_dict
    @param cmd: called cmd command
    @param site: specified site
    @param service_dep_ordered: services list to process
    @param module: services module
    @returns: Allowed or not to proceed operation <cmd> on <site>
    """
    # Only warn if inconsistent
    check_deps_consistency(sm_dict, service_dep_ordered, sm_dict.get_available_sites())
    check_sequence_consistency(sm_dict, service_dep_ordered, sm_dict.get_available_sites())

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
    @param cmd: called cmd command
    @param site: specified site
    @param service_dep_ordered: services list to process
    @param module: services module
    @returns: Allowed or not to proceed operation <cmd> on <site>
    """

    if not check_site_ssl_available(site, sm_dict) or \
            not check_services_on_sites(service_dep_ordered, [site], sm_dict) or \
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
    @param cmd: called cmd command
    @param site: specified site
    @param services_to_run: --run-services option value
    @param module: services module
    @returns: Allowed or not to proceed operation <cmd> on <site>
    """

    service_dep_ordered = [s for s in services_to_run
                           if sm_dict[sm_dict.get_available_sites()[0]]['services'].get(s, {})
                           .get("module", settings.default_module) == module] \
        if services_to_run else sm_dict.globals[module]['service_dep_ordered']

    if cmd not in validation_func:
        raise NotValid

    validation_func[cmd](sm_dict, cmd, site, service_dep_ordered, module)
    return service_dep_ordered
