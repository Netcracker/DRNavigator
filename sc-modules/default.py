import logging
import threading
import time
from sm_client import run


def seq(sm_dict, procedure, services_to_run, all_services, running_services, done_services, failed_services, ignored_services, after_stateful, force, no_wait):

    while not all(
            elem in (done_services + failed_services + ignored_services + after_stateful) for elem in all_services):

        for service_name in all_services:
            if service_name not in (done_services + running_services + failed_services + ignored_services):
                after_services = sm_dict["services"][service_name]["after"]
                if "stateful" in after_services:
                    logging.info(f"Founded service {service_name} which should be executed after stateful services")
                    after_stateful.append(service_name)
                    continue

                # Set service as failed when any of dependencies is failed
                if any(elem in failed_services for elem in after_services) and service_name not in failed_services:
                    logging.error(f"Service {service_name} marked as failed due to dependencies")
                    failed_services.append(service_name)
                    continue

                if service_name not in services_to_run:
                    ignored_services.append(service_name)

                else:
                    # Run service if it is not in running, failed or done lists
                    if all(elem in (done_services + ignored_services) for elem in after_services) or after_services == [
                        '']:
                        thread = threading.Thread(target=run,
                                                  args=(service_name,
                                                        procedure,
                                                        force,
                                                        no_wait))
                        thread.name = f"Thread: {service_name}"
                        thread.start()

        if len(done_services) != 0:
            logging.debug('done_services = %s' % done_services)
        if len(ignored_services) != 0:
            logging.debug('ignored_services = %s' % ignored_services)
        if len(running_services) != 0:
            logging.debug('running_services = %s' % running_services)
        if len(failed_services) != 0:
            logging.debug('failed_services = %s' % failed_services)

        time.sleep(5)

    threads = list()
    for service in after_stateful:
        thread = threading.Thread(target=run,
                                  args=(service,
                                        procedure,
                                        force,
                                        no_wait))
        thread.name = f"Thread: {service}"
        threads.append(thread)
        thread.start()

    for thread in threads:
        thread.join()


