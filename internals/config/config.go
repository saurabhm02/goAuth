package config

import (
	"fmt"
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
}

type DatabaseConfig struct {
	DSN       string `yaml:"dsn"`
	UserTable string `yaml:"user_table"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
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
		c.Projects[id] = p
	}
}

// applyEnvOverrides fills DSN from env when empty (e.g. DATABASE_DSN for project "default").
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
