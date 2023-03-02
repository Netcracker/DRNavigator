from graphlib import TopologicalSorter

from sm_client.data import settings


class SMConf(dict):  # global config.yaml and some RO config parameters and args, cmd manipulation

    def get_active_site(self, cmd, site):
        if cmd in ["active", "move"]:
            return site
        elif cmd in ["stop", "standby", 'disable']:
            return self.get_opposite_site(site)
        else:
            return None

    def get_opposite_site(self, site):
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
        else:  # "active","standby" are modes.
            return site_cmd

    @staticmethod
    def get_modules():
        mod_list = set()
        for elem in settings.module_flow:
            for module in elem.keys():
                mod_list.add(module)
        return list(mod_list)


class SMClusterState:
    def __init__(self, site=None):
        def init_default(site_name, modules: [] = None):
            if len(self.sm.keys()) == 2:
                raise ValueError("Only two sites in clusters are supported")
            self.sm[site_name] = {}
            self.sm[site_name]["services"] = {}
            self.sm[site_name]["return_code"] = None
            self.sm[site_name]["status"] = False  # ServiceDRStatus
            if not modules:
                modules = [settings.default_module]
            for module in modules:
                self.sm[site_name][module] = {}
                self.sm[site_name][module]["deps_issue"] = None
                self.sm[site_name][module]["service_dep_ordered"] = []
                self.sm[site_name][module]["ts"] = None

        if not site or (site and isinstance(site, str)):  # @todo rework
            if site and site not in settings.sm_conf.keys():
                raise ValueError("Unknown site name")
            self.sm = {}
            for cur_site in [s for s in settings.sm_conf.keys() if site == s] if site else settings.sm_conf.keys():
                init_default(cur_site, settings.sm_conf.get_modules())
        elif isinstance(site, dict):  # for dev/testing purposes
            self.sm = {}
            for key in site.keys():
                init_default(key)
                for k, v in site[key].items():
                    self.sm[key][k] = v

    def __getitem__(self, key):
        return self.sm[key]

    def __setitem__(self, key, item):
        self.sm[key] = item

    def __str__(self):
        return str(self.sm)

    def keys(self):
        return self.sm.keys()

    def get_dr_operation_sequence(self, serv, procedure, site) -> [[], []]:  # move active, stop standby
        """ Get DR operation(cmd) site sequence in the correct order for specific service for provided DR procedure
        @returns: [['site1','standby'],['site2','active']] - default in case sequence is empty
        @todo to rework when sm_dict[site|opposite]['services'][serv] serv is not present on one site
        """
        opposite_site = settings.sm_conf.get_opposite_site(site)
        site_sequence = []
        if procedure == 'move':  # switchover
            mode, = self.sm[site]['services'][serv]['sequence'][0:1] or ['standby']
            if mode == 'standby':
                site_sequence = [[opposite_site, 'standby'], [site, 'active']]
            elif mode == 'active':
                site_sequence = [[site, 'active'], [opposite_site, 'standby']]
        elif procedure == 'stop':  # failover
            mode, = self.sm[opposite_site]['services'][serv]['sequence'][0:1] or ['standby']
            if mode == 'standby':
                site_sequence = [[site, 'standby'], [opposite_site, 'active']]
            elif mode == 'active':
                site_sequence = [[opposite_site, 'active'], [site, 'standby']]
        else:
            raise Exception("Wrong command")

        return site_sequence

    def get_available_sites(self) -> []:
        """ Return list of available sites """
        return [site for site in self.sm.keys() if self.sm[site]['status']]

    def get_services_list_for_ok_site(self) -> []:
        final_set = set()
        for site in self.sm.values():
            final_set = final_set.union(set(list(site['services'].keys())))
        return list(final_set)

    def get_module_services(self, site, module) -> []:
        module_list = []
        for serv in self.sm[site]['services'].keys():
            if self.sm[site]['services'][serv].get('module') and module == self.sm[site]['services'][serv]['module']:
                module_list.append(serv)
        return module_list

    def make_ignored_services(self, service_dep_ordered: list) -> []:
        """ Make list of services which are not intended to run, ignored."""
        ignored_list = []
        for site in self.sm.keys():
            for serv in self.sm[site]['services']:
                if serv not in service_dep_ordered and serv not in settings.ignored_services:
                    ignored_list.append(serv)
        return list(set(ignored_list))


class ServiceDRStatus:
    def __getitem__(self, key):
        return self.__getattribute__(key)

    def __init__(self, data: dict = None):  # {'services':{service_name:{}}}
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
        self.status = serv["status"] if serv.get("status") in ["running", "done", "failed"] else "--"
        self.message = serv["message"] if serv.get("message") else ""
        # https://github.com/Netcracker/DRNavigator/blob/b4161fb15271485974abf5862e7272abc386fbc8/modules/stateful.py#L16

    def is_ok(self):
        if self.healthz in ['down', 'degraded', '--'] or self.status in ['failed']:
            return False
        return True

    def sortout_service_results(self):
        """ Put service name in appropriate list(failed or done) """
        if self.is_ok():  # return Ok - done_service
            if self.service not in settings.failed_services:
                settings.done_services.append(self.service) if self.service not in settings.done_services else None
        else:
            settings.failed_services.append(self.service) if self.service not in settings.failed_services else None
            settings.done_services.remove(self.service) if self.service in settings.done_services else None


class NotValid(Exception):
    """ Raised when it is not possible to process specified command on current cluster state"""
    pass


class TopologicalSorter2(TopologicalSorter):
    """ added method to get successors of specific node """
    def successors(self, node):
        for i in self._node2info.values():
            if i.node == node:
                return i.successors
        return None
