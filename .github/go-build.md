# Go Build

## Overview

The Go Build workflow is designed to build and test Go projects with comprehensive coverage reporting. It builds the Go project, runs tests with coverage analysis, and uploads the coverage report to SonarCloud for quality analysis.

## Trigger

This workflow is triggered when:
- Code is pushed to the `main` branch
- A pull request is opened, synchronized, or reopened

## Workflow Details

### Jobs

#### `build`
- **Runner**: `ubuntu-latest`
- **Purpose**: Builds the Go project, runs tests, and uploads coverage reports

### Steps

1. **Checkout Repository**
   - Uses `actions/checkout@v4`
   - Fetches complete repository history (`fetch-depth: 0`)

2. **Set up Go**
   - Uses `actions/setup-go@v4`
   - Configures Go version 1.23

3. **Build**
   - Runs `go build -v ./...`
   - Builds all Go packages in the repository with verbose output

4. **Test with coverage**
   - Runs `go test -v ./... -coverprofile coverage.out`
   - Executes all tests with verbose output
   - Generates coverage report in `coverage.out` file

5. **Upload coverage report to SonarCloud**
   - Uses `SonarSource/sonarqube-scan-action@v5`
   - Uploads coverage report to SonarCloud for quality analysis
   - Configures SonarCloud project settings

## Configuration

### Required Variables
- `SONAR_PROJECT_KEY`: SonarCloud project key
- `SONAR_ORGANIZATION`: SonarCloud organization name
- `SONAR_HOST_URL`: SonarCloud host URL

### Required Secrets
- `SONAR_TOKEN`: SonarCloud authentication token

### Permissions
- `contents: read` - For reading repository contents

## Usage

This workflow is particularly useful for:
- Building and testing Go applications
- Ensuring code quality through test coverage
- Integrating with SonarCloud for continuous quality monitoring
- Automating the build and test process for Go projects

## Features

- **Go 1.23 support**: Uses the latest stable Go version
- **Comprehensive testing**: Runs all tests in the repository
- **Coverage analysis**: Generates detailed coverage reports
- **SonarCloud integration**: Uploads coverage data for quality analysis
- **Verbose output**: Provides detailed build and test information

## Categories
- Go

## Labels
- go
- build
- test
- coverage
- sonarcloud
