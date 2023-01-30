"""
pytest sm-client common commands tests
python3 -u -m pytest  ./tests/integration -k NotExistServicesTestCase
"""

import pytest
import logging
import os
import test_utils

test_dir = os.path.dirname(__file__)
docker_config_dir = "/resources/service-a-b-cluster-with-diff-names"

template_env = {
    "sites": {
        "site_1": {
            "exposed_ports": {
                "service": {
                    "serviceASite1": 9001,
                    "serviceB": 9002
                },
                "site_manager": 9011
            },
            "token": "12345"
        },
        "site_2": {
            "exposed_ports": {
                "service": {
                    "serviceASite2": 9003,
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
class NotExistServicesTestCase:

    def test_init_statuses(self, config_dir, capfd):
        logging.info("TEST INIT STATUSES")
        test_utils.check_statuses(capfd, template_env, lambda site, service:
                                   {"healthz": "up", "status": "done", "message": "",
                                    "mode": "active" if template_env["active_site"] == site else "standby"}
                                  if (service != "serviceASite1" or site != "site_2") and
                                     (service != "serviceASite2" or site != "site_1") else
                                   {"healthz": "--", "status": "--", "message": "Service doesn't exist", "mode": "--"})


    def test_move_statuses(self, config_dir, capfd):
        logging.info("TEST MOVE STATUSES")

        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "move", "site_2"],
            expected_exit_code=1)

    def test_stop_statuses(self, config_dir, capfd):
        logging.info("TEST MOVE STATUSES")

        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "stop", "site_1"],
            expected_exit_code=1)