# Dependency Review

## Overview

The Dependency Review workflow performs comprehensive dependency analysis on pull requests and can also be triggered manually. It checks for vulnerabilities, license compliance, and provides OpenSSF Scorecard information to ensure the security and compliance of dependencies.

## Trigger

This workflow is triggered when:
- A pull request is created or updated
- Manually via `workflow_dispatch`

## Workflow Details

### Jobs

#### `dependency-review`
- **Runner**: `ubuntu-latest`
- **Purpose**: Performs dependency review analysis with different configurations based on trigger type

### Steps

1. **Checkout Repository**
   - Uses `actions/checkout@v4`
   - Fetches complete repository history (`fetch-depth: 0`)

2. **Get the first commit SHA** (Manual trigger only)
   - Runs only when triggered manually via `workflow_dispatch`
   - Identifies the first commit in the repository history
   - Sets `first_commit_sha` environment variable

3. **Dependency Review (manual)**
   - Runs only when triggered manually via `workflow_dispatch`
   - Uses `actions/dependency-review-action@v4`
   - Compares from the first commit to the current reference
   - Enables all checks with warning-only mode

4. **Dependency Review (pull_request)**
   - Runs only when triggered by pull request events
   - Uses `actions/dependency-review-action@v4`
   - Compares the base and head of the pull request
   - Enables all checks with standard enforcement

## Configuration

### Features Enabled
- **OpenSSF Scorecard**: Shows security scorecard information
- **Vulnerability Check**: Identifies known security vulnerabilities
- **License Check**: Validates license compliance
- **Warning Mode**: For manual triggers, issues warnings instead of failing the workflow

### Permissions
- `contents: read` - For reading repository contents and dependency information

## Usage

This workflow is particularly useful for:
- Ensuring dependency security in pull requests
- Identifying vulnerable dependencies before they are merged
- Checking license compliance of dependencies
- Getting comprehensive dependency analysis for the entire repository history

## Security Features

- **Vulnerability scanning**: Identifies known CVEs in dependencies
- **License compliance**: Ensures dependencies meet license requirements
- **OpenSSF Scorecard**: Provides security scoring for dependencies
- **Comprehensive analysis**: Reviews all dependencies in the repository

## Categories
- CI/CD
- Automation

## Labels
- security
- dependencies
- compliance
