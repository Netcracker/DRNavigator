import logging
import warnings

import pytest

from sm_client.data.structures import *
from sm_client.initialization import init_and_check_config
from tests.selftest.sm_client.common.test_utils import *


def test_SMClusterState_init():
    init_and_check_config(args_init())

    assert SMClusterState()
    assert SMClusterState("k8s-1")
    with pytest.raises(ValueError) as e:
        SMClusterState("not valid site")
    assert str(e.value) in "Unknown site name"
    assert "services" and "deps_issue" and settings.default_module in SMClusterState("k8s-2")["k8s-2"]
    assert "services" and "deps_issue" and settings.default_module in \
           SMClusterState({"k8s-3": {"services": {"serv1": {}}, "status": False}, "k8s-1": {}})["k8s-3"]
    sm_dict = SMClusterState({'k8s-1':
        {"services": {
            "serv1": {"module": 'stateful'},
            "serv2": {"module": 'stateful'},
            "serv3": {"module": 'notstateful'}}}})

    assert sm_dict.get_module_services('k8s-1', 'stateful') == ['serv1', 'serv2'] and \
           sm_dict.get_module_services('k8s-1', 'notstateful') == ['serv3']

    init_and_check_config(args_init(test_wrong_config_path))
    with pytest.raises(ValueError) as e:
        SMClusterState()
    assert str(e.value) in "Only two sites in clusters are supported"


def test_unexist_config_file_init():
    init_and_check_config(args_init())

    with pytest.raises(SystemExit) as pytest_wrapped_e:
        init_and_check_config(args_init("config_test_fake.yaml"))
        assert pytest_wrapped_e.type == SystemExit
        assert pytest_wrapped_e.value.code == 1


def test_init_and_check_config(caplog):
    """ Test file logging capabilities, -o file.log """
    args = args_init()
    for args.output in ["output.log", "not_exist_file", "/tmp/not_exist_file", "~/not_exist_file"]:
        with caplog.at_level(logging.CRITICAL):
            init_and_check_config(args)
            assert f"Cannot write to {args.output}" not in caplog.text
        os.remove(os.path.expanduser(args.output))

    import pwd  # keep this import here for Windows compatibility

    wrong_log_path = ["/", "~/", "./"]
    if pwd.getpwuid(os.getuid())[0] == 'root':
        warnings.warn(UserWarning("You use root user, can't check some test cases"))
    else:
        wrong_log_path.extend(["/not_exist_file", "/etc/passwd"])

    for args.output in wrong_log_path:
        with caplog.at_level(logging.CRITICAL):
            init_and_check_config(args)
            assert f"Cannot write to {args.output}" in caplog.text


def test_token_env_configuration(monkeypatch, caplog):
    # Fail in wrong configuration
    args = args_init(config=test_config_wrong_env_token_path)
    with caplog.at_level(logging.ERROR):
        assert not init_and_check_config(args)
        assert f"Wrong token configuration for site k8s-2: use string value or specify from_env parameter" in caplog.text

    # Fail when env var unexist
    args = args_init(config=test_config_env_token_path)
    with caplog.at_level(logging.ERROR):
        assert not init_and_check_config(args)
        assert f"Wrong token configuration for site k8s-2: specified env SM_TEST_TOKEN doesn't exist" in caplog.text

    # Set needed env
    monkeypatch.setenv('SM_TEST_TOKEN', '12345')

    # Check configuration
    args = args_init(config=test_config_env_token_path)
    assert init_and_check_config(args)
    assert settings.sm_conf['k8s-2']['token'] == '12345'
