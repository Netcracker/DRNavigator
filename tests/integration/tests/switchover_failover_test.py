"""
pytest sm-client common commands tests
python3 -u -m pytest  ./tests/integration -k SwitchoverFailoverTestCase
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
class SwitchoverFailoverTestCase:

    def test_init_statuses(self, config_dir, caplog):
        logging.info("TEST INIT STATUSES")
        test_utils.check_statuses(caplog, template_env, lambda site, service:
                                   {"healthz": "up", "status": "done", "message": "",
                                    "mode": "active" if template_env["active_site"] == site else "standby"})

    def test_move_to_another_site(self, config_dir, caplog):
        logging.info("TEST MOVE WITH STATEFUL SERVICES TO ANOTHER SITE")
        # Run move to another site
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "move", "site_2"])

        # Check status after move to another site
        test_utils.check_statuses(caplog, template_env, lambda site, service:
                                {"healthz": "up", "status": "done", "message": "",
                                "mode": "active" if "site_2" == site else "standby"})

    def test_stop_site(self, config_dir, caplog):
        logging.info("TEST STOP WITH STATEFUL SERVICES")
        # Run move to another site
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "stop", "site_2"])

        # Check status after move to another site
        test_utils.check_statuses(caplog, template_env, lambda site, service:
                                {"healthz": "up", "status": "done", "message": "",
                                "mode": "standby" if "site_2" == site else "active"})

    def test_move_only_second_service(self, config_dir, caplog):
        logging.info("TEST MOVE SECOND SERVICE TO ANOTHER SITE")
        # Run move to another site
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v",
                "--run-services", "serviceB", "move", "site_2"], expected_exit_code=1)

    def test_move_only_first_service(self, config_dir, caplog):
        logging.info("TEST MOVE FIRST SERVICE TO ANOTHER SITE")
        # Run move to another site
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v",
                "--skip-services", "serviceB", "move", "site_2"])

        # Check status after move to another site
        test_utils.check_statuses(caplog, template_env, lambda site, service:
                {"healthz": "up", "status": "done", "message": "",
                 "mode": "active" if ("site_1" == site) != ("serviceA" == service) else "standby"})

    def test_stop_only_second_service(self, config_dir, caplog):
        logging.info("TEST STOP SECOND SERVICE TO ANOTHER SITE")
        # Run move to another site
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v",
                "--run-services", "serviceB", "stop", "site_2"], expected_exit_code=1)

    def test_stop_only_first_service(self, config_dir, caplog):
        logging.info("TEST STOP FIRST SERVICE TO ANOTHER SITE")
        # Run move to another site
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v",
                "--skip-services", "serviceB", "stop", "site_2"])

        # Check status after move to another site
        test_utils.check_statuses(caplog, template_env, lambda site, service:
                {"healthz": "up", "status": "done", "message": "",
                 "mode": "active" if "site_1" == site else "standby"})
