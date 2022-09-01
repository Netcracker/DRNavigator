#!/usr/bin/env python3.9
"""
pytest based unit test
python3  -m pytest -o log_cli=true -s -v test_smclient.py <-k  test_name*>
TODO
 - use mock instead of http call
"""
import json
import pytest
from smclient import *
from http import HTTPStatus
import http.server
import ssl
import threading
import yaml



def pytest_namespace():
    return {'site_name':None}


def args_init(config=None):
    args=argparse.ArgumentParser
    args.verbose=True
    args.insecure=True
    args.config= config if config else "config_test.yaml"
    args.run_services=""
    args.skip_services=""
    init_and_check_config(args)
    return args


@pytest.mark.skipif(not os.path.exists("config.yaml"),
                    reason="Need to populate config.yaml https://github.com/Netcracker/DRNavigator/blob/main/documentation/Architecture.md#configuration-file")
def test_process_service__status_ok(caplog):
    """ SUCCESS basic general success case without SSL verification
    """
    caplog.set_level(logging.DEBUG)
    args_init("config.yaml")
    with open("config.yaml", 'r') as stream:
        try:
            parsed_yaml=yaml.safe_load(stream)
        except yaml.YAMLError as exc:
            print(exc)
    site=[i["name"] for i in parsed_yaml["sites"]][0]
    print(f"Using {site}")
    pytest.site_name=site
    json_body_s, ret, code=sm_process_service(site, "site-manager", "status")
    assert ret == True \
           and type(json_body_s) is dict \
           and json.loads('"' + str(json_body_s) + '"') \
           and json_body_s["services"], "Returned: dict, valid JSON, JSON contains valid response"


@pytest.mark.depends(on=['test_process_service__status_ok'])
def test_process_service__rw_ok(caplog):
    """ SUCCESS basic general success case without SSL verification
    """
    args_init("config.yaml")
    json_body_s, ret, code=sm_process_service(pytest.site_name if hasattr(pytest, "site_name") else "k8s-1",
                                           "cluster-replicator", "status")
    assert ret == True and \
           type(json_body_s) is dict and \
           json.loads('"' + str(json_body_s) + '"') and \
           code == HTTPStatus.OK, \
        "Returned: non empty dict, valid JSON, JSON contains valid response"



def test_sm_process_service(mocker,caplog):
    args_init()
    test_resp = {'services': {'test1': {'healthz': 'up', 'mode': 'active', 'status': 'done'}}}
    caplog.set_level(logging.DEBUG)
    fake_resp=mocker.Mock()
    fake_resp.json=mocker.Mock(return_value=test_resp)
    fake_resp.status_code=HTTPStatus.OK

    mocker.patch("smclient.requests.Session.post", return_value=fake_resp)

    json_body_s, ret, code = sm_process_service("k8s-1","test1", "active")

    assert json_body_s['services'] == {'test1': {'healthz': 'up', 'mode': 'active', 'status': 'done'}} and \
        ret == True and \
        code == HTTPStatus.OK

def test_make_ignored_services():
    sm_dict={"sites":
        {"site1":{"services":{
            "serv1":{"sequence":["standby", "active"]},
            "serv2":{"sequence":["standby", "active"]},
            "serv3":{"sequence":["standby", "active"]},
            "serv4":{"sequence":["standby", "active"]}}},
        "site2":{"services":{
                "serv1":{"sequence":["standby", "active"]},
                "serv2":{"sequence":["standby", "active"]},
                "serv3":{"sequence":["standby", "active"]}}}}}

    ignore1=data_make_ignored_services(sm_dict, ['serv1', 'serv3'])
    ignore2=data_make_ignored_services(sm_dict, ['serv1', 'serv2', 'serv3'])
    ignore3=data_make_ignored_services(sm_dict, [])
    assert set(ignore1) == {'serv2', 'serv4'} and \
           ignore2 == ['serv4'] and \
           set(ignore3) == {'serv1', 'serv2', 'serv3', 'serv4'}


