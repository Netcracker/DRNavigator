# How to develop sm-client

<!-- TOC -->
* [How to develop sm-client](#how-to-develop-sm-client)
  * [sm-client local build](#sm-client-local-build)
  * [Site-manager local docker-compose](#site-manager-local-docker-compose)
    * [Limitations](#limitations)
    * [Files purposes](#files-purposes)
    * [Run steps](#run-steps)
  * [Sm-client selftests](#sm-client-selftests)
    * [Run steps](#run-steps-1)
    * [New test case creation](#new-test-case-creation)
  * [Site-manager integration tests](#site-manager-integration-tests)
    * [Run steps](#run-steps-2)
    * [New test case creation](#new-test-case-creation-1)
  * [Site-manager cloud tests](#site-manager-cloud-tests)
    * [Run steps](#run-steps-3)
    * [New test case creation](#new-test-case-creation-2)
<!-- TOC -->

## sm-client local build

To use sm-client locally, as a binary, you should have Python installed to your environment.

Build steps:
1. Download sm-client requirements:
   ```bash
   pip install -r requirements-sc.txt
   ```
2. Check sm-client:
   ```bash
   ./sm-client -h
   ```

## Site-manager local docker-compose

`./test/docker-compose` contains some files, to start docker-compose for testing site-manager or sm-client. 
Using docker-compose, developer can easily emulate DR cluster with different configuration.  
By default it contains 4 services:
* `stateful` services:
  * serviceA
  * serviceB (after serviceA)
  * serviceC
* `custom_module` services:
  * customService

And 2 sites:
* `site-1` (active by default);
* `site-2` (standby by default);

`site-manager` works in http mode with `12345` token on both sites;

### Limitations

You should have some free ports, to run dev docker-compose. Default ports:

| container     | site-1 port | site-2 port |
|---------------|-------------|-------------|
| serviceA      | 9001        | 9005        |
| serviceB      | 9002        | 9006        |
| serviceC      | 9003        | 9007        |
| customService | 9004        | 9008        |
| site-manager  | 9010        | 9011        |

If you have to change opened port for some container, you should fix following files:
* `ports` section for needed container `./tests/docker-compose/docker-compose.yaml`;
* site-manager url in `./tests/docker-compose/sm-client-config.yaml` for needed site-manager container, 
if you change its port;

### Files purposes
1. `./tests/docker-compose/docker-compose.yaml` is main docker-compose file, that contains containers configuration. 
You can find mode information about docker-compose config in [official documentation](https://docs.docker.com/compose/);
2. `./tests/docker-compose/site-manager-config.yaml` and `./tests/docker-compose/site-manager-config-2.yaml` is 
configuration files for site-managers on `site-1` and `site-2`. They collect information about services CRs 
(instead of getting it from kube-api) and test token.
3. `./tests/docker-compose/sm-client-config.yaml` is sm-client configuration;

### Run steps

1. Build site-manager and sm-dummy images from repository directory:
   ```bash
   docker build -t site-manager --no-cache site-manager
   docker build -t sm-dummy --no-cache tests/sm-dummy
   ```

2. Run docker-compose:
   ```bash
   docker compose -f ./tests/docker-compose/docker-compose.yaml up --detach
   ```

3. Download sm-client requirements:
   ```bash
   pip install -r requirements-sc.txt
   ```

4. Run `sm-client` with `./tests/docker-compose/sm-client-config.yaml` as sm-client config. For-example, `status`:
   ```bash
   ./sm-client -c ./tests/docker-compose/sm-client-config.yaml -v status
   ```

5. You can do requests directly to container from docker-compose, all of them are exposed their ports to `localhost`. 
For that you can do request on container port. For example for site-manager on `site-1`:
   ```bash
   curl -H "Authorization: Bearer 12345" localhost:9010/sitemanager
   ```

6. Open docker-compose logs to see site-manager or sm-dummy logs:
   ```bash
   docker compose -f ./tests/docker-compose/docker-compose.yaml logs
   ```

7. Stop docker-compose, if you want to rebuild site-manager or sm-dummy, or return it to initial state:
   ```bash
   docker compose -f ./tests/docker-compose/docker-compose.yaml down
   ```

## Sm-client selftests

### Run steps

1. Install requirements for sm-client and tests:
   ```
   pip install -r requirements-sc-test.txt -r requirements-sc.txt
   ```

2. Run tests:
   ```
   python -u -m pytest  ./tests/selftest
   ```

### New test case creation

1. Go to [test directory](../../tests/selftest) .
2. Create your own script file in this directory or use one of existed scripts;
3. Locate your tests functions in this script;

## Site-manager integration tests

### Run steps

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
   You also can run only specific test case from [test folder](../../tests/integration/tests). For that run following command:
   ```
   python -u -m pytest  ./tests/integration -k <class name>
   ```

3. After tests finishing you can see logs/configurations for every test in `test-dump` folder.

### New test case creation

1. Go to [resources](../../tests/integration/tests/resources) and check available configuration for cluster emulation. Every of them contains some .j2 files for jinja parsing: docker-compose, sm/sm-client configurations, secrets, etc. You can use one of them or create your own in separate folder.
2. Create new python script in [test folder](../../tests/integration/tests).
3. Specify following parameters in this script:  
    3.1. `test_dir` is absolute path to this script;  
    3.2. `docker_config_dir` is relative path to used resource directory (chosen in step 1);  
    3.3. `template_env` is environment object for jinja parsing. This object should contain exposed ports for all containers from docker-compose, and tokens for SMs on different sites. You can also add your own environments, but don't forget to use it in templates;  
    See [examples](../../tests/integration/tests/read_statuses_test.py) for additional information;
4. Create test class with `*TestCase` name and add for it some annotations:
   ```python
   @pytest.mark.usefixtures('config_dir')
   @pytest.mark.usefixtures('prepare_docker_compose')
   class SomeClassTestCase:
        <Test case functions>
   ```
   `config_dir` annotation is needed for creating new directory for this class in `test-dump` and for jinja parsing of resource directory files there.
   `prepare_docker_compose` annotation is needed for docker-compose up/down during running tests from this test case.
5. Create tests functions inside this class. For that you can use userful [test utils](../../tests/test_utils.py) or write your own runs and checks.

## Site-manager cloud tests

### Run steps

1. Install requirements for sm-client and tests:
    ```
    pip install -r requirements-sc-test.txt -r requirements-sc.txt
    ```
   
2. Install cluster  
    For the tests to work, you need to install a cluster, you can use the command below
    ```
    pip install kubemarine
    kubemarine install -c cluster.yaml 
    ```
    Example [cluster.yaml](https://github.com/Netcracker/DRNavigator/blob/main/ci/cluster.yaml)
    **Note**: ths cluster should include site-manager and sm-dummy services, installed as a plugin.

3. Run tests:
    ```
    python -u -m pytest  ./tests/cloud_test --kubeconfig="<Path kubeconfig>"
    ```
    You also can run only specific test case from [test folder](../../tests/cloud_test/tests). For that run following command:
    ```
    python -u -m pytest  ./tests/cloud_test --kubeconfig="<Path kubeconfig>" -k <class name>
    ```
   
4. After tests finishing you can see logs/configurations for every test in `cloud-dump` folder.

### New test case creation

1. Go to the [resources](../../tests/cloud_test/tests/resources) section and check the available configuration for the cluster. They contain some file .j2 for parsing jinja: sm/sm-client configurations, secrets. You can use it or create your own in a folder.
2. Create new python script in [test folder](../../tests/cloud_test/tests).
3. Specify following parameters in this script:  
    3.1. `test_dir` is absolute path to this script;   
    3.2. `config_dir` is environment object for jinja parsing. 
    See [examples](../../tests/cloud_test/tests/sm-client_test.py) for additional information;
4. Create test class with `*TestCase` name and add for it some annotations:
   ```python
    @pytest.mark.usefixtures('config_dir')
    @pytest.mark.usefixtures('sm_env')
    @pytest.mark.usefixtures('config_ingress_service')
   class SomeClassTestCase:
        <Test case functions>
   ```
   `config_dir` annotation is needed for creating new directory for this class in `cloud-dump` and for jinja parsing of resource directory files there.
   `sm_env` annotation is needed for creating sm environment
   `config_ingress_service` annotation is needed creating services ingress environment
5. Create tests functions inside this class. For that you can use userful [test utils](../../tests) or write your own runs and checks.
