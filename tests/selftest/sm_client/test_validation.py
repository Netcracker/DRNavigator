import logging
from http import HTTPStatus

import pytest

from sm_client.data.structures import *
from sm_client.initialization import init_and_check_config
from sm_client.validation import validate_operation
from tests.selftest.sm_client.common.test_utils import *


def test_validate_operation(caplog):
    init_and_check_config(args_init())

    # Init state
    sm_dict = SMClusterState()
    sm_dict["k8s-1"] = {"status": True, "return_code": None}
    sm_dict["k8s-2"] = {"status": False, "return_code": None}
    sm_dict.globals = {"stateful": {"service_dep_ordered": [], "deps_issue": False, "ts": TopologicalSorter2}}

    # Check active
    assert validate_operation(sm_dict, "active", "k8s-1") == []
    with pytest.raises(NotValid):
        assert validate_operation(sm_dict, "active", "k8s-2")

    # Check standby
    assert validate_operation(sm_dict, "standby", "k8s-1") == []
    with pytest.raises(NotValid):
        assert validate_operation(sm_dict, "standby", "k8s-2")

    # Check disable
    assert validate_operation(sm_dict, "disable", "k8s-1") == []
    with pytest.raises(NotValid):
        assert validate_operation(sm_dict, "disable", "k8s-2")

    # Check return
    assert validate_operation(sm_dict, "return", "k8s-1") == []
    with pytest.raises(NotValid):
        assert validate_operation(sm_dict, "return", "k8s-2")

    # Check move
    with pytest.raises(NotValid):
        assert validate_operation(sm_dict, "move", "k8s-1")
    with pytest.raises(NotValid):
        assert validate_operation(sm_dict, "move", "k8s-2")

    # Check stop
    with pytest.raises(NotValid):
        assert validate_operation(sm_dict, "stop", "k8s-1")
    assert validate_operation(sm_dict, "stop", "k8s-2") == []

    sm_dict_run_services = SMClusterState()

    sm_dict_run_services["k8s-1"] = {"status": True, "return_code": None, "services": {"serv1": {}}}
    sm_dict_run_services["k8s-2"] = {"status": False, "return_code": None, "services": {"serv1": {}}}
    sm_dict_run_services.globals = {"stateful": {"service_dep_ordered": [], "deps_issue": False,
                                                 "ts": TopologicalSorter2}}

    with caplog.at_level(logging.WARNING):
        caplog.clear()
        validate_operation(sm_dict_run_services, "stop", "k8s-2", ["fake1", "serv1"])
        assert "Service 'fake1' does not exist on 'k8s-1' site" in caplog.text

    sm_dict_run_stop = SMClusterState()
    sm_dict_run_stop["k8s-2"] = {"status": True, "return_code": None, "services": {"serv1": {}}}
    sm_dict_run_stop["k8s-1"] = {"status": True, "return_code": None, "services": {"serv1": {}}}
    sm_dict_run_stop.globals = {"stateful": {"service_dep_ordered": ["serv1"], "deps_issue": True,
                                             "ts": TopologicalSorter2}}
    with caplog.at_level(logging.WARNING):
        caplog.clear()
        validate_operation(sm_dict_run_stop, "stop", "k8s-2", ["serv1"])
        assert "Ignoring dependency issues for stop command" in caplog.text

    # Check services_to_run for different modules
    sm_dict_services_to_run = SMClusterState()
    sm_dict_services_to_run["k8s-2"] = {"status": True, "return_code": None, "services":
        {"serv1": {"module": "stateful"}, "serv2": {"module": "custom_module"}, "serv3": {"module": "custom_module_2"}}}
    sm_dict_services_to_run["k8s-1"] = {"status": True, "return_code": None, "services":
        {"serv1": {"module": "stateful"}, "serv2": {"module": "custom_module"}, "serv3": {"module": "custom_module_2"}}}
    sm_dict_services_to_run.globals = {"stateful": {"service_dep_ordered": ["serv1"], "deps_issue": False,
                                             "ts": TopologicalSorter2},
                                "custom_module": {"service_dep_ordered": ["serv2"], "deps_issue": False,
                                                  "ts": TopologicalSorter2},
                                "custom_module_2": {"service_dep_ordered": ["serv3"], "deps_issue": False,
                                                  "ts": TopologicalSorter2}}

    assert validate_operation(sm_dict_services_to_run, "move", "k8s-1", ["serv1", "serv3"], "stateful") == ["serv1"]
    assert validate_operation(sm_dict_services_to_run, "move", "k8s-1", ["serv1", "serv3"], "custom_module") == []


