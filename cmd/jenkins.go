package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// overrideConfigPath is a test-only seam: when set, configPath() returns it
// instead of computing the real ~/.groovy-check/config.toml path. Production
// code never sets this.
var overrideConfigPath string

type fileConfig struct {
	JenkinsURL  string `toml:"jenkins_url"`
	JenkinsUser string `toml:"jenkins_user"`
}

type jenkinsConfig struct {
	URL        string
	User       string
	Token      string
	TokenFrom  string // "env" or "missing"
	ConfigPath string
}

func configPath() string {
	if overrideConfigPath != "" {
		return overrideConfigPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".groovy-check", "config.toml")
}

// loadJenkinsConfig resolves Jenkins connection details with no built-in
// default: URL and user must come from ~/.groovy-check/config.toml or
// $JENKINS_URL/$JENKINS_USER. This keeps groovy-check usable against any
// Jenkins instance rather than assuming one particular organization's setup.
func loadJenkinsConfig() jenkinsConfig {
	cfg := jenkinsConfig{
		TokenFrom:  "missing",
		ConfigPath: configPath(),
	}

	if u, err := user.Current(); err == nil {
		cfg.User = u.Username
	}

	if cfg.ConfigPath != "" {
		var fc fileConfig
		if _, err := toml.DecodeFile(cfg.ConfigPath, &fc); err == nil {
			if fc.JenkinsURL != "" {
				cfg.URL = fc.JenkinsURL
			}
			if fc.JenkinsUser != "" {
				cfg.User = fc.JenkinsUser
			}
		}
	}

	if url := os.Getenv("JENKINS_URL"); url != "" {
		cfg.URL = url
	}
	if u := os.Getenv("JENKINS_USER"); u != "" {
		cfg.User = u
	}
	if t := os.Getenv("JENKINS_TOKEN"); t != "" {
		cfg.Token = t
		cfg.TokenFrom = "env"
	}

	return cfg
}

func (c jenkinsConfig) requireToken() error {
	if c.Token == "" {
		return fmt.Errorf(
			"no Jenkins API token found (checked $JENKINS_TOKEN, then %s)\n"+
				"get one from <jenkins-url>/me/configure, then:\n"+
				"  export JENKINS_TOKEN=<token>\n"+
				"  export JENKINS_USER=%s   # if different from your local username",
			c.ConfigPath, c.User,
		)
	}
	return nil
}
