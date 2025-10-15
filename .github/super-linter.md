# Lint codebase (Super-linter)

This workflow executes several linters on changed files based on languages used in your codebase whenever you push a code or open a pull request.
You can adjust the behavior by adding/modifying configuration files for super-linter itself and for individual linters.

For more information, see:
[Super-linter](https://github.com/super-linter/super-linter)

Configuration file for super-linter example:
[.github/super-linter.env](https://github.com/Netcracker/.github/blob/main/.github/super-linter.env)
Configuration files for individual linters should be placed in .github/linters

**To add the super-linter workflow into your repository** just copy the [prepared file](https://github.com/Netcracker/.github/blob/main/workflow-templates/super-linter.yaml) into `.github/workflows` directory of your repository.

> Super-linter workflow uses the centralized configuration from the [Netcracker/.github](https://github.com/Netcracker/.github/tree/main/config/linters) repository.
> One can override/add any linter configuration and super-linter environment by adding corresponding file into target repository.
> For example to override super-linter environment, just add `.github/super-linter.env` file to your repostory and set the env you need. The same for individual linter configuration.
