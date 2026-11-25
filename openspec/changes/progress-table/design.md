# Design

## Context

See proposal.md - Why. sm-client already tracks per-service results in `settings` (`done_services`,
`failed_services`, `warned_services`, `skipped_due_deps_services`) and processes services in parallel threads,
module by module, through `process_ts_services`, which waits on a result queue. Before this change, it logged per-service
polling only at DEBUG level, interleaved across threads, and a plain summary at the end.

dr-operations runs sm-client inside GitLab CI jobs; the job log is the only place the operator sees.

## Goals / Non-Goals

**Goals:**
- A readable, compact table in the log while the procedure runs and at the end, also on failure.
- No change to the site-manager API, exit codes or the existing `Summary` output.

**Non-Goals:**
- Live status for `status`/`list` (they already print a full status table).
- A separate status job in dr-operations that polls site-manager during the procedure.
- Per-step or per-module columns, timings.

## Decisions

1. **Print the table from inside sm-client, not from a separate pipeline job.**
   sm-client is the only component that knows which services are queued vs running in this run. A parallel
   `sm-client status` job can't tell "not started yet" from "done in a previous run", can't run while the
   procedure job holds the pipeline stage, and doesn't help on failure.
   Alternatives: a manual re-triggerable status job in dr-operations; an automatic polling job. Both rejected for the
   reasons above and the extra pipeline complexity.

2. **Columns `Service | State` only, with a counts line above.**
   Extra columns (#, module, step, elapsed) made the table too wide and noisy in CI logs with many services.
   Failure messages go on lines below the table instead of a column, because they can be long.

3. **Final state comes from the existing result lists in `settings`; the tracker only adds started/finished.**
   This keeps one source of truth for results and avoids changing `processing.py` result handling.
   A new module `sm_client/progress.py` (`ProgressTracker`, one shared `tracker`) holds the state behind a lock,
   since service threads and the main loop update it.

4. **Reprint every 60 s by waiting on the result queue with a timeout.**
   `wait_for_service_result` replaces the blocking `queue.get()`: on timeout it reprints the table and waits again.
   No extra thread or timer. The interval is a constant (`REPEAT_INTERVAL = 60`), not an option, to keep the CLI
   small.

5. **No `Progress` table when the last service finishes.**
   It would be identical to `Final status`, which follows immediately.

6. **Log at INFO through the existing logger, opt-out with `--no-progress-table`.**
   INFO is printed without `-v` and goes to the `--output` file. Opt-out instead of opt-in, so dr-operations and
   other callers get the table without changes.

## Risks / Trade-offs

- [Services with more than one step in the module flow (e.g. a replicator with `standby` first and `active` at the
  end) show `done` after the first step] → `Final status` is always correct; documented as known behavior.
- [A short step that starts after the last queued service and takes less than 60 s is never shown as `running`] →
  Accepted: it avoids printing the same table twice at the end. `Final status` shows the result.
- [Log volume grows with many services: one table per finished service] → Accepted by the requester; can be turned off
  with `--no-progress-table`.
- [For a service that fails on health without its own error message, the line below the table shows the service's
  message, not the reason (e.g. healthz `down`)] → Follow-up improvement, outside this change.

## Migration Plan

None. The table is on by default; callers that don't want it pass `--no-progress-table`. Rollback = previous
sm-client image.
