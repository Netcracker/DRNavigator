# Maven Snapshot Deploy

## Overview

The Maven Snapshot Deploy workflow is used to deploy Maven snapshot artifacts to GitHub packages or Maven Central repository. It is triggered on pushes to branches other than main and release branches, providing continuous deployment of development builds.

## Trigger

This workflow is triggered when:
- Code is pushed to any branch except:
  - `main`
  - `**release*` (any release branch)
  - `prettier/**`
  - `dependabot/**`
- Paths are not ignored (excludes `docs/**`, `README.md`, `.github/**`)
- Manually via `workflow_dispatch`

## Workflow Details

### Jobs

#### `deploy`
- **Runner**: `ubuntu-latest`
- **Purpose**: Deploys Maven snapshot artifacts to the configured repository

### Steps

1. **Checkout code**
   - Uses `actions/checkout@v4`
   - Checks out the repository code

2. **Deploy Maven Snapshot**
   - Uses `netcracker/qubership-workflow-hub/actions/maven-snapshot-deploy@v1.0.4`
   - Deploys snapshot artifacts to the specified target store

## Configuration

### Maven Snapshot Deploy Settings
- **Java version**: 21
- **Target store**: 'github' (or 'central' for Maven Central repository)
- **Maven command**: 'deploy' (default)
- **Additional Maven args**: Configurable additional arguments
- **Maven username**: Uses `github.actor` for GitHub packages
- **Maven token**: Uses `github.token` for GitHub packages

### Optional Configuration (for Maven Central)
- **Maven username**: `${{ secrets.MAVEN_USER }}`
- **Maven token**: `${{ secrets.MAVEN_PASSWORD }}`
- **GPG private key**: `${{ secrets.MAVEN_GPG_PRIVATE_KEY }}`
- **GPG passphrase**: `${{ secrets.MAVEN_GPG_PASSPHRASE }}`

### Permissions
- `packages: write` - Required for GitHub packages deployment
- `contents: read` - For reading repository contents

## Usage

This workflow is particularly useful for:
- Continuous deployment of development builds
- Testing Maven artifacts before release
- Providing snapshot versions for testing
- Automating deployment to GitHub packages or Maven Central
- Supporting development workflows with immediate artifact availability

## Features

- **Smart triggering**: Only deploys from development branches
- **Path filtering**: Ignores documentation and configuration changes
- **Multiple targets**: Supports both GitHub packages and Maven Central
- **GPG signing**: Supports signed artifacts for Maven Central
- **Configurable Maven**: Allows custom Maven commands and arguments
- **Automatic versioning**: Handles snapshot versioning automatically

## Categories
- Automation
- Maven
- Utilities

## Labels
- maven
- snapshot
- deploy
- continuous-deployment
