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
class DryRunTestCase:

    def test_init_statuses(self, config_dir, capfd):
        logging.info("TEST INIT STATUSES")
        check_statuses(capfd, template_env, lambda site, service:
                                   {"healthz": "up", "status": "done", "message": "",
                                    "mode": "active" if template_env["active_site"] == site else "standby"})

    def test_dry_run_activate(self, config_dir, capfd):
        logging.info("TEST DRY RUN ACTIVATE")
        # Run activate
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "--dry-run", "-v", "active", "site_2"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                                   {"healthz": "up", "status": "done", "message": "",
                                    "mode": "active" if template_env["active_site"] == site else "standby"})

    def test_dry_run_passivate(self, config_dir, capfd):
        logging.info("TEST DRY RUN PASSIVATE")
        # Run passivate
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "--dry-run", "-v", "standby", "site_1"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                                   {"healthz": "up", "status": "done", "message": "",
                                    "mode": "active" if template_env["active_site"] == site else "standby"})

    def test_dry_run_switchover(self, config_dir, capfd):
        logging.info("TEST DRY RUN SWITCHOVER")
        # Run move to another site
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "--dry-run", "-v", "move", "site_2"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                                   {"healthz": "up", "status": "done", "message": "",
                                    "mode": "active" if template_env["active_site"] == site else "standby"})

    def test_dry_run_failover(self, config_dir, capfd):
        logging.info("TEST DRY RUN FAILOVER")
        # Run move to another site
        run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "--dry-run", "-v", "stop", "site_1"])

        # Check status
        check_statuses(capfd, template_env, lambda site, service:
                                   {"healthz": "up", "status": "done", "message": "",
                                    "mode": "active" if template_env["active_site"] == site else "standby"})
