"""
pytest sm-client common commands tests
python3 -u -m pytest  ./tests/cloud_test -k SitemanagerTestCase
"""

import pytest
import os
import logging
import urllib3
import requests
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
            status_code = requests.get(f"https://{sm_env['host_name']}", verify=config_dir['template_env']['sites']['site_1']['ca_cert']).status_code
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
                status_code = requests.get(f"http://{ingress_services}").status_code
            except Exception as e:
                status_code = 0
            assert status_code in [200, 204]

    def test_status_service(self, sm_env, config_ingress_service, use_auth=True, verify=False):
        url = 'https://' + sm_env['host_name'] + "/sitemanager"
        token = sm_env['token-sm']
        if token and use_auth:
            headers = {
                "Authorization": f"Bearer {token}"
            }
        else:
            headers = {}

        for ingress_service in config_ingress_service.keys():
            http_body = {"procedure": "status", "run-service": ingress_service}
            try:
                if any(http_body):
                    resp = requests.post(url, json=http_body,  headers=headers, verify=verify)
                else:
                    resp = requests.get(url,  headers=headers, verify=verify)
                logging.debug(f"REST response: {resp.json()}")
            except Exception:
                resp.status_code = 0
            assert resp.status_code in [200, 204]


