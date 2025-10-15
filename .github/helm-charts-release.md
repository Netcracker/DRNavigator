# Helm Charts Release

## Overview

The Helm Charts Release workflow is used to release Helm charts and Docker images. It includes a dry run stage to check the build process without actually publishing artifacts. The workflow is triggered manually and allows users to specify the release version. It also creates a GitHub release after successful deployment.

## Trigger

This workflow is triggered manually via `workflow_dispatch` with a required release version input.

## Workflow Details

### Jobs

#### `check-tag`
- **Runner**: `ubuntu-latest`
- **Purpose**: Verifies if the specified release tag already exists
- **Uses**: `netcracker/qubership-workflow-hub/actions/tag-action@v1.0.4`

#### `load-docker-build-components`
- **Runner**: `ubuntu-latest`
- **Purpose**: Loads and validates Docker build configuration
- **Outputs**: 
  - `component`: Docker components configuration
  - `platforms`: Supported platforms configuration

#### `docker-check-build`
- **Runner**: `ubuntu-latest`
- **Purpose**: Performs Docker build checks for each component
- **Strategy**: Matrix strategy based on components
- **Dependencies**: Requires `load-docker-build-components` and `check-tag` jobs

### Steps

#### check-tag Job
1. **Check if tag exists**
   - Validates if the specified release tag already exists
   - Prevents duplicate releases

#### load-docker-build-components Job
1. **Checkout code**
   - Uses `actions/checkout@v4`
   - Checks out the repository code

2. **Load Docker Configuration**
   - Validates the Docker build configuration file
   - Extracts components and platforms information
   - Ensures configuration structure is valid

#### docker-check-build Job
1. **Get version for current component**
   - Sets environment variables for the current component

2. **Docker build**
   - Uses `netcracker/qubership-workflow-hub/actions/docker-action@v1.0.4`
   - Performs Docker build for each component

## Configuration

### Required Input Parameters
- `release` (string, required): The release version to publish

### Required Configuration
- [`.github/helm-charts-release-config.yaml`](../../config/examples/helm-charts-release-config.yaml): Configuration for the release process
- [`.github/docker-build-config.json`](../../config/examples/docker-build-config.json): Docker build configuration
- [`.github/assets-config.yml`](../../config/examples/assets-config.yml): Assets configuration
- [`.github/release-drafter-config.yml`](../../config/examples/release-drafter-config.yml): GitHub release drafter configuration
- Create a `gh-pages` branch.

### Permissions
- `contents: write` - For creating releases and updating files
- `packages: write` - For publishing Docker images

### Concurrency
- Prevents multiple releases from running simultaneously
- Cancels in-progress releases when a new one is started

## Usage

This workflow is particularly useful for:
- Releasing Helm charts with associated Docker images
- Ensuring build quality through dry-run stages
- Automating the complete release process
- Creating GitHub releases with proper assets

## Features

- **Dry-run support**: Test builds without publishing
- **Tag validation**: Prevents duplicate releases
- **Matrix builds**: Supports multiple Docker components
- **Configuration validation**: Ensures all required configs are valid
- **Concurrent execution control**: Prevents release conflicts
- **GitHub release integration**: Creates releases automatically

## Categories
- Automation

## Labels
- helm
