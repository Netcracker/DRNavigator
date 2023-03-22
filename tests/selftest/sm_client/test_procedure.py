import logging

import smclient
from sm_client.data import settings
from sm_client.data.structures import SMClusterState, TopologicalSorter2
from sm_client.initialization import init_and_check_config
from sm_client.processing import run_dr_or_site_procedure
from tests.selftest.sm_client.common.test_utils import args_init

# Global parameter for mock_sm_process_service.
# {service -> list of failed sites}
service_failed_site = {}


def mock_sm_process_service(site, service, site_cmd, no_wait=True, force=False):
    """
    Function for mock to sm_client.processing.sm_process_service
    """
    mode = 'active' if site == 'k8s-1' else 'standby'
    status = 'done' if site not in service_failed_site.get(service, []) else 'failed'
    return {"services": {service: {'healthz': 'up', 'mode': mode, 'status': status}}}, True, 200


def create_sm_dict(stateful_services_deps: dict, notstateful_services_deps: dict) -> SMClusterState:
    """
    Function, that creates sm_dict with specified stateful and notstateful services
    """
    settings.module_flow = [{"notstateful": ["standby"]}, {"stateful": None}, {"notstateful": ["active"]}]
    sm_dict = SMClusterState()
    sm_dict["k8s-1"] = {"services": {}, "status": True}
    sm_dict["k8s-2"] = {"services": {}, "status": True}

    stateful_ts = TopologicalSorter2()
    for serv, deps in stateful_services_deps.items():
        sm_dict["k8s-1"]["services"][serv] = {"timeout": 1, "sequence": ['standby', 'active'],
                                              "allowedStandbyStateList": "up",
                                              "module": "stateful"}
        sm_dict["k8s-2"]["services"][serv] = {"timeout": 1, "sequence": ['standby', 'active'],
                                              "allowedStandbyStateList": "up",
                                              "module": "stateful"}
        stateful_ts.add(serv, *deps)
    stateful_ts.prepare()

    notstateful_ts = TopologicalSorter2()
    for serv, deps in notstateful_services_deps.items():
        sm_dict["k8s-1"]["services"][serv] = {"timeout": 1, "sequence": ['standby', 'active'],
                                              "allowedStandbyStateList": "up",
                                              "module": "notstateful"}
        sm_dict["k8s-2"]["services"][serv] = {"timeout": 1, "sequence": ['standby', 'active'],
                                              "allowedStandbyStateList": "up",
                                              "module": "notstateful"}
        notstateful_ts.add(serv, *deps)
    notstateful_ts.prepare()

    sm_dict.globals = {"stateful": {"ts": stateful_ts}, "notstateful": {"ts": notstateful_ts}}
    return sm_dict


def rerun_process_module_service(sm_dict, cmd, site):
    # Clear result lists
    settings.done_services.clear()
    settings.failed_services.clear()
    settings.warned_services.clear()
    settings.skipped_due_deps_services.clear()

    # Run
    run_dr_or_site_procedure(sm_dict, cmd, site)


def test_switchover_with_failed_services(mocker, caplog):
    caplog.set_level(logging.INFO)
    smclient.args = args_init()
    init_and_check_config(args_init())
    caplog.set_level(logging.DEBUG)
    mocker.patch("sm_client.processing.sm_process_service", side_effect=mock_sm_process_service)
    global service_failed_site

    sm_dict = create_sm_dict({"serv1": [], "serv2": ["serv1"]}, {"ns-serv1": [], "ns-serv2": ["ns-serv1"]})
    service_failed_site = {"serv1": ["k8s-1", "k8s-2"]}

    # Test when first notstateful service fails on standby site
    service_failed_site = {"ns-serv1": ["k8s-2"]}
    rerun_process_module_service(sm_dict, "move", "k8s-1")
    assert ["ns-serv1"] == settings.failed_services
    assert [] == settings.done_services
    assert [] == settings.warned_services
    assert ["ns-serv2", "serv1", "serv2"] == settings.skipped_due_deps_services

    # Test when second notstateful service fails on standby site
    service_failed_site = {"ns-serv2": ["k8s-2"]}
    rerun_process_module_service(sm_dict, "move", "k8s-1")
    assert ["ns-serv2"] == settings.failed_services
    assert [] == settings.done_services
    assert [] == settings.warned_services
    assert ["serv1", "serv2", "ns-serv1"] == settings.skipped_due_deps_services

    # Test when first stateful service fails on standby site
    service_failed_site = {"serv1": ["k8s-2"]}
    rerun_process_module_service(sm_dict, "move", "k8s-1")
    assert ["serv1"] == settings.failed_services
    assert [] == settings.done_services
    assert [] == settings.warned_services
    assert ["serv2", "ns-serv1", "ns-serv2"] == settings.skipped_due_deps_services

    # Test when second stateful service fails on standby site
    service_failed_site = {"serv2": ["k8s-2"]}
    rerun_process_module_service(sm_dict, "move", "k8s-1")
    assert ["serv2"] == settings.failed_services
    assert ["serv1"] == settings.done_services
    assert [] == settings.warned_services
    assert ["ns-serv1", "ns-serv2"] == settings.skipped_due_deps_services

    # Test when first stateful service fails on active site
    service_failed_site = {"serv1": ["k8s-1"]}
    rerun_process_module_service(sm_dict, "move", "k8s-1")
    assert ["serv1"] == settings.failed_services
    assert [] == settings.done_services
    assert [] == settings.warned_services
    assert ["serv2", "ns-serv1", "ns-serv2"] == settings.skipped_due_deps_services

    # Test when second stateful service fails on active site
    service_failed_site = {"serv2": ["k8s-1"]}
    rerun_process_module_service(sm_dict, "move", "k8s-1")
    assert ["serv2"] == settings.failed_services
    assert ["serv1"] == settings.done_services
    assert [] == settings.warned_services
    assert ["ns-serv1", "ns-serv2"] == settings.skipped_due_deps_services

    # Test when first notstateful services fails on active site
    service_failed_site = {"ns-serv1": ["k8s-1"]}
    rerun_process_module_service(sm_dict, "move", "k8s-1")
    assert ["ns-serv1"] == settings.failed_services
    assert ["serv1", "serv2"] == settings.done_services
    assert [] == settings.warned_services
    assert ["ns-serv2"] == settings.skipped_due_deps_services

    # Test when second notstateful service fails on active site
    service_failed_site = {"ns-serv2": ["k8s-1"]}
    rerun_process_module_service(sm_dict, "move", "k8s-1")
    assert ["ns-serv2"] == settings.failed_services
    assert ["ns-serv1", "serv1", "serv2"] == settings.done_services
    assert [] == settings.warned_services
    assert [] == settings.skipped_due_deps_services


