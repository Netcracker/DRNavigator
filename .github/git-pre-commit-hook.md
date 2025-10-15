# Configure local pre-commit hook

As we aggreed to use [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) and our GitHub contract does not support pre-commit hooks, it is suggested to configure local git pre-commit scripts to check commit messages follows conventional commits. Follow next instructions to install and configure pre-commit with usefull pre-commit checks, including conventional commits check.

## 1. Install [pre-commit](https://pre-commit.com/):

```bash
pip install pre-commit
```

## 2. Add configuration to your git repo `.pre-commit-config.yaml` (usefull checks included, see included repos for more details):

```yaml
default_install_hook_types:
  - pre-commit
  - commit-msg

repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v5.0.0
    hooks:
      - id: trailing-whitespace
      - id: check-case-conflict
      - id: check-executables-have-shebangs
      - id: no-commit-to-branch
      - id: destroyed-symlinks
      - id: mixed-line-ending
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files

  - repo: https://github.com/compilerla/conventional-pre-commit
    rev: v4.2.0
    hooks:
      - id: conventional-pre-commit
        stages: [commit-msg]
        args: [--strict, --verbose]
```

## 3. Install the pre-commit script:

```bash
pre-commit install --install-hooks
```
