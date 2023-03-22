# consts
default_module = 'stateful'
readonly_cmd = ("status", "list")
dr_procedures = ("move", "stop") + readonly_cmd  # DR procedures: switchover, failover
site_cmds = ("active", "standby", "return", "disable") + readonly_cmd  # per site commands

# Parameters from config
sm_conf: dict
module_flow = [{'stateful': None}]  # custom DR sequence per module
state_restrictions = {}
FRONT_HTTP_AUTH = False
SERVICE_DEFAULT_TIMEOUT = None

# Parameters from command line
done_services = []
ignored_services = []
failed_services = []
warned_services = []
skipped_due_deps_services = []
not_finished_due_deps_services = []
run_services = []
skip_services = []
force = False
