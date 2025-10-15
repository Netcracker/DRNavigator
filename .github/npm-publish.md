# npm Publish

## Overview

The npm Publish workflow can be triggered in two ways:
1. **Automatically on push** to any branch (runs in dry-run mode for safety)
2. **Manually via workflow_dispatch** with configurable parameters

It performs a comprehensive npm package release process including dependency management, version updates, building, testing, and publishing to npm registries. When triggered by push, it automatically runs in dry-run mode to prevent accidental publishing.

## Trigger

This workflow can be triggered in two ways:

1. **Push trigger**: Automatically runs on any push to any branch
   - Automatically sets `dry-run: true` for safety
   - Uses default parameter values
   - Ideal for testing the release process

2. **Manual trigger**: Can be manually triggered via workflow_dispatch
   - Allows full customization of all parameters
   - Can set `dry-run: false` for actual publishing
   - Suitable for production releases

## Workflow Details

### Jobs

#### `npm-build-publish`
- **Runner**: `ubuntu-latest`
- **Purpose**: Builds and publishes npm packages with comprehensive version management
- **Environment**: Uses `GITHUB_TOKEN` for npm authentication

### Steps

1. **Checkout repository**
   - Uses `actions/checkout@v4`
   - Checks out the repository code

2. **Show branch**
   - Displays the current branch information

3. **Setup Node.js**
   - Uses `actions/setup-node@v4`
   - Configures Node.js version and registry settings
   - Sets up scope for scoped packages

4. **Install dependencies**
   - Runs `npm ci --legacy-peer-deps`
   - Installs project dependencies

5. **Check if project is a Lerna monorepo**
   - Detects if the project uses Lerna for monorepo management
   - Sets environment variable for subsequent steps

6. **Update dependencies (if required)**
   - Updates @netcracker dependencies when specified
   - Only runs when `update-nc-dependency` is true

7. **Update package version**
   - Updates version in lerna.json (for Lerna projects)
   - Updates version in package.json (for standard npm projects)

## Configuration

### Required Input Parameters
- `version` (string, required): Release version to publish

### Optional Input Parameters
- `scope` (string, optional): npm scope (default: "@netcracker")
- `node-version` (string, optional): Node.js version (default: "22.x")
- `registry-url` (string, optional): npm registry URL (default: `<https://npm.pkg.github.com>`)
- `update-nc-dependency` (boolean, optional): Update @netcracker dependencies (default: false)
- `dry-run` (boolean, optional): Run in dry-run mode (default: false, auto-true on push)
- `dist-tag` (string, optional): npm distribution tag (default: "latest")
- `branch_name` (string, optional): Branch name (default: "main")

### Permissions
- `contents: write` - For updating package files and committing changes
- `packages: write` - For publishing to npm registries

### Operation Modes

#### Push Mode (Automatic)
- **Trigger**: Any push to any branch
- **dry-run**: Automatically set to `true`
- **Parameters**: Use default values
- **Purpose**: Safe testing of the release process

#### Manual Mode (workflow_dispatch)
- **Trigger**: Manual workflow execution
- **dry-run**: Configurable (default: `false`)
- **Parameters**: Fully customizable
- **Purpose**: Production releases with custom configuration

## Usage

This workflow is particularly useful for:
- **Testing release process**: Automatically runs on every push in safe dry-run mode
- **Publishing npm packages**: Manual execution for actual releases
- **Managing Lerna monorepo releases**: Supports both standard and monorepo projects
- **Automating version updates**: Handles dependency and version management
- **Ensuring package quality**: Builds and tests before publishing
- **Supporting scoped packages**: Configurable npm scope support
- **Safe development workflow**: Prevents accidental publishing during development

## Features

- **Dual trigger modes**: Automatic push trigger and manual workflow dispatch
- **Safety-first approach**: Automatic dry-run mode on push to prevent accidental publishing
- **Lerna support**: Handles both standard npm and Lerna monorepo projects
- **Dependency management**: Updates @netcracker dependencies when needed
- **Version management**: Automatically updates package versions
- **Multiple registries**: Supports GitHub Packages and npm registry
- **Scoped packages**: Supports scoped package publishing
- **Distribution tags**: Configurable npm distribution tags
- **Comprehensive testing**: Builds and tests before publishing
- **Flexible configuration**: Full parameter customization for manual execution

## Categories
- JavaScript
- Node.js
- Automation
- npm

## Labels
- npm
- publish
- JavaScript
- Node.js
