import pytest
import logging
import os
import test_utils
import yaml
test_dir = os.path.dirname(__file__)
config_dir = "/resources/test"
@pytest.mark.usefixtures('config_dir')
@pytest.mark.usefixtures('sm_ingress')
@pytest.mark.usefixtures('config_ingress_service')
class SMTestCase:
    def test_status(self, config_dir):
        logging.info("TEST STATUS COMMAND")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(config_dir[0], 'sm_config.yaml'), "-v", "status"])
    def test_list_section(self, config_dir):
        logging.info("TEST LIST COMMAND")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(config_dir[0], 'sm_config.yaml'), "-v", "list"])

    def test_activate_all_sites(self, config_dir, sm_ingress, config_ingress_service, capfd):
        logging.info("TEST ACTIVATE ALL WITH STATEFUL SERVICES")
        # Run activate
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(config_dir[0], 'sm_config.yaml'), "-v", "active", "site_1"])

        test_utils.check_statuses(capfd, config_dir, sm_ingress, config_ingress_service, lambda site, service:
                                {"healthz": "up", "mode": "active", "status": "done", "message": "I'm OK"})

    def test_passivate_all_sites(self, config_dir, sm_ingress, config_ingress_service, capfd):
        logging.info("TEST PASSIVATE ALL WITH STATEFUL SERVICES")
        # Run passivate
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(config_dir[0], 'sm_config.yaml'), "-v", "standby", "site_1"])

        test_utils.check_statuses(capfd, config_dir, sm_ingress, config_ingress_service, lambda site, service:
                                {"healthz": "up", "mode": "standby", "status": "done", "message": "I'm OK"})

    def test_maintenance_standby_site(self, config_dir, sm_ingress, config_ingress_service, capfd):
        logging.info("TEST MAINTENANCE ACTIVE SITE WITH STATEFUL SERVICES")
        # Run maintenance
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(config_dir[0], 'sm_config.yaml'), "-v", "disable", "site_1"])

        # Check status
        test_utils.check_statuses(capfd, config_dir, sm_ingress, config_ingress_service, lambda site, service:
                                {"healthz": "up", "mode": "disable" if "site_1" == site else "standby",
                                 "status": "done", "message": "I'm OK"})

    def test_return_first_site(self, config_dir, sm_ingress, config_ingress_service, capfd):
        logging.info("TEST RETURN FIRST SITE WITH STATEFUL SERVICES")
        # Run return
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(config_dir[0], 'sm_config.yaml'), "-v", "return", "site_1"])

        # Check status
        test_utils.check_statuses(capfd, config_dir, sm_ingress, config_ingress_service, lambda site, service:
                                {"healthz": "up", "mode": "standby" if "site_1" == site else "standby",
                                 "status": "done", "message": "I'm OK"})

