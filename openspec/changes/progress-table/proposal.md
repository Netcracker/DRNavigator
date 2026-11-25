# Proposal

Ticket: CPCAP-17157

## Why

During a DR procedure, sm-client logs only a summary before and after the run. Operators can't see from the job log
which services are already done, which are still running and which are waiting, and a failed run ends with plain
lists instead of a status table. This makes long switchovers hard to follow and failed runs slow to troubleshoot.

## What Changes

- sm-client logs a per-service progress table (`Service | State`, with a counts line above it) during DR procedures
  (`move`, `stop`) and per-site procedures (`active`, `standby`, `return`, `disable`):
  - at start (`Progress at start`);
  - each time a service finishes while others are still running or queued (`Progress`);
  - every 60 seconds while no service finishes (`Progress`);
  - at the end, on success and on failure (`Final status`), before the `Summary` lines.
- Failure and warning messages from site-manager are printed below the table.
- New option `--no-progress-table` disables the table.
- The table goes through the logger, so it is also written to the `--output` file.
- `status` and `list` commands are not changed.

No breaking changes: the existing `Summary` output, exit codes and API calls stay the same.

## Capabilities

### New Capabilities

- `dr-procedure-progress`: per-service progress reporting in the sm-client log while a DR or per-site procedure runs.

### Modified Capabilities

None.

## Impact

- Code: `sm-client` (start/final table, `--no-progress-table`), `sm_client/processing.py` (hooks, periodic reprint),
  new `sm_client/progress.py`, `sm_client/data/settings.py`.
- Docs: `documentation/public/architecture.md` (section "Progress table", CLI usage).
- Tests: `tests/selftest/sm_client/test_progress.py`.
- No change to site-manager, its API, CRDs or Helm chart.
- Consumers: dr-operations pipelines run sm-client and show the table in the job log. The dr-operations side
  (status jobs after a failed run) is a separate change in that repository under the same ticket.
