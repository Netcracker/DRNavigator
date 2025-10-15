# Maven project release workflow

Maven project release workflow is used to make a GitHub release and publish released artifacts into Maven Central or GitHub packages.
The workflow consists of several sequential jobs:

1. Checks if the tag already exists.
2. Builds current snapshot versions (dry run step)
3. Prepares release using Maven (maven-release-plugin)
4. Builds and deploys the project using Maven (maven-release-plugin).
5. Builds and publishes a Docker image if there is Dockerfile.
6. Create GitHub release

To make it use one need to do several preparations in the project.

## Step 1: Prepare pom.xml

First of all please make sure the `pom.xml` file prepared to build source code and java doc jars alongside with main artifact. You need to sign all publishing artifacts with PGP key too. The instruction on how to do it can be found here [Prepare your project to publish into Maven Central](https://github.com/Netcracker/.github/blob/main/docs/maven-publish-pom-preparation_doc.md)

## Step 2: Maven release workflow

Copy the [prepared file](https://github.com/Netcracker/.github/blob/main/workflow-templates/maven-release-v2.yaml) into `.github/workflows` directory of your repository.

This workflow is designed to be run manually. It has four input parameters on manual execution:

- `Version type to release` -- a string represents version type to release: `patch`, `minor` or `major`
- `Maven profile to use` -- maven profile can be `central` or `github`.
- `Additional maven arguments to pass` -- additional arguments to pass to `mvn` command.
- `Build Docker image` -- build and publish docker image to GitHub packages if Dockerfile exists

## Step 3: Add configuration file for GitHub release

Upload [prepared configuration file](https://github.com/Netcracker/.github/blob/main/config/examples/release-drafter-config.yml) to your repository in `.github` folder. You can customize it in future for your needs.

## Step 4: Prepare actions secrets

The workflow needs several secrets to be prepared to work properly. For `Netcracker` repositories all of them already prepared, configured and available for use. You can find them in table [The organization level secrets and vars used in actions](#the-organization-level-secrets-and-vars-used-in-actions). Detailed instructions on how to generate a GPG key and set up secrets in a GitHub repository can be found in [this document](https://github.com/Netcracker/.github/blob/main/docs/maven-publish-secrets_doc.md).
