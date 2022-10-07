#!/usr/bin/env python3.9
"""
pytest based integration unit test which require live external environment
python3  -m pytest -o log_cli=true -s -v test_smclient_integration.py <-k  test_name*>

"""
import json
import pytest
from smclient import *
from http import HTTPStatus
import http.server
import ssl
import threading
import yaml


def pytest_namespace():
    return {'site_name': None}


def args_init(config=None):
    args = argparse.ArgumentParser
    args.verbose = True
    args.insecure = True
    args.config = config if config else "config_test.yaml"
    args.run_services = ""
    args.skip_services = ""
    init_and_check_config(args)
    return args

@pytest.mark.skipif(not os.path.exists("config.yaml"),
                    reason="Need to populate config.yaml https://github.com/Netcracker/DRNavigator/blob/main/documentation/Architecture.md#configuration-file")
def test_process_service__status_ok(caplog):
    """ SUCCESS basic general success case without SSL verification
    """
    caplog.set_level(logging.DEBUG)
    args_init("config.yaml")
    with open("config.yaml", 'r') as stream:
        try:
            parsed_yaml=yaml.safe_load(stream)
        except yaml.YAMLError as exc:
            print(exc)
    site=[i["name"] for i in parsed_yaml["sites"]][0]
    print(f"Using {site}")
    pytest.site_name=site
    json_body_s, ret, code=sm_process_service(site, "site-manager", "status")
    assert ret is True \
           and type(json_body_s) is dict \
           and json.loads('"' + str(json_body_s) + '"') \
           and json_body_s["services"], "Returned: dict, valid JSON, JSON contains valid response"


@pytest.mark.depends(on=['test_process_service__status_ok'])
def test_process_service__rw_ok(caplog):
    """ SUCCESS basic general success case without SSL verification
    """
    args_init("config.yaml")
    json_body_s, ret, code=sm_process_service(pytest.site_name if hasattr(pytest, "site_name") else "k8s-1",
                                           "cluster-replicator", "status")
    assert ret is True and \
           type(json_body_s) is dict and \
           json.loads('"' + str(json_body_s) + '"') and \
           code == HTTPStatus.OK, \
        "Returned: non empty dict, valid JSON, JSON contains valid response"