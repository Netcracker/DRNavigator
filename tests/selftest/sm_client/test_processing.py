import http
import json
import logging
import ssl
import threading
from http import HTTPStatus

import pytest

import smclient
from common.utils import io_make_http_json_request
from sm_client.data.structures import *
from sm_client.initialization import init_and_check_config
from sm_client.processing import sm_process_service, thread_result_queue, process_ts_services, \
    sm_poll_service_required_status, sm_process_service_with_polling, process_module_services
from tests.selftest.sm_client.common.test_utils import *


def test_sm_process_service(mocker, caplog):
    init_and_check_config(args_init())
    test_resp = {'services': {'test1': {'healthz': 'up', 'mode': 'active', 'status': 'done'}}}
    caplog.set_level(logging.DEBUG)
    fake_resp = mocker.Mock()
    fake_resp.json = mocker.Mock(return_value=test_resp)
    fake_resp.status_code = HTTPStatus.OK

    mocker.patch("common.utils.requests.Session.post", return_value=fake_resp)

    json_body_s, ret, code = sm_process_service("k8s-1", "test1", "active")

    assert json_body_s['services'] == {'test1': {'healthz': 'up', 'mode': 'active', 'status': 'done'}} and \
           ret is True and \
           code is HTTPStatus.OK


def test_io_http_json_request():
    """ SUCCESS basic general success case with SSL verification """
    ret, json_body, http_code = io_make_http_json_request("https://api.github.com", verify=True)
    assert http_code == HTTPStatus.OK and \
           type(json_body) is dict and \
           json.loads('"' + str(json_body) + '"'), "Returned: 200 OK, dict, JSON"

    """ SUCCESS basic general success case without SSL verification """
    os.environ.setdefault('DEBUG', 'True')  # set DEBUG which is used to print warning
    with pytest.warns(Warning, match=r"Unverified HTTPS request is being made") as record:
        ret, json_body, http_code = io_make_http_json_request("https://api.github.com", verify=False)

    assert http_code == HTTPStatus.OK and \
           json.loads('"' + str(json_body) + '"') and \
           len(record) > 1, "Returned: 200 OK, dict, JSON, SSL warning "

    """ FAIL in case not JSON returned with 200 OK """
    ret, json_body, http_code = io_make_http_json_request("https://www.github.com", verify=True)

    assert http_code == False and \
           type(json_body) is dict and \
           bool(dict), "Returned: False and empy dict"

    """ FAIL in case 404 """
    ret, json_body, http_code = io_make_http_json_request("https://api.github.com/page_does_not_exist", verify=True)
    assert http_code == HTTPStatus.NOT_FOUND and \
           type(json_body) is dict and \
           json.loads('"' + str(json_body) + '"') and \
           bool(dict), "Returned: 404 OK and nont empty JSON dict"

    """ FAIL in case SSL verification fails """
    os.system("openssl req -new -x509 -keyout self-signed-fake.pem -out self-signed-fake.pem "
              "-days 365 -nodes -subj \"/C=US/ST=Denial/L=Springfield/O=Dis/CN=www.example.com\"")
    httpd = http.server.HTTPServer(('localhost', 4443), http.server.SimpleHTTPRequestHandler)
    httpd.socket = ssl.wrap_socket(httpd.socket, certfile='self-signed-fake.pem', server_side=True)
    thread = threading.Thread(target=httpd.handle_request)
    thread.start()
    ret, body, code = io_make_http_json_request("https://localhost:4443", verify=True,
                                                retry=0)  # thread handles only one request
    thread.join()
    os.remove("self-signed-fake.pem")
    assert ret is False and \
           code is ssl.SSLErrorNumber.SSL_ERROR_SSL.__int__() and \
           not bool(body), "empty body with specific SSL error"

    ret, json_body, http_code = io_make_http_json_request("https://absent-site", verify=True)
    assert http_code is False and \
           json_body == {} and \
           ret is False


