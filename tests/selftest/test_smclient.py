#!/usr/bin/env python3.9
"""
pytest based unit test
python3  -m pytest -o log_cli=true -s -v tests/selftest/test_smclient.py <-k  test_name*>
"""
import json

import pytest

import smclient
from smclient import *
from http import HTTPStatus
import http.server
import ssl
import threading
import os
import pwd
import warnings

test_config_path=os.path.abspath("tests/selftest/resources/config_test.yaml")
test_wrong_config_path=os.path.abspath("tests/selftest/resources/config_test_wrong.yaml")


def pytest_namespace():
    return {'site_name':None}


def args_init(config=None):
    args=argparse.ArgumentParser
    args.verbose=True
    args.insecure=True
    args.config=config if config else test_config_path
    args.run_services=""
    args.skip_services=""
    args.output=None
    args.force = False
    args.command="version"
    return args


def test_sm_process_service(mocker, caplog):
    init_and_check_config(args_init())
    test_resp={'services':{'test1':{'healthz':'up', 'mode':'active', 'status':'done'}}}
    caplog.set_level(logging.DEBUG)
    fake_resp=mocker.Mock()
    fake_resp.json=mocker.Mock(return_value=test_resp)
    fake_resp.status_code=HTTPStatus.OK

    mocker.patch("utils.requests.Session.post", return_value=fake_resp)

    json_body_s, ret, code=sm_process_service("k8s-1", "test1", "active")

    assert json_body_s['services'] == {'test1':{'healthz':'up', 'mode':'active', 'status':'done'}} and \
           ret is True and \
           code is HTTPStatus.OK


def test_make_ignored_services():
    sm_dict=SMClusterState(
        {"site1":{"services":{
            "serv1":{"sequence":["standby", "active"]},
            "serv2":{"sequence":["standby", "active"]},
            "serv3":{"sequence":["standby", "active"]},
            "serv4":{"sequence":["standby", "active"]}}},
            "site2":{"services":{
                "serv1":{"sequence":["standby", "active"]},
                "serv2":{"sequence":["standby", "active"]},
                "serv3":{"sequence":["standby", "active"]}}}})

    ignore1=sm_dict.make_ignored_services(['serv1', 'serv3'])
    assert set(ignore1) == {'serv2', 'serv4'}

    ignore2=sm_dict.make_ignored_services(['serv1', 'serv2', 'serv3'])
    assert ignore2 == ['serv4']

    ignore3=sm_dict.make_ignored_services([])
    assert set(ignore3) == {'serv1', 'serv2', 'serv3', 'serv4'}


def test_make_ordered_services_to_process():
    sm_dict=SMClusterState({
        "site1":{"services":{
            "a":{"after":[], "before":[]},
            "b":{"after":[], "before":["a"]},
            "c":{"after":["b"], "before":[]},
            "d":{"after":["c"], "before":[]},
            "e":{"after":[], "before":["a"]},
            "f":{"after":[], "before":["c"]},
            #            "g":{"after":[],"before":["b"]},
        }},
        "site2":{"services":{
            "a":{"after":[], "before":[]},
            "b":{"after":[], "before":["a"]},
            "c":{"after":["b"], "before":[]},
            "d":{"after":["c"], "before":[]},
            "e":{"after":[], "before":["a"]},
            "f":{"after":[], "before":["c"]},
            #            "g":{"after":[], "before":["b"]},
        }}})
    sorted_list, code, _=make_ordered_services_to_process(sm_dict, "site2")
    assert sorted_list == ['b', 'e', 'f', 'a', 'c', 'd'] and code is True

    sm_dict_one_site=SMClusterState({
        "siteN":{"services":{
            "a":{"after":[], "before":[]},
            "b":{"after":[], "before":["a"]},
            "c":{"after":["b"], "before":[]},
            "d":{"after":["c"], "before":[]},
            "e":{"after":[], "before":["a"]},
        }}})
    sorted_list2, code2, _=make_ordered_services_to_process(sm_dict_one_site, "siteN")
    assert sorted_list2 == ['b', 'e', 'c', 'a', 'd'] and code2

    sm_dict_absent_deps=SMClusterState({
        "site_with_absent_deps":{"services":{
            "a":{"after":[], "before":[]},
            "b":{"after":["z"], "before":["a"]},
            "c":{"after":["a"], "before":["f"]},
        }}})
    sorted_list3, code3, _=make_ordered_services_to_process(sm_dict_absent_deps, "site_with_absent_deps")
    assert sorted_list3 == ['b', 'a', 'c'] and code3 is False

    sm_dict_wrong_deps=SMClusterState({
        "site_with_wrong_deps":{"services":{
            "a":{"after":[], "before":[]},
            "b":{"after":["a"], "before":["a"]},
            "c":{"after":["a"], "before":["b"]},
        }}})
    sorted_list4, code4, _=make_ordered_services_to_process(sm_dict_wrong_deps, "site_with_wrong_deps")
    assert sorted_list4 == [] and code4 is False


