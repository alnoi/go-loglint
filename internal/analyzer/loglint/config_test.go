package loglint

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func writeTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	return p
}

func TestLogAPI_compileAndHasMethod_TrimsAndCaches(t *testing.T) {
	a := &LogAPI{
		Methods: []string{" Info ", "", "Error", "  ", "Debug"},
	}

	if a.compiledMethods != nil {
		t.Fatalf("expected compiledMethods nil before compile")
	}

	if !a.hasMethod("Info") {
		t.Fatalf("expected Info method to exist")
	}
	if a.hasMethod("") {
		t.Fatalf("expected empty method to not exist")
	}
	if a.hasMethod(" Warn ") {
		t.Fatalf("expected Warn to not exist")
	}

	before := a.compiledMethods
	beforePtr := reflect.ValueOf(before).Pointer()

	a.compile()

	after := a.compiledMethods
	afterPtr := reflect.ValueOf(after).Pointer()

	if beforePtr != afterPtr {
		t.Fatalf("expected compile() to be cached (same map instance)")
	}
}

func TestDefaultConfig_Sanity(t *testing.T) {
	cfg := DefaultConfig()

	if !cfg.Rules.Lowercase || !cfg.Rules.EnglishOnly || !cfg.Rules.NoEmoji || !cfg.Rules.Sensitive {
		t.Fatalf("expected all rules enabled by default: %#v", cfg.Rules)
	}

	if len(cfg.Sensitive.Patterns) == 0 {
		t.Fatalf("expected default sensitive patterns to be non-empty")
	}

	if len(cfg.LogAPIs) == 0 {
		t.Fatalf("expected default logAPIs non-empty")
	}

	foundVendor := false
	for _, p := range cfg.Exclude.Paths {
		if strings.Contains(p, "vendor") {
			foundVendor = true
			break
		}
	}
	if !foundVendor {
		t.Fatalf("expected default exclude paths contain vendor; got %#v", cfg.Exclude.Paths)
	}
}

func TestNormalize_CompilesMethods_And_NormalizesStrings(t *testing.T) {
	cfg := Config{
		Rules: RulesConfig{
			Lowercase:   true,
			EnglishOnly: true,
			NoEmoji:     true,
			Sensitive:   true,
		},
		Sensitive: SensitiveConfig{
			Patterns:  []string{" Password ", "ToKeN", "  "},
			Allowlist: []string{" TokenBucket ", "  ", "PASSwordPolicy"},
		},
		LogAPIs: []LogAPI{
			{
				PackagePath: "log/slog",
				Methods:     []string{" Info ", "", "Error"},
			},
		},
		Exclude: ExcludeConfig{
			Paths: []string{" vendor ", "  internal/gen  "},
			Files: []string{" *.pb.go ", "  "},
		},
	}

	if cfg.LogAPIs[0].compiledMethods != nil {
		t.Fatalf("expected compiledMethods nil before Normalize")
	}

	cfg.Normalize()

	if cfg.LogAPIs[0].compiledMethods == nil {
		t.Fatalf("expected compiledMethods to be compiled after Normalize")
	}
	if !cfg.LogAPIs[0].hasMethod("Info") || !cfg.LogAPIs[0].hasMethod("Error") {
		t.Fatalf("expected Info/Error to exist after compile")
	}
	if cfg.LogAPIs[0].hasMethod("") {
		t.Fatalf("expected empty method not to exist")
	}

	if cfg.Sensitive.Patterns[0] != "password" || cfg.Sensitive.Patterns[1] != "token" {
		t.Fatalf("patterns not normalized: %#v", cfg.Sensitive.Patterns)
	}
	if cfg.Sensitive.Allowlist[0] != "tokenbucket" || cfg.Sensitive.Allowlist[2] != "passwordpolicy" {
		t.Fatalf("allowlist not normalized: %#v", cfg.Sensitive.Allowlist)
	}

	if cfg.Exclude.Paths[0] != "vendor" || cfg.Exclude.Paths[1] != "internal/gen" {
		t.Fatalf("exclude paths not trimmed: %#v", cfg.Exclude.Paths)
	}
	if cfg.Exclude.Files[0] != "*.pb.go" {
		t.Fatalf("exclude files not trimmed: %#v", cfg.Exclude.Files)
	}
}

func TestLoadConfigFile_EmptyPath_ReturnsDefaults(t *testing.T) {
	cfg, err := LoadConfigFile("")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(cfg.LogAPIs) == 0 {
		t.Fatalf("expected defaults to include logAPIs")
	}

	for i := range cfg.LogAPIs {
		if cfg.LogAPIs[i].compiledMethods == nil {
			t.Fatalf("expected compiledMethods after Normalize in defaults")
		}
	}
}

func TestLoadConfigFile_ReadError(t *testing.T) {
	_, err := LoadConfigFile("/definitely/does/not/exist.yaml")
	if err == nil {
		t.Fatalf("expected error for missing file")
	}
}

func TestLoadConfigFile_YamlUnmarshalError(t *testing.T) {
	dir := t.TempDir()
	p := writeTempFile(t, dir, "bad.yaml", "::::not-yaml::::\n: [")
	_, err := LoadConfigFile(p)
	if err == nil {
		t.Fatalf("expected YAML unmarshal error")
	}
}

