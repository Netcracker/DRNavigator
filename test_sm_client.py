#!/usr/bin/env python3
"""
pytest based unit test
SSL_UNTRUSTED_SITE="https://accessible_ssl_untrusted.com" python3  -m pytest test_sm_client.py
"""
import json
import os

import pytest
import requests
from smclient import io_http_json_request
from http import HTTPStatus


def test_io_http_json_request_ok():
    """ SUCCESS basic general success case with SSL verification
    """
    json_body, http_code = io_http_json_request("https://api.github.com",verify_ssl=True)
    assert http_code == HTTPStatus.OK and \
           type(json_body) is dict and \
           json.loads('"'+str(json_body)+'"'), "Returned: 200 OK, dict, JSON"

def test_io_http_json_request_ok_warning():
    """ SUCCESS basic general success case without SSL verification
    """
    with pytest.warns(Warning) as record:
        json_body, http_code = io_http_json_request("https://api.github.com",verify_ssl=False)

    assert http_code == HTTPStatus.OK and \
           json.loads('"'+str(json_body)+'"') and \
           record, "Returned: 200 OK, dict, JSON, SSL warning "

def test_io_http_json_request_200_not_json():
    """ FAIL in case not JSON returned with 200 OK
    """
    json_body, http_code = io_http_json_request("https://www.github.com",verify_ssl=True)
    assert http_code == False and \
           type(json_body) is dict and \
           bool(dict), "Returned: False and empy dict"

def test_io_http_json_request_404():
    """ FAIL in case 404
    """
    json_body, http_code = io_http_json_request("https://api.github.com/page_does_not_exist",verify_ssl=True)

    assert http_code == HTTPStatus.NOT_FOUND and \
           type(json_body) is dict and \
           bool(dict),"Returned: 404 OK and empty dict"


def test_io_http_json_request_ssl_fails():
    """ FAIL in case SSL verification fails and validated
    """
    exception = None
    try:
        io_http_json_request(os.getenv("SSL_UNTRUSTED_SITE",default="https://ssluntrusted.com"),verify_ssl=True)
    except requests.exceptions.SSLError as e:
        exception = e
    assert exception # exception doesn't happen