def test_io_http_json_request():
    """ SUCCESS basic general success case with SSL verification """
    ret, json_body, http_code=io_make_http_json_request("https://api.github.com", verify=True)
    assert http_code == HTTPStatus.OK and \
           type(json_body) is dict and \
           json.loads('"' + str(json_body) + '"'), "Returned: 200 OK, dict, JSON"

    """ SUCCESS basic general success case without SSL verification """
    os.environ.setdefault('DEBUG', 'True')  # set DEBUG which is used to print warning
    with pytest.warns(Warning, match=r"Unverified HTTPS request is being made") as record:
        ret, json_body, http_code=io_make_http_json_request("https://api.github.com", verify=False)

    assert http_code == HTTPStatus.OK and \
           json.loads('"' + str(json_body) + '"') and \
           len(record) > 1, "Returned: 200 OK, dict, JSON, SSL warning "

    """ FAIL in case not JSON returned with 200 OK """
    ret, json_body, http_code=io_make_http_json_request("https://www.github.com", verify=True)

    assert http_code == False and \
           type(json_body) is dict and \
           bool(dict), "Returned: False and empy dict"

    """ FAIL in case 404 """
    ret, json_body, http_code=io_make_http_json_request("https://api.github.com/page_does_not_exist", verify=True)
    assert http_code == HTTPStatus.NOT_FOUND and \
           type(json_body) is dict and \
           json.loads('"' + str(json_body) + '"') and \
           bool(dict), "Returned: 404 OK and nont empty JSON dict"

    """ FAIL in case SSL verification fails """
    os.system("openssl req -new -x509 -keyout self-signed-fake.pem -out self-signed-fake.pem "
              "-days 365 -nodes -subj \"/C=US/ST=Denial/L=Springfield/O=Dis/CN=www.example.com\"")
    httpd=http.server.HTTPServer(('localhost', 4443), http.server.SimpleHTTPRequestHandler)
    httpd.socket=ssl.wrap_socket(httpd.socket, certfile='self-signed-fake.pem', server_side=True)
    thread=threading.Thread(target=httpd.handle_request)
    thread.start()
    ret, body, code=io_make_http_json_request("https://localhost:4443", verify=True,
                                              retry=0)  # thread handles only one request
    thread.join()
    os.remove("self-signed-fake.pem")
    assert ret is False and \
           code is ssl.SSLErrorNumber.SSL_ERROR_SSL.__int__() and \
           not bool(body), "empty body with specific SSL error"

    ret, json_body, http_code=io_make_http_json_request("https://absent-site", verify=True)
    assert http_code is False and \
           json_body == {} and \
           ret is False


