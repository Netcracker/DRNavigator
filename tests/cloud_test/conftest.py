import base64
import logging
import os
import shutil

import pytest
from jinja2 import Template
from kubernetes import client, config


def pytest_addoption(parser):
    parser.addoption("--kubeconfig", action="store", default="", help="path kubeconfig")


@pytest.fixture(scope='session', name='kubeconfig')
def prepare_kubeconfig(request):
    kubeconfig = {}

    kubeconfig_file = request.config.getoption("--kubeconfig")

    if kubeconfig_file != "" and os.path.isfile(kubeconfig_file):
        config.load_kube_config(config_file=kubeconfig_file)

        _, kubeconfig["current_context"] = config.list_kube_config_contexts(config_file=kubeconfig_file)
        kubeconfig["namespace_sm"] = kubeconfig["current_context"].get('namespace', 'site-manager')
        kubeconfig["namespace_services"] = kubeconfig["current_context"].get('namespace', 'test-services')
    else:
        logging.error("Kubeconfig empty or is not a file")
        os._exit(1)

    logging.info(f"Kubeconfig load")

    yield kubeconfig


@pytest.fixture(scope='session', name='sm_env')
def prepare_sm_env(kubeconfig):
    sm_env = {}

    try:
        token_sm = client.CoreV1Api().read_namespaced_secret('sm-auth-sa-token', kubeconfig['namespace_sm'])
        sm_env['token-sm'] = base64.b64decode(token_sm.data['token']).decode()
        sm_ingress = client.NetworkingV1Api().read_namespaced_ingress('site-manager', kubeconfig['namespace_sm'])
        sm_env["host_name"] = sm_ingress.spec.tls[0].hosts[0]
        secret_sm = client.CoreV1Api().read_namespaced_secret('sm-certs', kubeconfig['namespace_sm'])
        sm_env["ca_crt"] = base64.b64decode(secret_sm.data['ca.crt']).decode()

    except Exception as e:
        logging.error("Can not get sm-auth-sa token, sm-ingress-name and ca.crt: \n %s" % str(e))
        os._exit(1)

    logging.info(f"Token, sm-ingress-name and ca.crt collect")

    yield sm_env


@pytest.fixture(scope='session', name='config_ingress_service')
def info_services(kubeconfig):
    services_ingress = {}
    try:
        service_a_ingress = client.NetworkingV1Api().read_namespaced_ingress('service-a',
                                                                             kubeconfig['namespace_services'])
        services_ingress['service-a.test-services'] = service_a_ingress.spec.rules[0].host
        service_b_ingress = client.NetworkingV1Api().read_namespaced_ingress('service-b',
                                                                             kubeconfig['namespace_services'])
        services_ingress['service-b.test-services'] = service_b_ingress.spec.rules[0].host
    except Exception as e:
        logging.error("Can not get ingress services : \n %s" % str(e))
        os._exit(1)

    yield services_ingress


@pytest.fixture(scope='class', name='config_dir')
def prepare_configs(request, sm_env):
    config_test = {}

    # Create class directory
    try:
        config_test["tmp_dir"] = os.path.abspath(os.path.join("cloud-dump", request.cls.__name__))
        if os.path.isdir(config_test["tmp_dir"]):
            shutil.rmtree(config_test["tmp_dir"], ignore_errors=True)
        os.mkdir(config_test["tmp_dir"])
        logging.debug(f"Tmp directory: {config_test['tmp_dir']}")

        # Get directory with configs
        test_dir = getattr(request.module, "test_dir")
        config_dir = getattr(request.module, "config_dir")
        config_dir = os.path.abspath(test_dir + config_dir)
        ca_crt = open(os.path.abspath(config_test["tmp_dir"] + '/ca.crt'), 'w').write(sm_env['ca_crt'])

        config_test["template_env"] = {
            "sites": {
                "site_1": {
                    "link": sm_env['host_name'],
                    "token": sm_env['token-sm'],
                    "ca_cert": os.path.abspath(config_test["tmp_dir"] + '/ca.crt')
                },
            }
        }
        # Convert configuration to tmp dir
        for file_name in os.listdir(config_dir):
            abs_path_source = os.path.join(config_dir, file_name)
            abs_path_target = os.path.join(config_test["tmp_dir"], file_name.removesuffix('.j2'))

            if os.path.isdir(abs_path_source):
                shutil.copytree(abs_path_source, abs_path_target)
            elif not abs_path_source.endswith('.j2'):
                shutil.copyfile(abs_path_source, abs_path_target)
            else:
                open(abs_path_target, 'w').write(
                    Template(open(abs_path_source).read()).render(env=config_test["template_env"]))
    except Exception as e:
        logging.error("Can not get config_test : \n %s" % str(e))
        os._exit(1)

    yield config_test
