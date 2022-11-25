"""
pytest sm-client common commands tests
python3 -u -m pytest  ./tests/integration -k RestrictionsTestCase
"""

import pytest
import logging
import os
import test_utils

test_dir = os.path.dirname(__file__)
docker_config_dir = "/resources/service-a-b-cluster"

template_env = {
    "sites": {
        "site_1": {
            "exposed_ports": {
                "service": {
                    "serviceA": 9001,
                    "serviceB": 9002
                },
                "site_manager": 9011
            },
            "token": "12345"
        },
        "site_2": {
            "exposed_ports": {
                "service": {
                    "serviceA": 9003,
                    "serviceB": 9004
                },
                "site_manager": 9012
            },
            "token": "54321"
        }
    },
    "active_site": "site_1",
    "service_timeout": 0,
    "config_dir": "<is filled in config_dir fixture>",
    "os_path_sep": "<is filled in config_dir fixture>"
}


@pytest.mark.usefixtures('config_dir')
@pytest.mark.usefixtures('prepare_docker_compose')
@pytest.mark.usefixtures('wait_services_until_healthy')
class RestrictionsTestCase:

    def test_init_statuses(self, config_dir, capfd):
        logging.info("TEST INIT STATUSES")
        test_utils.check_statuses(capfd, template_env, lambda site, service:
                                   {"healthz": "up", "status": "done", "message": "",
                                    "mode": "active" if template_env["active_site"] == site else "standby"})

    def test_passivate_all_sites(self):
        logging.info("TEST PASSIVATE ALL WHEN IT'S RESTRICTED")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config-with-restrictions.yaml'),
             "-v", "standby", "site_1"], expected_exit_code=1)

    def test_activate_all_sites(self):
        logging.info("TEST ACTIVATE ALL WHEN IT'S RESTRICTED")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config-with-restrictions.yaml'),
             "-v", "active", "site_2"], expected_exit_code=1)

    def test_activate_sites_for_not_restricted_services(self):
        logging.info("TEST ACTIVATE NOT RESTRICTED SERVICES")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config-with-restrictions.yaml'),
             "-v", "--run-services", "serviceA", "active", "site_2"])

    def test_activate_all_sites_with_skip_restrictions(self):
        logging.info("TEST ACTIVATE ALL WITH RESTRICTIONS SKIPPING")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config-with-restrictions.yaml'),
             "-v", "-r", "active", "site_2"])

    def test_passivate_all_sites_with_skip_restrictions(self):
        logging.info("TEST PASSIVATE ALL WITH RESTRICTIONS SKIPPING")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config-with-restrictions.yaml'),
             "-v", "-r", "standby", "site_2"])
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config-with-restrictions.yaml'),
             "-v", "-r", "standby", "site_1"])
