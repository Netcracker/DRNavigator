# Dev Docker Build

## Overview

The Dev Docker Build workflow is designed for building and publishing Docker images during development. It dynamically generates artifact names, prepares tags, and supports a dry-run mode for testing. The workflow is triggered manually and allows customization of artifact names and tags.

## Trigger

This workflow is triggered manually via `workflow_dispatch` with customizable input parameters.

## Workflow Details

### Jobs

#### `perform-version`
- **Runner**: `ubuntu-latest`
- **Purpose**: Generates metadata and prepares tags for the Docker build
- **Outputs**: 
  - `metadata`: Generated metadata result
  - `tags`: Prepared tags for the Docker image

#### `docker-build`
- **Runner**: Uses reusable workflow
- **Purpose**: Builds and publishes the Docker image
- **Dependencies**: Requires `perform-version` job completion

### Steps

#### perform-version Job
1. **Checkout code**
   - Uses `actions/checkout@v2`
   - Checks out the repository code

2. **Create name**
   - Uses `netcracker/qubership-workflow-hub/actions/metadata-action@v1.0.4`
   - Generates metadata based on [`.github/metadata-config.yml`](../../config/examples/metadata-config.yml) configuration

3. **Prepare tags**
   - Combines base tag from metadata with optional extra tags
   - Handles cases where extra tags are provided or not

4. **Summary step**
   - Outputs metadata and tags information to workflow summary

#### docker-build Job
- Uses reusable workflow `netcracker/qubership-workflow-hub/.github/workflows/docker-publish.yml@v1.0.3`
- Passes the generated metadata and tags to the reusable workflow

## Configuration

### Input Parameters
- `artifact-name` (string, optional): Custom artifact name (defaults to repository name if not provided)
- `tags` (string, optional): Additional tags for the Docker image
- `dry-run` (boolean, optional): Enable dry-run mode for testing (default: false)

### Required Files
- [`.github/metadata-config.yml`](../../config/examples/metadata-config.yml): Configuration file for metadata generation

### Permissions
- `contents: read` - For reading repository contents
- `packages: write` - For publishing Docker images (inherited from reusable workflow)

## Usage

This workflow is particularly useful for:
- Development builds of Docker images
- Testing Docker build configurations
- Customizing artifact names and tags
- Dry-run testing before actual builds

## Features

- **Dynamic metadata generation**: Automatically generates appropriate names and versions
- **Flexible tagging**: Supports custom tags in addition to generated ones
- **Dry-run support**: Test builds without actually publishing images
- **Reusable workflow integration**: Leverages the established Docker publish workflow

## Categories
- CI/CD
- Automation

## Labels
- dev
- docker
