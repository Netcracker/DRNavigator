# Cleanup Old Docker Container Versions

## Overview

The Cleanup Old Docker Container Versions workflow is designed to clean up old Docker container versions in a GitHub repository. It helps maintain repository hygiene by removing outdated container images based on configurable criteria.

## Trigger

This workflow is triggered when:
- Scheduled weekly on Sunday at midnight (cron: "0 0 * * 0")
- Manually via `workflow_dispatch` with customizable input parameters

## Workflow Details

### Jobs

#### `cleanup`
- **Runner**: `ubuntu-latest`
- **Purpose**: Cleans up old Docker container versions from GitHub Container Registry (GHCR)

### Steps

1. **Summary**
   - Displays workflow execution details including:
     - Event type
     - Actor (who triggered the workflow)
     - Threshold days for cleanup
     - Included/excluded tag patterns
     - Dry-run mode status

2. **Login to GHCR**
   - Uses `docker/login-action@v3`
   - Authenticates with GitHub Container Registry
   - Requires `GH_RWD_PACKAGE_TOKEN` secret

3. **Run Container Package Cleanup Action**
   - Uses `netcracker/qubership-workflow-hub/actions/container-package-cleanup@v1.0.4`
   - Performs the actual cleanup operation
   - Requires `GH_ACCESS_TOKEN` secret

## Configuration

### Input Parameters (Manual Trigger)
- `threshold-days` (string, optional): Number of days to keep container versions (default: "7")
- `included-tags` (string, optional): Tags to include for deletion (default: "")
- `excluded-tags` (string, optional): Tags to exclude from deletion (default: "release*")
- `dry-run` (string, optional): Enable dry-run mode (default: "false")

### Required Secrets
- `GH_RWD_PACKAGE_TOKEN`: Token for GitHub Container Registry authentication
- `GH_ACCESS_TOKEN`: Token for package cleanup operations

### Permissions
- `contents: read` - For reading repository contents

## Usage

This workflow is particularly useful for:
- Maintaining clean GitHub Container Registry repositories
- Removing outdated development and test images
- Preserving important release tags
- Testing cleanup operations in dry-run mode before execution

## Safety Features

- **Dry-run mode**: Test cleanup operations without actually deleting images
- **Tag exclusions**: Protect important tags (like releases) from deletion
- **Configurable thresholds**: Set custom retention periods
- **Scheduled execution**: Automated cleanup without manual intervention

## Categories
- CI/CD
- Automation

## Labels
- cleanup
- docker
- maven
- ghcr
