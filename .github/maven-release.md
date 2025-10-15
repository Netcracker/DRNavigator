# Maven Release

## Overview

The Maven Release workflow is designed to be triggered when a release is marked as a full release. It performs a comprehensive release process including version management, building, testing, tagging, and deployment to Maven repositories.

## Trigger

This workflow is triggered manually via `workflow_dispatch` with configurable input parameters.

## Workflow Details

### Jobs

#### `check-tag`
- **Runner**: `ubuntu-latest`
- **Purpose**: Validates input parameters and checks if the release tag already exists
- **Uses**: `netcracker/qubership-workflow-hub/actions/tag-checker@v1.0.4`

#### `update-pom-version`
- **Runner**: `ubuntu-latest`
- **Purpose**: Updates the version in pom.xml and commits changes
- **Dependencies**: Requires `check-tag` job completion
- **Outputs**: `artifact_id` from configuration

### Steps

#### check-tag Job
1. **Input parameters**
   - Displays version and Java version in workflow summary

2. **Checkout code**
   - Uses `actions/checkout@v4`
   - Checks out the repository code

3. **Check if tag exists**
   - Validates if the release tag already exists
   - Prevents duplicate releases

4. **Output result**
   - Logs tag existence status and tag name

5. **Fail if tag exists**
   - Terminates workflow if tag already exists

#### update-pom-version Job
1. **Checkout code**
   - Uses `actions/checkout@v4`
   - Fetches complete repository history

2. **Update pom.xml**
   - Updates version information in pom.xml file

## Configuration

### Required Input Parameters
- `version` (string, required): Release version (e.g., 1.0.0)

### Optional Input Parameters
- `java-version` (string, optional): Java version to use (default: "21")
- `build-docker` (boolean, optional): Release docker image if Dockerfile exists (default: true)
- `profile` (choice, required): Release mode - 'github' or 'central' (default: 'central')
- `dry-run` (boolean, optional): Enable dry-run mode (default: false)

### Permissions
- `contents: write` - For updating pom.xml and creating releases
- `packages: write` - For publishing Maven artifacts

## Usage

This workflow is particularly useful for:
- Automating Maven project releases
- Publishing artifacts to Maven Central or GitHub Packages
- Creating GitHub releases with proper versioning
- Building and publishing Docker images alongside Maven artifacts
- Ensuring release quality through comprehensive testing

## Features

- **Version management**: Automatically updates pom.xml with release version
- **Tag validation**: Prevents duplicate releases
- **Multiple deployment targets**: Supports Maven Central and GitHub Packages
- **Docker integration**: Optionally builds and publishes Docker images
- **Dry-run support**: Test releases without actual deployment
- **Comprehensive testing**: Runs tests before deployment

## Categories
- Java
- Automation
- Maven

## Labels
- maven
- release
- java
- central
