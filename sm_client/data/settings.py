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

# Result filter
done_services = []              # Services, that were successfully done
ignored_services = []           # Services, that were ignored because of --run-services or --skip-services option
failed_services = []            # Services, that were failed and broke the procedure
warned_services = []            # Services, that were failed, but didn't break the procedure (e.g. failed standby part for failover)
skipped_due_deps_services = []  # Services, that were skipped, because they depend on failed ones or not finished because of previous flow step failed

# Parameters from command line
run_services = []
skip_services = []
force = False
