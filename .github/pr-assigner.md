# PR Assigner

## Overview

The PR Assigner workflow automatically assigns reviewers to pull requests based on configuration files or CODEOWNERS. It uses the PR Assigner GitHub Action to streamline the review process and ensure appropriate reviewers are assigned to each pull request.

## Trigger

This workflow is triggered when:
- A pull request is opened
- A pull request is reopened
- A pull request is synchronized (updated)

## Workflow Details

### Jobs

#### `assign`
- **Runner**: `ubuntu-latest`
- **Purpose**: Automatically assigns reviewers to pull requests
- **Uses**: `netcracker/qubership-workflow-hub/actions/pr-assigner@v1.0.4`

#### `warning-message`
- **Runner**: `ubuntu-latest`
- **Purpose**: Displays warning for pull requests from forks
- **Condition**: Only runs for pull requests from external forks

### Steps

#### assign Job
1. **Checkout Repository**
   - Uses `actions/checkout@v4`
   - Checks out the repository code

2. **PR Auto-Assignment**
   - Uses the PR assigner action
   - Reads configuration from [`.github/pr-assigner-config.yml`](../../config/examples/pr-assigner-config.yml)
   - Shuffles assignments with factor 2 for distribution

#### warning-message Job
1. **Warning**
   - Displays warning message for fork pull requests
   - Informs that automatic assignment is not possible for forks

## Configuration

### Required Files
- [`.github/pr-assigner-config.yml`](../../config/examples/pr-assigner-config.yml): Configuration file for PR assignment rules

### Assignment Settings
- **Configuration path**: `.github/pr-assigner-config.yml`
- **Shuffle factor**: 2 (distributes assignments evenly)
- **Assignment method**: Based on configuration file or CODEOWNERS

### Permissions
- `pull-requests: write` - For assigning reviewers to pull requests
- `contents: read` - For reading configuration files

## Usage

This workflow is particularly useful for:
- Automating reviewer assignment for pull requests
- Ensuring consistent review coverage
- Reducing manual overhead in pull request management
- Distributing review workload evenly among team members
- Maintaining review quality through appropriate assignments

## Features

- **Automatic assignment**: Assigns reviewers based on configuration
- **Fork detection**: Handles pull requests from external forks appropriately
- **Shuffle distribution**: Ensures even distribution of review assignments
- **Configuration-based**: Uses flexible configuration files
- **CODEOWNERS integration**: Can use existing CODEOWNERS files

## Categories
- Automation
- Pull Request
