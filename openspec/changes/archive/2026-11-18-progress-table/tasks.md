# Tasks

## 1. Progress tracking

- [x] 1.1 Add `sm_client/progress.py` with `ProgressTracker` (states, counts line, `Service | State` table, messages below the table, `Next update…` line) and verify `test_service_states`, `test_service_running_again_in_next_step`, `test_render`, `test_next_update_hint_only_while_running` pass
- [x] 1.2 Hook the tracker into `sm_client/processing.py` (service started/finished, table when a service finishes while others remain) and verify `test_table_printed_while_processing` and `test_no_progress_table_after_last_service` pass
- [x] 1.3 Reprint the table every 60 s while waiting for a service result (`wait_for_service_result`) and verify `test_table_repeated_while_service_is_running` and `test_seconds_until_repeat` pass

## 2. CLI

- [x] 2.1 Print `Progress at start` and `Final status` in `sm-client` around DR and per-site procedures, and verify the final table appears on a failed run (environment test, failure run)
- [x] 2.2 Add `--no-progress-table` (`settings.progress_table`) and verify `test_no_progress_table_flag` and `test_print_table_disabled` pass
- [x] 2.3 Document the table and the option in `documentation/public/architecture.md` (section "Progress table", CLI usage) and verify the example matches real output

## 3. Integration

- [x] 3.1 Run the local CI (`--python`: lint and unit tests) and GitHub checks (lint, unit, integration, cloud tests) and verify all pass
- [x] 3.2 Run all dr-operations pipelines (switchover and failover both ways, passivate, activate, a failure run) with the branch image on a two-site OpenShift environment and verify the tables match the spec

## Workflow follow-up

- Archive the change in the PR after verification, so the spec lands in `openspec/specs/dr-procedure-progress/`.
