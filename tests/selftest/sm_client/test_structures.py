import pytest

from sm_client.data.structures import *
from sm_client.initialization import init_and_check_config
from tests.selftest.sm_client.common.test_utils import *


def test_get_available_sites(caplog):
    init_and_check_config(args_init())
    sm_dict = SMClusterState()
    sm_dict["k8s-1"] = {"status": True}
    sm_dict["k8s-2"] = {"status": True}
    list1 = sm_dict.get_available_sites()
    assert list1 == ["k8s-1", "k8s-2"]

    sm_dict["k8s-2"] = {"status": False}
    list2 = sm_dict.get_available_sites()
    assert list2 == ["k8s-1"]

    sm_dict["k8s-1"] = {"status": False}
    list3 = sm_dict.get_available_sites()
    assert list3 == []


def test_make_ignored_services():
    sm_dict = SMClusterState(
        {"site1": {"services": {
            "serv1": {"sequence": ["standby", "active"]},
            "serv2": {"sequence": ["standby", "active"]},
            "serv3": {"sequence": ["standby", "active"]},
            "serv4": {"sequence": ["standby", "active"]}}},
            "site2": {"services": {
                "serv1": {"sequence": ["standby", "active"]},
                "serv2": {"sequence": ["standby", "active"]},
                "serv3": {"sequence": ["standby", "active"]}}}})

    ignore1 = sm_dict.make_ignored_services(['serv1', 'serv3'])
    assert set(ignore1) == {'serv2', 'serv4'}

    ignore2 = sm_dict.make_ignored_services(['serv1', 'serv2', 'serv3'])
    assert ignore2 == ['serv4']

    ignore3 = sm_dict.make_ignored_services([])
    assert set(ignore3) == {'serv1', 'serv2', 'serv3', 'serv4'}


def test_ServiceDRStatus_init():
    stat = ServiceDRStatus('test', {'services': {'test': {'healthz': 'up', 'mode': 'disable', 'status': 'done'}}})
    assert stat.status in "done" and stat.healthz in 'up' and stat.mode in 'disable'

    assert not ServiceDRStatus('test', {'services': {'test': {'healthz': 'up'}}}).healthz in "degraded"

    assert ServiceDRStatus('test', {'services': {'test': {'mode': 'disable'}}})['mode'] in "disable"

    assert ServiceDRStatus('test', {'services': {'test': {}}})['mode'] in "--"

    assert ServiceDRStatus('test', {'services': {'test': {}}}).service in "test"

    assert not ServiceDRStatus('test', {'services': {'test': {}}}).is_ok() and \
           ServiceDRStatus('test', {'services': {'test': {'healthz': 'up'}}}).is_ok()

    with pytest.raises(Exception):
        assert ServiceDRStatus('test')
        assert ServiceDRStatus('test', {'services': {}})

    stat = ServiceDRStatus('absent-service', {'message': 'You defined service that does not exist in cluster',
                            'wrong-service': 'absent-service'})
    assert stat.service in 'absent-service' and stat.message and not stat.is_ok()


def test_get_dr_operation_sequence():
    """ DR commands sequence calculation check
      """
    init_and_check_config(args_init())
    sm_dict = SMClusterState()
    sm_dict["k8s-1"] = {"services": {
        "serv1": {"sequence": ["standby", "active"]},
        "serv2": {"sequence": ["standby", "active"]}}}

    sm_dict["k8s-2"] = {"services": {
        "serv1": {"sequence": ["standby", "active"]},
        "serv2": {"sequence": ["standby", "active"]}}}

    # switchover to site2
    assert [['k8s-1', 'standby'], ['k8s-2', 'active']] == sm_dict.get_dr_operation_sequence('serv1', 'move', 'k8s-2')
    # failover to site2
    assert [['k8s-1', 'standby'], ['k8s-2', 'active']] == sm_dict.get_dr_operation_sequence('serv1', 'stop', 'k8s-1')
    # failover to site1
    assert [['k8s-2', 'standby'], ['k8s-1', 'active']] == sm_dict.get_dr_operation_sequence('serv1', 'stop', 'k8s-2')

    sm_dict2 = SMClusterState()
    sm_dict2["k8s-1"] = {"services": {
        "serv1": {"sequence": ["active", "standby"]}}}
    sm_dict2["k8s-2"] = {"services": {
        "serv1": {"sequence": ["active", "active"]}}}

    # switchover to site2
    assert [['k8s-2', 'active'], ['k8s-1', 'standby']] == sm_dict2.get_dr_operation_sequence('serv1', 'move', 'k8s-2')

    # empty sequence, default is used
    sm_dict3 = SMClusterState()
    sm_dict3["k8s-1"] = {"services": {
        "serv1": {"sequence": []}}}
    sm_dict3["k8s-2"] = {"services": {
        "serv1": {"sequence": []}}}

    assert [['k8s-2', 'standby'], ['k8s-1', 'active']] == sm_dict3.get_dr_operation_sequence('serv1', 'move', 'k8s-1')

    # wrong command
    sm_dict4 = SMClusterState()
    sm_dict4["k8s-1"] = {"services": {
        "serv1": {"sequence": []}}}

    with pytest.raises(Exception):
        sm_dict4.get_dr_operation_sequence('serv1', 'wrong_command', 'k8s-1')
