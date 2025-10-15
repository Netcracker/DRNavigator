# Add/Update License Headers

The workflow will add or update license headers in all source code files in the repository.

## Features

- **Customizable license headers**: The workflow supports any license header that can be specified in a text file.
- **Support for multiple file types**: The workflow can handle various file types, including Java, Go, Markdown, and more.
- **Robust file processing**: The workflow processes all files in the repository, including those in subdirectories.
- **Configurable file patterns**: You can configure the workflow to process only specific file patterns.
- **Dry run mode**: The workflow can be run in dry run mode to preview changes without modifying the repository.

## Usage

1. Copy the [prepared workflow file](../../workflow-templates/license-header.yml) into the `.github/workflows` directory of your repository.
2. Create a configuration file [`.licenserc.yaml`](../../config/examples/.licenserc.yaml) in the repository root with the desired license header configuration.
3. Configure the workflow by modifying the `.licenserc.yaml` file. You can specify the license type, the file patterns to process, and whether to run the workflow in dry run mode.
4. Trigger the workflow by pushing changes to the repository or by using the `workflow_dispatch` event.

> Full documentation on `.licenserc.yaml` configuration can be found in [Apache SkyWalking Eyes documentation](https://github.com/apache/skywalking-eyes).

## Categories

- Automation
- Utilities
