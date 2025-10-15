# Docker Project Release Workflow

Docker project release workflow is used to make a GitHub release and publish released artifacts into GitHub docker registry.
The workflow consists of several sequential jobs:

1. Checks if the tag already exists.
2. Creates a new tag with name "v${version}.
3. [Builds and publishes a Docker image](https://github.com/Netcracker/qubership-workflow-hub/blob/main/docs/reusable/docker-publish.md).
4. Create GitHub release

## Step 1: Docker release workflow

Copy the [prepared file](https://github.com/Netcracker/.github/blob/main/workflow-templates/docker-release.yaml) into `.github/workflows` directory of your repository.

This workflow is designed to be run manually. It has four input parameters on manual execution:

- `Release version` -- a string represents version number of the release
- `Dry run` -- if selected the workflow will go through all the steps, but will not publish anything.

## Step 2: Add configuration file for GitHub release

Upload [prepared configuration file](https://github.com/Netcracker/.github/blob/main/config/examples/release-drafter-config.yml) to your repository in `.github` folder. You can customize it in future for your needs.

## Some more info on input parameters

The [underlying reusable workflow](https://github.com/Netcracker/qubership-workflow-hub/blob/main/.github/workflows/docker-publish.yml) accepts more input parameters then the release workflow. For simplicity the template workflow provides reasonable defaults which are suitable in most cases. The default usage scenario assumes the following conditions:

- The resulting artifact name is the same as repository name
- Artifact built from tag name equal to `v${version}`
- There are no pre-built artifacts. It means that Dockerfile has a "build from sources" stage
- There is only one Dockerfile and it placed in the root of repository
- Build context for Dockerfile is `.`

If the default case isn't your's -- please read [the detailed description](https://github.com/Netcracker/qubership-workflow-hub/blob/main/docs/reusable/docker-publish.md#detailed-description-of-variables) of the input parameters and customize them according to your needs.
