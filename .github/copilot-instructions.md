# Netcracker/.github Repository - GitHub Workflow Templates

The Netcracker/.github repository is a special GitHub organization repository that provides standardized CI/CD workflow templates, configuration files, and documentation for repositories across the Netcracker/Qubership organization. This is a template and configuration repository, not a traditional software project.

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Working Effectively

### Essential Setup and Validation
- Bootstrap the repository validation tools:
  - `npm install markdownlint-cli2 markdown-link-check --no-save` -- installs validation tools locally (7 seconds)
  - NEVER CANCEL: Validation tools install takes 5-10 seconds. Set timeout to 30+ seconds.
  - Add `node_modules/` to `.gitignore` to avoid committing dependencies

### Core Validation Commands (The "Build" Process)
- **YAML Linting**: `yamllint -c .github/linters/.yaml-lint.yml .` -- validates all YAML files (0.5 seconds)
- **Markdown Linting**: `npx markdownlint-cli2 --config .github/linters/.markdownlint.json "**/*.md"` -- validates markdown files (3 seconds) 
- **Link Checking**: `npx markdown-link-check README.md` -- validates links in README (1.5 seconds)
- **Full Validation**: Run all three commands sequentially -- completes in under 10 seconds total
- NEVER CANCEL: All validation commands complete in under 10 seconds. Set timeout to 30+ seconds.

### Common Validation Issues and Solutions
- YAML files may have indentation warnings - these are style issues, not critical errors
- Markdown files often have line length violations (120 char limit) - fix by breaking long lines
- Some workflow templates intentionally have YAML validation issues - document but don't fix unless requested

## Repository Structure and Navigation

### Key Directories
- **`workflow-templates/`**: GitHub Actions workflow templates for various languages and use cases
  - Contains 30+ templates including Maven, NPM, Python, Go, Docker workflows
  - Each template has corresponding `.properties.json` file for GitHub's template system
- **`config/examples/`**: Example configuration files for various tools and workflows
  - Includes examples for Docker builds, release drafting, PR labeling, etc.
- **`docs/workflows/`**: Detailed documentation for each workflow template
  - One markdown file per workflow explaining usage, requirements, and configuration
- **`.github/`**: Organization-level GitHub configuration
  - Workflow files that run on this repository itself
  - Linter configurations in `.github/linters/`
  - Issue templates, security policies, contributor guidelines

### Important Files
- **`README.md`**: Main documentation listing all available workflow templates
- **`.github/linters/`**: Linter configuration files
  - `.yaml-lint.yml`: YAML validation rules
  - `.markdownlint.json`: Markdown validation rules
  - `.checkov.yaml`: Security scanning configuration
- **`workflow-templates/*.yaml`**: The actual workflow template files
- **`config/examples/`**: Configuration file examples for copying to other repositories

## Validation Scenarios

### After Making Changes to Workflow Templates
1. **Always run full validation**: `yamllint -c .github/linters/.yaml-lint.yml . && npx markdownlint-cli2 --config .github/linters/.markdownlint.json "*.md"`
2. **Test specific workflow syntax**: YAML linting will catch most GitHub Actions syntax issues
3. **Validate documentation**: Check that any new templates have corresponding documentation in `docs/workflows/`
4. **Link validation**: Run `npx markdown-link-check README.md` to ensure all links work

### After Making Changes to Documentation
1. **Markdown validation**: `npx markdownlint-cli2 --config .github/linters/.markdownlint.json "**/*.md"`
2. **Link checking**: Test links with `npx markdown-link-check [filename].md`
3. **Table of contents sync**: Ensure README.md table matches actual available workflows

### Manual Testing Scenarios
- **Workflow Template Usage**: Copy a template to a test repository and verify it works
- **Configuration Examples**: Validate example config files are syntactically correct
- **Documentation Accuracy**: Verify workflow documentation matches actual template implementation

## Common Tasks

### Adding a New Workflow Template
1. Create workflow YAML file in `workflow-templates/`
2. Create corresponding `.properties.json` file for GitHub template system
3. Add documentation file in `docs/workflows/`
4. Update README.md table with new workflow
5. Add example configuration files to `config/examples/` if needed
6. Run validation: `yamllint -c .github/linters/.yaml-lint.yml workflow-templates/[new-file].yaml`

### Updating Existing Templates
1. Edit workflow file in `workflow-templates/`
2. Update corresponding documentation in `docs/workflows/`
3. Update README.md if description changes
4. Run full validation to ensure no syntax errors

### Configuration File Management
- Example configurations in `config/examples/` should be valid and current
- Always validate JSON/YAML syntax when updating config examples
- Test configurations with actual workflows when possible

## Repository Characteristics

- **No Traditional Build Process**: This repository contains templates and documentation, not compiled code
- **Validation is the "Build"**: The closest equivalent to building is running the linting and validation tools
- **Fast Operations**: All validation completes in under 10 seconds
- **GitHub Actions Focus**: Primary purpose is providing reusable GitHub Actions workflows
- **Multi-Language Support**: Templates support Java/Maven, Node.js/NPM, Python, Go, Docker, and more
- **Organization-Wide Impact**: Changes here affect CI/CD across all Netcracker repositories

## Time Expectations

- **Tool Installation**: 5-10 seconds for npm packages
- **YAML Validation**: 0.5 seconds for entire repository
- **Markdown Validation**: 2-3 seconds for all files
- **Link Checking**: 1-2 seconds per file
- **Full Validation Suite**: Under 10 seconds total
- NEVER CANCEL: All operations complete quickly. Use 30+ second timeouts to be safe.

## Known Issues and Limitations

- Some workflow templates intentionally have YAML style warnings (indentation)
- Link checking may fail on external URLs due to network restrictions
- Super-linter Docker image may not be accessible in all environments
- ActionLint tool installation may fail due to network restrictions
- Node modules should never be committed (use .gitignore)

## Reference Information

### Repository Root Contents
```
.editorconfig          # Editor configuration
.gitattributes         # Git file handling rules  
.github/               # Organization GitHub config and workflows
CODEOWNERS            # Code review assignments
LICENSE               # Apache 2.0 license
README.md             # Main documentation
config/               # Configuration examples and tools
docs/                 # Workflow documentation
workflow-templates/   # GitHub Actions workflow templates
```

### Essential Commands Reference
```bash
# Install validation tools (7 seconds)
npm install markdownlint-cli2 markdown-link-check --no-save

# Validate YAML files (0.5 seconds)
yamllint -c .github/linters/.yaml-lint.yml .

# Validate Markdown files (3 seconds)
npx markdownlint-cli2 --config .github/linters/.markdownlint.json "**/*.md"

# Check links (1.5 seconds)
npx markdown-link-check README.md

# Full validation sequence (under 10 seconds)
yamllint -c .github/linters/.yaml-lint.yml . && \
npx markdownlint-cli2 --config .github/linters/.markdownlint.json "*.md" && \
npx markdown-link-check README.md
```