import os
import pytest
import logging
import shutil

from jinja2 import Template


@pytest.fixture(scope='class', name='config_dir')
def prepare_configs(request, tmpdir_factory):
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
    os.system(f"docker-compose -f {os.path.join(config_dir, 'docker-compose.yaml')} up --detach --wait")

    # Run tests
    yield

    # Collect logs from docker-compose
    os.system(f"docker-compose -f {os.path.join(config_dir, 'docker-compose.yaml')} logs > "
              f"{os.path.join(config_dir, 'docker_logs.log')}")

    # Docker-compose down
    logging.info("Docker compose down")
    os.system(f"docker-compose -f {os.path.join(config_dir, 'docker-compose.yaml')} down")
