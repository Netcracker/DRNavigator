# Qubership GitHub Workflow Templates

In this repository in folder [workflow-templates](./workflow-templates/) you can find GitHub Actions workflow templates for common CI/CD, release, and automation tasks across Qubership and Netcracker repositories. These templates help standardize and accelerate automation for various languages and use cases.

## Table of contents

- [Qubership GitHub Workflow Templates](#qubership-github-workflow-templates)
  - [Table of contents](#table-of-contents)
  - [How to Use](#how-to-use)
  - [Available Workflow Templates](#available-workflow-templates)
  - [References](#references)
    - [Qubership Workflow Hub Documentation](#qubership-workflow-hub-documentation)
    - [GitHub Actions Documentation](#github-actions-documentation)
    - [Maven project configuration for release](#maven-project-configuration-for-release)
    - [Organization level secrets reference](#organization-level-secrets-reference)
    - [Git local pre-commit hook](#git-local-pre-commit-hook)
  - [Configuration files examples](#configuration-files-examples)

## How to Use

1. **Copy the desired workflow YAML file** from this folder into your repository's `.github/workflows/` directory.
2. **Review and update required secrets and configuration files** as described in the comments at the top of each workflow file. Example config files are available in `config/examples/`.
3. **Customize input parameters** (such as version, tags, or build options) as needed for your project.
4. **Commit and push** the workflow to your repository.

## Available Workflow Templates

| Workflow Name                | Description                                                                                 | Typical Use Case / Trigger                | Workflow File |
|------------------------------|---------------------------------------------------------------------------------------------|-------------------------------------------|---------------|
| [**Add License Headers**](./docs/workflows/license-header.md) | Checks or adds license header into source code files.<br>Requires a [`.licenserc.yaml`](./config/examples/.licenserc.yaml) config file in the root folder. | On `push` and `workflow_dispatch` events | [license-header.yml](./workflow-templates/license-header.yml) |
| [**Automatic PR Labeler**](./docs/workflows/automatic-pr-labeler.md)     | Automatically label PRs based on conventional commit messages. <br>Requires a [auto-labeler-config.yaml](./config/examples/auto-labeler-config.yaml) config file in the `.github` folder | On PR events                              | [automatic-pr-labeler.yaml](./workflow-templates/automatic-pr-labeler.yaml) |
| [**CDXGen**](./docs/workflows/cdxgen.md)                   | Generate SBOM (Software Bill of Materials) and vulnerability scan report                    | On push to main, manual trigger           | [cdxgen.yaml](./workflow-templates/cdxgen.yaml) |
| [**Check Go Modules Licenses**](./docs/workflows/check-license.md) | Check the licenses of Go modules in the repository using a configurable allowlist. <br>Fails if any module has a disallowed or missing license. <br>Requires a [.wwhrd.yml](./config/examples/.wwhrd.yml) config file in the repository root. | On push | [check-license.yaml](./workflow-templates/check-license.yaml) |
| [**CLA Assistant**](./docs/workflows/cla.md)            | Check if PR authors have signed the Contributor License Agreement                           | On PR events                              | [cla.yaml](./workflow-templates/cla.yaml) |
| [**Cleanup Old Docker**](./docs/workflows/cleanup-old-docker-container.md)       | Clean up old Docker container versions in GitHub Packages                                   | Scheduled (cron), manual trigger          | [cleanup-old-docker-container.yaml](./workflow-templates/cleanup-old-docker-container.yaml) |
| [**Dependency Review**](./docs/workflows/dependency-review.md)        | Analyze new and updated dependencies for vulnerabilities, license issues, and OpenSSF Scorecard results on PRs and manual runs. | On pull request, manual trigger           | [dependency-review.yaml](./workflow-templates/dependency-review.yaml) |
| [**Dev Docker Build**](./docs/workflows/dev-docker-build.md)         | Build and publish Docker images for development, supports dry-run and custom tags. <br>Can be configured by [.github/docker-build-config.json](./config/examples/docker-build-config.json) file. [More info](https://github.com/Netcracker/qubership-workflow-hub/tree/main/actions/docker-action#example-configuration) | Manual trigger (workflow_dispatch)        | [dev-docker-build.yml](./workflow-templates/dev-docker-build.yml) |
| [**Dev Maven Docker Build**](./docs/workflows/dev-mvn-docker-build.md)   | Development build for Maven projects, with Docker image build and artifact publishing       | Manual trigger (workflow_dispatch)        | [dev-mvn-docker-build.yml](./workflow-templates/dev-mvn-docker-build.yml) |
| [**Docker Release**](./docs/workflows/docker-release.md)           | Build and publish Docker images to GitHub Container Registry, create GitHub release         | On push to main, manual trigger           | [docker-release.yaml](./workflow-templates/docker-release.yaml) |
| [**Go Build**](./docs/workflows/go-build.md)                 | Build and test Go projects, upload coverage to SonarCloud                                   | On push to main, on pull request          | [go-build.yaml](./workflow-templates/go-build.yaml) |
| [**Helm Charts Release**](./docs/workflows/helm-charts-release.md)      | Release Helm charts and Docker images, create GitHub release. <br>Requires a lot of configuration. Please read workflow file comments.<br>Configuration examples:<br>[.github/helm-charts-release-config.yaml](./config/examples/helm-charts-release-config.yaml)<br>[.github/docker-build-config.json](./config/examples/docker-build-config.json)<br>[.github/release-drafter-config.yml](./config/examples/release-drafter-config.yml) | Manual trigger (workflow_dispatch)        | [helm-charts-release.yaml](./workflow-templates/helm-charts-release.yaml) |
| [**Link Checker**](./docs/workflows/link-checker.md)             | Check Markdown files for broken links using lychee                                          | On push, manual trigger                   | [link-checker.yaml](./workflow-templates/link-checker.yaml) |
| [**Lint Codebase**](./docs/workflows/super-linter.md)           | Lint codebase using GitHub Super-Linter. Runs multiple linters on changed files for supported languages. <br>See [.github/super-linter.env](.github/super-linter.env) and [.github/linters/](.github/linters/) for configuration. | On push, pull request, manual trigger     | [super-linter.yaml](./workflow-templates/super-linter.yaml) |
| [**Maven Release v2**](./docs/workflows/maven-release-v2.md)         | Enhanced Maven release with dry-run, Docker build, and GitHub release support. <br>Requires `pom.xml` [configuration](./docs/maven-publish-pom-preparation_doc.md) and [.github/release-drafter-config.yml](./config/examples/release-drafter-config.yml) config file. | Manual trigger (workflow_dispatch)        | [maven-release-v2.yaml](./workflow-templates/maven-release-v2.yaml) |
| [**Maven Release**](./docs/workflows/maven-release.md)            | Release and upload Java artifacts to Maven Central or GitHub Packages, create GitHub release | Manual trigger (workflow_dispatch)        | [maven-release.yaml](./workflow-templates/maven-release.yaml) |
| [**Maven Snapshot Deploy**](./docs/workflows/maven-snapshot-deploy.md)    | Deploy Maven snapshot artifacts to GitHub Packages or Maven Central                         | On push to non-main/non-release branches  | [maven-snapshot-deploy.yaml](./workflow-templates/maven-snapshot-deploy.yaml) |
| [**PR Assigner**](./docs/workflows/pr-assigner.md)              | Automatically assign reviewers to PRs based on config or CODEOWNERS                        | On PR events                              | [pr-assigner.yml](./workflow-templates/pr-assigner.yml) |
| [**PR Conventional Commits**](./docs/workflows/pr-conventional-commits.md)  | Check if PR commits follow conventional commit messages                                     | On PR events                              | [pr-conventional-commits.yaml](./workflow-templates/pr-conventional-commits.yaml) |
| [**PR Lint Title**](./docs/workflows/pr-lint-title.md)            | Lint PR titles to ensure they follow conventional commit strategy                           | On PR events                              | [pr-lint-title.yaml](./workflow-templates/pr-lint-title.yaml) |
| [**Profanity Filter**](./docs/workflows/profanity-filter.md)         | Check PRs, issues, and comments for profanity                                               | On PR/issue/comment events                | [profanity-filter.yaml](./workflow-templates/profanity-filter.yaml) |
| [**NPM Publish**](./docs/workflows/npm-publish.md)              | Publish npm packages to GitHub Packages or npm registry, supports Lerna monorepos            | Manual trigger (workflow_dispatch)        | [npm-publish.yaml](./workflow-templates/npm-publish.yaml) |
| [**Python Release**](./docs/workflows/python-release.md)           | Publish Python packages to PyPI and create GitHub release                                   | Manual trigger (workflow_dispatch)        | [python-release.yaml](./workflow-templates/python-release.yaml) |
| [**Scorecard supply-chain security**](./docs/workflows/ossf-scorecard.md) | Generates and optionaly publishes OSSF scorecard of the repository | On push to `main` and weekly schedule | [ossf-scorecard.yaml](./workflow-templates/ossf-scorecard.yaml) |
| [**SBOM to Release**](./docs/workflows/sbom-to-release.md)          | Generate SBOM and upload it as a GitHub Release asset                                       | On release, manual trigger                | [sbom-to-release.yaml](./workflow-templates/sbom-to-release.yaml) |

## References

### [Qubership Workflow Hub Documentation](https://github.com/netcracker/qubership-workflow-hub)

### [GitHub Actions Documentation](https://docs.github.com/en/actions)

### [Maven project configuration for release](./docs/maven-publish-pom-preparation_doc.md)

### [Organization level secrets reference](https://github.com/Netcracker/qubership-workflow-hub?tab=readme-ov-file#the-organization-level-secrets-and-vars-used-in-actions)

### [Git local pre-commit hook](./docs/git-pre-commit-hook.md)

## Configuration files examples

- [Automatic PR Labeler](./config/examples/auto-labeler-config.yaml)
- [Docker build](./config/examples/docker-build-config.json)
- [Docker metadata](./config/examples/metadata-config.yml)
- [Helm charts release](./config/examples/helm-charts-release-config.yaml)
- [Release notes auto-generation](./config/examples/release-drafter-config.yml)
- [Release assets upload](./config/examples/assets-config.yml)


---
For questions or improvements, please open an issue or PR in this repository.