def test_validate_services_exist_validation(mocker, caplog):
    init_and_check_config(args_init())

    sm_dict = SMClusterState()
    sm_dict["k8s-1"] = {"status": True, "return_code": None, "services": {"serv1": {}, "serv2": {}}}
    sm_dict["k8s-2"] = {"status": True, "return_code": None, "services": {"serv1": {}, "serv3": {}}}
    sm_dict.globals = {"stateful": {"service_dep_ordered": [], "deps_issue": False, "ts": TopologicalSorter2}}

    # Check active
    assert validate_operation(sm_dict, "active", "k8s-1", ["serv1"])
    with caplog.at_level(logging.WARNING):
        caplog.clear()
        with pytest.raises(NotValid):
            assert validate_operation(sm_dict, "active", "k8s-1", ["serv1", "serv2", "serv3"])
        assert "Service 'serv3' does not exist on 'k8s-1' site" in caplog.text

    # Check standby
    assert validate_operation(sm_dict, "standby", "k8s-1", ["serv1"])
    with caplog.at_level(logging.WARNING):
        caplog.clear()
        with pytest.raises(NotValid):
            assert validate_operation(sm_dict, "standby", "k8s-1", ["serv1", "serv2", "serv3"])
        assert "Service 'serv3' does not exist on 'k8s-1' site" in caplog.text

    # Check disable
    assert validate_operation(sm_dict, "disable", "k8s-1", ["serv1"])
    with caplog.at_level(logging.WARNING):
        caplog.clear()
        with pytest.raises(NotValid):
            assert validate_operation(sm_dict, "disable", "k8s-1", ["serv1", "serv2", "serv3"])
        assert "Service 'serv3' does not exist on 'k8s-1' site" in caplog.text

    # Check return
    assert validate_operation(sm_dict, "return", "k8s-1", ["serv1"])
    with caplog.at_level(logging.WARNING):
        caplog.clear()
        with pytest.raises(NotValid):
            assert validate_operation(sm_dict, "return", "k8s-1", ["serv1", "serv2", "serv3"])
        assert "Service 'serv3' does not exist on 'k8s-1' site" in caplog.text

    # Check move
    with caplog.at_level(logging.WARNING):
        caplog.clear()
        with pytest.raises(NotValid):
            assert validate_operation(sm_dict, "move", "k8s-1", ["serv1", "serv2", "serv3"])
        assert "Service 'serv3' does not exist on 'k8s-1' site" in caplog.text
        assert "Service 'serv2' does not exist on 'k8s-2' site" in caplog.text

    # Check stop
    with caplog.at_level(logging.WARNING):
        caplog.clear()
        assert validate_operation(sm_dict, "stop", "k8s-1", ["serv1", "serv2", "serv3"])
        assert "Service 'serv3' does not exist on 'k8s-1' site" in caplog.text
        assert "Service 'serv2' does not exist on 'k8s-2' site" in caplog.text


def test_validate_restrictions(mocker, caplog):
    init_and_check_config(args_init(test_restrictions_config_path))
    sm_dict = SMClusterState()
    sm_dict["k8s-1"] = {"status": True, "return_code": None, "services": {"serv1": {}, "serv2": {}}}
    sm_dict["k8s-2"] = {"status": True, "return_code": None, "services": {"serv1": {}, "serv2": {}}}
    sm_dict.globals = {"stateful": {"service_dep_ordered": ["serv1", "serv2"], "deps_issue": False,
                                    "ts": TopologicalSorter2}}

    # Test standby-standby restriction for all services
    test_resp = {'services': {'serv1': {'healthz': 'up', 'mode': 'standby', 'status': 'done'}}}
    fake_resp = mocker.Mock()
    fake_resp.json = mocker.Mock(return_value=test_resp)
    fake_resp.status_code = HTTPStatus.OK
    mocker.patch("common.utils.requests.Session.post", return_value=fake_resp)
    with pytest.raises(NotValid):
        assert validate_operation(sm_dict, "standby", "k8s-1")

    # Test active-active restriction for specific service
    test_resp = {'services': {'serv2': {'healthz': 'up', 'mode': 'active', 'status': 'done'}}}
    fake_resp = mocker.Mock()
    fake_resp.json = mocker.Mock(return_value=test_resp)
    fake_resp.status_code = HTTPStatus.OK
    mocker.patch("common.utils.requests.Session.post", return_value=fake_resp)
    with pytest.raises(NotValid):
        caplog.clear()
        assert validate_operation(sm_dict, "active", "k8s-1")
        assert "final state {'k8s-1': 'active', 'k8s-2': 'active'} for service serv2 is restricted" in caplog.text

    # Test active-active restriction for not restricted service
    assert validate_operation(sm_dict, "active", "k8s-1", ["serv1"])


