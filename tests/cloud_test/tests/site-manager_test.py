"""
pytest sm-client common commands tests_for_integration
python3 -u -m pytest  ./tests_for_integration/cloud_test -k SitemanagerTestCase
"""

import logging
import os

import pytest
import requests
import urllib3
from urllib3.exceptions import InsecureRequestWarning

test_dir = os.path.dirname(__file__)
config_dir = "/resources/"


@pytest.mark.usefixtures('config_dir')
@pytest.mark.usefixtures('sm_env')
@pytest.mark.usefixtures('config_ingress_service')
class SitemanagerTestCase:
    def test_sm_connectivity(self, config_dir, sm_env):
        logging.info("TEST CONNECTIVITY SITE-MANAGER")

        # Disable warnings about unsecure tls connection
        urllib3.disable_warnings(InsecureRequestWarning)
        # Check site-manager connectivity
        logging.info(f"Check site-manager for connectivity")
        status_code = requests.get(f"https://{sm_env['host_name']}",
                                   verify=config_dir['template_env']['sites']['site_1']['ca_cert']).status_code
        assert status_code in [200, 204]

    def test_services_connectivity(self, config_ingress_service):
        logging.info("TEST CONNECTIVITY SERVICES")

        # Disable warnings about unsecure tls connection
        urllib3.disable_warnings(InsecureRequestWarning)
        # Check services connectivity
        logging.info(f"Check services for connectivity")

        for service_name, ingress_services in config_ingress_service.items():
            status_code = requests.get(f"http://{ingress_services}").status_code
            assert status_code in [200, 204]

    def test_sm_without_ca(self, sm_env):
        logging.info("TEST SITE-MANAGER WITHOUT CA.CRT")

        # Disable warnings about unsecure tls connection
        urllib3.disable_warnings(InsecureRequestWarning)
        # Check site-manager connectivity
        logging.info(f"Check site-manager for connectivity")
        with pytest.raises(requests.exceptions.SSLError) as e:
            requests.get(f"https://{sm_env['host_name']}/sitemanager", verify=True)

    def test_sm_incorrect_token(self, sm_env, config_dir):
        logging.info("TEST SITE-MANAGER WITH WRONG TOKEN")
        url = 'https://' + sm_env['host_name'] + "/sitemanager"
        headers = {
            "Authorization": f"Bearer test"
        }

        resp = requests.get(url, headers=headers, verify=config_dir['template_env']['sites']['site_1']['ca_cert'])
        assert resp.status_code == 403

    def test_sm_empty_token(self, sm_env, config_dir):
        logging.info("TEST SITE-MANAGER WITH EMPTY TOKEN")
        url = 'https://' + sm_env['host_name'] + "/sitemanager"

        resp = requests.get(url, verify=config_dir['template_env']['sites']['site_1']['ca_cert'])
        assert resp.status_code == 401

    def test_sm_check_incorrect_services(self, sm_env, config_dir):
        logging.info("TEST SITE-MANAGER WITH WRONG SERVICES")
        url = 'https://' + sm_env['host_name'] + "/sitemanager"
        token = sm_env['token-sm']
        headers = {
            "Authorization": f"Bearer {token}"
        }
        http_body = {"procedure": "status", "run-service": "not-exist"}
        resp = requests.post(url, json=http_body, headers=headers,
                             verify=config_dir['template_env']['sites']['site_1']['ca_cert'])
        resp_json = resp.json()
        assert "Service doesn't exist" == resp_json['message']
        assert resp.status_code == 400

    def test_sm_check_incorrect_procedure(self, sm_env, config_dir):
        logging.info("TEST SITE-MANAGER WITH WRONG PROCEDURE")
        url = 'https://' + sm_env['host_name'] + "/sitemanager"
        token = sm_env['token-sm']
        headers = {
            "Authorization": f"Bearer {token}"
        }

        http_body = {"procedure": "lock", "run-service": "service-a.test-services"}
        resp = requests.post(url, json=http_body, headers=headers,
                             verify=config_dir['template_env']['sites']['site_1']['ca_cert'])
        resp_json = resp.json()
        assert "You should define procedure from list: [list status active standby disable]" == resp_json[
            'message']
        assert resp.status_code == 400

    def test_check_services(self, sm_env, config_dir, config_ingress_service):
        logging.info("TEST CHECK SERVICES STATUS")

        url = 'https://' + sm_env['host_name'] + "/sitemanager"
        token = sm_env['token-sm']
        headers = {
            "Authorization": f"Bearer {token}"
        }
        # Disable warnings about unsecure tls connection
        urllib3.disable_warnings(InsecureRequestWarning)
        # Check services connectivity
        logging.info(f"Check status services")

        for service_name, ingress_services in config_ingress_service.items():
            http_body = {"procedure": "status", "run-service": service_name}
            resp_status = requests.get(f"http://{ingress_services}/sitemanager", headers=headers,
                                       verify=config_dir['template_env']['sites']['site_1']['ca_cert'])
            resp_healthz = requests.get(f"http://{ingress_services}/healthz", headers=headers,
                                        verify=config_dir['template_env']['sites']['site_1']['ca_cert'])
            resp_sm = requests.post(url, json=http_body, headers=headers,
                             verify=config_dir['template_env']['sites']['site_1']['ca_cert'])

            resp_status_json = resp_status.json()
            resp_healthz_json = resp_healthz.json()
            resp_sm_json = resp_sm.json()

            service_status = {"healthz": resp_healthz_json.get("status", "--"),
                              "mode": resp_status_json.get("mode", "--"),
                              "status": resp_status_json.get("status", "--"),
                              "message": resp_status_json.get("message", "")}
            assert service_status == resp_sm_json['services'][service_name]
            assert resp_status.status_code, resp_healthz.status_code in [200, 204]

    def test_check_services_in_sm(self, sm_env, config_dir, config_ingress_service):
        logging.info("TEST CHECK SERVICES ON SITE-MANAGER")
        url = 'https://' + sm_env['host_name'] + "/sitemanager"
        token = sm_env['token-sm']
        headers = {
            "Authorization": f"Bearer {token}"
        }
        resp = requests.get(url, headers=headers,
                            verify=config_dir['template_env']['sites']['site_1']['ca_cert'])

        resp_status_code = resp.status_code
        assert resp_status_code in [200, 204]
        resp_json = resp.json()
        for service_name in config_ingress_service.keys():
            assert service_name in resp_json['services']
