import pytest
import logging
import os
import test_utils
import yaml
test_dir = os.path.dirname(__file__)
config_dir = "/resources/test"
@pytest.mark.usefixtures('config_dir')
class SMTestCase:
    def test_status(self, config_dir):
        logging.info("TEST STATUS COMMAND")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(config_dir[0], 'sm_config.yaml'), "-v", "-k", "status"])
    def test_list_section(self, config_dir):
        logging.info("TEST LIST COMMAND")
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(config_dir[0], 'sm_config.yaml'), "-v", "-k", "list"])

    def test_activate_all_sites(self, config_dir):
        logging.info("TEST ACTIVATE ALL WITH STATEFUL SERVICES")
        # Run activate
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(config_dir[0], 'sm_config.yaml'), "-v", "-k", "active", "site_1"])

    def test_passivate_all_sites(self, config_dir):
        logging.info("TEST PASSIVATE ALL WITH STATEFUL SERVICES")
        # Run passivate
        test_utils.run_sm_client_command_with_exit(
            ["--config", os.path.join(config_dir[0], 'sm_config.yaml'), "-v", "-k", "standby", "site_1"])

