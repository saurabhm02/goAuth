package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"goAuth/internals/types"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Projects map[string]ProjectConfig `yaml:"projects"`
}

type ProjectConfig struct {
	Database DatabaseConfig `yaml:"database"`
	OTP      bool           `yaml:"otp"`
}

type DatabaseConfig struct {
	DSN       string `yaml:"dsn"`
	UserTable string `yaml:"user_table"`
	OtpTable  string `yaml:"otp_table"` // used when project has otp: true
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	expandedData := os.ExpandEnv(string(data))
	if err := yaml.Unmarshal([]byte(expandedData), &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	applyDefaults(&cfg)
	applyEnvOverrides(&cfg)
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func applyDefaults(c *Config) {
	for id, p := range c.Projects {
		if p.Database.UserTable == "" {
			p.Database.UserTable = "users"
		}
		if p.Database.OtpTable == "" {
			p.Database.OtpTable = "otp"
		}
		c.Projects[id] = p
	}
}

// applyEnvOverrides fills DSN from env when empty; optionally appends sslrootcert if DATABASE_SSL_ROOT_CERT is set.
func applyEnvOverrides(c *Config) {
	for id, p := range c.Projects {
		if p.Database.DSN == "" {
			key := types.EnvDatabaseDSN
			if id != types.DefaultProjectID {
				key = "DATABASE_DSN_" + strings.ToUpper(strings.ReplaceAll(id, "-", "_"))
			}
			if v := os.Getenv(key); v != "" {
				p.Database.DSN = v
			}
		}
		if ca := os.Getenv(types.EnvDatabaseSSLRootCert); ca != "" {
			dsn := p.Database.DSN
			caParam := "sslrootcert=" + url.QueryEscape(ca)
			if strings.Contains(dsn, "?") {
				dsn += "&" + caParam
			} else {
				dsn += "?" + caParam
			}
			p.Database.DSN = dsn
		}
		c.Projects[id] = p
	}
}

func validate(c *Config) error {
	if len(c.Projects) == 0 {
		return fmt.Errorf("at least one project must be defined in projects")
	}
	for projectID, p := range c.Projects {
		if p.Database.DSN == "" {
			return fmt.Errorf("project %s: database dsn is required", projectID)
		}
	}
	return nil
}
