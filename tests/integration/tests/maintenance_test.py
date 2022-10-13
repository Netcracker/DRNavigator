"""
pytest sm-client common commands tests
python3 -u -m pytest  ./tests/integration -k MaintenanceTestCase
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
class MaintenanceTestCase:

    def test_init_statuses(self, config_dir, capfd):
        logging.info("TEST INIT STATUSES")
        test_utils.check_statuses(capfd, template_env, lambda site, service:
                                   {"healthz": "up", "status": "done", "message": "",
                                    "mode": "active" if template_env["active_site"] == site else "standby"})

    def test_maintenance_active_site(self, config_dir, capfd):
        logging.info("TEST MAINTENANCE ACTIVE SITE WITH STATEFUL SERVICES")
        # Run activate
        test_utils.run_sm_client_command(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "mntc", "site_1"])

        # Check status after move to another site
        test_utils.check_statuses(capfd, template_env, lambda site, service:
                                {"healthz": "up", "status": "done", "message": "",
                                 "mode": "disable" if "site_1" == site else "standby"})

    def test_maintenance_standby_site(self, config_dir, capfd):
        logging.info("TEST MAINTENANCE STANDBY SITE WITH STATEFUL SERVICES")
        # Run activate
        test_utils.run_sm_client_command(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "mntc", "site_2"])

        # Check status after move to another site
        test_utils.check_statuses(capfd, template_env, lambda site, service:
                                {"healthz": "up", "status": "done", "message": "",
                                 "mode": "disable"})

    def test_return_first_site(self, config_dir, capfd):
        logging.info("TEST RETURN FIRST SITE WITH STATEFUL SERVICES")
        # Run activate
        test_utils.run_sm_client_command(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "return", "site_1"])

        # Check status after move to another site
        test_utils.check_statuses(capfd, template_env, lambda site, service:
                                {"healthz": "up", "status": "done", "message": "",
                                 "mode": "disable" if "site_2" == site else "standby"})

    def test_return_second_site(self, config_dir, capfd):
        logging.info("TEST RETURN SECOND SITE WITH STATEFUL SERVICES")
        # Run activate
        test_utils.run_sm_client_command(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "-v", "return", "site_2"])

        # Check status after move to another site
        test_utils.check_statuses(capfd, template_env, lambda site, service:
                                {"healthz": "up", "status": "done", "message": "",
                                 "mode": "standby"})