def test_get_sequence():
    """ DR commands sequence calculation check
    """
    sm_dict={"sites":
        {"site1":{"services":{
            "serv1":{"sequence":["standby", "active"]},
            "serv2":{"sequence":["standby", "active"]}, }},
            "site2":{"services":{
                "serv1":{"sequence":["standby", "active"]},
                "serv2":{"sequence":["standby", "active"]}}}}}

    sm_dict2={"sites":
        {"site1":{"services":{
            "serv1":{"sequence":["active", "standby"]}}},
            "site2":{"services":{
                "serv1":{"sequence":["active", "active"]}}}}}

    seq_1=data_get_dr_operation_sequence(sm_dict, 'serv1', 'move', 'site2')  # switchover to site2
    seq_2=data_get_dr_operation_sequence(sm_dict, 'serv1', 'stop', 'site1')  # failover to site2
    seq_3=data_get_dr_operation_sequence(sm_dict, 'serv1', 'stop', 'site2')  # failover to site1

    seq2_1=data_get_dr_operation_sequence(sm_dict2, 'serv1', 'move', 'site2')  # switchover to site2
    assert seq_1 == [['site1', 'standby'], ['site2', 'active']] and \
           seq_2 == [['site1', 'standby'], ['site2', 'active']] and \
           seq_3 == [['site2', 'standby'], ['site1', 'active']] and \
           seq2_1 == [['site2', 'active'], ['site1', 'standby']]


def test_make_ordered_services_to_process():
    sm_dict={"sites":{
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
        }}}}
    sm_dict_one_site={"sites":{
        "siteN":{"services":{
            "a":{"after":[], "before":[]},
            "b":{"after":[], "before":["a"]},
            "c":{"after":["b"], "before":[]},
            "d":{"after":["c"], "before":[]},
            "e":{"after":[], "before":["a"]},
        }}}}
    sm_dict_absent_deps={"sites":{
        "site_with_absent_deps":{"services":{
            "a":{"after":[], "before":[]},
            "b":{"after":["z"], "before":["a"]},
            "c":{"after":["a"], "before":["f"]},
        }}}}
    sm_dict_wrong_deps={"sites":{
        "site_with_wrong_deps":{"services":{
            "a":{"after":[], "before":[]},
            "b":{"after":["a"], "before":["a"]},
            "c":{"after":["a"], "before":["b"]},
        }}}}
    sm_dict_diff_lists_intersect={"sites":{
        "site1":{"services":{
            "a":{"after":[], "before":[]},
            "b":{"after":[], "before":["a"]},
            "c":{"after":["d"], "before":[]},
            "d":{"after":[], "before":["c"]},
        }, "status":True},
        "site2":{"services":{
            "c":{"after":["d"], "before":[]},
            "d":{"after":[], "before":["c"]},
            "e":{"after":[], "before":["a"]},
            "f":{"after":[], "before":["c"]},
        }, "status":True}}}
    sorted_list, code, _ = make_ordered_services_to_process(sm_dict)
    sorted_list2, code2, _ = make_ordered_services_to_process(sm_dict_one_site)
    sorted_list3, code3, _ = make_ordered_services_to_process(sm_dict_absent_deps)
    sorted_list4, code4, _ = make_ordered_services_to_process(sm_dict_wrong_deps)
    #sorted_list5, code5, _ = make_ordered_services_to_process(sm_dict_diff_lists_intersect)

    assert sorted_list == ['b','e','f','a','c','d'] and code == True and \
           sorted_list2 == ['b','e','c','a','d'] and code2 == True and \
           sorted_list3 == ['b', 'a', 'c'] and code3 == False and \
           sorted_list4 == [] and code4 == False
           #sorted_list5 == ["d", "c"] and code5 == False


def test_io_http_json_request_ok():
    """ SUCCESS basic general success case with SSL verification
    """
    ret, json_body, http_code=io_make_http_json_request("https://api.github.com", verify=True)
    assert http_code == HTTPStatus.OK and \
           type(json_body) is dict and \
           json.loads('"' + str(json_body) + '"'), "Returned: 200 OK, dict, JSON"


def test_io_http_json_request_ok_warning():
    """ SUCCESS basic general success case without SSL verification
    """
    os.environ.setdefault('DEBUG', 'True')  # set DEBUG which is used to print warning
    with pytest.warns(Warning, match=r"Unverified HTTPS request is being made") as record:
        ret, json_body, http_code=io_make_http_json_request("https://api.github.com", verify=False)

    assert http_code == HTTPStatus.OK and \
           json.loads('"' + str(json_body) + '"') and \
           len(record) > 1, "Returned: 200 OK, dict, JSON, SSL warning "


