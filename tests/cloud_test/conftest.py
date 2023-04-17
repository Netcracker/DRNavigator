import os
import pytest
import logging
import shutil
import base64
from kubernetes import client, config
from jinja2 import Template

def pytest_addoption(parser):
    parser.addoption("--kubeconfig", action="store", default="", help="path kubeconfig")

@pytest.fixture(scope='class',name='kubeconfig')
def prepare_kubeconfig(request):
    KUBECONFIG_FILE = request.config.getoption("--kubeconfig")

    if KUBECONFIG_FILE != "":
        config.load_kube_config(config_file=KUBECONFIG_FILE)

        _, current_context = config.list_kube_config_contexts(config_file=KUBECONFIG_FILE)
        namespace = current_context['context'].get('namespace', 'site-manager')

    else:
        config.load_incluster_config()
        with open("/var/run/secrets/kubernetes.io/serviceaccount/namespace") as open_file:
            namespace = open_file.read()

    logging.info(f"Kubeconfig load")

    yield current_context, namespace
@pytest.fixture(scope='class', name='sm_env')
def prepare_sm_env(kubeconfig, api_watch=False):

    if not api_watch:

        try:
            service_account = client.CoreV1Api().read_namespaced_service_account("sm-auth-sa", kubeconfig[1])
            secret_name = [s for s in service_account.secrets if 'token' in s.name][0].name
            btoken = client.CoreV1Api().read_namespaced_secret(name=secret_name, namespace=kubeconfig[1]).data['token']
            token = base64.b64decode(btoken).decode()
            btoken = client.CoreV1Api().read_namespaced_secret(name=secret_name, namespace=kubeconfig[1]).data['ca.crt']
            ca_crt = base64.b64decode(btoken).decode()
        except Exception as e:
            logging.error("Can not get sm-auth-sa token and ca.crt: \n %s" % str(e))
            os._exit(1)

    logging.info(f"Token and ca.crt collect")

    yield token, ca_crt

@pytest.fixture(scope='class',name='sm_ingress')
def prepare_sm_ingress(kubeconfig, api_watch=False):

    if not api_watch:

        try:
            sm_ingress = client.NetworkingV1Api().read_namespaced_ingress('site-manager', kubeconfig[1])
            host_name = sm_ingress.spec.tls[0].hosts[0]
            secret_sm = client.CoreV1Api().read_namespaced_secret('sm-certs', kubeconfig[1])
            sm_ca_crt = base64.b64decode(secret_sm.data['ca.crt']).decode()
        except Exception as e:
            logging.error("Can not get ingress : \n %s" % str(e))
            os._exit(1)

    yield host_name, sm_ca_crt

@pytest.fixture(scope='class', name='config_dir')
def prepare_configs(request, sm_ingress, sm_env):

    # Create class directory
    tmp_dir = os.path.abspath(os.path.join("cloud-dump", request.cls.__name__))
    if os.path.isdir(tmp_dir):
        shutil.rmtree(tmp_dir, ignore_errors=True)
    os.mkdir(tmp_dir)
    logging.debug(f"Tmp directory: {tmp_dir}")

    # Get directory with configs
    test_dir = getattr(request.module, "test_dir")
    config_dir = getattr(request.module, "config_dir")
    config_dir = os.path.abspath(test_dir + config_dir)
    ca_crt = open(os.path.abspath(tmp_dir + '/ca.crt'), 'w').write(sm_env[1])
    sm_ca_crt = open(os.path.abspath(tmp_dir + '/sm_ca.crt'), 'w').write(sm_ingress[1])
    template_env = {
        "sites": {
            "site_1": {
                "link": sm_ingress[0],
                "token": sm_env[0],
                "ca_cert": os.path.abspath(tmp_dir + '/ca.crt')
            },
        }
    }
    # Convert configuration to tmp dir
    for file_name in os.listdir(config_dir):
        abs_path_source = os.path.join(config_dir, file_name)
        abs_path_target = os.path.join(tmp_dir, file_name.removesuffix('.j2'))

        if os.path.isdir(abs_path_source):
            shutil.copytree(abs_path_source, abs_path_target)
        elif not abs_path_source.endswith('.j2'):
            shutil.copyfile(abs_path_source, abs_path_target)
        else:
            open(abs_path_target, 'w').write(Template(open(abs_path_source).read()).render(env=template_env))

    yield tmp_dir, template_env




