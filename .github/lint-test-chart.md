# Lint and Test Charts

## Overview

The Lint and Test Charts workflow validates Helm charts through linting and testing processes. It uses chart-testing tools to ensure Helm charts are properly formatted, follow best practices, and can be successfully installed in a Kubernetes cluster.

## Trigger

This workflow is triggered when:
- A pull request is created or updated
- Manually via `workflow_dispatch` with optional input parameters

## Workflow Details

### Jobs

#### `lint-test`
- **Runner**: `ubuntu-latest`
- **Purpose**: Lints and tests Helm charts using chart-testing tools

### Steps

1. **Checkout**
   - Uses `actions/checkout@v3`
   - Fetches complete repository history (`fetch-depth: 0`)

2. **Set up Helm**
   - Uses `azure/setup-helm@v4.2.0`
   - Installs Helm version 3.17.0

3. **Set up Python**
   - Uses `actions/setup-python@v5.3.0`
   - Installs Python 3.x with latest available version

4. **Set up chart-testing**
   - Uses `helm/chart-testing-action@v2.7.0`
   - Installs chart-testing tools

5. **Run chart-testing (list-changed)**
   - Identifies which charts have changed compared to the target branch
   - Sets output variable to determine if linting/testing should proceed

6. **Run chart-testing (lint)**
   - Runs linting on changed charts
   - Condition: Only runs if charts have changed or manual trigger is used

7. **Create kind cluster**
   - Uses `helm/kind-action@v1.12.0`
   - Creates a local Kubernetes cluster for testing
   - Condition: Only runs if charts have changed or manual trigger is used

8. **Run chart-testing (install)**
   - Installs and tests charts in the kind cluster
   - Condition: Only runs if charts have changed or manual trigger is used

## Configuration

### Input Parameters (Manual Trigger)
- `lint-n-test` (boolean, optional): Force run lint and test charts (default: false)

### Required Tools
- **Helm**: Version 3.17.0
- **Python**: 3.x
- **chart-testing**: Latest version from chart-testing-action
- **kind**: Kubernetes cluster for testing

## Usage

This workflow is particularly useful for:
- Validating Helm chart syntax and structure
- Ensuring charts follow Helm best practices
- Testing chart installation in a real Kubernetes environment
- Catching issues before charts are deployed to production
- Maintaining chart quality standards

## Features

- **Smart execution**: Only processes changed charts on pull requests
- **Comprehensive testing**: Lints charts and tests installation
- **Real cluster testing**: Uses kind cluster for realistic testing
- **Manual override**: Can force execution via manual trigger
- **Latest tooling**: Uses current versions of Helm and chart-testing

## Categories
- Helm
- Testing
- CI/CD

## Labels
- helm
- charts
- testing
- linting
