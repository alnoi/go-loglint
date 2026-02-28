package main

// TODO: suggestions, duplicated level content

import (
	"errors"

	"golang.org/x/tools/go/analysis"
	"gopkg.in/yaml.v3"

	"github.com/alnoi/go-loglint/internal/analyzer/loglint"
)

func New(conf any) ([]*analysis.Analyzer, error) {
	if conf == nil {
		return []*analysis.Analyzer{loglint.Analyzer}, nil
	}

	m, ok := conf.(map[string]any)
	if !ok {
		return nil, errors.New("loglint: invalid settings type")
	}

	var cfg *loglint.Config

	if v, ok := m["configPath"]; ok {
		if path, ok := v.(string); ok && path != "" {
			loaded, err := loglint.LoadConfigFile(path)
			if err != nil {
				return nil, err
			}
			cfg = loaded
		} else {
			cfg = loglint.DefaultConfig()
		}
		delete(m, "configPath")
	} else {
		cfg = loglint.DefaultConfig()
	}

	b, err := yaml.Marshal(m)
	if err != nil {
		return nil, err
	}
	if len(b) != 0 {
		if err := yaml.Unmarshal(b, cfg); err != nil {
			return nil, err
		}
	}

	cfg.Normalize()
	loglint.SetConfig(cfg)

	return []*analysis.Analyzer{loglint.Analyzer}, nil
}
