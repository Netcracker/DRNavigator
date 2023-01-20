#!/usr/bin/env python3.9
"""
pytest based unit test
python3  -m pytest -o log_cli=true -s -v tests/selftest/test_site_manager.py <-k  test_name*>
"""
from http import HTTPStatus

import pytest

from server_utils import *


'''
serviceD------------|-------------|
                    V             V
serviceC -----> serviceB ----> serviceA
  |                               ^
  |_______________________________|
'''

test_endpoints = {
    "serviceEndpoint": "http://stub-endpoint8080/sitemanager",
    "healthzEndpoint": "http://stub-endpoint8080/health"
}

test_sm_dict = {
    "services": {
        "serviceA": {
            "before": [],
            "after": [],
            "parameters": test_endpoints
        },
        "serviceB": {
            "before": [],
            "after": ["serviceA"],
            "parameters": test_endpoints
        },
        "serviceC": {
            "before": ["serviceB"],
            "after": ["serviceA"],
            "parameters": test_endpoints
        },
        "serviceD": {
            "before": ["serviceA", "serviceB"],
            "after": [],
            "parameters": test_endpoints
        }
    }
}


def test_status_with_deps(mocker):
    # Mock endpoint from services
    test_resp = {'mode': 'active', 'status': 'done'}
    fake_resp = mocker.Mock()
    fake_resp.json = mocker.Mock(return_value=test_resp)
    fake_resp.status_code = HTTPStatus.OK

    mocker.patch("utils.requests.Session.get", return_value=fake_resp)

    # Check services with deps
    serviceAStatus = get_status_with_deps("serviceA", test_sm_dict, with_deps=True)['serviceA']
    assert serviceAStatus['deps']
    assert len(serviceAStatus['deps']['before']) == 0
    assert len(serviceAStatus['deps']['after']) == 0

    serviceBStatus = get_status_with_deps("serviceB", test_sm_dict, with_deps=True)['serviceB']
    assert serviceBStatus['deps']
    assert len(serviceBStatus['deps']['before']) == 0
    assert len(serviceBStatus['deps']['after']) == 1
    assert "serviceA" in serviceBStatus['deps']['after']
    assert serviceBStatus['deps']['after']["serviceA"] == serviceAStatus

    serviceCStatus = get_status_with_deps("serviceC", test_sm_dict, with_deps=True)['serviceC']
    assert serviceCStatus['deps']
    assert len(serviceCStatus['deps']['before']) == 1
    assert "serviceB" in serviceCStatus['deps']['before']
    assert serviceCStatus['deps']['before']["serviceB"] == serviceBStatus
    assert len(serviceCStatus['deps']['after']) == 1
    assert "serviceA" in serviceCStatus['deps']['after']
    assert serviceCStatus['deps']['after']["serviceA"] == serviceAStatus

    serviceDStatus = get_status_with_deps("serviceD", test_sm_dict, with_deps=True)['serviceD']
    assert serviceDStatus['deps']
    assert len(serviceDStatus['deps']['before']) == 2
    assert "serviceA" in serviceDStatus['deps']['before']
    assert serviceDStatus['deps']['before']["serviceA"] == serviceAStatus
    assert "serviceB" in serviceDStatus['deps']['before']
    assert serviceDStatus['deps']['before']["serviceB"] == serviceBStatus
    assert len(serviceDStatus['deps']['after']) == 0


def test_status_without_deps(mocker):
    # Mock endpoint from services
    test_resp = {'mode': 'active', 'status': 'done'}
    fake_resp = mocker.Mock()
    fake_resp.json = mocker.Mock(return_value=test_resp)
    fake_resp.status_code = HTTPStatus.OK

    mocker.patch("utils.requests.Session.get", return_value=fake_resp)

    # Check services without deps
    serviceAStatus = get_status_with_deps("serviceA", test_sm_dict)['serviceA']
    assert 'deps' not in serviceAStatus

    serviceBStatus = get_status_with_deps("serviceB", test_sm_dict)['serviceB']
    assert 'deps' not in serviceBStatus

    serviceCStatus = get_status_with_deps("serviceC", test_sm_dict)['serviceC']
    assert 'deps' not in serviceCStatus

    serviceDStatus = get_status_with_deps("serviceD", test_sm_dict)['serviceD']
    assert 'deps' not in serviceDStatus


def test_not_exist_deps(mocker):
    # Mock endpoint from services
    test_resp = {'mode': 'active', 'status': 'done'}
    fake_resp = mocker.Mock()
    fake_resp.json = mocker.Mock(return_value=test_resp)
    fake_resp.status_code = HTTPStatus.OK

    mocker.patch("utils.requests.Session.get", return_value=fake_resp)

    # Add service with wrong dependency
    test_sm_dict['services']['serviceE'] = {
            "before": ["serviceF"],
            "after": [],
            "parameters": test_endpoints
    }
    test_sm_dict['services']['serviceF'] = {
        "before": [],
        "after": ["not-exist"],
        "parameters": test_endpoints
    }

    # Check, that wrong deps don't affect other services
    serviceDStatus = get_status_with_deps("serviceD", test_sm_dict, with_deps=True)['serviceD']
    assert serviceDStatus['deps']

    # Check, that status is returned without deps
    serviceEStatus = get_status_with_deps("serviceE", test_sm_dict)['serviceE']
    assert 'deps' not in serviceEStatus
    serviceFStatus = get_status_with_deps("serviceF", test_sm_dict)['serviceF']
    assert 'deps' not in serviceFStatus

    # Check, that exception throws with deps
    with pytest.raises(utils.ProcedureException) as e_procedure:
        get_status_with_deps("serviceE", test_sm_dict, with_deps=True)

    assert e_procedure.value.output
    assert e_procedure.value.output["message"] == "Dependency defined in CR doesn't exist"
    assert e_procedure.value.output["wrong-service"] == "not-exist"
    assert e_procedure.value.output["problem-cr"] == "serviceF"


def test_cycles_deps(mocker):
    # Mock endpoint from services
    test_resp = {'mode': 'active', 'status': 'done'}
    fake_resp = mocker.Mock()
    fake_resp.json = mocker.Mock(return_value=test_resp)
    fake_resp.status_code = HTTPStatus.OK

    mocker.patch("utils.requests.Session.get", return_value=fake_resp)

    # Add services with cycles
    test_sm_dict['services']['serviceE'] = {
            "before": ["serviceF"],
            "after": [],
            "parameters": test_endpoints
    }
    test_sm_dict['services']['serviceF'] = {
        "before": ["serviceE"],
        "after": [],
        "parameters": test_endpoints
    }

    # Check, that wrong deps don't affect other services
    serviceDStatus = get_status_with_deps("serviceD", test_sm_dict, with_deps=True)['serviceD']
    assert serviceDStatus['deps']

    # Check, that status is returned without deps
    serviceEStatus = get_status_with_deps("serviceE", test_sm_dict)['serviceE']
    assert 'deps' not in serviceEStatus
    serviceFStatus = get_status_with_deps("serviceF", test_sm_dict)['serviceF']
    assert 'deps' not in serviceFStatus

    # Check, that exception throws with deps
    with pytest.raises(utils.ProcedureException) as e_procedure:
        get_status_with_deps("serviceE", test_sm_dict, with_deps=True)
        assert e_procedure.output
        assert e_procedure.output["message"] == "Found cycle in service dependencies"
        assert e_procedure.output["wrong_service"] == "serviceE"
        assert e_procedure.output["cycled-services"] == ["serviceE", "serviceF", "serviceE"]