# Python Project Release Workflow

Python project release workflow is used to make a GitHub release and publish released artifacts into PyPi.

## Step 1: Python release workflow

Copy the [prepared file](https://github.com/Netcracker/.github/blob/main/workflow-templates/python-release.yaml) into `.github/workflows` directory of your repository.

This workflow is designed to be run manually. It has six input parameters on manual execution:

- `Specify version (optional)` -- a string represents the version number (optional). If empty the version number will be created automatically.
- `Python version to use` -- a string represents Python version to use to build artifacts (e.g., `3.11`)
- `Poetry version bump` -- which part of SemVer version to bump (e.g., `patch`, `minor`, `major`)'
- `Additional poetry build parameters` -- additional poetry build parameters
- `Run pytest` -- to run pytests or not (true/false).
- `Parameters for pytest` -- additional parameters to pass into pytests. By default value is `--maxfail=3 -v`

## Step 2: Prepare secrets

The Python Release workflow require PyPi API token. You need to get it from PyPi and add into your repositories Actions secrets. The name of the secret should be `PYPI_API_TOKEN`.

## Step 3: Add configuration file for GitHub release

Upload [prepared configuration file](https://github.com/Netcracker/.github/blob/main/config/examples/release-drafter-config.yml) to your repository in `.github` folder. You can customize it in future for your needs.
