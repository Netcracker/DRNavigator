"""Module, that contains differents classes and structures, that are used in other modules"""
from graphlib import TopologicalSorter


class SMConf(dict):  # global config.yaml and some RO config parameters and args, cmd manipulation
    """config.yaml content"""
    def get_active_site(self, cmd, site):
        """Returns active site after procedure processing"""
        if cmd in ["active", "move"]:
            return site
        if cmd in ["stop", "standby", 'disable']:
            return self.get_opposite_site(site)
        return None

    def get_opposite_site(self, site):
        """Returns opposite site"""
        if site not in self.keys():
            return None
        opposite_site = None

        for s in self.keys():
            if s != site:
                opposite_site = s
        return opposite_site

    @staticmethod
    def convert_sitecmd_to_dr_mode(site_cmd):
        """ Converts <site_cmd>  into DR mode ["active","standby","disable"]
        """
        if site_cmd == "return":
            return "standby"
        # "active","standby" are modes.
        return site_cmd

    @staticmethod
    def get_modules():
        """Returns module list"""
        from sm_client.data import settings

        mod_list = set()
        for elem in settings.module_flow:
            for module in elem.keys():
                mod_list.add(module)
        return list(mod_list)


class SMClusterState:
    """Cluster content"""
    def __init__(self, site=None):
        from sm_client.data import settings

        def init_default(site_name):
            if len(self.sm.keys()) == 2:
                raise ValueError("Only two sites in clusters are supported")
            self.sm[site_name] = {}
            self.sm[site_name]["services"] = {}
            self.sm[site_name]["return_code"] = None
            self.sm[site_name]["status"] = False  # ServiceDRStatus

        if not site or (site and isinstance(site, str)):  # @todo rework
            if site and site not in settings.sm_conf.keys():
                raise ValueError("Unknown site name")
            self.sm = {}
            for cur_site in [s for s in settings.sm_conf.keys() if site == s] if site else settings.sm_conf.keys():
                init_default(cur_site)
        elif isinstance(site, dict):  # for dev/testing purposes
            self.sm = {}
            for key in site.keys():
                init_default(key)
                for k, v in site[key].items():
                    self.sm[key][k] = v

        # Init global service order and global ts
        self.globals = {}
        for module in settings.sm_conf.get_modules():
            self.globals[module] = {}
            self.globals[module]["deps_issue"] = None
            self.globals[module]["service_dep_ordered"] = []
            self.globals[module]["ts"] = None

    def __getitem__(self, key):
        return self.sm[key]

    def __setitem__(self, key, item):
        self.sm[key] = item

    def __str__(self):
        return str(self.sm)

    def keys(self):
        """Dict keys() function"""
        return self.sm.keys()

    def items(self):
        """Dict items() function"""
        return self.sm.items()

    def get_dr_operation_sequence(self, serv, procedure, site) -> list:  # move active, stop standby
        """ Get DR operation(cmd) site sequence in the correct order for specific service for provided DR procedure
        @returns: [['site1','standby'],['site2','active']] - default in case sequence is empty
        @todo to rework when sm_dict[site|opposite]['services'][serv] serv is not present on one site
        """
        from sm_client.data import settings
        opposite_site = settings.sm_conf.get_opposite_site(site)
        site_sequence = []
        if procedure == 'move':  # switchover
            mode, = self.sm[site]['services'][serv]['sequence'][0:1] or ['standby']
            if mode == 'standby':
                site_sequence = [[opposite_site, 'standby'], [site, 'active']]
            elif mode == 'active':
                site_sequence = [[site, 'active'], [opposite_site, 'standby']]
        elif procedure == 'stop':  # failover
            site_to_check = opposite_site if serv in self.sm[opposite_site]['services'] else site
            mode, = self.sm[site_to_check]['services'][serv]['sequence'][0:1] or ['standby']
            if mode == 'standby':
                site_sequence = [[site, 'standby'], [opposite_site, 'active']]
            elif mode == 'active':
                site_sequence = [[opposite_site, 'active'], [site, 'standby']]
        else:
            raise Exception("Wrong command")

        return site_sequence

    def get_available_sites(self) -> list:
        """ Return list of available sites """
        return [site for site in self.sm.keys() if self.sm[site]['status']]

    def get_services_list_for_ok_site(self) -> list:
        """Returns list of services for available sites"""
        final_set: set = set()
        for site in self.sm.values():
            final_set = final_set.union(set(list(site['services'].keys())))
        return list(final_set)

    def get_module_services(self, site, module) -> list:
        """Get services list for specified module on site"""
        module_list = []
        for serv in self.sm[site]['services'].keys():
            if self.sm[site]['services'][serv].get('module') and module == self.sm[site]['services'][serv]['module']:
                module_list.append(serv)
        return module_list

    def make_ignored_services(self, service_dep_ordered: list) -> list:
        """ Make list of services which are not intended to run, ignored."""
        from sm_client.data import settings
        ignored_list = []
        for site in self.sm.keys():
            for serv in self.sm[site]['services']:
                if serv not in service_dep_ordered and serv not in settings.ignored_services:
                    ignored_list.append(serv)
        return list(set(ignored_list))