def test_runservice_engine(caplog):
    def process_node(node):
        node = ServiceDRStatus({'services': {node: {}}})
        if node.service in test_failed_services:
            node.healthz = 'down'
        else:
            node.healthz = 'up'
        thread_result_queue.put(node)

    caplog.set_level(logging.INFO)
    ts = TopologicalSorter2()
    ts.add("aa")
    ts.add("bb1", "bb")
    ts.add("cc")
    ts.add("cc1", "cc")
    test_failed_services = ['bb']
    ts.prepare()
    """ ------------ """

    process_ts_services(ts, process_node)

    logging.info(f"failed_services: {settings.failed_services}")
    logging.info(f"done_services: {settings.done_services}")
    assert settings.done_services in [['aa', 'cc', 'cc1'], ['cc', 'aa', 'cc1'], ['cc', 'cc1', 'aa']] \
           and settings.failed_services == ['bb', 'bb1']


def test_sm_poll_service_required_status(mocker, caplog):
    init_and_check_config(args_init())

    test_resp = {'services': {'serv1': {'healthz': 'up', 'mode': 'active', 'status': 'done'}}}
    caplog.set_level(logging.DEBUG)
    fake_resp = mocker.Mock()
    fake_resp.json = mocker.Mock(return_value=test_resp)
    fake_resp.status_code = HTTPStatus.OK

    mocker.patch("common.utils.requests.Session.post", return_value=fake_resp)

    # default timeout
    sm_dict = SMClusterState()
    sm_dict["k8s-1"] = {"services": {
        "serv1": {"timeout": None}}}
    with caplog.at_level(logging.INFO):
        caplog.clear()
        dr_status = sm_poll_service_required_status("k8s-1", "serv1", "active", sm_dict)
        assert f"{settings.SERVICE_DEFAULT_TIMEOUT} seconds left until timeout" in caplog.text
        assert dr_status.is_ok()

    # service specific timeout
    sm_dict = SMClusterState()
    sm_dict["k8s-1"] = {"services": {
        "serv1": {"timeout": 100}}}
    with caplog.at_level(logging.INFO):
        caplog.clear()
        sm_poll_service_required_status("k8s-1", "serv1", "active", sm_dict)
        assert "100 seconds left until timeout" in caplog.text

    # service specific timeout occured
    sm_dict["k8s-1"] = {"services": {
        "serv1": {"timeout": 1}}}
    with caplog.at_level(logging.INFO):
        caplog.clear()
        sm_poll_service_required_status("k8s-1", "serv1", "disable", sm_dict)
        assert "Timeout expired" in caplog.text

    # healthz up
    test_resp = {'services': {'serv1': {'healthz': 'up', 'mode': 'active', 'status': 'done'}}}
    fake_resp.json = mocker.Mock(return_value=test_resp)
    with caplog.at_level(logging.INFO):
        caplog.clear()
        assert sm_poll_service_required_status("k8s-1", "serv1", "active", sm_dict).is_ok() and \
               "Expected state" in caplog.text

    # polling successful 'healthz':'down' 'mode':'standby' "allowedStandbyStateList":["down"]
    test_resp = {'services': {'serv1': {'healthz': 'down', 'mode': 'standby', 'status': 'done'}}}
    fake_resp.json = mocker.Mock(return_value=test_resp)
    sm_dict = SMClusterState()
    sm_dict["k8s-1"] = {"services": {
        "serv1": {"timeout": 100,
                  "allowedStandbyStateList": ["down"]}}}
    assert 'down' in sm_poll_service_required_status("k8s-1", "serv1", "standby", sm_dict).healthz
    # @todo need to detect True for service_status_polling in case standby

    # 'healthz':'down'
    test_resp = {'services': {'serv1': {'healthz': 'down', 'mode': 'active', 'status': 'done'}}}
    fake_resp.json = mocker.Mock(return_value=test_resp)
    assert not sm_poll_service_required_status("k8s-1", "serv1", "active", sm_dict).is_ok()

    # 'status':'failed'}
    test_resp = {'services': {'serv1': {'healthz': 'up', 'mode': 'active', 'status': 'failed'}}}
    fake_resp.json = mocker.Mock(return_value=test_resp)
    assert not sm_poll_service_required_status("k8s-1", "serv1", "active", sm_dict).is_ok()

    # 'healthz':'down' 'status':'running'}
    sm_dict["k8s-1"] = {"services": {
        "serv1": {"timeout": 1}}}
    test_resp = {'services': {'serv1': {'healthz': 'down', 'mode': 'active', 'status': 'running'}}}
    fake_resp.json = mocker.Mock(return_value=test_resp)
    with caplog.at_level(logging.INFO):
        caplog.clear()
        assert not sm_poll_service_required_status("k8s-1", "serv1", "active", sm_dict).is_ok() and \
               "Error state" not in caplog.text


