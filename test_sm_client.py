#!/usr/bin/env python3
"""
pytest based unit test
python3  -m pytest -s -v test_sm_client.py <-k  test_name*>
TODO
 - use mock instead of http call
"""
import json
import logging
import os

import pytest
from smclient import io_http_json_request,\
    init_and_check_config,service_process,get_DR_operation_sequence,make_ordered_services,parse_command_line
from http import HTTPStatus
import http.server
import ssl
import threading
import yaml

def pytest_namespace():
    return {'site_name': None}

@pytest.mark.skipif(not os.path.exists("config.yaml"),
                    reason="Need to populate config.yaml https://github.com/Netcracker/DRNavigator/blob/main/documentation/Architecture.md#configuration-file")
def test_process_service__status_ok(caplog):
    """ SUCCESS basic general success case without SSL verification
    """
    import argparse

    caplog.set_level(logging.DEBUG)
    args = argparse.ArgumentParser
    args.verbose = True
    args.insecure = True
    args.config = "config.yaml"
    args.run_services = ""
    args.skip_services = ""
    init_and_check_config(args)
    with open("config.yaml",'r') as stream:
        try:
            parsed_yaml=yaml.safe_load(stream)
        except yaml.YAMLError as exc:
            print(exc)
    site = [i["name"] for i in parsed_yaml["sites"]][0]
    print(f"Using {site}")
    pytest.site_name = site
    json_body_s,code = service_process(site,"site-manager","status")
    assert type(json_body_s) is dict \
           and json.loads('"' + str(json_body_s) + '"') \
           and json_body_s["services"] \
           and code == True , \
        "Returned: dict, valid JSON, JSON contains valid response"

#@pytest.mark.depends(on=['test_process_service__status_ok'])
def test_process_service__rw_ok(caplog):
    """ SUCCESS basic general success case without SSL verification
    """
    import argparse

    caplog.set_level(logging.DEBUG)
    args = argparse.ArgumentParser
    args.verbose = True
    args.insecure = True
    args.config = "config.yaml"
    args.run_services = ""
    args.skip_services = ""
    init_and_check_config(args)
    json_body_s,code = service_process(pytest.site_name if hasattr(pytest,"site_name") else "k8s-1",
                                       "dmsh-pg-test-site-manager","status")
    assert  json_body_s \
            and type(json_body_s) is dict \
            and json.loads('"' + str(json_body_s) + '"') \
            and code == True, \
        "Returned: non empty dict, valid JSON, JSON contains valid response"

def test_get_sequence():
    """ DR commands sequence calculation check
    """
    sm_dict = { "sites":
        {"site1":{"services":{
                          "serv1":{"sequence":["standby","active"]}}},
         "site2":{"services":{
                          "serv1":{"sequence":["standby","active"]}}}}}

    sm_dict2 = { "sites":
        { "site1":{ "services":{
            "serv1":{ "sequence":["active","standby"] } } },
            "site2":{ "services":{
                "serv1":{ "sequence":["active","active"] } } } } }

    seq_1 = get_DR_operation_sequence(sm_dict, 'serv1', 'move', 'site2') # switchover to site2
    seq_2 = get_DR_operation_sequence(sm_dict, 'serv1', 'stop', 'site1')  # failover to site2
    seq_3 = get_DR_operation_sequence(sm_dict, 'serv1', 'stop', 'site2')  # failover to site1

    seq2_1 = get_DR_operation_sequence(sm_dict2, 'serv1', 'move', 'site2')  # switchover to site2
    assert seq_1 == [['site1','standby'],['site2','active']] and \
           seq_2 == [['site1','standby'],['site2','active']] and \
           seq_3 == [['site2','standby'],['site1','active']] and \
           seq2_1 == [['site2','active'],['site1','standby']]

def test_make_ordered_services():
    sm_dict = {"sites":
        {"site1":{"services":{
            "a":{"after":["e"]},
            "b":{"after":[]},
            "c":{"after":["b"]},
            "d":{"after":["c"]},
            "e":{"after":["d"]},
            "f":{"after":["a"]}}},
         "site2":{"services":{
             "a":{"after":["e"]},
             "b":{"after":[]},
             "c":{"after":["b"]},
             "d":{"after":["c"]},
             "e":{"after":["d"]},
             "f":{"after":["a"]}
         }}}}
    sm_dict_wrong = {"sites":
        {"site1":{"services":{
            "a":{"after":["e"]},
            "b":{"after":[]},
            "c":{"after":["g"]},
            "d":{"after":[""]},
            "e":{"after":["d"]},
            "f":{"after":["a"]}}},
         "site2":{"services":{
             "a":{"after":["e"]},
             "b":{"after":["z"]},
             "c":{"after":["g"]},
             "d":{"after":[]},
             "e":{"after":["d"]},
             "f":{"after":["az"]}
         }}}}


    sorted_list,code = make_ordered_services(sm_dict)
    assert sorted_list == ['b','c','d','e','a','f'] and  \
           code == True

def test_io_http_json_request_ok():
    """ SUCCESS basic general success case with SSL verification
    """
    json_body, http_code = io_http_json_request("https://api.github.com",verify=True)
    assert http_code == HTTPStatus.OK and \
           type(json_body) is dict and \
           json.loads('"'+str(json_body)+'"'), "Returned: 200 OK, dict, JSON"

def test_io_http_json_request_ok_warning():
    """ SUCCESS basic general success case without SSL verification
    """
    os.environ.setdefault('DEBUG', 'True')  # set DEBUG which is used to print warning
    with pytest.warns(Warning,match=r"Unverified HTTPS request is being made") as record:
        json_body, http_code = io_http_json_request("https://api.github.com",verify=False)

    assert http_code == HTTPStatus.OK and \
           json.loads('"'+str(json_body)+'"') and \
           len(record)>1, "Returned: 200 OK, dict, JSON, SSL warning "

def test_io_http_json_request_200_not_json():
    """ FAIL in case not JSON returned with 200 OK
    """
    json_body, http_code = io_http_json_request("https://www.github.com",verify=True)

    assert http_code == False and \
           type(json_body) is dict and \
           bool(dict), "Returned: False and empy dict"

def test_io_http_json_request_404():
    """ FAIL in case 404
    """
    json_body, http_code = io_http_json_request("https://api.github.com/page_does_not_exist",verify=True)
    assert http_code == HTTPStatus.NOT_FOUND and \
           type(json_body) is dict and \
           json.loads('"' + str(json_body) + '"') and \
           bool(dict),"Returned: 404 OK and nont empty JSON dict"


def test_io_http_json_request_ssl_fails():
    """ FAIL in case SSL verification fails
    """
    os.system("openssl req -new -x509 -keyout self-signed-fake.pem -out self-signed-fake.pem "
              "-days 365 -nodes -subj \"/C=US/ST=Denial/L=Springfield/O=Dis/CN=www.example.com\" &>/dev/null")
    exception = None
    httpd = http.server.HTTPServer(('localhost', 4443), http.server.SimpleHTTPRequestHandler)
    httpd.socket = ssl.wrap_socket(httpd.socket, certfile='self-signed-fake.pem', server_side=True)
    thread = threading.Thread(target=httpd.handle_request)
    thread.start()
    body,code = io_http_json_request("https://localhost:4443",verify=True,retry=0) #thread handles only one request
    thread.join()
    os.remove("self-signed-fake.pem")
    assert code == ssl.SSLErrorNumber.SSL_ERROR_SSL.__int__() \
            and not bool(body), "empty body with specific SSL error"

