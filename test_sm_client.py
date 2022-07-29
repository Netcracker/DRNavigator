#!/usr/bin/env python3
"""
pytest based unit test
python3  -m pytest -s -k test_sm_client.py
TODO
 - use mock instead of http call
"""
import json
import logging
import os
from logging import INFO

import pytest
import requests
from smclient import io_http_json_request,config_checks_init,process_service
from http import HTTPStatus
import http.server
import ssl
import threading

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
    config_checks_init(args)
    json_body_s,code = process_service("k8s-2","site-manager","status")
    assert type(json_body_s) is dict \
           and json.loads('"' + str(json_body_s) + '"') \
           and json_body_s["services"] \
           and code == True , \
        "Returned: dict, valid JSON, JSON contains valid response"

def test_process_service__rw_ok(caplog):
    """ SUCCESS basic general success case without SSL verification
    """
    import argparse

    caplog.set_level(logging.DEBUG)
    args = argparse.ArgumentParser
    args.verbose = True
    args.insecure = True
    args.config = "config.yaml"
    args.run_services=""
    args.skip_services=""
    config_checks_init(args)
    json_body_s,code = process_service("k8s-2","kafka-kafka-service","status")
    assert  json_body_s \
            and type(json_body_s) is dict \
            and json.loads('"' + str(json_body_s) + '"') \
            and code == True, \
        "Returned: non empty dict, valid JSON, JSON contains valid response"

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

