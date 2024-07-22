import logging
import os
import re

import pytest
import requests

import smclient


def run_sm_client_command_with_exit(args, expected_exit_code=0):
    with pytest.raises(SystemExit) as pytest_wrapped_e:
        smclient.main(args)
    assert pytest_wrapped_e.type == SystemExit
    assert pytest_wrapped_e.value.code == expected_exit_code


# TODO: temporary solution
# returns dict[site-name][service][status/mode/healthz/message]
def parse_status_table(capfd):
    out, _ = capfd.readouterr()
    # Print is needed, because readouterr() reset buffer and captured content doesn't appear in logs
    print(out)

    logging.info("Parse status table...")

    # Get site names
    lines = out.split("\n")
    site_names = re.split(r' *\| *', lines[1])[2:-1]
    result = {key: {} for key in site_names}
    logging.info(f"init value: {result}")
    service_name = ""
    statuses = {key: {} for key in site_names}

    # Get services
    for line in lines[5:-2]:
        row_content = re.split(r' *\| *', line)[1:-1]
        if row_content[0] != "":
            service_name = row_content[0]
        for site_name, status_row in zip(site_names, row_content[1:]):
            if service_name not in statuses[site_name]:
                statuses[site_name][service_name] = status_row
            elif status_row != "":
                statuses[site_name][service_name] += f" {status_row}"

    for site_name in statuses:
        for service_name in statuses[site_name]:
            status = re.split(r' / *', statuses[site_name][service_name])
            result[site_name][service_name] = dict(zip(["mode", "status", "healthz", "message"], status))

    logging.info(f"Parse status table results: {result}")
    return result


def check_status_from_service(site_name, service_name, service_url, expected_status, token=None):
    if token is None:
        headers = {}
    else:
        headers = {
            "Content-type": "application/json",
            "Accept": "application/json",
            "Authorization": f"Bearer {token}"
        }
    failed = False
    try:
        healthz_resp = requests.get(service_url + "/healthz", timeout=1).json()
    except requests.exceptions.RequestException as e:
        healthz_resp = {}
    try:
        status_resp = requests.get(service_url + "/sitemanager", headers=headers, timeout=1).json()
    except requests.exceptions.RequestException as e:
        status_resp = {}
        failed = True
    service_status = {"healthz": healthz_resp.get("status", "--"),
                      "mode": status_resp.get("mode", "--"),
                      "status": status_resp.get("status", "--"),
                      "message": status_resp.get("message", "")}
    logging.debug(f"Check status from service {service_name} in site {site_name}, received: {service_status}")
    assert service_status["healthz"] == expected_status["healthz"]
    assert service_status["mode"] == expected_status["mode"]
    assert service_status["status"] == expected_status["status"]
    if not failed:
        assert service_status["message"] == expected_status["message"]


def check_status_from_site_manager(site_name, service_name, sm_url, token, verify, expected_answer):
    headers = {
        "Content-type": "application/json",
        "Accept": "application/json",
        "Authorization": f"Bearer {token}"
    }
    data = {"procedure": "status", "run-service": service_name}
    try:
        service_answer = requests.post(sm_url + "/sitemanager", json=data, headers=headers, verify=verify,
                                       timeout=60).json()
    except requests.exceptions.RequestException as e:
        service_answer = {"services": {service_name: {"healthz": "--", "mode": "--", "status": "--", "message": ""}}}
    logging.debug(f"Check status from service {service_name} in site {site_name}, received: {service_answer}")
    if "wrong-service" in expected_answer:
        assert service_answer["wrong-service"] == expected_answer["wrong-service"]
    else:
        assert service_answer == expected_answer

def check_status_from_sm_client(status_dict, site_name, service_name, expected_status):
    assert status_dict[site_name][service_name] == expected_status


def get_services_to_check(template_env):
    services_to_check = []
    for site, config in template_env["sites"].items():
        services_to_check += [service for service in config["exposed_ports"]["service"]
                              if service not in services_to_check]
    return services_to_check


def check_statuses(capfd, template_env, expected_status_func: lambda site_name, service_name: dict):
    services_to_check = get_services_to_check(template_env)
    run_sm_client_command_with_exit(["--config", os.path.join(template_env['config_dir'], 'sm-client-config.yaml'),
                                     "-v", "status"])
    sm_client_statuses_dict = parse_status_table(capfd)
    for site, config in template_env["sites"].items():
        for service in services_to_check:
            port = config["exposed_ports"]["service"].get(service)
            expected_status = expected_status_func(site, service)
            if port:
                check_status_from_service(site, service, f"http://localhost:{port}", expected_status)
            expected_answer_from_sm = {"services": {service: expected_status}} \
                if expected_status["mode"] != "--" else {"wrong-service": service}
            check_status_from_site_manager(site, service,
                                           sm_url=f"http://localhost:{config['exposed_ports']['site_manager']}",
                                           token=config["token"],
                                           verify=False,
                                           expected_answer=expected_answer_from_sm)
            check_status_from_sm_client(sm_client_statuses_dict, site, service, expected_status)


def check_statuses_for_cloud_test(capfd, config_dir, sm_env, config_ingress_service,
                                  expected_status_func: lambda site_name, service_name: dict):
    run_sm_client_command_with_exit(["--config", os.path.join(config_dir['tmp_dir'], 'sm_config.yaml'),
                                     "-v", "status"])
    sm_client_statuses_dict = parse_status_table(capfd)
    for service, ingress in config_ingress_service.items():
        expected_status = expected_status_func("site_1", service)
        check_status_from_service("site_1", service, f"http://{ingress}", expected_status,
                                  sm_env['site-manager-sa-token'])
        expected_answer_from_sm = {"services": {service: expected_status}} \
            if expected_status.get("message") != "Service doesn't exist" else \
            {"message": "Service doesn't exist", "wrong-service": service}
        check_status_from_site_manager("site_1", service,
                                       sm_url=f"https://{sm_env['host_name']}",
                                       token=sm_env['sm-auth-sa-token'],
                                       verify=False,
                                       expected_answer=expected_answer_from_sm)
        check_status_from_sm_client(sm_client_statuses_dict, "site_1", service, expected_status)