def test_failover_with_failed_services(mocker, caplog):
    caplog.set_level(logging.INFO)
    smclient.args = args_init()
    init_and_check_config(args_init())
    caplog.set_level(logging.DEBUG)
    mocker.patch("sm_client.processing.sm_process_service", side_effect=mock_sm_process_service)
    global service_failed_site

    sm_dict = create_sm_dict({"serv1": [], "serv2": ["serv1"]}, {"ns-serv1": [], "ns-serv2": ["ns-serv1"]})
    service_failed_site = {"serv1": ["k8s-1", "k8s-2"]}

    # Test when second notstateful services fails on standby site
    service_failed_site = {"ns-serv2": ["k8s-2"]}
    rerun_process_module_service(sm_dict, "stop", "k8s-2")
    assert [] == settings.failed_services
    assert ["ns-serv1", "serv1", "serv2"] == settings.done_services
    assert ["ns-serv2"] == settings.warned_services
    assert [] == settings.skipped_due_deps_services

    # Test when all notstateful services fail on standby site
    service_failed_site = {"ns-serv1": ["k8s-2"], "ns-serv2": ["k8s-2"]}
    rerun_process_module_service(sm_dict, "stop", "k8s-2")
    assert [] == settings.failed_services
    assert ["serv1", "serv2"] == settings.done_services
    assert ["ns-serv1", "ns-serv2"] == settings.warned_services
    assert [] == settings.skipped_due_deps_services

    # Test when all stateful services fail on standby site
    service_failed_site = {"serv1": ["k8s-2"], "serv2": ["k8s-2"]}
    rerun_process_module_service(sm_dict, "stop", "k8s-2")
    assert [] == settings.failed_services
    assert ["ns-serv1", "ns-serv2"] == settings.done_services
    assert ["serv1", "serv2"] == settings.warned_services
    assert [] == settings.skipped_due_deps_services

    # Test when all services fail on standby site
    service_failed_site = {"ns-serv1": ["k8s-2"], "ns-serv2": ["k8s-2"], "serv1": ["k8s-2"], "serv2": ["k8s-2"]}
    rerun_process_module_service(sm_dict, "stop", "k8s-2")
    assert [] == settings.failed_services
    assert [] == settings.done_services
    assert ["ns-serv1", "ns-serv2", "serv1", "serv2"] == settings.warned_services
    assert [] == settings.skipped_due_deps_services

    # Test when first stateful service fails on active site and first notstateful service fails on standby site
    service_failed_site = {"serv1": ["k8s-1"], "ns-serv1": ["k8s-2"]}
    rerun_process_module_service(sm_dict, "stop", "k8s-2")
    assert ["serv1"] == settings.failed_services
    assert [] == settings.done_services
    assert [] == settings.warned_services
    assert ["serv2", "ns-serv1", "ns-serv2"] == settings.skipped_due_deps_services

    # Test when second stateful service fails on active site and first stateful service fails on standby site
    service_failed_site = {"serv2": ["k8s-1"], "serv1": ["k8s-2"]}
    rerun_process_module_service(sm_dict, "stop", "k8s-2")
    assert ["serv2"] == settings.failed_services
    assert [] == settings.done_services
    assert ["serv1"] == settings.warned_services
    assert ["ns-serv1", "ns-serv2"] == settings.skipped_due_deps_services

    # Test when first notstateful service fails on standby site and second not stateful service fails on active site
    service_failed_site = {"ns-serv1": ["k8s-2"], "ns-serv2": ["k8s-1"]}
    rerun_process_module_service(sm_dict, "stop", "k8s-2")
    assert ["ns-serv2"] == settings.failed_services
    assert ["serv1", "serv2"] == settings.done_services
    assert ["ns-serv1"] == settings.warned_services
    assert [] == settings.skipped_due_deps_services

    # Test when first notstateful service fails on both sites
    service_failed_site = {"ns-serv1": ["k8s-1", "k8s-2"]}
    rerun_process_module_service(sm_dict, "stop", "k8s-2")
    assert ["ns-serv1"] == settings.failed_services
    assert ["serv1", "serv2"] == settings.done_services
    assert [] == settings.warned_services
    assert ["ns-serv2"] == settings.skipped_due_deps_services
