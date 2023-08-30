# Status

In progress.

# Dry-Run Mode Design Documentation

## Overview

Dry-run mode is a feature in the SM Client that allows users to simulate the execution of a procedure without actually performing any actions. This is particularly useful for users who want to understand the sequence of services to be processed, their dependencies, and the order in which they will be executed, without making any changes to the actual environment.

## Purpose

The purpose of the dry-run is to provide users with insights into the execution flow of a procedure and help them verify the order in which services will be processed. This helps users identify any potential issues, validate dependencies, and make informed decisions before actually executing the procedure.

## Implementation

### Command Line Interface (CLI) Usage

To execute the command in dry-run, users need to include the `--dry-run` flag when running the SM Client. For example:
./sm-client --dry-run -v move cluster-2

### Workflow

1. When the dry-run is enabled using the CLI flag, the SM Client performs the following steps:

2. The utility validates the command, site, and services to ensure they are valid and compatible with the requested operation.

3. The utility generates an ordered list of services that need to be processed based on dependencies and other factors.

4. It creates a deep copy of the dependency graph and state information to preserve the original data.

5. The utility uses the `print_service_order` function to display the service order with dependencies. This function prints a visual representation of the order in which services will be processed and their dependency relationships.

6. It then logs the processing order, indicating the dependency relationships between services. A dependency indicator (`->`) is added before each service name, except for the first service in the list.

7. Finally, the utility informs the user that dry-run mode is active and that the procedure will not be executed.

### Benefits

- Users can understand the order in which services will be processed during a procedure.
- Dependency relationships are clearly visualized, helping users identify potential issues.
- Users can make informed decisions and validate the procedure's expected behavior before execution.
- Provides a safe environment for testing and verification without affecting the actual environment.

## Conclusion

Dry-run enhances the usability and safety of the SM Client utility by allowing users to preview the execution sequence and dependencies of a procedure. This feature empowers users to ensure the accuracy and reliability of their operations before applying changes to the real environment.
