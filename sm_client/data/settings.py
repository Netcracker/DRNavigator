# consts

from sm_client.data.structures import SMConf

default_module = 'stateful'
readonly_cmd = ("status", "list")
dr_procedures = ("move", "stop") + readonly_cmd  # DR procedures: switchover, failover
site_cmds = ("active", "standby", "return", "disable") + readonly_cmd  # per site commands

# Parameters from config
sm_conf: SMConf
module_flow = [{'stateful': None}]  # custom DR sequence per module
state_restrictions: dict = {}
FRONT_HTTP_AUTH = False
SERVICE_DEFAULT_TIMEOUT = None

# Parameters from command line
done_services: list = []
ignored_services: list = []
failed_services: list = []
run_services: list = []
skip_services: list = []
force = False
