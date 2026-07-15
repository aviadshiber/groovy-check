// Package lintconfig provides a locked-down default CodeNarc config for
// groovy-check's lint command, so npm-groovy-lint never auto-discovers (and
// therefore never require()s / executes) a .groovylintrc.js or similar file
// from the tree being linted.
package lintconfig

import _ "embed"

//go:embed default.json
var DefaultConfig []byte
