import copy
import logging
from graphlib import CycleError
from typing import Tuple, Optional

from sm_client.data.structures import *


def make_ordered_services_to_process(sm_dict: SMClusterState, site: str = None, services_to_process: list = None,
                                     module = settings.default_module ) -> Tuple[list, bool, Optional[TopologicalSorter2]]:
    """ Make ordered and validated services list from all sites in sm_dict
    @returns: ordered list or empty if not possible to assemble(integrity issue), True or False  in case minor issue
        TopologicalSorter object in case success
    @todo[3]: cross site validation
    """
    def build_after_before_graph(dep_list: dict) -> TopologicalSorter2:
        """Assemble directed graph using AFTER and BEFORE dependency
        @param dep_list:[]: services list with deps in special format, see @note
        @returns: TopologicalSorter graph
        @todo: [3] multiple dependency service support
        @note:  format of input list  [{'serv2':{'after':['serv1'],'before':['serv3']}},...]
        """
        ts = TopologicalSorter2()
        for item in dep_list.keys():
            ts.add(item)
            if dep_list[item].get('after') and dep_list[item]['after'][0] in [ii for ii in dep_list.keys()]:
                ts.add(item, dep_list[item]['after'][0]) #if dep[AFTER,BEFORE] is present and exist in the dep_list - add it
            if dep_list[item].get('before') and dep_list[item]['before'][0] in [ii for ii in dep_list.keys()]:
                ts.add(dep_list[item]['before'][0], item)
        return ts

    def after_before_check(ll: dict) -> {}:
        """ Check AFTER and BEFORE dependency integrity
        @returns: dict of services if any with non exist deps , see format in @note
        @note: the return dict format {service:{'before':wrong_dep}}}
        """
        wrong_dep_list={}
        for i in ll.keys():
            wrong_deps = [dep for dep in ll[i]['after'] if dep not in ll]
            if wrong_deps:
                if not wrong_dep_list.get(i):
                    wrong_dep_list[i] = {}
                wrong_dep_list[i]['after'] = wrong_deps

            wrong_deps = [dep for dep in ll[i]['before'] if dep not in ll]
            if wrong_deps:
                if not wrong_dep_list.get(i):
                    wrong_dep_list[i] = {}
                    wrong_dep_list[i]['before'] = wrong_deps

        return wrong_dep_list

    ret = True
    integrity_error = False
    used_sites = [site] if site is not None else sm_dict.get_available_sites()
    services_with_deps = {}

    for site_name in used_sites:
        for serv, serv_conf in sm_dict[site_name]['services'].items():
            # leave only services listed in service_to_process, not skipped/ignored and belong to module
            if serv_conf.get('module', "") not in module or \
                    services_to_process and serv not in services_to_process or \
                    settings.ignored_services and serv in settings.ignored_services:
                continue
            if serv not in services_with_deps:
                services_with_deps[serv] = {'before': [], 'after': []}
            # Add new service dependencies
            services_with_deps[serv]['before'] += \
                [dep for dep in serv_conf['before'] if dep not in services_with_deps[serv]['before']]
            services_with_deps[serv]['after'] += \
                [dep for dep in serv_conf['after'] if dep not in services_with_deps[serv]['after']]

    # collect sorted ordered service list
    service_lists = []
    try:
        service_lists = [i for i in build_after_before_graph(services_with_deps).static_order()]
        for service, depends in after_before_check(services_with_deps).items():  # check deps
            for depend in depends:
                logging.warning(f"Site {site if site else 'merging'}. Service: {service} has nonexistent "
                                f"{depend} dependencies: {depends[depend]}")
                ret = False
    except CycleError as e:
        logging.error(f"Site {site if site else 'merging'} has integrity issues: %s",e)
        integrity_error = True

    if integrity_error:
        return [], False, type(None) # return error, integrity issue

    # check services equality on all sites
    ts = build_after_before_graph(services_with_deps)
    ts.prepare()
    return service_lists, ret, ts
