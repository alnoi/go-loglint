package loglint

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Rules     RulesConfig     `yaml:"rules"`
	Sensitive SensitiveConfig `yaml:"sensitive"`
	LogAPIs   []LogAPI        `yaml:"logAPIs"`
	Exclude   ExcludeConfig   `yaml:"exclude"`
}

type RulesConfig struct {
	Lowercase   bool `yaml:"lowercase"`
	EnglishOnly bool `yaml:"englishOnly"`
	NoEmoji     bool `yaml:"noEmoji"`
	Sensitive   bool `yaml:"sensitive"`
}

type SensitiveConfig struct {
	Patterns  []string `yaml:"patterns"`
	Allowlist []string `yaml:"allowlist"`
}

type ExcludeConfig struct {
	Paths []string `yaml:"paths"`
	Files []string `yaml:"files"`
}

type LogAPI struct {
	PackagePath     string   `yaml:"packagePath"`
	ReceiverPkgPath string   `yaml:"receiverPkgPath"`
	ReceiverType    string   `yaml:"receiverType"`
	Methods         []string `yaml:"methods"`

	compiledMethods map[string]struct{} `yaml:"-"`
}

func (a *LogAPI) compile() {
	if a.compiledMethods != nil {
		return
	}
	m := make(map[string]struct{}, len(a.Methods))
	for _, n := range a.Methods {
		n = strings.TrimSpace(n)
		if n == "" {
			continue
		}
		m[n] = struct{}{}
	}
	a.compiledMethods = m
}

func (a *LogAPI) hasMethod(name string) bool {
	a.compile()
	_, ok := a.compiledMethods[name]
	return ok
}

func DefaultConfig() *Config {
	logMethods := []string{
		"Debug", "Info", "Warn", "Error",
		"DPanic", "Panic", "Fatal",
		"Debugf", "Infof", "Warnf", "Errorf", "DPanicf", "Panicf", "Fatalf",
		"Debugw", "Infow", "Warnw", "Errorw", "DPanicw", "Panicw", "Fatalw",
	}

	return &Config{
		Rules: RulesConfig{
			Lowercase:   true,
			EnglishOnly: true,
			NoEmoji:     true,
			Sensitive:   true,
		},
		Sensitive: SensitiveConfig{
			Patterns: []string{
				"password", "passwd", "token", "secret", "api_key", "apikey", "authorization", "bearer",
			},
			Allowlist: []string{},
		},
		LogAPIs: []LogAPI{
			{PackagePath: "log/slog", ReceiverPkgPath: "log/slog", ReceiverType: "Logger", Methods: logMethods},
			{ReceiverPkgPath: "go.uber.org/zap", ReceiverType: "Logger", Methods: logMethods},
			{ReceiverPkgPath: "go.uber.org/zap", ReceiverType: "SugaredLogger", Methods: logMethods},
		},
		Exclude: ExcludeConfig{
			Paths: []string{"vendor" + string(filepath.Separator)},
			Files: []string{"*.pb.go", "*_gen.go"},
		},
	}
}

func (c *Config) Normalize() {
	for i := range c.LogAPIs {
		c.LogAPIs[i].compile()
	}
	for i := range c.Sensitive.Patterns {
		c.Sensitive.Patterns[i] = strings.ToLower(strings.TrimSpace(c.Sensitive.Patterns[i]))
	}
	for i := range c.Sensitive.Allowlist {
		c.Sensitive.Allowlist[i] = strings.ToLower(strings.TrimSpace(c.Sensitive.Allowlist[i]))
	}
	for i := range c.Exclude.Paths {
		c.Exclude.Paths[i] = strings.TrimSpace(c.Exclude.Paths[i])
	}
	for i := range c.Exclude.Files {
		c.Exclude.Files[i] = strings.TrimSpace(c.Exclude.Files[i])
	}
}

func LoadConfigFile(path string) (*Config, error) {
	if strings.TrimSpace(path) == "" {
		cfg := DefaultConfig()
		cfg.Normalize()
		return cfg, nil
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	type fileCfg struct {
		Rules     *RulesConfig     `yaml:"rules"`
		Sensitive *SensitiveConfig `yaml:"sensitive"`
		LogAPIs   []LogAPI         `yaml:"logAPIs"`
		Exclude   *ExcludeConfig   `yaml:"exclude"`
	}

	var fc fileCfg
	if err := yaml.Unmarshal(b, &fc); err != nil {
		return nil, err
	}

	cfg := DefaultConfig()

	if fc.Rules != nil {
		cfg.Rules = *fc.Rules
	}
	if fc.Sensitive != nil {
		if len(fc.Sensitive.Patterns) > 0 {
			cfg.Sensitive.Patterns = fc.Sensitive.Patterns
		}
		if len(fc.Sensitive.Allowlist) > 0 {
			cfg.Sensitive.Allowlist = fc.Sensitive.Allowlist
		}
	}
	if len(fc.LogAPIs) > 0 {
		cfg.LogAPIs = fc.LogAPIs
	}
	if fc.Exclude != nil {
		if len(fc.Exclude.Paths) > 0 {
			cfg.Exclude.Paths = fc.Exclude.Paths
		}
		if len(fc.Exclude.Files) > 0 {
			cfg.Exclude.Files = fc.Exclude.Files
		}
	}

	if len(cfg.LogAPIs) == 0 {
		return nil, errors.New("config: logAPIs is empty")
	}

	cfg.Normalize()
	return cfg, nil
}