def test_sm_process_service_with_polling(mocker, caplog):
    smclient.args = args_init()
    init_and_check_config(args_init())
    test_resp = {'services': {'serv1': {'healthz': 'up', 'mode': 'active', 'status': 'failed'}}}
    caplog.set_level(logging.DEBUG)
    fake_resp = mocker.Mock()
    fake_resp.json = mocker.Mock(return_value=test_resp)
    fake_resp.status_code = HTTPStatus.OK

    mocker.patch("common.utils.requests.Session.post", return_value=fake_resp)

    # custom timeout
    sm_dict = SMClusterState()
    sm_dict["k8s-1"] = {
        "services": {
            "serv1": {"timeout": 1,
                      "sequence": ['active', 'standby']}},
        "status": True}
    with caplog.at_level(logging.INFO):
        caplog.clear()
        sm_process_service_with_polling("serv1", "k8s-1", "move", sm_dict)
        service_response = thread_result_queue.get()
        service_response.sortout_service_results()
        assert 'serv1' in settings.failed_services
        assert "Service serv1 failed on k8s-1, skipping it on another site" in caplog.text

    # timeout expired fail
    test_resp = {'services': {'serv2': {'healthz': 'up', 'mode': 'standby', 'status': 'done'}}}
    fake_resp.json = mocker.Mock(return_value=test_resp)
    mocker.patch("common.utils.requests.Session.post", return_value=fake_resp)
    sm_dict = SMClusterState()
    sm_dict["k8s-1"] = {
        "services": {"serv2": {"timeout": 1}},
        "status": True}
    with caplog.at_level(logging.INFO):
        caplog.clear()
        sm_process_service_with_polling("serv2", "k8s-1", "active", sm_dict)
        service_response = thread_result_queue.get()
        service_response.sortout_service_results()
        assert 'serv2' in settings.failed_services


def test_process_module_services(mocker, caplog):
    caplog.set_level(logging.INFO)
    smclient.args = args_init()
    init_and_check_config(args_init())
    test_resp = {'services': {'serv1': {'healthz': 'up', 'mode': 'active', 'status': 'done'}}}
    caplog.set_level(logging.DEBUG)
    fake_resp = mocker.Mock()
    fake_resp.json = mocker.Mock(return_value=test_resp)
    fake_resp.status_code = HTTPStatus.OK

    mocker.patch("common.utils.requests.Session.post", return_value=fake_resp)

    settings.done_services.clear()
    sm_dict = SMClusterState()
    sm_dict["k8s-1"] = {
        "services": {
            "serv1": {"timeout": 1,
                      "sequence": ['standby', 'active']}},
        "stateful": {"ts": None},
        "status": False}
    ts = TopologicalSorter2()
    ts.add("serv1")
    ts.prepare()
    sm_dict["k8s-2"] = {
        "services": {
            "serv1": {"timeout": 1,
                      "sequence": ['standby', 'active'],
                      "allowedStandbyStateList": "up"}},
        "stateful": {"ts": ts},
        "status": True}

    process_module_services("stateful", "", "stop", "k8s-1", sm_dict)
    assert "serv1" in settings.done_services

    settings.done_services.clear()
    ts = TopologicalSorter2()
    ts.add("serv1")
    ts.prepare()
    sm_dict["k8s-2"]["stateful"]["ts"] = ts
    process_module_services("stateful", "", "active", "k8s-2", sm_dict)
    assert "serv1" in settings.done_services
