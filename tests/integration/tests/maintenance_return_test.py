"""
pytest sm-client common commands tests
python3 -u -m pytest  ./tests/integration -k MaintenanceReturnTestCase
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
class MaintenanceReturnTestCase:

    def test_init_statuses(self, config_dir, capfd):
        logging.info("TEST INIT STATUSES")
        check_statuses(capfd, template_env, lambda site, service:
                                   {"healthz": "up", "status": "done", "message": "",
                                    "mode": "active" if template_env["active_site"] == site else "standby"})

    def test_maintenance_active_site(self, config_dir, capfd):
        logging.info("TEST MAINTENANCE ACTIVE SITE WITH STATEFUL SERVICES")
        # Run maintenance
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "disable", "site_1"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                                {"healthz": "up", "status": "done", "message": "",
                                 "mode": "disable" if "site_1" == site else "standby"})

    def test_maintenance_standby_site(self, config_dir, capfd):
        logging.info("TEST MAINTENANCE STANDBY SITE WITH STATEFUL SERVICES")
        # Run maintenance
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "disable", "site_2"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                                {"healthz": "up", "status": "done", "message": "",
                                 "mode": "disable"})

    def test_return_first_site(self, config_dir, capfd):
        logging.info("TEST RETURN FIRST SITE WITH STATEFUL SERVICES")
        # Run return
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "return", "site_1"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                                {"healthz": "up", "status": "done", "message": "",
                                 "mode": "disable" if "site_2" == site else "standby"})

    def test_return_second_site(self, config_dir, capfd):
        logging.info("TEST RETURN SECOND SITE WITH STATEFUL SERVICES")
        # Run return
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "return", "site_2"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                                {"healthz": "up", "status": "done", "message": "",
                                 "mode": "standby"})

    def test_maintenance_second_service(self, config_dir, capfd):
        logging.info("TEST MAINTENANCE SECOND SERVICE")
        # Run maintenance
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v",
             "--run-services", "serviceB", "disable", "site_2"], expected_exit_code=1)

    def test_return_second_service(self, config_dir, capfd):
        logging.info("TEST RETURN SECOND SERVICE")
        # Run return
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v",
             "--run-services", "serviceB", "return", "site_2"], expected_exit_code=1)

    def test_maintenance_first_service(self, config_dir, capfd):
        logging.info("TEST MAINTENANCE FIRST SERVICE")
        # Run maintenance
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v",
             "--skip-services", "serviceB", "disable", "site_2"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                            {"healthz": "up", "status": "done", "message": "",
                            "mode": "standby" if ("site_1" == site) or ("serviceB" == service) else "disable"})

    def test_return_first_service(self, config_dir, capfd):
        logging.info("TEST RETURN FIRST SERVICE")
        # Run return
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v",
             "--skip-services", "serviceB", "return", "site_2"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                            {"healthz": "up", "status": "done", "message": "",
                             "mode": "standby"})
