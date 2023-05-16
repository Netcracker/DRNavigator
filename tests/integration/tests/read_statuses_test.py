"""
pytest sm-client common commands tests
python3 -u -m pytest  ./tests/integration -k ReadStatusesTestCase
"""

import pytest
import logging
import os
from tests.test_utils import check_statuses, run_sm_client_command_with_exit

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
class ReadStatusesTestCase:

    def test_help_section(self):
        logging.info("TEST HELP COMMAND")
        run_sm_client_command_with_exit(["-h"])

    def test_version(self):
        logging.info("TEST VERSION COMMAND")
        run_sm_client_command_with_exit(["version"])

    def test_list_without_config(self):
        logging.info("TEST LIST COMMAND WITHOUT CONFIG")
        run_sm_client_command_with_exit(["list"], expected_exit_code=1)

    def test_list_section(self, config_dir):
        logging.info("TEST LIST COMMAND")
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "list"])

    def test_status_without_config(self):
        logging.info("TEST STATUS COMMAND WITHOUT CONFIG")
        run_sm_client_command_with_exit(["status"], expected_exit_code=1)

    def test_status(self, capfd):
        logging.info("TEST STATUS COMMAND")
        check_statuses(capfd, template_env, lambda site, service:
                                {"healthz": "up", "status": "done", "message": "",
                                "mode": "active" if template_env["active_site"] == site else "standby"})


