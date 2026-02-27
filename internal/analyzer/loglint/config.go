package loglint

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	// Rules toggles (enable/disable). No extra parameters beyond that.
	Rules RulesConfig `yaml:"rules"`

	// Sensitive data detection (rule 4).
	Sensitive SensitiveConfig `yaml:"sensitive"`

	// What is considered a logging call.
	LogAPIs []LogAPI `yaml:"logAPIs"`

	// Exclusions for files/paths.
	Exclude ExcludeConfig `yaml:"exclude"`
}

type RulesConfig struct {
	Lowercase   bool `yaml:"lowercase"`
	EnglishOnly bool `yaml:"englishOnly"`
	NoEmoji     bool `yaml:"noEmoji"`
	Sensitive   bool `yaml:"sensitive"`
}

type SensitiveConfig struct {
	// Substrings to match in identifiers/selectors (case-insensitive).
	Patterns []string `yaml:"patterns"`

	// If identifier/selectors contain any allowlisted substring, we skip reporting
	// (case-insensitive). Useful to suppress intentional names like "tokenBucket".
	Allowlist []string `yaml:"allowlist"`
}

type ExcludeConfig struct {
	// If file path contains any of these substrings, we skip the file.
	Paths []string `yaml:"paths"`

	// Glob patterns matched against the base filename (e.g. "*_gen.go", "*.pb.go").
	Files []string `yaml:"files"`
}

// LogAPI describes how to recognize logging calls.
// NOTE: Methods are stored as a slice for YAML friendliness and compiled into a set at runtime.
type LogAPI struct {
	// PackagePath matches package-level calls: <pkg>.<Method>(...)
	// Example: "log/slog" for slog.Info(...)
	PackagePath string `yaml:"packagePath"`

	// ReceiverPkgPath + ReceiverType match method calls on a receiver:
	// <recv>.<Method>(...) where <recv> has named type <ReceiverPkgPath>.<ReceiverType>.
	// Example: ReceiverPkgPath="go.uber.org/zap", ReceiverType="Logger".
	ReceiverPkgPath string `yaml:"receiverPkgPath"`
	ReceiverType    string `yaml:"receiverType"`

	// Methods is the list of method/function names considered logging for this API.
	Methods []string `yaml:"methods"`

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

func DefaultConfig() Config {
	// Unified set of log method names we recognize across supported loggers.
	// Safe to use as a superset: if a given logger doesn't have a method, there simply won't be such calls in AST.
	logMethods := []string{
		// Common across many loggers
		"Debug", "Info", "Warn", "Error",

		// zap Logger variants
		"DPanic", "Panic", "Fatal",

		// zap SugaredLogger variants
		"Debugf", "Infof", "Warnf", "Errorf", "DPanicf", "Panicf", "Fatalf",
		"Debugw", "Infow", "Warnw", "Errorw", "DPanicw", "Panicw", "Fatalw",
	}

	return Config{
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
			// slog package-level: slog.Info(...)
			{PackagePath: "log/slog", Methods: logMethods},
			// slog logger methods: logger.Info(...)
			{ReceiverPkgPath: "log/slog", ReceiverType: "Logger", Methods: logMethods},

			// zap logger methods: logger.Info(...)
			{ReceiverPkgPath: "go.uber.org/zap", ReceiverType: "Logger", Methods: logMethods},
			// zap sugared logger methods: sugar.Infof(...), sugar.Infow(...)
			{ReceiverPkgPath: "go.uber.org/zap", ReceiverType: "SugaredLogger", Methods: logMethods},
		},
		Exclude: ExcludeConfig{
			Paths: []string{"vendor" + string(filepath.Separator)},
			Files: []string{"*.pb.go", "*_gen.go"},
		},
	}
}

// Normalize prepares config for runtime use (compile method sets, normalize patterns).
func (c *Config) Normalize() {
	for i := range c.LogAPIs {
		c.LogAPIs[i].compile()
	}
	// Normalize patterns/allowlist by trimming and lowercasing (case-insensitive matching is done at use sites).
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

// LoadConfigFile loads configuration from a YAML file and merges it over defaults.
// Merge rules:
// - zero-value booleans override defaults (so specify them explicitly in YAML)
// - non-empty slices override defaults; empty/missing slices keep defaults.
func LoadConfigFile(path string) (Config, error) {
	if strings.TrimSpace(path) == "" {
		cfg := DefaultConfig()
		cfg.Normalize()
		return cfg, nil
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	// Use a separate type to distinguish "missing slice" from "empty slice".
	type fileCfg struct {
		Rules     *RulesConfig     `yaml:"rules"`
		Sensitive *SensitiveConfig `yaml:"sensitive"`
		LogAPIs   []LogAPI         `yaml:"logAPIs"`
		Exclude   *ExcludeConfig   `yaml:"exclude"`
	}

	var fc fileCfg
	if err := yaml.Unmarshal(b, &fc); err != nil {
		return Config{}, err
	}

	cfg := DefaultConfig()

	if fc.Rules != nil {
		cfg.Rules = *fc.Rules
	}
	if fc.Sensitive != nil {
		// Patterns/allowlist: if provided as non-empty, override; if empty, keep defaults.
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
		// Paths/files: if provided as non-empty, override; if empty, keep defaults.
		if len(fc.Exclude.Paths) > 0 {
			cfg.Exclude.Paths = fc.Exclude.Paths
		}
		if len(fc.Exclude.Files) > 0 {
			cfg.Exclude.Files = fc.Exclude.Files
		}
	}

	// Basic sanity: must have some log APIs to match on.
	if len(cfg.LogAPIs) == 0 {
		return Config{}, errors.New("config: logAPIs is empty")
	}

	cfg.Normalize()
	return cfg, nil
}
