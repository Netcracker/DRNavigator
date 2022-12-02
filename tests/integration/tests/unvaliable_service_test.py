"""
pytest sm-client common commands tests
python3 -u -m pytest  ./tests/integration -k UnvaliableServiceTestCase
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
class UnvaliableServiceTestCase:

    def test_init_statuses(self, config_dir, capfd):
        logging.info("Pause service to emulate not working")
        os.system(f"docker-compose -f {os.path.join(config_dir, 'docker-compose.yaml')} pause serviceB-site-2")

        logging.info("TEST INIT STATUSES")
        test_utils.check_statuses(capfd, template_env, lambda site, service: {
            "healthz": "up" if "site_1" == site or "serviceA" == service else "--",
            "status": "done" if "site_1" == site or "serviceA" == service else "--",
            "mode": "active" if "site_1" == site else "standby" if service == "serviceA" else "--",
            "message": ""})

    def test_passivate_working_site(self):
        logging.info("TEST PASSIVATE WORKING SITE")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "standby", "site_1"])

    def test_activate_working_site(self):
        logging.info("TEST ACTIVATE WORKING SITE")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "active", "site_1"])

    def test_passivate_unvaliable_site(self):
        logging.info("TEST PASSIVATE PROBLEM SITE")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "standby", "site_2"],
            expected_exit_code=1)

    def test_activate_unvaliable_site(self):
        logging.info("TEST ACTIVATE PROBLEM SITE")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "active", "site_2"],
            expected_exit_code=1)

    def test_activate_working_service_unvaliable_site(self):
        logging.info("TEST ACTIVATE WORKING SERVICE ON PROBLEM SITE")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'),
             "--run-services", "serviceA", "active", "site_2"])

    def test_passivate_working_service_unvaliable_site(self):
        logging.info("TEST PASSIVATE WORKING SERVICE ON PROBLEM SITE")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'),
             "--run-services", "serviceA", "standby",
             "site_2"])
             
    def test_stop_working_site(self):
        logging.info("TEST STOP WORKING SITE")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "stop", "site_1"],
            expected_exit_code=1)

    def test_stop_working_service_working_site(self):
        logging.info("TEST STOP WORKING SERVICE ON WORKING SITE")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'),
             "--run-services", "serviceA", "stop", "site_1"])

    def test_move_working_site(self):
        logging.info("TEST MOVE WORKING SITE")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "move", "site_1"],
            expected_exit_code=1)

    def test_move_working_service_working_site(self):
        logging.info("TEST MOVE WORKING SERVICE ON WORKING SITE")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'),
             "--run-services", "serviceA", "move", "site_1"])

    def test_stop_unvaliable_site(self):
        logging.info("TEST STOP PROBLEM SITE")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "stop", "site_2"])

    def test_move_unvaliable_site(self):
        logging.info("TEST MOVE PROBLEM SITE")
        # Run passivate
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "move", "site_2"],
            expected_exit_code=1)

    def test_move_working_service_unvaliable_site(self):
        logging.info("TEST MOVE WORKING SERVICE ON PROBLEM SITE")
        # Run passivate
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'),
             "--run-services", "serviceA", "move", "site_2"])

    def test_mntc_working_site(self):
        logging.info("TEST MAINTENANCE WORKING SITE")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "disable", "site_1"])

    def test_return_working_site(self):
        logging.info("TEST RETURN WORKING SITE")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "return", "site_1"])

    def test_mntc_unvaliable_site(self):
        logging.info("TEST MAINTENANCE PROBLEM SITE")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "disable", "site_2"],
            expected_exit_code=1)

    def test_return_unvaliable_site(self):
        logging.info("TEST RETURN PROBLEM SITE")
        # Run passivate
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'), "return", "site_2"],
            expected_exit_code=1)

    def test_mntc_unvaliable_site(self):
        logging.info("TEST MAINTENANCE PROBLEM SITE")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'),
             "--run-services", "serviceA", "disable", "site_2"])

    def test_return_unvaliable_site(self):
        logging.info("TEST RETURN PROBLEM SITE")
        # Run passivate
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'),
             "--run-services", "serviceA", "return", "site_2"])
