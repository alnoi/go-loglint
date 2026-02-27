package loglint

import (
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	td := analysistest.TestData()

	tests := []struct {
		name string
		cfg  string
		pkgs []string
	}{
		{
			name: "basic",
			cfg:  "cfg/default.yaml",
			pkgs: []string{"basic"},
		},
		{
			name: "apis",
			cfg:  "cfg/custom_apis.yaml",
			pkgs: []string{"apis"},
		},
		{
			name: "constmsg",
			cfg:  "cfg/default.yaml",
			pkgs: []string{"constmsg"},
		},
		{
			name: "sensitive",
			cfg:  "cfg/only_sensitive.yaml",
			pkgs: []string{"sensitive"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath = filepath.Join(td, tt.cfg)
			analysistest.Run(t, td, Analyzer, tt.pkgs...)
		})
	}
}
