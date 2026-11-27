# Spec Delta

## Purpose

Lets operators follow a running DR or per-site procedure from the sm-client log: which services are done, running,
waiting or failed, both while the procedure runs and after it ends.

## ADDED Requirements

### Requirement: Progress table during procedures
sm-client SHALL log a per-service progress table while it runs a DR procedure (`move`, `stop`) or a per-site
procedure (`active`, `standby`, `return`, `disable`). The table SHALL list every processed service in processing
order, with columns `Service` and `State`.

#### Scenario: Table at start
- **WHEN** a DR or per-site procedure starts
- **THEN** sm-client logs a table titled `Progress at start` with every processed service in state `queued`

#### Scenario: Status and list commands
- **WHEN** sm-client runs `status` or `list`
- **THEN** no progress table is logged

### Requirement: Service states
Each service in the table SHALL have exactly one state: `queued` (not started), `running` (a step is in progress),
`done`, `failed`, `warned` (failed without breaking the procedure), or `skipped (dependency)` (not processed because
a service it depends on, or a previous module, failed).

#### Scenario: Service is processed
- **WHEN** sm-client sends the request for a service and is polling its status
- **THEN** the service is shown as `running`

#### Scenario: Service finished successfully
- **WHEN** a service finishes its step successfully
- **THEN** the service is shown as `done`

#### Scenario: Dependent service is skipped
- **WHEN** a service fails and another service depends on it
- **THEN** the dependent service is shown as `skipped (dependency)`

### Requirement: Counts line
Above each table, sm-client SHALL log a line with the title, the procedure name, the number of finished services out
of all services, and the count of services in each state.

#### Scenario: Counts during a run
- **WHEN** 4 services are processed and 1 is done, 1 failed, 1 running and 1 queued
- **THEN** the line reads `Progress: procedure move, 2/4 services finished (done 1, failed 1, warned 0, skipped 0, running 1, queued 1)`

### Requirement: Table when a service finishes
sm-client SHALL log the table (title `Progress`) each time a service finishes while other services are still
running or queued. It SHALL NOT log it when the last service finishes, because the final table follows.

#### Scenario: Service finishes while others remain
- **WHEN** a service finishes and at least one other service is running or queued
- **THEN** sm-client logs the table with the new state of that service

#### Scenario: Last service finishes
- **WHEN** the last running service finishes and no service is queued
- **THEN** sm-client logs no `Progress` table, only the `Final status` table

### Requirement: Periodic table while services run
While a service is running and no table was logged for 60 seconds, sm-client SHALL log the table again, so
long-running services still show progress. While a service is running, the table SHALL end with the line
`Next update when a service finishes, or in 60s at the latest.`

#### Scenario: Long-running service
- **WHEN** a service runs for more than 60 seconds and no other service finishes
- **THEN** sm-client logs the table every 60 seconds with that service as `running`

#### Scenario: Next update line
- **WHEN** the table is logged and no service is `running`
- **THEN** the `Next update…` line is not printed

### Requirement: Final status table
At the end of a DR or per-site procedure, on success and on failure, sm-client SHALL log the table titled
`Final status` before the `Summary` lines. The exit code SHALL stay as before (1 if any service failed).

#### Scenario: Successful procedure
- **WHEN** all services finish as `done`
- **THEN** sm-client logs `Final status` with every service `done`, then the `Summary` lines, and exits with 0

#### Scenario: Failed procedure
- **WHEN** a service fails during the procedure
- **THEN** sm-client logs `Final status` showing that service as `failed` and its dependents as `skipped (dependency)`, then the `Summary` lines, and exits with 1

### Requirement: Failure messages below the table
For each `failed` or `warned` service that has a message from site-manager, sm-client SHALL log the line
`<service> <state>: <message>` below the table.

#### Scenario: Failed service with a message
- **WHEN** a service fails and site-manager returns a message for it
- **THEN** the line `<service> failed: <message>` follows the table

### Requirement: Table goes to the log
The table SHALL be written through the sm-client logger at INFO level, so it appears in the console without `-v`
and in the file set by `--output`.

#### Scenario: Output file
- **WHEN** sm-client runs a procedure with `-o <file>`
- **THEN** every progress table is also written to `<file>`

### Requirement: Disable the table
sm-client SHALL accept the option `--no-progress-table`. When it is set, no progress table SHALL be logged, and the
rest of the output SHALL be unchanged.

#### Scenario: Option set
- **WHEN** sm-client runs `move` with `--no-progress-table`
- **THEN** no progress table is logged, and the `Summary` lines and exit code are the same as without the table
