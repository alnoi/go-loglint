package loglint

import (
	"go/ast"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const AnalyzerName = "loglint"

var Analyzer = &analysis.Analyzer{
	Name: AnalyzerName,
	Doc:  "checks log message conventions for slog and zap",
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
	Run: run,
}

var overrideCfg *Config
var configPath string

func init() {
	Analyzer.Flags.StringVar(&configPath, "config", "", "path to loglint YAML config file")
}

func run(pass *analysis.Pass) (any, error) {
	var cfg *Config
	if overrideCfg != nil {
		cfg = overrideCfg
	} else {
		loaded, err := LoadConfigFile(configPath)
		if err != nil {
			return nil, err
		}
		cfg = loaded
	}

	excludedByName := buildExcludedFiles(pass, cfg)

	ins := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	ins.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call := n.(*ast.CallExpr)

		if f := pass.Fset.File(call.Pos()); f != nil && excludedByName[f.Name()] {
			return
		}

		if _, ok := isLogCall(pass, call, cfg); !ok {
			return
		}

		if cfg.Rules.Sensitive {
			if pos, why, ok := containsSensitive(call.Args, cfg); ok {
				pass.Reportf(pos, "loglint(%s): %s", RuleSensitive, why)
			}
		}

		msg, pos, ok := extractMessage(pass, call.Args)
		if !ok {
			return
		}

		violations := validateMessage(msg, cfg)
		for _, v := range violations {
			pass.Reportf(pos, "loglint(%s): %s", v.Rule, v.Message)
		}
	})

	return nil, nil
}

func SetConfig(cfg *Config) {
	overrideCfg = cfg
}

func buildExcludedFiles(pass *analysis.Pass, cfg *Config) map[string]bool {
	out := make(map[string]bool, len(pass.Files))
	for _, f := range pass.Files {
		tf := pass.Fset.File(f.Pos())
		if tf == nil {
			continue
		}
		name := tf.Name()

		if containsAnyPathPart(name, cfg.Exclude.Paths) {
			out[name] = true
			continue
		}
		if matchAnyGlob(filepath.Base(name), cfg.Exclude.Files) {
			out[name] = true
			continue
		}

		out[name] = false
	}
	return out
}

func matchAnyGlob(name string, globs []string) bool {
	for _, g := range globs {
		g = strings.TrimSpace(g)
		if g == "" {
			continue
		}
		if ok, _ := filepath.Match(g, name); ok {
			return true
		}
	}
	return false
}

func containsAnyPathPart(path string, parts []string) bool {
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.Contains(path, p) {
			return true
		}
	}
	return false
}
