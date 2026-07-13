# groovy-check

Lint Groovy and Jenkinsfile code locally, and validate declarative pipeline
syntax against a live Jenkins server — from a single command-line tool.

`groovy-check` wraps two independent checks: local CodeNarc static analysis
(via [npm-groovy-lint](https://github.com/nvuillam/vscode-groovy-lint)), and
the Jenkins-maintained
[`pipeline-model-converter/validate`](https://www.jenkins.io/blog/2018/11/07/Validate-Jenkinsfile/)
endpoint. Point it at any Jenkins instance via config or environment
variables — no assumptions about your setup baked in.

## Features

- **`lint`** — local static analysis, no network or Jenkins access required
- **`validate`** — authoritative declarative-pipeline syntax check against a real Jenkins controller
- **`request`** — raw authenticated Jenkins API escape hatch (GET/HEAD by default; `--allow-write` for anything else)
- **`doctor`** — reports tool/config/auth status before you run into a confusing error
- `--json` on every command for machine-readable output

## Installation

### Homebrew (macOS and Linux)

```bash
brew install aviadshiber/tap/groovy-check
```

### Go Install

```bash
go install github.com/aviadshiber/groovy-check@latest
```

### Download Binary

Download the latest release from [GitHub Releases](https://github.com/aviadshiber/groovy-check/releases/latest).

### From Source

```bash
git clone https://github.com/aviadshiber/groovy-check.git
cd groovy-check
make build   # produces ./groovy-check
```

## Usage

```bash
# Check what's configured and available
groovy-check doctor

# Local lint — no auth needed
groovy-check lint path/to/file.groovy
groovy-check lint path/to/jenkins-shared-library/vars

# Validate a Jenkinsfile against a live Jenkins server
export JENKINS_URL=https://jenkins.example.com
export JENKINS_TOKEN=<your-api-token>
groovy-check validate path/to/Jenkinsfile

# Raw authenticated API call
groovy-check request GET /crumbIssuer/api/json
```

## Configuration

Precedence: `$JENKINS_TOKEN`/`$JENKINS_USER`/`$JENKINS_URL` environment
variables, then `~/.groovy-check/config.toml`:

```toml
jenkins_url = "https://jenkins.example.com"
jenkins_user = "alice"
```

Get a Jenkins API token from `<jenkins-url>/me/configure`.

## JSON output

- `doctor --json` → `{npm_groovy_lint, npx_available, jenkins_url, jenkins_user, token_source, jenkins_reachable, config_path}`
- `lint --json` → passes through npm-groovy-lint's own `--output json` schema
- `validate --json` → `{ok: bool, message: string}`
- `version --json` → `{version, commit, date}`
- Errors go to stderr as plain text (not JSON-wrapped) and set a non-zero exit code.

## What this does NOT catch

Neither `lint` nor `validate` understands the Jenkins **CPS transform**
(`workflow-cps` runtime closure/field-resolution behavior used by Jenkins
Shared Libraries). Both operate on static Groovy syntax/AST. A closure that
captures an enclosing-class field inside `withCredentials`, a custom retry
helper, or similar steps can pass both of these checks and still fail at
runtime under real Jenkins. Treat `groovy-check` as a fast pre-check, not a
substitute for testing shared-library changes against a real Jenkins
controller.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Note: this repository restricts
merges to `main` to the maintainer — all changes, including the
maintainer's own, go through a pull request; direct pushes and force-pushes
to `main` are blocked.

## License

[MIT](LICENSE)
