"""Flank configuration functions (specified in --args section)"""
from prometheus_flask_exporter.multiprocess import GunicornInternalPrometheusMetrics  # type: ignore


def child_exit(server, worker):
    """Mark process dead for prometheus metrics"""
    GunicornInternalPrometheusMetrics.mark_process_dead_on_child_exit(worker.pid)
