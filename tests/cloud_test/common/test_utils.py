
import pytest
import smclient


def run_sm_client_command_with_exit(args, expected_exit_code=0):
    with pytest.raises(SystemExit) as pytest_wrapped_e:
        smclient.main(args)
    assert pytest_wrapped_e.type == SystemExit
    assert pytest_wrapped_e.value.code == expected_exit_code