def test_io_http_json_request_200_not_json():
    """ FAIL in case not JSON returned with 200 OK
    """
    ret, json_body, http_code=io_make_http_json_request("https://www.github.com", verify=True)

    assert http_code == False and \
           type(json_body) is dict and \
           bool(dict), "Returned: False and empy dict"


def test_io_http_json_request_404():
    """ FAIL in case 404
    """
    ret, json_body, http_code=io_make_http_json_request("https://api.github.com/page_does_not_exist", verify=True)
    assert http_code == HTTPStatus.NOT_FOUND and \
           type(json_body) is dict and \
           json.loads('"' + str(json_body) + '"') and \
           bool(dict), "Returned: 404 OK and nont empty JSON dict"


def test_io_http_json_request_ssl_fails():
    """ FAIL in case SSL verification fails
    """
    os.system("openssl req -new -x509 -keyout self-signed-fake.pem -out self-signed-fake.pem "
              "-days 365 -nodes -subj \"/C=US/ST=Denial/L=Springfield/O=Dis/CN=www.example.com\" &>/dev/null")
    exception=None
    httpd=http.server.HTTPServer(('localhost', 4443), http.server.SimpleHTTPRequestHandler)
    httpd.socket=ssl.wrap_socket(httpd.socket, certfile='self-signed-fake.pem', server_side=True)
    thread=threading.Thread(target=httpd.handle_request)
    thread.start()
    ret, body, code=io_make_http_json_request("https://localhost:4443", verify=True,
                                              retry=0)  # thread handles only one request
    thread.join()
    os.remove("self-signed-fake.pem")
    assert ret == False and \
           code == ssl.SSLErrorNumber.SSL_ERROR_SSL.__int__() and \
           not bool(body), "empty body with specific SSL error"


def test_poll_service_required_status():
    assert False


def test_validate_operation(caplog):
    args_init()

    sm_dict={"sites":{"k8s-1":{"status":True, "return_code":None}}}
    ret=validate_operation(sm_dict, ([], True), "active", "k8s-1")

    sm_dict_site_not_available_bySSL={"sites":{"k8s-1":{"status":False, "return_code":1}}}
    ret2=validate_operation(sm_dict_site_not_available_bySSL, ([], True), "active", "k8s-1")

    sm_dict_site_not_available={"sites":{"k8s-2":{"status":False, "return_code":None}}}
    ret3=validate_operation(sm_dict_site_not_available, ([], True), "active", "k8s-2")

    sm_dict_status={"sites":{"k8s-1":{"status":True, "return_code":None}, "k8s-2":{"status":False, "return_code":None}}}
    ret4=validate_operation(sm_dict_status, ([], True), "status", None)

    sm_dict_move_fail={
        "sites":{"k8s-1":{"status":True, "return_code":None}, "k8s-2":{"status":False, "return_code":None}}}
    ret5=validate_operation(sm_dict_move_fail, ([], True), "move", "k8s-2")

    sm_dict_run_services={
        "sites":{"k8s-1":{"status":True, "services":{"serv1":{}}}, "k8s-2":{"status":False, "services":{"serv1":{}}}}}
    with caplog.at_level(logging.WARNING):
        caplog.clear()
        validate_operation(sm_dict_run_services, ([], False), "stop", "k8s-2", ["fake1", "serv1"])
        assert "Service fake1 does not exist in k8s-1 kubernetes cluster" in caplog.text

    sm_dict_move_fail={
        "sites":{"k8s-1":{"status":True, "return_code":None}, "k8s-2":{"status":True, "return_code":None}}}
    ret6=validate_operation(sm_dict_move_fail, ([], False), "move", "k8s-2")

    assert ret == True and ret2 == False and ret3 == False and ret4 == True and \
           ret5 == False and ret6 == False


def test_main():
    assert False


def test_get_available_sites():
    sm_dict={"sites":{"k8s-1":{"status":True}, "k8s-2":{"status":True}}}
    list1 = data_get_available_sites(sm_dict)

    sm_dict2={"sites":{"k8s-1":{"status":True}, "k8s-2":{"status":False}}}
    list2 = data_get_available_sites(sm_dict2)

    sm_dict3={"sites":{"k8s-1":{"status":False}, "k8s-2":{"status":False}}}
    list3 = data_get_available_sites(sm_dict3)
    assert list1 == ["k8s-1","k8s-2"] and list2 == ["k8s-1"] and list3 == []