def test_validate_operation(caplog):
    init_and_check_config(args_init())

    sm_dict=SMClusterState("k8s-1")
    sm_dict["k8s-1"]={"status":True, "return_code":None, "service_dep_ordered":[], "deps_issue":False,
                      "ts":TopologicalSorter2}

    assert validate_operation(sm_dict, "active", "k8s-1")

    sm_dict_site_not_available=SMClusterState("k8s-1")
    sm_dict_site_not_available["k8s-1"]={"status":False, "return_code":None, "service_dep_ordered":[],
                                         "deps_issue":False,
                                         "ts":TopologicalSorter2}
    with pytest.raises(NotValid):
        assert validate_operation(sm_dict_site_not_available, "active", "k8s-1")

    init_and_check_config(args_init())
    sm_dict_status=SMClusterState()
    sm_dict_status["k8s-1"]={"status":True, "return_code":None, "service_dep_ordered":[], "deps_issue":False,
                             "ts":TopologicalSorter2}
    sm_dict_status["k8s-2"]={"status":True, "return_code":None, "service_dep_ordered":[], "deps_issue":False,
                             "ts":TopologicalSorter2}
    assert validate_operation(sm_dict_status, "status", None)

    sm_dict_move_fail=SMClusterState()
    sm_dict_move_fail["k8s-1"]={"status":True, "return_code":None, "service_dep_ordered":[], "deps_issue":False,
                                "ts":TopologicalSorter2}

    sm_dict_move_fail["k8s-2"]={"status":False, "return_code":None, "service_dep_ordered":[], "deps_issue":False,
                                "ts":TopologicalSorter2}
    with pytest.raises(NotValid):
        assert validate_operation(sm_dict_move_fail, "move", "k8s-2")

    sm_dict_run_services=SMClusterState()

    sm_dict_run_services["k8s-1"]={"status":True, "return_code":None, "service_dep_ordered":[], "deps_issue":False,
                                   "ts":TopologicalSorter2,
                                   "services":{"serv1":{}}}

    sm_dict_run_services["k8s-2"]={"status":False, "return_code":None, "service_dep_ordered":[], "deps_issue":False,
                                   "ts":TopologicalSorter2,
                                   "services":{"serv1":{}}}
    with caplog.at_level(logging.WARNING):
        caplog.clear()
        validate_operation(sm_dict_run_services, "stop", "k8s-2", ["fake1", "serv1"])
        assert "Service 'fake1' does not exist on 'k8s-1' site" in caplog.text


def test_get_available_sites(caplog):
    init_and_check_config(args_init())
    sm_dict=SMClusterState()
    sm_dict["k8s-1"]={"status":True}
    sm_dict["k8s-2"]={"status":True}
    list1=sm_dict.get_available_sites()
    assert list1 == ["k8s-1", "k8s-2"]

    sm_dict["k8s-2"]={"status":False}
    list2=sm_dict.get_available_sites()
    assert list2 == ["k8s-1"]

    sm_dict["k8s-1"]={"status":False}
    list3=sm_dict.get_available_sites()
    assert list3 == []


def test_SMClusterState_init():
    init_and_check_config(args_init())

    assert SMClusterState()
    assert SMClusterState("k8s-1")
    with pytest.raises(ValueError) as e:
        SMClusterState("not valid site")
    assert str(e.value) in "Unknown site name"
    assert "services" and "dep_issue" in SMClusterState("k8s-2")["k8s-2"]
    assert "services" and "dep_issue" in SMClusterState({"k8s-3":{"services":{"serv1":{}},
                                                                  "status":False},
                                                         "k8s-1":{}})["k8s-3"]
    sm_dict=SMClusterState()
    sm_dict["k8s-1"]={"services":{
        "serv1":{"module":'stateful'},
        "serv2":{"module":'stateful'},
        "serv3":{"module":'notstateful'}}}
    assert sm_dict.get_module_services('k8s-1', 'stateful') == ['serv1', 'serv2'] and \
           sm_dict.get_module_services('k8s-1', 'notstateful') == ['serv3']

    init_and_check_config(args_init(test_wrong_config_path))
    with pytest.raises(ValueError) as e:
        SMClusterState()
    assert str(e.value) in "Only two sites in clusters are supported"


