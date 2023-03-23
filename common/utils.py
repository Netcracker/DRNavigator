#!/usr/bin/env python3
import logging
import ssl
from typing import Tuple, Dict

import requests.packages
import time
import os
import yaml
import base64
from kubernetes import client, config, watch
from requests.adapters import HTTPAdapter, Retry
from urllib3.exceptions import InsecureRequestWarning


SM_GET_REQUEST_TIMEOUT = int(os.environ.get("SM_GET_REQUEST_TIMEOUT", 10))
SM_POST_REQUEST_TIMEOUT = int(os.environ.get("SM_POST_REQUEST_TIMEOUT", 30))


class ProcedureException(Exception):
    """
    Raised when something went wrong during procedure
    @param dict output: output that should be returned if problem appears
    """

    def __init__(self, output):
        super().__init__(output)
        self.output = output


def io_make_http_json_request(url="", token=None, verify=True, http_body:dict=None, retry=3, use_auth=True) -> Tuple[bool, Dict, int]:
    """ Sends GET/POST request to service
    @param string url: the URL to service operator
    @param token: Bearer token
    @param verify: Server side SSL verification
    @param retry: the number of retries
    @param http_body: the dictionary with procedure and list of services
    @returns: True/False, Dict with not empty json body in case Ok/{}, HTTP_CODE/
    IO SSL codes: ssl.SSLErrorNumber.SSL_ERROR_SSL/SSLErrorNumber.SSL_ERROR_EOF
    """
    if not os.getenv("DEBUG"):
        # Disable warnings about self-signed certificates from requests library
        requests.packages.urllib3.disable_warnings(InsecureRequestWarning)

    if token and use_auth:
        headers = {"Authorization": f"Bearer {token}"}
    else:
        headers = {}
    if not http_body:
        http_body = {}
    logging.debug(f"REST url: {url}")
    logging.debug(f"REST data: {http_body}")

    session = requests.Session()
    retries = Retry(total=retry)
    session.mount('https://', HTTPAdapter(max_retries=retries))
    session.mount('http://', HTTPAdapter(max_retries=retries))

    logging.getLogger("urllib3").setLevel(logging.CRITICAL)

    try:
        if any(http_body):
            resp = session.post(url, json=http_body, timeout=SM_POST_REQUEST_TIMEOUT, headers=headers, verify=verify)
        else:
            resp = session.get(url, timeout=SM_GET_REQUEST_TIMEOUT, headers=headers, verify=verify)
        logging.debug(f"REST response: {resp.json()}")
        return True, resp.json() if resp.json() else {}, resp.status_code # return ANY content with HTTP code

    except requests.exceptions.SSLError as e:
        logging.error("SSL certificate verify failed")
        #TODO in more accurate manner error handling manner
        if "SSLCertVerificationError" in str(e.args): ## SSL Verification fails ; SSLCertVerificationError(1, '[SSL: CERTIFICATE_VERIFY_FAILED] certificate verify failed: unable to get local issuer certificate (_ssl.c:1125)')
            return False, {}, ssl.SSLErrorNumber.SSL_ERROR_SSL.__int__() # - 1
        elif "SSLEOFError" in str(e.args): # SSL connect error, SSL resource is not accessible vi ha-proxy  ; SSLEOFError(8, 'EOF occurred in violation of protocol (_ssl.c:1091)')
            #TODO need a test for this case
            return False, {}, ssl.SSLErrorNumber.SSL_ERROR_EOF.__int__() # - 8
    except requests.exceptions.JSONDecodeError as e:
        logging.error("Wrong JSON data received %s", e)
    except requests.exceptions.RequestException as e:
        logging.error("General request error %s", e)
    except Exception as e:
        logging.error("General error %s",e)

    return False,{},False