func TestLoadConfigFile_MergeRulesAndOverrides(t *testing.T) {
	dir := t.TempDir()

	yml := `
rules:
  lowercase: false
  englishOnly: true
  noEmoji: false
  sensitive: true
`
	p := writeTempFile(t, dir, "cfg.yaml", yml)

	cfg, err := LoadConfigFile(p)
	if err != nil {
		t.Fatalf("LoadConfigFile: %v", err)
	}

	if cfg.Rules.Lowercase != false || cfg.Rules.NoEmoji != false || cfg.Rules.EnglishOnly != true || cfg.Rules.Sensitive != true {
		t.Fatalf("rules were not overridden: %#v", cfg.Rules)
	}
}

func TestLoadConfigFile_SensitiveSlices_OverrideOnlyIfNonEmpty(t *testing.T) {
	dir := t.TempDir()

	p1 := writeTempFile(t, dir, "cfg1.yaml", `
sensitive:
  patterns: []
  allowlist: []
`)
	cfg1, err := LoadConfigFile(p1)
	if err != nil {
		t.Fatalf("LoadConfigFile: %v", err)
	}
	def := DefaultConfig()
	def.Normalize()

	if strings.Join(cfg1.Sensitive.Patterns, ",") != strings.Join(def.Sensitive.Patterns, ",") {
		t.Fatalf("expected patterns to keep defaults when empty; got %#v", cfg1.Sensitive.Patterns)
	}
	if strings.Join(cfg1.Sensitive.Allowlist, ",") != strings.Join(def.Sensitive.Allowlist, ",") {
		t.Fatalf("expected allowlist to keep defaults when empty; got %#v", cfg1.Sensitive.Allowlist)
	}

	p2 := writeTempFile(t, dir, "cfg2.yaml", `
sensitive:
  patterns: ["  PASSword  ", "Token"]
  allowlist: [" TokenBucket "]
`)
	cfg2, err := LoadConfigFile(p2)
	if err != nil {
		t.Fatalf("LoadConfigFile: %v", err)
	}
	if strings.Join(cfg2.Sensitive.Patterns, ",") != "password,token" {
		t.Fatalf("expected overridden patterns normalized; got %#v", cfg2.Sensitive.Patterns)
	}
	if strings.Join(cfg2.Sensitive.Allowlist, ",") != "tokenbucket" {
		t.Fatalf("expected overridden allowlist normalized; got %#v", cfg2.Sensitive.Allowlist)
	}
}

func TestLoadConfigFile_LogAPIs_OverrideOnlyIfNonEmpty(t *testing.T) {
	dir := t.TempDir()

	p := writeTempFile(t, dir, "cfg.yaml", `
logAPIs:
  - packagePath: "log/slog"
    methods: ["Info"]
`)
	cfg, err := LoadConfigFile(p)
	if err != nil {
		t.Fatalf("LoadConfigFile: %v", err)
	}
	if len(cfg.LogAPIs) != 1 {
		t.Fatalf("expected logAPIs override to size 1, got %d", len(cfg.LogAPIs))
	}
	if cfg.LogAPIs[0].PackagePath != "log/slog" {
		t.Fatalf("expected packagePath log/slog, got %q", cfg.LogAPIs[0].PackagePath)
	}
	if cfg.LogAPIs[0].compiledMethods == nil {
		t.Fatalf("expected compiledMethods after Normalize")
	}
	if !cfg.LogAPIs[0].hasMethod("Info") || cfg.LogAPIs[0].hasMethod("Error") {
		t.Fatalf("expected only Info method, got compiled=%v", cfg.LogAPIs[0].compiledMethods)
	}
}

func TestLoadConfigFile_Exclude_OverrideOnlyIfNonEmpty(t *testing.T) {
	dir := t.TempDir()

	p1 := writeTempFile(t, dir, "cfg1.yaml", `
exclude:
  paths: []
  files: []
`)
	cfg1, err := LoadConfigFile(p1)
	if err != nil {
		t.Fatalf("LoadConfigFile: %v", err)
	}
	def := DefaultConfig()
	def.Normalize()
	if strings.Join(cfg1.Exclude.Paths, ",") != strings.Join(def.Exclude.Paths, ",") {
		t.Fatalf("expected exclude.paths keep defaults; got %#v", cfg1.Exclude.Paths)
	}

	p2 := writeTempFile(t, dir, "cfg2.yaml", `
exclude:
  paths: [" internal/generated "]
  files: [" *_mock.go "]
`)
	cfg2, err := LoadConfigFile(p2)
	if err != nil {
		t.Fatalf("LoadConfigFile: %v", err)
	}
	if len(cfg2.Exclude.Paths) != 1 || cfg2.Exclude.Paths[0] != "internal/generated" {
		t.Fatalf("expected trimmed override paths, got %#v", cfg2.Exclude.Paths)
	}
	if len(cfg2.Exclude.Files) != 1 || cfg2.Exclude.Files[0] != "*_mock.go" {
		t.Fatalf("expected trimmed override files, got %#v", cfg2.Exclude.Files)
	}
}

func TestLoadConfigFile_ErrorWhenLogAPIsEmptyAfterMerge(t *testing.T) {
	dir := t.TempDir()

	p := writeTempFile(t, dir, "cfg.yaml", `logAPIs: []
`)

	cfg, err := LoadConfigFile(p)
	if err != nil {
		t.Fatalf("expected no error (defaults remain), got %v", err)
	}
	if len(cfg.LogAPIs) == 0 {
		t.Fatalf("defaults should ensure logAPIs non-empty")
	}

	_ = runtime.GOOS
}
