"""
pytest sm-client common commands tests
python3 -u -m pytest  ./tests/integration -k ActivatePassivateWithCustomModuleTestCase
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
class ActivatePassivateWithCustomModuleTestCase:

    def test_init_statuses(self, config_dir, capfd):
        logging.info("TEST INIT STATUSES")
        check_statuses(capfd, template_env, lambda site, service:
                                   {"healthz": "up", "status": "done", "message": "",
                                    "mode": "active" if template_env["active_site"] == site else "standby"})

    def test_activate_all_sites(self, config_dir, capfd):
        logging.info("TEST ACTIVATE ALL WITH CUSTOM MODULE SERVICES")
        # Run activate
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "active", "site_2"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                                {"healthz": "up", "status": "done", "message": "", "mode": "active"})

    def test_passivate_site(self, config_dir, capfd):
        logging.info("TEST PASSIVATE WITH CUSTOM MODULE SERVICES")
        # Run passivate
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "standby", "site_2"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                                {"healthz": "up", "status": "done", "message": "",
                                 "mode": "standby" if "site_2" == site else "active"})

    def test_passivate_all_sites(self, config_dir, capfd):
        logging.info("TEST PASSIVATE ALL WITH CUSTOM MODULE SERVICES")
        # Run passivate
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "standby", "site_1"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                                {"healthz": "up", "status": "done", "message": "","mode": "standby"})

    def test_activate_site(self, config_dir, capfd):
        logging.info("TEST ACTIVATE WITH CUSTOM MODULE SERVICES")
        # Run activate
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "active", "site_1"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                                {"healthz": "up", "status": "done", "message": "",
                                 "mode": "active" if "site_1" == site else "standby"})

    def test_activate_second_service(self, config_dir, capfd):
        logging.info("TEST ACTIVATE CUSTOM MODULE SERVICE")
        # Run activate
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v",
             "--run-services", "custom_module_service", "active", "site_2"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                {"healthz": "up", "status": "done", "message": "",
                 "mode": "active" if ("site_1" == site) or ("custom_module_service" == service) else "standby"})

    def test_passivate_second_service(self, config_dir, capfd):
        logging.info("TEST PASSIVATE CUSTOM MODULE SERVICE")
        # Run passivate
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v",
             "--run-services", "custom_module_service", "standby", "site_2"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                {"healthz": "up", "status": "done", "message": "",
                 "mode": "active" if "site_1" == site else "standby"})

    def test_activate_first_service(self, config_dir, capfd):
        logging.info("TEST ACTIVATE STATEFUL SERVICE")
        # Run activate
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v",
             "--skip-services", "custom_module_service,custom_module_service2", "active", "site_2"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                            {"healthz": "up", "status": "done", "message": "",
                            "mode": "active" if ("site_1" == site) or ("stateful_service" == service) else "standby"})

    def test_passivate_first_service(self, config_dir, capfd):
        logging.info("TEST PASSIVATE STATEFUL SERVICE")
        # Run passivate
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v",
             "--skip-services", "custom_module_service,custom_module_service2", "standby", "site_2"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                            {"healthz": "up", "status": "done", "message": "",
                             "mode": "active" if "site_1" == site else "standby"})
