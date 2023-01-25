#!/usr/bin/env python3.9
"""
pytest based unit test
python3  -m pytest -o log_cli=true -s -v tests/selftest/test_site_manager.py <-k  test_name*>
"""
from http import HTTPStatus

import pytest

from server_utils import *


'''
serviceD--------------------------|
   V                              V
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
            "before": ["serviceA", "serviceC"],
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
    serviceAStatus = get_status_with_deps("serviceA", test_sm_dict, with_deps=True)
    assert serviceAStatus.keys() == {'serviceA'}
    assert serviceAStatus['serviceA']['deps']
    assert serviceAStatus['serviceA']['deps']['before'] == test_sm_dict['services']['serviceA']['before']
    assert serviceAStatus['serviceA']['deps']['after'] == test_sm_dict['services']['serviceA']['after']

    serviceBStatus = get_status_with_deps("serviceB", test_sm_dict, with_deps=True)
    assert serviceBStatus.keys() == {'serviceA', 'serviceB'}
    assert serviceBStatus['serviceA'] == serviceAStatus['serviceA']
    assert serviceBStatus['serviceB']['deps']
    assert serviceBStatus['serviceB']['deps']['before'] == test_sm_dict['services']['serviceB']['before']
    assert serviceBStatus['serviceB']['deps']['after'] == test_sm_dict['services']['serviceB']['after']

    serviceCStatus = get_status_with_deps("serviceC", test_sm_dict, with_deps=True)
    assert serviceCStatus.keys() == {'serviceA', 'serviceB', 'serviceC'}
    assert serviceCStatus['serviceA'] == serviceAStatus['serviceA']
    assert serviceCStatus['serviceB'] == serviceBStatus['serviceB']
    assert serviceCStatus['serviceC']['deps']
    assert serviceCStatus['serviceC']['deps']['before'] == test_sm_dict['services']['serviceC']['before']
    assert serviceCStatus['serviceC']['deps']['after'] == test_sm_dict['services']['serviceC']['after']

    serviceDStatus = get_status_with_deps("serviceD", test_sm_dict, with_deps=True)
    assert serviceDStatus.keys() == {'serviceA', 'serviceB', 'serviceC', 'serviceD'}
    assert serviceDStatus['serviceA'] == serviceAStatus['serviceA']
    assert serviceDStatus['serviceB'] == serviceBStatus['serviceB']
    assert serviceDStatus['serviceC'] == serviceCStatus['serviceC']
    assert serviceDStatus['serviceD']['deps']
    assert serviceDStatus['serviceD']['deps']['before'] == test_sm_dict['services']['serviceD']['before']
    assert serviceDStatus['serviceD']['deps']['after'] == test_sm_dict['services']['serviceD']['after']


def test_status_without_deps(mocker):
    # Mock endpoint from services
    test_resp = {'mode': 'active', 'status': 'done'}
    fake_resp = mocker.Mock()
    fake_resp.json = mocker.Mock(return_value=test_resp)
    fake_resp.status_code = HTTPStatus.OK

    mocker.patch("utils.requests.Session.get", return_value=fake_resp)

    # Check services without deps
    serviceAStatus = get_status_with_deps("serviceA", test_sm_dict)
    assert serviceAStatus.keys() == {'serviceA'}
    assert 'deps' not in serviceAStatus['serviceA']

    serviceBStatus = get_status_with_deps("serviceB", test_sm_dict)
    assert serviceBStatus.keys() == {'serviceB'}
    assert 'deps' not in serviceBStatus['serviceB']

    serviceCStatus = get_status_with_deps("serviceC", test_sm_dict)
    assert serviceCStatus.keys() == {'serviceC'}
    assert 'deps' not in serviceCStatus['serviceC']

    serviceDStatus = get_status_with_deps("serviceD", test_sm_dict)
    assert serviceDStatus.keys() == {'serviceD'}
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
    serviceDStatus = get_status_with_deps("serviceD", test_sm_dict, with_deps=True)
    assert serviceDStatus.keys() == {'serviceA', 'serviceB', 'serviceC', 'serviceD'}
    assert serviceDStatus['serviceD']['deps']

    # Check, that status is returned without deps
    serviceEStatus = get_status_with_deps("serviceE", test_sm_dict)
    assert serviceEStatus.keys() == {'serviceE'}
    assert 'deps' not in serviceEStatus['serviceE']
    serviceFStatus = get_status_with_deps("serviceF", test_sm_dict)
    assert serviceFStatus.keys() == {'serviceF'}
    assert 'deps' not in serviceFStatus['serviceF']

    # Check, that exception throws with deps
    with pytest.raises(utils.ProcedureException) as e_procedure:
        get_status_with_deps("serviceE", test_sm_dict, with_deps=True)

    assert e_procedure.value.output
    assert e_procedure.value.output["message"] == "Dependency defined in CR doesn't exist"
    assert e_procedure.value.output["wrong-service"] == "not-exist"
    assert e_procedure.value.output["problem-cr"] == "serviceF"

def test_cr_cycles(mocker):
    # Mock endpoint from services
    test_resp = {'mode': 'active', 'status': 'done'}
    fake_resp = mocker.Mock()
    fake_resp.json = mocker.Mock(return_value=test_resp)
    fake_resp.status_code = HTTPStatus.OK

    mocker.patch("utils.requests.Session.get", return_value=fake_resp)

    # Add services with CR cycles
    test_sm_dict['services']['serviceE'] = {
            "before": ["serviceF"],
            "after": [],
            "parameters": test_endpoints
    }
    test_sm_dict['services']['serviceF'] = {
        "before": [],
        "after": ["serviceE"],
        "parameters": test_endpoints
    }

    serviceEStatus = get_status_with_deps("serviceE", test_sm_dict, with_deps=True)
    assert serviceEStatus.keys() == {'serviceE', 'serviceF'}
    assert serviceEStatus['serviceE']['deps']
    assert serviceEStatus['serviceE']['deps']['before'] == test_sm_dict['services']['serviceE']['before']
    assert serviceEStatus['serviceE']['deps']['after'] == test_sm_dict['services']['serviceE']['after']
    assert serviceEStatus['serviceF']['deps']
    assert serviceEStatus['serviceF']['deps']['before'] == test_sm_dict['services']['serviceF']['before']
    assert serviceEStatus['serviceF']['deps']['after'] == test_sm_dict['services']['serviceF']['after']

def test_deps_cycles(mocker):
    # Mock endpoint from services
    test_resp = {'mode': 'active', 'status': 'done'}
    fake_resp = mocker.Mock()
    fake_resp.json = mocker.Mock(return_value=test_resp)
    fake_resp.status_code = HTTPStatus.OK

    mocker.patch("utils.requests.Session.get", return_value=fake_resp)

    # Add services with CR cycles
    test_sm_dict['services']['serviceE'] = {
            "before": ["serviceF"],
            "after": ["serviceF"],
            "parameters": test_endpoints
    }
    test_sm_dict['services']['serviceF'] = {
        "before": [],
        "after": [],
        "parameters": test_endpoints
    }

    serviceEStatus = get_status_with_deps("serviceE", test_sm_dict, with_deps=True)
    assert serviceEStatus.keys() == {'serviceE', 'serviceF'}
    assert serviceEStatus['serviceE']['deps']
    assert serviceEStatus['serviceE']['deps']['before'] == test_sm_dict['services']['serviceE']['before']
    assert serviceEStatus['serviceE']['deps']['after'] == test_sm_dict['services']['serviceE']['after']
    assert serviceEStatus['serviceF']['deps']
    assert serviceEStatus['serviceF']['deps']['before'] == test_sm_dict['services']['serviceF']['before']
    assert serviceEStatus['serviceF']['deps']['after'] == test_sm_dict['services']['serviceF']['after']
