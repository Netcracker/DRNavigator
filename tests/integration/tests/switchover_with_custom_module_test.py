"""
pytest sm-client common commands tests
python3 -u -m pytest  ./tests/integration -k SwitchoverWithCustomModuleTestCase
"""

import pytest
import logging
import os
import test_utils

test_dir = os.path.dirname(__file__)
docker_config_dir = "/resources/service-with-custom-module-cluster"

template_env = {
    "sites": {
        "site_1": {
            "exposed_ports": {
                "service": {
                    "stateful_service": 9001,
                    "custom_module_service": 9002
                },
                "site_manager": 9011
            },
            "token": "12345"
        },
        "site_2": {
            "exposed_ports": {
                "service": {
                    "stateful_service": 9003,
                    "custom_module_service": 9004
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
class SwitchoverWithCustomModuleTestCase:

    def test_init_statuses(self, config_dir, capfd):
        logging.info("TEST INIT STATUSES WITH CUSTOM MODULE SERVICES")
        test_utils.check_statuses(capfd, template_env, lambda site, service:
                                   {"healthz": "up", "status": "done", "message": "",
                                    "mode": "active" if template_env["active_site"] == site else "standby"})

    def test_move_to_another_site(self, config_dir, capfd):
        logging.info("TEST MOVE  WITH CUSTOM MODULE SERVICES TO ANOTHER SITE")
        # Run move to another site
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "move", "site_2"])

        # Check status after move to another site
        test_utils.check_statuses(capfd, template_env, lambda site, service:
                                {"healthz": "up", "status": "done", "message": "",
                                "mode": "active" if "site_2" == site else "standby"})
