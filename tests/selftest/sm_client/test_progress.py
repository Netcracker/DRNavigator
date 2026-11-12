import logging
import time

import pytest

import smclient
from sm_client.data import settings
from sm_client.data.structures import ServiceDRStatus, TopologicalSorter2
from sm_client.initialization import init_and_check_config
from sm_client.processing import process_ts_services, thread_result_queue
from sm_client.progress import ProgressTracker, tracker, \
    STATE_QUEUED, STATE_RUNNING, STATE_DONE, STATE_FAILED, STATE_WARNED, STATE_SKIPPED
from tests.selftest.sm_client.common.test_utils import args_init


@pytest.fixture(autouse=True)
def clean_results():
    """Progress state is derived from the global result lists, so reset them around every test"""
    lists = (settings.done_services, settings.failed_services, settings.warned_services,
             settings.skipped_due_deps_services)
    for result_list in lists:
        result_list.clear()
    tracker.reset()
    yield
    for result_list in lists:
        result_list.clear()
    tracker.reset()


def test_service_states():
    progress = ProgressTracker()
    progress.start("move", ["queued", "running", "done", "failed", "warned", "skipped"])

    for service in ("running", "done", "failed", "warned"):
        progress.service_started(service)
    for service in ("done", "failed", "warned", "skipped"):
        progress.service_finished(service)
    settings.done_services.append("done")
    settings.failed_services.append("failed")
    settings.warned_services.append("warned")
    settings.skipped_due_deps_services.append("skipped")

    assert progress.state("queued") == STATE_QUEUED
    assert progress.state("running") == STATE_RUNNING
    assert progress.state("done") == STATE_DONE
    assert progress.state("failed") == STATE_FAILED
    assert progress.state("warned") == STATE_WARNED
    assert progress.state("skipped") == STATE_SKIPPED


def test_service_running_again_in_next_step():
    """A service done in one module step and started again in the next one is shown as running"""
    progress = ProgressTracker()
    progress.start("move", ["serv1"])
    progress.service_started("serv1")
    progress.service_finished("serv1")
    settings.done_services.append("serv1")
    assert progress.state("serv1") == STATE_DONE

    progress.service_started("serv1")
    assert progress.state("serv1") == STATE_RUNNING


def test_render():
    progress = ProgressTracker()
    progress.start("move", ["serv1", "serv2", "serv3"])
    progress.service_started("serv1")
    progress.service_finished("serv1")
    settings.done_services.append("serv1")
    progress.service_started("serv2")
    progress.service_finished("serv2", "operator CR is Failed")
    settings.failed_services.append("serv2")

    lines = progress.render("Final status").splitlines()

    assert lines[0] == ("Final status: procedure move, 2/3 services finished "
                        "(done 1, failed 1, warned 0, skipped 0, running 0, queued 1)")
    table = "\n".join(lines[1:-1])
    assert "| Service" in table and "| State" in table
    for expected in ("serv1", "serv2", "serv3", STATE_DONE, STATE_FAILED, STATE_QUEUED):
        assert expected in table
    # only the two columns are rendered
    assert lines[2].count("|") == 3
    # failure reason goes below the table, not into a column
    assert "operator CR is Failed" not in table
    assert lines[-1] == "serv2 failed: operator CR is Failed"


def test_next_update_hint_only_while_running():
    progress = ProgressTracker()
    progress.start("move", ["serv1"])
    hint = "Next update when a service finishes, or in 60s at the latest."
    assert hint not in progress.render()  # nothing running yet

    progress.service_started("serv1")
    assert progress.render().splitlines()[-1] == hint

    progress.service_finished("serv1")
    settings.done_services.append("serv1")
    assert hint not in progress.render()


def test_print_table_disabled(caplog):
    caplog.set_level(logging.INFO)
    progress = ProgressTracker()
    progress.start("move", ["serv1"], enabled=False)
    progress.print_table("Final status")
    assert "Final status" not in caplog.text

    progress.start("move", ["serv1"], enabled=True)
    progress.print_table("Final status")
    assert "Final status: procedure move" in caplog.text


def test_table_printed_while_processing(caplog):
    """process_ts_services prints a table after each processed service, with the failed service's message"""
    init_and_check_config(args_init())
    caplog.set_level(logging.INFO)

    def process_node(serv):
        tracker.service_started(serv)
        data = {'services': {serv: {"message": "operator CR is Failed"} if serv == "bb" else {}}}
        node = ServiceDRStatus(data)
        node.service_status = serv != "bb"
        thread_result_queue.put(node)

    ts = TopologicalSorter2()
    ts.add("aa")
    ts.add("bb1", "bb")
    ts.prepare()

    tracker.start("active", ["aa", "bb", "bb1"])
    process_ts_services(ts, process_node)

    assert tracker.state("aa") == STATE_DONE
    assert tracker.state("bb") == STATE_FAILED
    assert tracker.state("bb1") == STATE_SKIPPED
    assert tracker.message["bb"] == "operator CR is Failed"
    # one table per processed service (aa, bb); skipped bb1 doesn't print its own table
    assert caplog.text.count("Progress: procedure active") == 2
    assert "bb failed: operator CR is Failed" in caplog.text


def test_table_repeated_while_service_is_running(caplog):
    """The table is printed again while a long-running service hasn't finished yet"""
    init_and_check_config(args_init())
    caplog.set_level(logging.INFO)

    def process_node(serv):
        tracker.service_started(serv)
        time.sleep(0.5)
        node = ServiceDRStatus({'services': {serv: {}}})
        node.service_status = True
        thread_result_queue.put(node)

    ts = TopologicalSorter2()
    ts.add("slow")
    ts.prepare()

    tracker.start("active", ["slow"])
    tracker.repeat_interval = 0.1
    process_ts_services(ts, process_node)

    tables = caplog.text.split("Progress: procedure active")[1:]
    # several repeated tables while running; none after the last service finished (final table covers it)
    assert len(tables) >= 2
    assert all("running 1" in table and "| running |" in table for table in tables)


def test_no_progress_table_after_last_service(caplog):
    """When the last service finishes, only the final status table is printed, not the same table twice"""
    caplog.set_level(logging.INFO)
    progress = ProgressTracker()
    progress.start("active", ["serv1", "serv2"])

    progress.service_started("serv1")
    progress.service_finished("serv1")
    settings.done_services.append("serv1")
    progress.print_progress()  # serv2 still queued: printed
    assert caplog.text.count("Progress: procedure active") == 1

    progress.service_started("serv2")
    progress.service_finished("serv2")
    settings.done_services.append("serv2")
    progress.print_progress()  # all finished: skipped
    progress.print_table("Final status")
    assert caplog.text.count("Progress: procedure active") == 1
    assert caplog.text.count("Final status: procedure active") == 1


def test_seconds_until_repeat():
    progress = ProgressTracker()
    assert progress.seconds_until_repeat() is None  # disabled: wait without timeout
    progress.start("move", ["serv1"])
    assert 0 < progress.seconds_until_repeat() <= progress.repeat_interval

    # any printed table (e.g. after a service finished) restarts the countdown
    progress.last_printed -= 50
    assert progress.seconds_until_repeat() <= progress.repeat_interval - 50
    progress.print_table()
    assert progress.seconds_until_repeat() > progress.repeat_interval - 1


def test_no_progress_table_flag():
    assert smclient.parse_command_line(["move", "site-1"]).no_progress_table is False
    assert smclient.parse_command_line(["--no-progress-table", "move", "site-1"]).no_progress_table is True
