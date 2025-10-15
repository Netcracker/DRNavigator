# Dev Mvn Docker Build

## Overview

The Dev Mvn Docker Build workflow is designed for development builds that combine Maven package building with Docker image creation. It automatically builds Maven packages with Java configuration, generates artifacts with dynamic naming and tags, and publishes Docker images. The workflow supports customizable Java version and dry-run mode.

## Trigger

This workflow is triggered manually via `workflow_dispatch` with customizable input parameters.

## Workflow Details

### Jobs

#### `dev-build`
- **Runner**: Uses reusable workflow
- **Purpose**: Builds Maven packages and uploads artifacts
- **Uses**: `netcracker/qubership-workflow-hub/.github/workflows/maven-publish.yml@v1.0.7`

#### `perform-version`
- **Runner**: `ubuntu-latest`
- **Purpose**: Generates metadata and prepares tags for the Docker build
- **Dependencies**: Requires `dev-build` job completion
- **Outputs**: 
  - `metadata`: Generated metadata result
  - `tags`: Prepared tags for the Docker image

#### `docker-build`
- **Runner**: Uses reusable workflow
- **Purpose**: Builds and publishes the Docker image
- **Dependencies**: Requires `perform-version` job completion
- **Uses**: `netcracker/qubership-workflow-hub/.github/workflows/docker-publish.yml@v1.0.7`

### Steps

#### dev-build Job
- Uses the Maven publish reusable workflow
- Configurable Maven command and Java version
- Uploads artifacts for use in Docker build

#### perform-version Job
1. **Checkout code**
   - Uses `actions/checkout@v2`
   - Checks out the repository code

2. **Create name**
   - Uses `netcracker/qubership-workflow-hub/actions/metadata-action@v1.0.7`
   - Generates metadata for naming and versioning

3. **Prepare tags**
   - Combines base tag from metadata with optional extra tags
   - Handles cases where extra tags are provided or not

4. **Summary step**
   - Outputs metadata and tags information to workflow summary

#### docker-build Job
- Uses the Docker publish reusable workflow
- Downloads artifacts from the Maven build
- Publishes Docker image with generated tags

## Configuration

### Input Parameters
- `maven-command` (string, optional): Maven command to execute (default: "--batch-mode package -Dgpg.skip=true")
- `artifact-name` (string, optional): Custom artifact name (defaults to repository name if not provided)
- `tags` (string, optional): Additional tags for the Docker image
- `java-version` (string, optional): Java version to use (default: "21")
- `dry-run` (boolean, optional): Enable dry-run mode for testing (default: false)

### Required Secrets
- `GITHUB_TOKEN`: Token for Maven package operations

### Permissions
- `contents: read` - For reading repository contents
- `packages: write` - For publishing Maven packages and Docker images

## Usage

This workflow is particularly useful for:
- Development builds that require both Maven and Docker
- Testing Maven builds with different Java versions
- Creating Docker images from Maven artifacts
- Customizing build parameters for development testing

## Features

- **Maven integration**: Builds Maven packages with configurable commands
- **Java version flexibility**: Supports different Java versions
- **Artifact management**: Uploads Maven artifacts for Docker builds
- **Dynamic metadata generation**: Automatically generates appropriate names and versions
- **Flexible tagging**: Supports custom tags in addition to generated ones
- **Dry-run support**: Test builds without actually publishing images

## Categories
- CI/CD
- Automation

## Labels
- dev
- docker
