# Auto-Assign Project to Issue

## Overview

The Auto-Assign Project to Issue workflow automatically assigns a project to an issue based on the issue's labels. This workflow template can be used to streamline project management by ensuring that issues are categorized and tracked in the appropriate projects.

## Trigger

This workflow is triggered when:
- An issue is opened
- An issue is edited
- An issue is reopened

## Workflow Details

### Jobs

#### `add-to-project`
- **Runner**: `ubuntu-latest`
- **Purpose**: Adds the issue to a specified GitHub project

### Steps

1. **Add issue to project**
   - Uses `actions/add-to-project@v1.0.2`
   - Adds the issue to the project specified by `vars.PROJECT`
   - Requires `ADD_PROJECT_TO_ISSUE_PAT` secret for authentication

2. **Log info**
   - Logs the issue number and title
   - Logs the project number that the issue was added to

## Configuration

### Required Variables
- `PROJECT`: The project number to assign issues to

### Required Secrets
- `ADD_PROJECT_TO_ISSUE_PAT`: Personal Access Token with permissions to add issues to projects

## Usage

This workflow is designed to work with GitHub's beta project feature. When an issue is created or modified, it will automatically be added to the specified project for better organization and tracking.

## Categories
- Automation

## Labels
- issues
- projects
