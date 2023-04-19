"""
pytest sm-client common commands tests
python3 -u -m pytest  ./tests/cloud_test -k SitemanagerTestCase
"""

import logging
import os

import pytest
import requests
import urllib3
from urllib3.exceptions import InsecureRequestWarning

test_dir = os.path.dirname(__file__)
config_dir = "/resources/test"


@pytest.mark.usefixtures('config_dir')
@pytest.mark.usefixtures('sm_env')
@pytest.mark.usefixtures('config_ingress_service')
class SitemanagerTestCase:
    def test_wait_sm_until_connectivity(self, config_dir, sm_env):
        logging.info("TEST CONNECTIVITY SITE-MANAGER")

        # Disable warnings about unsecure tls connection
        urllib3.disable_warnings(InsecureRequestWarning)
        # Check site-manager connectivity
        logging.info(f"Check site-manager for connectivity")
        try:
            status_code = requests.get(f"https://{sm_env['host_name']}",
                                       verify=config_dir['template_env']['sites']['site_1']['ca_cert']).status_code
        except Exception as e:
            status_code = 0
        assert status_code in [200, 204]

    def test_wait_services_until_connectivity(self, config_ingress_service):
        logging.info("TEST CONNECTIVITY SERVICES")

        # Disable warnings about unsecure tls connection
        urllib3.disable_warnings(InsecureRequestWarning)
        # Check services connectivity
        logging.info(f"Check services for connectivity")

        for service_name, ingress_services in config_ingress_service.items():
            try:
                status_code = requests.get(f"http://{ingress_services}/healthz").status_code
            except Exception as e:
                status_code = 0
            assert status_code in [200, 204]

    def test_wait_sm_until_connectivity_incorrect_ca(self, sm_env):
        logging.info("TEST CONNECTIVITY SITE-MANAGER WITH WRONG CA.CRT")

        # Disable warnings about unsecure tls connection
        urllib3.disable_warnings(InsecureRequestWarning)
        # Check site-manager connectivity
        logging.info(f"Check site-manager for connectivity")
        try:
            status_code = requests.get(f"https://{sm_env['host_name']}/sitemanager", verify="").status_code
        except Exception as e:
            status_code = 0
        assert status_code in [401]

    def test_wait_sm_until_connectivity_incorrect_token(self, sm_env, config_dir):
        logging.info("TEST CONNECTIVITY SITE-MANAGER WITH WRONG TOKEN")
        url = 'https://' + sm_env['host_name'] + "/sitemanager"
        token = sm_env['token-sm']
        if token:
            headers = {
                "Authorization": f"Bearer test"
            }
        else:
            headers = {}
        try:
            resp = requests.get(url, headers=headers, verify=config_dir['template_env']['sites']['site_1']['ca_cert'])
        except Exception:
            resp = 0
        assert resp.status_code in [403]

    def test_wait_sm_until_check_incorrect_services(self, sm_env, config_dir):
        logging.info("TEST CONNECTIVITY SITE-MANAGER WITH WRONG SERVICES")
        url = 'https://' + sm_env['host_name'] + "/sitemanager"
        token = sm_env['token-sm']
        if token:
            headers = {
                "Authorization": f"Bearer {token}"
            }
        else:
            headers = {}
        ingress_service = "service-c"
        http_body = {"procedure": "status", "run-service": ingress_service}
        try:
            resp = requests.post(url, json=http_body, headers=headers,
                                 verify=config_dir['template_env']['sites']['site_1']['ca_cert'])
        except Exception:
            resp.status_code = 0
        assert resp.status_code in [400]

    def test_check_services_in_sm(self, sm_env, config_dir, config_ingress_service):
        url = 'https://' + sm_env['host_name'] + "/sitemanager"
        token = sm_env['token-sm']
        if token:
            headers = {
                "Authorization": f"Bearer {token}"
            }
        else:
            headers = {}

        for ingress_service in config_ingress_service.keys():
            http_body = {"procedure": "status", "run-service": ingress_service}
            try:
                resp = requests.get(url, json=http_body, headers=headers,
                                    verify=config_dir['template_env']['sites']['site_1']['ca_cert'])
            except Exception:
                resp.status_code = 0
            assert ingress_service in resp.text
