"""
pytest sm-client common commands tests
python3 -u -m pytest  ./tests/integration -k MaintenanceReturnWihCustomModuleTestCase
"""

import pytest
import logging
import os
from tests.test_utils import check_statuses, run_sm_client_command_with_exit

test_dir = os.path.dirname(__file__)
docker_config_dir = "/resources/service-with-custom-module-cluster"

template_env = {
    "sites": {
        "site_1": {
            "exposed_ports": {
                "service": {
                    "stateful_service": 9001,
                    "custom_module_service": 9002,
                    "custom_module_service2": 9005
                },
                "site_manager": 9011
            },
            "token": "12345"
        },
        "site_2": {
            "exposed_ports": {
                "service": {
                    "stateful_service": 9003,
                    "custom_module_service": 9004,
                    "custom_module_service2": 9006
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
class MaintenanceReturnWihCustomModuleTestCase:

    def test_init_statuses(self, config_dir, capfd):
        logging.info("TEST INIT STATUSES")
        check_statuses(capfd, template_env, lambda site, service:
                                   {"healthz": "up", "status": "done", "message": "",
                                    "mode": "active" if template_env["active_site"] == site else "standby"})

    def test_maintenance_active_site(self, config_dir, capfd):
        logging.info("TEST MAINTENANCE ACTIVE SITE WITH CUSTOM MODULE SERVICES")
        # Run maintenance
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "disable", "site_1"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                                {"healthz": "up", "status": "done", "message": "",
                                 "mode": "disable" if "site_1" == site else "standby"})

    def test_maintenance_standby_site(self, config_dir, capfd):
        logging.info("TEST MAINTENANCE STANDBY SITE WITH CUSTOM MODULE SERVICES")
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

    def test_maintenance_custom_module_service(self, config_dir, capfd):
        logging.info("TEST MAINTENANCE CUSTOM MODULE SERVICE")
        # Run maintenance
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v",
             "--run-services", "custom_module_service,custom_module_service2", "disable", "site_2"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
            {"healthz": "up", "status": "done", "message": "",
             "mode": "standby" if ("site_1" == site) or ("stateful_service" == service) else "disable"})

    def test_return_custom_module_service(self, config_dir, capfd):
        logging.info("TEST RETURN CUSTOM MODULE SERVICE")
        # Run return
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v",
             "--run-services", "custom_module_service,custom_module_service2", "return", "site_2"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
            {"healthz": "up", "status": "done", "message": "",
             "mode": "standby"})

    def test_maintenance_stateful_service(self, config_dir, capfd):
        logging.info("TEST MAINTENANCE STATEFUL SERVICE")
        # Run maintenance
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v",
             "--skip-services", "custom_module_service,custom_module_service2", "disable", "site_2"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
        {"healthz": "up", "status": "done", "message": "",
         "mode": "standby" if ("site_1" == site) or (service in ['custom_module_service','custom_module_service2']) else "disable"})

    def test_return_stateful_service(self, config_dir, capfd):
        logging.info("TEST RETURN STATEFUL SERVICE")
        # Run return
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v",
             "--skip-services", "custom_module_service,custom_module_service2", "return", "site_2"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
        {"healthz": "up", "status": "done", "message": "",
         "mode": "standby"})
