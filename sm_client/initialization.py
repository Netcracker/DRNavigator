import logging
import os
import pathlib
import sys

import yaml

from common import utils
from sm_client.data.structures import *
from sm_client.processing import sm_process_service


def init_and_check_config(args) -> bool:
    """ Main entry point. Provides validations, config parsing and initialization
    @returns: True of False  """
    #@todo python version check; version print

    # Set verbosity for logging
    logger = logging.getLogger()
    logger.setLevel(logging.DEBUG)

    ch = logging.StreamHandler()
    if args.verbose:
        ch.setLevel(logging.DEBUG)
        ch.setFormatter(logging.Formatter("%(asctime)s [%(levelname)s] %(filename)s.%(funcName)s(%(lineno)d): %(message)s"))
    else:
        ch.setLevel(logging.INFO)
        ch.setFormatter(logging.Formatter("%(asctime)s [%(levelname)s] %(filename)s: %(message)s"))

    logging.basicConfig(handlers=[ch])
    if args.output:
        log_output = None
        if pathlib.Path(args.output).is_file():
            log_output = pathlib.Path(args.output).expanduser() if os.access(args.output, os.W_OK) else None
        elif os.access(pathlib.Path(args.output).expanduser().parent.resolve(), os.W_OK) and \
                not pathlib.Path(args.output).expanduser().is_dir(): # do not take into account provided dirs
            log_output = pathlib.Path(args.output).expanduser()

        if not log_output:
            logging.critical(f"Cannot write to {args.output} file. Printing stdout ...")
        else:
            fh = logging.FileHandler(log_output)
            fh.setLevel(logging.DEBUG)
            fh.setFormatter(logging.Formatter("%(asctime)s [%(levelname)s] %(filename)s.%(funcName)s(%(lineno)d): %(message)s"))
            fh.emit(logging.LogRecord("",logging.INFO,sys.argv[0], 0, args.command,None,None)) # add delimeter in logfile

            logger.addHandler(fh)

    logging.debug(f"Script arguments: {args}")

    # Define, check and load configuration file
    conf_file = os.path.join(os.path.dirname(os.path.abspath(__file__)), "../config.yaml") if args.config == "" else args.config
    if not os.path.isfile(conf_file):
        logging.fatal("You should define configuration file for site-manager or copy it to config.yaml in site-manager main directory")
        exit(1)

    try:
        conf_parsed = yaml.load(open(conf_file), Loader=yaml.FullLoader)
    except:
        logging.fatal("Can not parse configuration file!")
        return False

    logging.debug(f"Parsed config: {conf_parsed}")

    settings.FRONT_HTTP_AUTH = conf_parsed.get("sm-client", {}).get("http_auth", False)
    settings.SERVICE_DEFAULT_TIMEOUT = conf_parsed.get("sm-client", {}).get("service_default_timeout", 200)

    utils.SM_GET_REQUEST_TIMEOUT = conf_parsed.get("sm-client", {}).get("get_request_timeout", 10)
    utils.SM_POST_REQUEST_TIMEOUT = conf_parsed.get("sm-client", {}).get("post_request_timeout", 30)

    settings.ignored_services.clear()

    # Check services for running
    if args.run_services != '':
        settings.run_services = args.run_services.replace(',', ' ').replace('  ', ' ').split(' ')
    else:
        settings.run_services.clear()

    if args.skip_services != '':
        settings.skip_services = args.skip_services.replace(',', ' ').replace('  ', ' ').split(' ')
    else:
        settings.skip_services.clear()

    settings.failed_services.clear()
    settings.done_services.clear()
    settings.skipped_due_deps_services.clear()

    settings.force = args.force

    settings.state_restrictions = conf_parsed.get("restrictions", {}) if not args.ignore_restrictions else {}

    settings.module_flow = conf_parsed.get("flow", [{'stateful': None}])
    site_names = [i["name"] for i in conf_parsed["sites"]]
    settings.sm_conf = SMConf()
    for site in site_names:
        try:
            site_url = [ i["site-manager"] for i in conf_parsed["sites"] if i ["name"] == site ][0]
        except KeyError:
            logging.error("Check configuration file. Some of sites does not have 'site-manager' parameter")
            return False

        for i in conf_parsed["sites"]:
            if i["name"] != site:
                continue
            if isinstance(i.get("token", ""), dict):
                if "from_env" not in i.get("token", {}):
                    logging.error(f"Wrong token configuration for site {i['name']}: "
                                  f"use string value or specify from_env parameter")
                    return False
                site_token = os.environ.get(i["token"]["from_env"])
                if site_token is None:
                    logging.error(f"Wrong token configuration for site {i['name']}: "
                                  f"specified env {i['token']['from_env']} doesn't exist")
                    return False
            else:
                site_token = i.get("token", "")
        site_cacert = [ i.get("cacert", True) for i in conf_parsed["sites"] if i ["name"] == site ][0]

        if site_cacert != True and not os.path.isfile(site_cacert):
            logging.fatal(f"You should define correct path to CA certificate for site {site}")
            return False
        settings.sm_conf[site] = {}
        settings.sm_conf[site]["url"] = site_url
        settings.sm_conf[site]["token"] = site_token
        settings.sm_conf[site]["cacert"] = False if args.insecure else site_cacert

    # Check state restrictions
    for restrictions_list in settings.state_restrictions.values():
        if any(state_str.count('-') + 1 != len(settings.sm_conf) for state_str in restrictions_list):
            logging.error(f"Check configuration file. Some state restrictions don't suitable for the current number of sites")
            return False

    return True


def sm_get_cluster_state(site=None) -> SMClusterState:
    """ Get cluster status or per specific site and init sm_dict object
    """
    sm_dict = SMClusterState(site)
    for site_name in sm_dict.keys():
        response, ret, code = sm_process_service(site_name, "site-manager", "status")
        sm_dict[site_name]["return_code"] = code # HTTP or SSL
        sm_dict[site_name]["status"] = ret
        sm_dict[site_name].update(response)
    return sm_dict
