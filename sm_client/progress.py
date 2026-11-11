"""Per-service progress table, printed while a DR or per-site procedure runs"""
import logging
import threading
import time

from prettytable import PrettyTable  # type: ignore

from sm_client.data import settings

STATE_QUEUED = "queued"
STATE_RUNNING = "running"
STATE_DONE = "done"
STATE_FAILED = "failed"
STATE_WARNED = "warned"
STATE_SKIPPED = "skipped (dependency)"

REPEAT_INTERVAL = 60  # seconds without a new table before the table is printed again


class ProgressTracker:
    """Tracks which services have started and finished in the procedure.
    The final state (done/failed/warned/skipped) is taken from the result lists in settings."""

    def __init__(self):
        self._lock = threading.Lock()
        self.enabled = False
        self.cmd = ""
        self.services: list = []
        self.started: set = set()
        self.finished: set = set()
        self.message: dict = {}
        self.repeat_interval: float = REPEAT_INTERVAL
        self.last_printed = 0.0

    def reset(self):
        """Clear all tracked data and disable printing"""
        with self._lock:
            self.enabled = False
            self.cmd = ""
            self.services = []
            self.started = set()
            self.finished = set()
            self.message = {}
            self.last_printed = 0.0

    def start(self, cmd: str, services: list, enabled: bool = True):
        """Begin tracking a procedure for the given ordered services"""
        self.reset()
        with self._lock:
            self.enabled = enabled
            self.cmd = cmd
            self.services = list(dict.fromkeys(services))
            self.last_printed = time.time()

    def seconds_until_repeat(self):
        """Seconds left before the table should be printed again, None when printing is disabled"""
        if not self.enabled:
            return None
        return max(0.01, self.repeat_interval - (time.time() - self.last_printed))

    def service_started(self, service: str):
        """Mark that the service started a step (it can start again in the next module step)"""
        with self._lock:
            self.started.add(service)
            self.finished.discard(service)
            if service not in self.services:
                self.services.append(service)

    def service_finished(self, service: str, message: str = ""):
        """Mark that the service finished processing in the current module step"""
        with self._lock:
            self.finished.add(service)
            if message:
                self.message[service] = message
            if service not in self.services:
                self.services.append(service)

    def state(self, service: str) -> str:
        """Current state of the service"""
        if service in self.started and service not in self.finished:
            return STATE_RUNNING
        if service in settings.failed_services:
            return STATE_FAILED
        if service in settings.skipped_due_deps_services:
            return STATE_SKIPPED
        if service in settings.warned_services:
            return STATE_WARNED
        if service in settings.done_services:
            return STATE_DONE
        return STATE_QUEUED

    def render(self, title: str = "") -> str:
        """Build the progress table as text"""
        with self._lock:
            states = {service: self.state(service) for service in self.services}
            counts = {state: list(states.values()).count(state)
                      for state in (STATE_DONE, STATE_FAILED, STATE_WARNED, STATE_SKIPPED, STATE_RUNNING, STATE_QUEUED)}
            finished = len(self.services) - counts[STATE_RUNNING] - counts[STATE_QUEUED]
            lines = [f"{title or 'Progress'}: procedure {self.cmd}, "
                     f"{finished}/{len(self.services)} services finished "
                     f"(done {counts[STATE_DONE]}, failed {counts[STATE_FAILED]}, warned {counts[STATE_WARNED]}, "
                     f"skipped {counts[STATE_SKIPPED]}, running {counts[STATE_RUNNING]}, "
                     f"queued {counts[STATE_QUEUED]})"]

            table = PrettyTable(["Service", "State"])
            table.align = {"Service": "l", "State": "l"}
            for service, state in states.items():
                table.add_row([service, state])
            lines.append(str(table))

            for service, state in states.items():
                if state in (STATE_FAILED, STATE_WARNED) and self.message.get(service):
                    lines.append(f"{service} {state}: {self.message[service]}")
            if counts[STATE_RUNNING]:
                lines.append(f"Next update when a service finishes, or in {int(self.repeat_interval)}s at the latest.")
            return "\n".join(lines)

    def print_table(self, title: str = ""):
        """Log the progress table if tracking is enabled"""
        if not self.enabled:
            return
        self.last_printed = time.time()
        logging.info("\n%s", self.render(title))


tracker = ProgressTracker()
