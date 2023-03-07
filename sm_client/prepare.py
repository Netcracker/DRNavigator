import copy
import logging
from graphlib import CycleError
from typing import Tuple, Optional

from sm_client.data.structures import *


def make_ordered_services_to_process(sm_dict: SMClusterState, site, services_to_process: list = None,
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
            if ll[i]['after'] and  ll[i]['after'][0] not in [ii for ii in ll.keys()]: #if wrong dep - collect them
                if not wrong_dep_list.get(i):
                    wrong_dep_list[i]={}
                wrong_dep_list[i]['after'] = ll[i]['after']

            if ll[i]['before'] and ll[i]['before'][0] not in [ii for ii in ll.keys()]: #if wrong dep - collect them
                if not wrong_dep_list.get(i):
                    wrong_dep_list[i]={}
                wrong_dep_list[i]['before'] = ll[i]['before']

        return wrong_dep_list

    ret = True
    integrity_error = False
    temp_dict = copy.deepcopy(sm_dict)

    # leave only services listed in service_to_process, not skipped/ignored and belong to module
    for serv in temp_dict[site]['services'].copy().keys():
        if temp_dict[site]['services'][serv].get('module',"") not in module:
            temp_dict[site]['services'].pop(serv, None)
        if services_to_process and serv not in services_to_process:
            temp_dict[site]['services'].pop(serv,None)
        if settings.ignored_services and serv in settings.ignored_services: # remove skipped/ignored
            temp_dict[site]['services'].pop(serv, None)

    # collect sorted ordered service list
    service_lists = []
    try:
        service_lists = [i for i in build_after_before_graph(temp_dict[site]['services']).static_order()]
        for service, depends in after_before_check(temp_dict[site]['services']).items():  # check deps
            for depend in depends:
                logging.warning(f"Site: {site}. Service: {service} has nonexistent "
                                f"{depend} dependency: {depends[depend]}")
                ret = False
    except CycleError as e:
        logging.error(f"Site {site} has integrity issues: %s",e)
        integrity_error = True

    if integrity_error:
        return [], False, type(None) # return error, integrity issue

    # check services equality on all sites
    ts = build_after_before_graph(temp_dict[site]['services'])
    ts.prepare()
    return service_lists, ret, ts
