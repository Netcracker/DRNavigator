# OSSF Scorecard

## Overview

The OSSF Scorecard workflow performs supply-chain security analysis using the Open Source Security Foundation (OSSF) Scorecard tool. It evaluates repositories for security best practices and provides detailed security scoring and recommendations.

## Trigger

This workflow is triggered when:
- Branch protection rules are created or deleted
- Scheduled weekly on Fridays at 20:26 UTC
- Code is pushed to the `main` branch

## Workflow Details

### Jobs

#### `analysis`
- **Runner**: `ubuntu-latest`
- **Purpose**: Performs comprehensive security analysis using OSSF Scorecard
- **Condition**: Only runs on default branch or pull requests

### Steps

1. **Checkout code**
   - Uses `actions/checkout@v4.2.2`
   - Checks out the repository code
   - Disables credential persistence for security

2. **Run analysis**
   - Uses `ossf/scorecard-action@v2.4.1`
   - Performs security analysis and generates SARIF report
   - Configurable authentication and publishing options

3. **Upload artifact**
   - Uses `actions/upload-artifact@v4.6.1`
   - Uploads SARIF results as workflow artifacts
   - Retention period: 5 days

4. **Upload to code-scanning**
   - Uses `github/codeql-action/upload-sarif@v3.29.0`
   - Uploads results to GitHub's code scanning dashboard

## Configuration

### Analysis Settings
- **Results format**: SARIF (Static Analysis Results Interchange Format)
- **Results file**: `results.sarif`
- **Publish results**: Disabled (can be enabled for public repositories)
- **File mode**: Optional git mode for repositories with .gitattributes

### Optional Authentication
- **Repository token**: `${{ secrets.SCORECARD_TOKEN }}` (for private repositories or branch protection checks)

### Permissions
- **Default**: `read-all` (read-only access)
- **Security events**: `write` (for uploading to code scanning)
- **ID token**: `write` (for publishing results and badges)

## Usage

This workflow is particularly useful for:
- Evaluating repository security posture
- Identifying security vulnerabilities and risks
- Monitoring security practices over time
- Providing security scoring and recommendations
- Integrating with GitHub's code scanning dashboard

## Security Features

- **Branch protection analysis**: Evaluates branch protection rules
- **Maintenance checks**: Ensures repositories are actively maintained
- **Supply chain security**: Identifies supply chain vulnerabilities
- **Comprehensive scoring**: Provides detailed security metrics
- **Code scanning integration**: Uploads results to GitHub's security dashboard

## Categories
- Security
- Compliance
- Monitoring

## Labels
- security
- scorecard
- ossf
- supply-chain