def test_unexist_config_file_init():
    init_and_check_config(args_init())

    with pytest.raises(SystemExit) as pytest_wrapped_e:
        init_and_check_config(args_init("config_test_fake.yaml"))
        assert pytest_wrapped_e.type == SystemExit
        assert pytest_wrapped_e.value.code == 1


def test_ServiceDRStatus_init():
    stat=ServiceDRStatus({'services':{'test':{'healthz':'up', 'mode':'disable', 'status':'done'}}})
    assert stat.status in "done" and stat.healthz in 'up' and stat.mode in 'disable'

    assert not ServiceDRStatus({'services':{'test':{'healthz':'up'}}}).healthz in "degraded"

    assert ServiceDRStatus({'services':{'test':{'mode':'disable'}}})['mode'] in "disable"

    assert ServiceDRStatus({'services':{'test':{}}})['mode'] in "--"

    assert ServiceDRStatus({'services':{'test':{}}}).service in "test"

    assert not ServiceDRStatus({'services':{'test':{}}}).is_ok() and \
           ServiceDRStatus({'services':{'test':{'healthz':'up'}}}).is_ok()

    with pytest.raises(Exception):
        assert ServiceDRStatus()
        assert ServiceDRStatus({'services':{}})

    stat=ServiceDRStatus({'message':'You defined service that does not exist in cluster',
                          'wrong-service':'absent-service'})
    assert stat.service in 'absent-service' and stat.message and not stat.is_ok()


def test_runservise_engine(caplog):
    def process_node(node):
        node=ServiceDRStatus({'services':{node:{}}})
        if node.service in test_failed_services:
            node.healthz='down'
        else:
            node.healthz='up'
        thread_result_queue.put(node)

    caplog.set_level(logging.INFO)
    ts=TopologicalSorter2()
    ts.add("aa")
    ts.add("bb1", "bb")
    ts.add("cc")
    ts.add("cc1", "cc")
    test_failed_services=['bb']
    ts.prepare()
    """ ------------ """

    process_ts_services(ts, process_node)

    logging.info(f"failed_services: {failed_services}")
    logging.info(f"done_services: {done_services}")
    assert done_services == ['aa', 'cc', 'cc1'] and failed_services == ['bb', 'bb1']


def test_get_dr_operation_sequence():
    """ DR commands sequence calculation check
      """
    init_and_check_config(args_init())
    sm_dict=SMClusterState()
    sm_dict["k8s-1"]={"services":{
        "serv1":{"sequence":["standby", "active"]},
        "serv2":{"sequence":["standby", "active"]}}}

    sm_dict["k8s-2"]={"services":{
        "serv1":{"sequence":["standby", "active"]},
        "serv2":{"sequence":["standby", "active"]}}}

    # switchover to site2
    assert [['k8s-1', 'standby'], ['k8s-2', 'active']] == sm_dict.get_dr_operation_sequence('serv1', 'move', 'k8s-2')
    # failover to site2
    assert [['k8s-1', 'standby'], ['k8s-2', 'active']] == sm_dict.get_dr_operation_sequence('serv1', 'stop', 'k8s-1')
    # failover to site1
    assert [['k8s-2', 'standby'], ['k8s-1', 'active']] == sm_dict.get_dr_operation_sequence('serv1', 'stop', 'k8s-2')

    sm_dict2=SMClusterState()
    sm_dict2["k8s-1"]={"services":{
        "serv1":{"sequence":["active", "standby"]}}}
    sm_dict2["k8s-2"]={"services":{
        "serv1":{"sequence":["active", "active"]}}}

    # switchover to site2
    assert [['k8s-2', 'active'], ['k8s-1', 'standby']] == sm_dict2.get_dr_operation_sequence('serv1', 'move', 'k8s-2')

    # empty sequence, default is used
    sm_dict3=SMClusterState()
    sm_dict3["k8s-1"]={"services":{
        "serv1":{"sequence":[]}}}
    sm_dict3["k8s-2"]={"services":{
        "serv1":{"sequence":[]}}}

    assert [['k8s-2', 'standby'], ['k8s-1', 'active']] == sm_dict3.get_dr_operation_sequence('serv1', 'move', 'k8s-1')

    # wrong command
    sm_dict4=SMClusterState()
    sm_dict4["k8s-1"]={"services":{
        "serv1":{"sequence":[]}}}

    with pytest.raises(Exception):
        sm_dict4.get_dr_operation_sequence('serv1', 'wrong_command', 'k8s-1')


