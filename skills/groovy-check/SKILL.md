---
name: groovy-check
description: Lint Groovy/Jenkinsfile code and validate declarative pipeline syntax against a live Jenkins server via the `groovy-check` CLI. Use when editing or reviewing Groovy files, Jenkinsfiles, or Jenkins Shared Library code, or when asked to lint/validate Groovy or check Jenkinsfile syntax.
---

# groovy-check

Wraps two independent Groovy/Jenkinsfile checks behind one CLI: local CodeNarc
linting (no network) and authoritative declarative-pipeline validation against
a real Jenkins server.

## Verify the CLI is installed

```bash
command -v groovy-check || brew install aviadshiber/tap/groovy-check
```

## Run doctor first

```bash
groovy-check --json doctor
```

Reports npm-groovy-lint availability, Jenkins URL/reachability, and whether a
Jenkins token is configured.

## Safe read path: local lint (no auth, no network)

```bash
groovy-check lint path/to/file.groovy
groovy-check lint path/to/jenkins-shared-library/vars
```

## Authoritative validation (needs Jenkins auth)

```bash
groovy-check validate path/to/Jenkinsfile
```

Auth: `$JENKINS_TOKEN`/`$JENKINS_USER`/`$JENKINS_URL` env vars, or
`~/.groovy-check/config.toml` (`jenkins_url`, `jenkins_user`). Get a token
from `<jenkins-url>/me/configure`.

## Raw escape hatch

```bash
groovy-check request GET /job/some-job/api/json
```

Non-GET/HEAD requires `--allow-write` — this hits a live Jenkins instance, so
only pass it when a write was explicitly requested.

## What NOT to treat this as a substitute for

`lint` and `validate` both do static Groovy/DSL analysis. **Neither
understands the Jenkins CPS transform** (`workflow-cps` runtime closure/
field-resolution behavior used by Jenkins Shared Libraries). A closure that
captures an enclosing-class field inside `withCredentials`, a custom retry
helper, or similar steps can pass both checks and still fail at runtime under
real Jenkins. Run `groovy-check lint`/`validate` as a fast pre-check, then
still test shared-library changes against a real Jenkins controller before
merging.

## Examples

```bash
groovy-check lint ./vars --json
groovy-check validate ./Jenkinsfile
groovy-check request GET /crumbIssuer/api/json
```
