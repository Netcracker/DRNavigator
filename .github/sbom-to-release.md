# SBOM to Release

## Overview

The SBOM to Release workflow generates Software Bill of Materials (SBOM) files and uploads them to GitHub releases as assets. It uses CycloneDX to create comprehensive SBOM files that document all dependencies and components in the software.

## Trigger

This workflow is triggered when:
- A release is published (released event)
- Manually via `workflow_dispatch` with a release tag input

## Workflow Details

### Jobs

#### `generate-sbom`
- **Runner**: `ubuntu-latest`
- **Purpose**: Generates SBOM files and uploads them to GitHub releases

### Steps

1. **Set RELEASE_VERSION**
   - Determines the release version from the release event or manual input
   - Sets environment variable for use in subsequent steps

2. **Checkout code**
   - Uses `actions/checkout@v4`
   - Checks out the specific release tag version

3. **Free some space**
   - Removes hosted tool cache to free up disk space
   - Ensures sufficient space for SBOM generation

4. **Generate BOM**
   - Uses CycloneDX cdxgen Docker image (v8.6.0)
   - Generates SBOM in JSON format
   - Includes license information (`FETCH_LICENSE=true`)
   - Cleans up Docker images after generation

5. **Upload SBOM file to GitHub Release**
   - Uses `AButler/upload-release-assets@v3.0`
   - Uploads the generated SBOM file to the GitHub release
   - Uses the repository name and release version in the filename

## Configuration

### Input Parameters (Manual Trigger)
- `release-tag` (string, required): Release tag to generate SBOM for

### SBOM Generation Settings
- **Tool**: CycloneDX cdxgen v8.6.0
- **Format**: JSON
- **License fetching**: Enabled
- **Output filename**: `{repository-name}_sbom.{version}.json`

### Permissions
- `contents: write` - For uploading assets to GitHub releases

## Usage

This workflow is particularly useful for:
- Generating comprehensive SBOM files for releases
- Documenting all dependencies and components
- Compliance and security auditing
- Supply chain transparency
- License compliance verification

## Features

- **Automatic triggering**: Runs automatically when releases are published
- **Manual execution**: Can be triggered manually for specific releases
- **Comprehensive SBOM**: Includes all dependencies and license information
- **Release integration**: Automatically uploads to GitHub releases
- **Space management**: Frees up disk space for large SBOM generation
- **Docker cleanup**: Removes Docker images after generation

## Categories
- Automation

## Labels
- sbom
- security
- compliance
- release
