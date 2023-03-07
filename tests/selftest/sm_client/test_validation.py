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
        assert validate_operation(sm_dict, "move", "k8s-1")

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
