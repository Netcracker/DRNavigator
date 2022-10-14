# Site-manager integration tests

### How to run:

1. Install requirements for sm-client and tests:
```
pip install -r requirements-sc-test.txt -r requirements-sc.txt
```

2. Run tests:

```
python -u -m pytest  ./tests/integration
```
During this running site-manager and sm-dummy images will be rebuilt. If you want to skip this action, add special flag `--skip-build`:
```
python -u -m pytest  ./tests/integration --skip-build
```
You also can run only specific test case from [test folder](./integration/tests). For that run following command:
```
python -u -m pytest  ./tests/integration -k <class name>
```

3. After tests finishing you can see logs/configurations for every test in `test-dump` folder.

### How to create your own test case:

1. Go to [resources](./integration/tests/resources) and check available configuration for cluster emulation. Every of them contains some .j2 files for jinja parsing: docker-compose, sm/sm-client configurations, secrets, etc. You can use one of them or create your own in separate folder.
2. Create new python script in [test folder](./integration/tests).
3. Specify following parameters in this script:  
    3.1. `test_dir` is absolute path to this script;  
    3.2. `docker_config_dir` is relative path to used resource directory (chosen in step 1);  
    3.3. `template_env` is environment object for jinja parsing. This object should contain exposed ports for all containers from docker-compose, and tokens for SMs on different sites. You can also add your own environments, but don't forget to use it in templates;  
    See [examples](./integration/tests/read_statuses_test.py#L11) for additional information;
4. Create test class with `*TestCase` name and add for it some annotations:
   ```python
   @pytest.mark.usefixtures('config_dir')
   @pytest.mark.usefixtures('prepare_docker_compose')
   class SomeClassTastCase:
        <Test case functions>
   ```
   `config_dir` annotation is needed for creating new directory for this class in `test-dump` and for jinja parsing of resource directory files there.
   `prepare_docker_compose` annotation is needed for docker-compose up/down during running tests from this test case.
5. Create tests functions inside this class. For that you can use userful [test utils](integration/common) or write your own runs and checks.
