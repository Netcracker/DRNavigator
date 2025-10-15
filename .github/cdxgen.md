# CDXGen

## Overview

The CDXGen workflow generates SBOM (Software Bill of Materials) files for the repository and vulnerability scan reports using CycloneDX. The workflow runs on push to the main branch and can also be triggered manually.

## Trigger

This workflow is triggered when:
- Code is pushed to the `main` branch
- Manually via `workflow_dispatch` with optional input parameters

## Workflow Details

### Jobs

#### `cdxgen`
- **Runner**: `ubuntu-latest`
- **Purpose**: Generates CycloneDX SBOM and vulnerability reports

#### `deploy-pages` (Conditional)
- **Runner**: `ubuntu-latest`
- **Purpose**: Deploys the generated report to GitHub Pages
- **Condition**: Only runs when `generate_cdx_report` input is true

### Steps

#### cdxgen Job
1. **cdxgen**
   - Uses `netcracker/qubership-workflow-hub/actions/cdxgen@v1.0.4`
   - Generates SBOM and vulnerability reports
   - Accepts `generate_cdx_report` input parameter

#### deploy-pages Job
1. **Deploy to GitHub Pages**
   - Uses `actions/deploy-pages@v4`
   - Deploys the generated report to GitHub Pages
   - Requires `id-token: write` and `pages: write` permissions

2. **Summary**
   - Outputs the GitHub Pages URL for the CycloneDX report

## Configuration

### Input Parameters (Manual Trigger)
- `generate_cdx_report` (boolean, optional): Whether to generate CycloneDX report and deploy to GitHub Pages

### Permissions
- `contents: read` - For reading repository contents
- `id-token: write` - For GitHub Pages deployment (conditional)
- `pages: write` - For GitHub Pages deployment (conditional)

### Environment
- `github-pages` - For GitHub Pages deployment

## Supported File Patterns
- `*/**/go.mod` - Go modules
- `*/**/pom.xml` - Maven projects
- `*/**/requirements.txt` - Python requirements
- `*/**/Dockerfile` - Docker containers
- `*/**/pyproject.toml` - Python projects

## Usage

This workflow is particularly useful for:
- Generating SBOM files for compliance and security audits
- Identifying vulnerabilities in dependencies
- Maintaining an up-to-date inventory of software components
- Deploying vulnerability reports to GitHub Pages for easy access

## Categories
- Go
- Maven
- Python
- Docker
- Automation
- Utilities

## Labels
- sbom
- security
- vulnerability-scan
- cyclonedx