def test_init_and_check_config(caplog):
    """ Test file logging capabilities, -o file.log """
    args=args_init()
    for args.output in ["output.log", "not_exist_file", "/tmp/not_exist_file", "~/not_exist_file"]:
        with caplog.at_level(logging.CRITICAL):
            init_and_check_config(args)
            assert f"Cannot write to {args.output}" not in caplog.text
        os.remove(os.path.expanduser(args.output))

    wrong_log_path = ["/", "~/", "./"]
    if pwd.getpwuid(os.getuid())[0] == 'root':
        warnings.warn(UserWarning("You use root user, can't check some test cases"))
    else:
        wrong_log_path.extend(["/not_exist_file", "/etc/passwd"])

    for args.output in wrong_log_path:
        with caplog.at_level(logging.CRITICAL):
            init_and_check_config(args)
            assert f"Cannot write to {args.output}" in caplog.text


def test_sm_poll_service_required_status(mocker, caplog):
    init_and_check_config(args_init())

    test_resp={'services':{'serv1':{'healthz':'up', 'mode':'active', 'status':'done'}}}
    caplog.set_level(logging.DEBUG)
    fake_resp=mocker.Mock()
    fake_resp.json=mocker.Mock(return_value=test_resp)
    fake_resp.status_code=HTTPStatus.OK

    mocker.patch("utils.requests.Session.post", return_value=fake_resp)

    # default timeout
    sm_dict=SMClusterState()
    sm_dict["k8s-1"]={"services":{
        "serv1":{"timeout":None}}}
    with caplog.at_level(logging.INFO):
        caplog.clear()
        dr_status = sm_poll_service_required_status("k8s-1", "serv1", "active", sm_dict)
        assert f"{smclient.SERVICE_DEFAULT_TIMEOUT} seconds left until timeout" in caplog.text
        assert dr_status.is_ok()

    # service specific timeout
    sm_dict=SMClusterState()
    sm_dict["k8s-1"]={"services":{
        "serv1":{"timeout":100}}}
    with caplog.at_level(logging.INFO):
        caplog.clear()
        sm_poll_service_required_status("k8s-1", "serv1", "active", sm_dict)
        assert "100 seconds left until timeout" in caplog.text

def test_sm_process_service_with_polling(mocker, caplog):
    smclient.args=args_init()
    init_and_check_config(args_init())
    test_resp={'services':{'serv1':{'healthz':'up', 'mode':'active', 'status':'running'}}}
    caplog.set_level(logging.DEBUG)
    fake_resp=mocker.Mock()
    fake_resp.json=mocker.Mock(return_value=test_resp)
    fake_resp.status_code=HTTPStatus.OK

    mocker.patch("utils.requests.Session.post", return_value=fake_resp)

    # custom timeout
    sm_dict=SMClusterState()
    sm_dict["k8s-1"]={
        "services":{
            "serv1":{"timeout":1,
                 "sequence":['active','standby']}},
        "status": True}
    with caplog.at_level(logging.INFO):
        caplog.clear()
        sm_process_service_with_polling("serv1", "k8s-1",  "move", sm_dict)
        service_response = thread_result_queue.get()
        service_response.sortout_service_results()
        assert 'serv1' in failed_services
        assert "Service serv1 failed on k8s-1, skipping it on another site" in caplog.text
