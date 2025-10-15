# Link Checker

## Overview

The Link Checker workflow validates all links in Markdown files within the repository. It uses the Lychee link checker to identify broken, inaccessible, or problematic links, helping maintain the quality and reliability of documentation.

## Trigger

This workflow is triggered when:
- Code is pushed to any branch
- Repository dispatch events occur
- Manually via `workflow_dispatch`

## Workflow Details

### Jobs

#### `linkChecker`
- **Runner**: `ubuntu-latest`
- **Purpose**: Checks all Markdown files for broken or problematic links

### Steps

1. **Checkout Repository**
   - Uses `actions/checkout@v4`
   - Checks out the repository code

2. **Link Checker**
   - Uses `lycheeverse/lychee-action@v2`
   - Scans all Markdown files (`./**/*.md`) for links
   - Provides detailed link validation with configurable settings

## Configuration

### Link Checker Settings
- **Target files**: `./**/*.md` (all Markdown files recursively)
- **Verbose output**: Enabled for detailed reporting
- **Progress display**: Disabled for cleaner output
- **User agent**: Chrome browser user agent for compatibility
- **Retry settings**: 5 retries with 30-second wait time
- **Acceptable status codes**: 100-103, 200-299, 429 (rate limited)
- **Cookie jar**: Uses `cookies.json` for session management
- **Private links**: Excludes all private/internal links
- **Output format**: Markdown format
- **Failure mode**: Fails the workflow on broken links

## Usage

This workflow is particularly useful for:
- Maintaining documentation quality
- Identifying broken external links
- Ensuring documentation accessibility
- Preventing link rot in repositories
- Validating documentation before releases

## Features

- **Comprehensive scanning**: Checks all Markdown files recursively
- **Robust retry logic**: Handles temporary network issues
- **Rate limit handling**: Accepts 429 status codes
- **Session management**: Uses cookies for authenticated links
- **Detailed reporting**: Provides verbose output for troubleshooting
- **Fail-fast behavior**: Stops on first broken link

## Categories
- Automation
- Utilities
