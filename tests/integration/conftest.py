import os
import time
import pytest
import logging
import shutil
import requests
import urllib3
from urllib3.exceptions import InsecureRequestWarning
from jinja2 import Template


def pytest_addoption(parser):
    parser.addoption("--skip-build", action="store_true", help="skip docker images build")


@pytest.fixture(scope='session', name='build_images')
def build_images(pytestconfig):
    if not pytestconfig.getoption("--skip-build"):
        logging.info("Build SM docker image")
        os.system(f"docker build --rm -f {os.path.abspath('site-manager-cr-controller/Dockerfile')} "
                  f"-t site-manager ./site-manager-cr-controller")

        logging.info("Build sm-dummy docker image")
        os.system(f"docker build --rm -f {os.path.abspath('tests/sm-dummy/Dockerfile')} -t sm-dummy .")


@pytest.fixture(scope='class', name='config_dir')
def prepare_configs(request, tmpdir_factory, build_images):
    # Create class directory
    tmp_dir = os.path.abspath(os.path.join("test-dump", request.cls.__name__))
    if os.path.isdir(tmp_dir):
        shutil.rmtree(tmp_dir, ignore_errors=True)
    os.mkdir(tmp_dir)
    logging.debug(f"Tmp directory: {tmp_dir}")

    # Get directory with configs
    test_dir = getattr(request.module, "test_dir")
    docker_config_dir = getattr(request.module, "docker_config_dir")
    config_dir = os.path.abspath(test_dir + docker_config_dir)

    # Get template env
    template_env = getattr(request.module, "template_env")
    template_env["config_dir"] = tmp_dir
    template_env["os_path_sep"] = os.sep

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

    # Run tests
    yield tmp_dir


@pytest.fixture(scope='class', name='prepare_docker_compose')
def prepare_docker_compose(config_dir):
    # Docker-compose up
    logging.info("Docker compose up")
    os.system(f"docker-compose -f {os.path.join(config_dir, 'docker-compose.yaml')} up --detach")

    # Run tests
    yield

    # Collect logs from docker-compose
    os.system(f"docker-compose -f {os.path.join(config_dir, 'docker-compose.yaml')} logs > "
              f"{os.path.join(config_dir, 'docker_logs.log')}")

    # Docker-compose down
    logging.info("Docker compose down")
    os.system(f"docker-compose -f {os.path.join(config_dir, 'docker-compose.yaml')} down")


@pytest.fixture(scope='class', name='wait_services_until_healthy')
def wait_services_until_healthy(request, prepare_docker_compose):
    # Get template env
    template_env = getattr(request.module, "template_env")

    attempt_count = 5
    sleep_seconds = 5
    all_services_started = True

    # Disable warnings about unsecure tls connection
    urllib3.disable_warnings(InsecureRequestWarning)

    # Check services health
    for attempt in range(attempt_count):
        logging.info(f"Check services for healthy, attempt {attempt + 1}...")
        all_services_started = True
        for site, config in template_env["sites"].items():
            for service, port in config["exposed_ports"]["service"].items():
                try:
                    status_code = requests.get(f"http://localhost:{port}/healthz").status_code
                except Exception as e:
                    status_code = 0
                if status_code not in [200, 204]:
                    logging.info(f"Service {service} on site {site} hasn't started yet")
                    all_services_started = False
            try:
                status_code = requests.get(f"http://localhost:{config['exposed_ports']['site_manager']}/health"
                                           ).status_code
            except Exception as e:
                status_code = 0
            if status_code not in [200, 204]:
                logging.info(f"Site-manager on site {site} hasn't started yet")
                all_services_started = False
        if all_services_started:
            break
        logging.info(f"Not all services started, sleep {sleep_seconds} seconds")
        time.sleep(sleep_seconds)

    if all_services_started:
        logging.info("All services are healthy")
    else:
        logging.error("Some services haven't started in expected time")

    # Run tests
    yield
