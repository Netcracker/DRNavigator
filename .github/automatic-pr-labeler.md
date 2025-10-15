# Automatic PR labels based on conventional commits

The workflow automatically label PR on it's open/reopen events. It checks all the commit messages for certain words and apply corresponding labels to PR. Those labels used by several workflows to generate release notes.

## Step 1: Create workflow file

Create a new workflow in your repository. Copy the [prepared file](https://github.com/Netcracker/.github/blob/main/workflow-templates/automatic-pr-labeler.yaml) into `.github/workflows` directory of your repository.

## Step 2: Add configuration file

Create a new configuration file `.github/auto-labeler-config.yaml`. Copy [prepared file](https://github.com/Netcracker/.github/blob/main/config/examples/auto-labeler-config.yaml)  into `.github` directory of your repository.

## Step 3: Follow the conventional commits messages strategy

The configuration file from [previous step](#step-2-add-configuration-file) defines the next rules for PR labeling based on words in **commit messages**:

| Commit message word(s)                                  | Label           |
| ------------------------------------------------------- | --------------- |
| 'FIX', 'Fix', 'fix', 'FIXED', 'Fixed', 'fixed'          | bug             |
| 'FEATURE', 'Feature', 'feature', 'FEAT', 'Feat', 'feat' | enhancement     |
| 'BREAKING CHANGE', 'BREAKING', 'MAJOR'                  | breaking-change |
| 'refactor','Refactor'                                   | refactor        |
| 'doc','docs','document','documentation'                 | documentation   |
| 'build','rebuild'                                       | build           |
| 'config', 'conf', 'configuration', 'configure'          | config          |

Labels on PRs used to generate release notes for GitHub releases. You can edit labels configuration and [release notes generation template](../../config/examples/release-drafter-config.yml) to extend or improve the default ones.