class ServiceDRStatus:
    """Service status"""
    def __getitem__(self, key):
        return self.__getattribute__(key)

    def __init__(self, data: dict = None, smdict = None, site: str = None,
                 mode: str = None, force = False, allow_failure = False):  # {'services':{service_name:{}}}
        if data and data.get("services") and isinstance(data['services'], dict):
            self.service = list(data['services'].keys())[0]
        elif data and data.get("wrong-service") and isinstance(data['wrong-service'], str):
            self.service = data['wrong-service']
        else:
            raise ValueError("Missing service name")
        serv = data['services'][self.service] if data.get('services') else data
        self.mode = serv['mode'] if serv.get("mode") in ["active", "standby", "disable"] else "--"
        self.nowait = serv["nowait"] if serv.get("nowait") else False
        self.healthz = serv["healthz"] if serv.get("healthz") in ["up", "down", "degraded"] else "--"
        self.status = serv["status"] if serv.get("status") in ["running", "done", "failed", "queue"] else "--"
        self.message = serv["message"] if serv.get("message") else ""

        # https://github.com/Netcracker/DRNavigator/blob/b4161fb15271485974abf5862e7272abc386fbc8/modules/stateful.py#L16

        def set_service_status(smdict, site, mode, force):
            if self.message == "Service doesn't exist":
                return True
            failed_healthz = ['down', 'degraded', '--']

            if mode and mode in 'standby' and smdict[site]['services'][self.service].get('allowedStandbyStateList'):
                failed_healthz = set(failed_healthz) - set(
                    smdict[site]['services'][self.service].get('allowedStandbyStateList'))

            return (self.healthz not in failed_healthz or force) and self.status not in ['failed']

        # separate service status field, since healthz may be treated differently depending on running mode - allowedStandbyStateList
        self.service_status = set_service_status(smdict, site, mode, force)

        self.allow_failure = allow_failure  # Will be changed to True for failover on standby site

    def is_ok(self):
        """Returns if service status is ok"""
        return self.service_status or self.allow_failure

    def sortout_service_results(self):
        """ Put service name in appropriate list(failed or done) """
        from sm_client.data import settings
        if self.service_status:  # return Ok - done_service
            if self.service not in settings.failed_services and self.service not in settings.warned_services \
                    and self.service not in settings.done_services:
                settings.done_services.append(self.service)
        elif self.allow_failure:
            if self.service not in settings.failed_services:
                if self.service not in settings.warned_services:
                    settings.warned_services.append(self.service)
                if self.service in settings.done_services:
                    settings.done_services.remove(self.service)
        else:
            if self.service not in settings.failed_services:
                settings.failed_services.append(self.service)
            if self.service in settings.warned_services:
                settings.warned_services.remove(self.service)
            if self.service in settings.done_services:
                settings.done_services.remove(self.service)


class NotValid(Exception):
    """ Raised when it is not possible to process specified command on current cluster state"""


class TopologicalSorter2(TopologicalSorter):
    """ added method to get successors of specific node """

    def successors(self, node):
        """Get node successors"""
        for i in self._node2info.values():
            if i.node == node:
                return i.successors
        return None
