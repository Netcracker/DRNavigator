import argparse
import os

test_config_path = os.path.abspath("tests/selftest/sm_client/resources/config_test.yaml")
test_wrong_config_path = os.path.abspath("tests/selftest/sm_client/resources/config_test_wrong.yaml")
test_restrictions_config_path = os.path.abspath("tests/selftest/sm_client/resources/config_test_with_restrictions.yaml")
test_config_env_token_path = os.path.abspath("tests/selftest/sm_client/resources/config_test_with_env_token.yaml")
test_config_wrong_env_token_path = os.path.abspath("tests/selftest/sm_client/resources/config_test_wrong_with_env_token.yaml")
config_path_wrong_states = os.path.abspath("tests/selftest/sm_client/resources/config_test_wrong_states.yaml")
config_path_correct_states = os.path.abspath("tests/selftest/sm_client/resources/config_test_correct_states.yaml")
config_path_states_without_brackets = os.path.abspath("tests/selftest/sm_client/resources/config_test_states_without_brackets.yaml")


def pytest_namespace():
    return {'site_name': None}


def args_init(config=None):
    args = argparse.ArgumentParser
    args.verbose = True
    args.insecure = True
    args.config = config if config else test_config_path
    args.run_services = ""
    args.skip_services = ""
    args.output = None
    args.force = False
    args.command = "version"
    args.ignore_restrictions = False
    args.site = None
    return args
