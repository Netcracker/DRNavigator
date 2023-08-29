"""Module, that contains parameters for other modules"""
from sm_client.data.structures import SMConf

# consts
default_module = 'stateful'
readonly_cmd = ("status", "list")
dr_processing_cmd = ("move", "stop")
dr_procedures = dr_processing_cmd + readonly_cmd  # DR procedures: switchover, failover
site_processing_cmd = ("active", "standby", "return", "disable")
site_cmds = site_processing_cmd + readonly_cmd  # per site commands

# Parameters from config
sm_conf: SMConf
module_flow = [{'stateful': None}]  # custom DR sequence per module
state_restrictions: dict = {}
FRONT_HTTP_AUTH = False
SERVICE_DEFAULT_TIMEOUT = None

# Result filter
done_services: list = []              # Services, that were successfully done
ignored_services: list = []           # Services, that were ignored because of --run-services or --skip-services option
failed_services: list = []            # Services, that were failed and broke the procedure
warned_services: list = []            # Services, that were failed, but didn't break the procedure (e.g. failed standby part for failover)
skipped_due_deps_services: list = []  # Services, that were skipped, because they depend on failed ones or not finished because of previous flow step failed

# Parameters from command line
run_services: list = []
skip_services: list = []
force = False
