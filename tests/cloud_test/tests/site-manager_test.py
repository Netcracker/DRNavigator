import pytest
import os
import logging
import urllib3
import requests
import time
from urllib3.exceptions import InsecureRequestWarning
test_dir = os.path.dirname(__file__)
config_dir = "/resources/test"
@pytest.mark.usefixtures('config_dir')
@pytest.mark.usefixtures('sm_env')
@pytest.mark.usefixtures('config_ingress_service')
class SitemanagerTestCase:
    def test_wait_services_until_healthy(self, config_dir, sm_env):
        attempt_count = 5
        sleep_seconds = 5
        all_services_started = True

        # Disable warnings about unsecure tls connection
        urllib3.disable_warnings(InsecureRequestWarning)
        # Check services health
        for attempt in range(attempt_count):
            logging.info(f"Check services for healthy, attempt {attempt + 1}...")
            all_services_started = True
            try:
                 status_code = requests.get(f"https://{sm_env['host_name']}", verify=config_dir['tmp_dir'] + "/ca.crt").status_code
            except Exception as e:
                status_code = 0
            if status_code not in [200, 204]:
                logging.info(f"Site-manager on site  hasn't started yet")
                all_services_started = False
            if all_services_started:
                break
            logging.info(f"Not all services started, sleep {sleep_seconds} seconds")
            time.sleep(sleep_seconds)

        if all_services_started:
            logging.info("All services are healthy")
        else:
            logging.error("Some services haven't started in expected time")

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
            except Exception as e:
                resp.status_code = 0
            if resp.status_code not in [200, 204]:
                logging.info(f"Site-manager could not get the service {ingress_service} status")
            else:
                logging.info(f"Site-manager could  get the service {ingress_service} status")