def test_deps_consistence_validation(mocker, caplog):
    init_and_check_config(args_init())

    # Check services_to_run for different modules
    sm_dict = SMClusterState()
    sm_dict["k8s-2"] = {"status": True, "return_code": None, "services":
        {"serv1": {}, "serv2": {"before": "serv1"}, "serv3": {"after": "serv1"}}}
    sm_dict["k8s-1"] = {"status": True, "return_code": None, "services":
        {"serv1": {}, "serv2": {}, "serv3": {}}}
    sm_dict.globals = {"stateful": {"service_dep_ordered": ["serv1", "serv2", "serv3"],
                                    "deps_issue": False, "ts": TopologicalSorter2}}

    # Check status
    with caplog.at_level(logging.WARNING):
        caplog.clear()
        assert validate_operation(sm_dict, "status", None)
        assert "Found inconsistent before dependencies for service serv2" in caplog.text
        assert "Found inconsistent after dependencies for service serv3" in caplog.text

    # Check list
    with caplog.at_level(logging.WARNING):
        caplog.clear()
        assert validate_operation(sm_dict, "list", None)
        assert "Found inconsistent before dependencies for service serv2" in caplog.text
        assert "Found inconsistent after dependencies for service serv3" in caplog.text

    # Check stop
    with caplog.at_level(logging.WARNING):
        caplog.clear()
        assert validate_operation(sm_dict, "stop", "k8s-1")
        assert "Found inconsistent before dependencies for service serv2" in caplog.text
        assert "Found inconsistent after dependencies for service serv3" in caplog.text

    # Check move
    with caplog.at_level(logging.WARNING):
        caplog.clear()
        with pytest.raises(NotValid):
            assert validate_operation(sm_dict, "move", "k8s-1")
        assert "Found inconsistent before dependencies for service serv2" in caplog.text
        assert "Found inconsistent after dependencies for service serv3" in caplog.text


def test_sequence_consistence_validation(mocker, caplog):
    init_and_check_config(args_init())

    # Check services_to_run for different modules
    sm_dict_deps = SMClusterState()
    sm_dict_deps["k8s-2"] = {"status": True, "return_code": None, "services":
        {"serv1": {"sequence": ["standby", "active"]}, "serv2": {"sequence": ["standby", "active"]}}}
    sm_dict_deps["k8s-1"] = {"status": True, "return_code": None, "services":
        {"serv1": {"sequence": ["standby", "active"]}, "serv2": {"sequence": ["active", "standby"]}}}
    sm_dict_deps.globals = {"stateful": {"service_dep_ordered": ["serv1", "serv2"],
                                    "deps_issue": False, "ts": TopologicalSorter2}}

    # Check status
    with caplog.at_level(logging.WARNING):
        caplog.clear()
        assert validate_operation(sm_dict_deps, "status", None)
        assert "Found inconsistent sequence for service serv2" in caplog.text

    # Check list
    with caplog.at_level(logging.WARNING):
        caplog.clear()
        assert validate_operation(sm_dict_deps, "list", None)
        assert "Found inconsistent sequence for service serv2" in caplog.text

    # Check stop
    with caplog.at_level(logging.WARNING):
        caplog.clear()
        assert validate_operation(sm_dict_deps, "stop", "k8s-1")
        assert "Found inconsistent sequence for service serv2" in caplog.text

    # Check move
    with caplog.at_level(logging.WARNING):
        caplog.clear()
        with pytest.raises(NotValid):
            assert validate_operation(sm_dict_deps, "move", "k8s-1")
        assert "Found inconsistent sequence for service serv2" in caplog.text