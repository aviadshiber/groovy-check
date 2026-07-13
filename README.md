# groovy-check

Local Groovy/Jenkinsfile linting plus authoritative declarative-pipeline validation
against a live Jenkins server. Built for `~/git/pipeline-libs/` (`vars/`, `src/`) but
works on any Groovy/Jenkinsfile path.

## Install

```bash
make install   # builds and copies to ~/.local/bin/groovy-check
```

## Commands

| Command | Network? | Auth? | What it does |
|---|---|---|---|
| `groovy-check doctor` | Jenkins reachability only | no | reports tool/config/auth status |
| `groovy-check lint [path]` | no | no | CodeNarc static analysis via `npx npm-groovy-lint` |
| `groovy-check validate <Jenkinsfile>` | yes | yes | POSTs to `<jenkins_url>/pipeline-model-converter/validate` (official Jenkins endpoint) |
| `groovy-check request <method> <path>` | yes | yes | raw authenticated Jenkins API call; non-GET/HEAD needs `--allow-write` |

All commands accept `--json` for machine-readable output.

## Auth

Precedence: `$JENKINS_TOKEN`/`$JENKINS_USER` env vars, then `~/.groovy-check/config.toml`
(`jenkins_url`, `jenkins_user`), then the `jenkins_url` default
(`https://ci.taboolasyndication.com`).

Get a token via the `taboola-service-auth` skill (this CLI never reads the keychain
directly — see `~/.claude/rules/agent-security.md`), then:

```bash
export JENKINS_TOKEN=<token>
```

## JSON output

- `doctor --json` → `{npm_groovy_lint, npx_available, jenkins_url, jenkins_user, token_source, jenkins_reachable, config_path}`
- `lint --json` → passes through npm-groovy-lint's own `--output json` schema
- `validate --json` → `{ok: bool, message: string}`
- Errors go to stderr as plain text (not JSON-wrapped) and set a non-zero exit code.

## What this does NOT catch

Neither `lint` nor `validate` understands the Jenkins **CPS transform**
(`workflow-cps` runtime closure/field-resolution behavior — see
`~/.claude/rules/jenkins-shared-library-cps.md`). Both operate on static Groovy
syntax/AST. A closure that captures an enclosing-class field inside
`withCredentials`/`Retrier.retry`/etc. can pass both of these checks and still fail
100% at runtime under real Jenkins (see DEV-227939/228250). The real-Jenkins sandbox
gate in that rule file is still mandatory before merging such changes.
