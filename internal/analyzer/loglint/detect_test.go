package loglint

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/analysis"
)

func typecheckFile(t *testing.T, src string) (*analysis.Pass, *ast.File, *token.FileSet) {
	t.Helper()

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "p.go", src, parser.ParseComments)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Uses:  make(map[*ast.Ident]types.Object),
		Defs:  make(map[*ast.Ident]types.Object),
	}

	conf := &types.Config{
		Importer: importer.Default(),
	}

	_, err = conf.Check("p", fset, []*ast.File{f}, info)
	if err != nil {
		t.Fatalf("typecheck: %v", err)
	}

	pass := &analysis.Pass{
		Fset:      fset,
		TypesInfo: info,
	}

	return pass, f, fset
}

func findFirstCall(t *testing.T, f *ast.File) *ast.CallExpr {
	t.Helper()

	var call *ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		if call != nil {
			return false
		}
		if c, ok := n.(*ast.CallExpr); ok {
			call = c
			return false
		}
		return true
	})
	if call == nil {
		t.Fatalf("no CallExpr found in AST")
	}
	return call
}

func mustParseExpr(t *testing.T, s string) ast.Expr {
	t.Helper()
	e, err := parser.ParseExpr(s)
	if err != nil {
		t.Fatalf("ParseExpr(%q): %v", s, err)
	}
	return e
}

func TestExtractMessage_EmptyArgs(t *testing.T) {
	msg, pos, ok := extractMessage(nil, nil)
	if ok || msg != "" || pos != token.NoPos {
		t.Fatalf("expected empty result, got msg=%q pos=%v ok=%v", msg, pos, ok)
	}
}

func TestExtractMessage_NoPassOrNoTypesInfo(t *testing.T) {
	args := []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"hi"`}}

	msg, pos, ok := extractMessage(nil, args)
	if ok || msg != "" || pos != token.NoPos {
		t.Fatalf("expected false without pass, got msg=%q pos=%v ok=%v", msg, pos, ok)
	}

	pass := &analysis.Pass{TypesInfo: nil}
	msg, pos, ok = extractMessage(pass, args)
	if ok || msg != "" || pos != token.NoPos {
		t.Fatalf("expected false without TypesInfo, got msg=%q pos=%v ok=%v", msg, pos, ok)
	}
}

func TestExtractMessage_ConstStringFromTypesInfo(t *testing.T) {
	pass, f, _ := typecheckFile(t, `
package p

import "log/slog"

func f() {
	slog.Info("Hello")
}
`)
	call := findFirstCall(t, f)

	msg, pos, ok := extractMessage(pass, call.Args)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if msg != "Hello" {
		t.Fatalf("expected %q, got %q", "Hello", msg)
	}
	if pos == token.NoPos {
		t.Fatalf("expected non-NoPos")
	}
}

func TestExtractMessage_NonConstString(t *testing.T) {
	pass, f, _ := typecheckFile(t, `
package p

import "log/slog"

func f() {
	msg := "Hello"
	slog.Info(msg)
}
`)
	call := findFirstCall(t, f)

	msg, pos, ok := extractMessage(pass, call.Args)
	if ok || msg != "" || pos != token.NoPos {
		t.Fatalf("expected non-const to fail, got msg=%q pos=%v ok=%v", msg, pos, ok)
	}
}

func TestContainsSensitive_ScansAllArgs(t *testing.T) {
	var cfg Config
	cfg.Sensitive.Patterns = []string{"password"}

	args := []ast.Expr{
		mustParseExpr(t, "x"),
		mustParseExpr(t, "password"),
		mustParseExpr(t, "y"),
	}

	pos, why, ok := containsSensitive(args, cfg)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if pos == token.NoPos {
		t.Fatalf("expected pos != NoPos")
	}
	if why == "" {
		t.Fatalf("expected non-empty why")
	}
}

func TestContainsSensitive_NoMatch(t *testing.T) {
	var cfg Config
	cfg.Sensitive.Patterns = []string{"password"}

	args := []ast.Expr{
		mustParseExpr(t, "x"),
		mustParseExpr(t, "user.name"),
		mustParseExpr(t, "zap.String(\"k\", v)"),
	}

	pos, why, ok := containsSensitive(args, cfg)
	if ok || pos != token.NoPos || why != "" {
		t.Fatalf("expected no match, got ok=%v pos=%v why=%q", ok, pos, why)
	}
}

func TestIsPkgSelector(t *testing.T) {
	pass, f, _ := typecheckFile(t, `
package p

import "log/slog"

func f() {
	slog.Info("x")
}
`)
	call := findFirstCall(t, f)
	sel := call.Fun.(*ast.SelectorExpr)

	if !isPkgSelector(pass, sel, "log/slog") {
		t.Fatalf("expected true for slog package selector")
	}
	if isPkgSelector(pass, sel, "go.uber.org/zap") {
		t.Fatalf("expected false for wrong package path")
	}

	sel2 := &ast.SelectorExpr{
		X:   &ast.BasicLit{Kind: token.INT, Value: "1"},
		Sel: ast.NewIdent("Info"),
	}
	if isPkgSelector(pass, sel2, "log/slog") {
		t.Fatalf("expected false when X is not Ident")
	}
	if isPkgSelector(nil, sel, "log/slog") {
		t.Fatalf("expected false when pass is nil")
	}
	if isPkgSelector(&analysis.Pass{TypesInfo: nil}, sel, "log/slog") {
		t.Fatalf("expected false when TypesInfo is nil")
	}
}

func TestTypeOf(t *testing.T) {
	pass, f, _ := typecheckFile(t, `
package p

import "log/slog"

func f() {
	logger := slog.Default()
	logger.Info("x")
}
`)
	var call *ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if sel, ok := c.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Info" {
				call = c
				return false
			}
		}
		return true
	})
	if call == nil {
		t.Fatalf("no logger.Info call found")
	}

	sel := call.Fun.(*ast.SelectorExpr)
	tp := typeOf(pass, sel.X)
	if tp == nil {
		t.Fatalf("expected non-nil type")
	}

	if typeOf(nil, sel.X) != nil {
		t.Fatalf("expected nil when pass is nil")
	}
	if typeOf(&analysis.Pass{TypesInfo: nil}, sel.X) != nil {
		t.Fatalf("expected nil when TypesInfo is nil")
	}
	if typeOf(pass, &ast.BasicLit{Kind: token.INT, Value: "1"}) != nil {
		t.Fatalf("expected nil when expr not in TypesInfo")
	}
}

func TestIsNamedTypeFromPkg(t *testing.T) {
	pass, f, _ := typecheckFile(t, `
package p

import "log/slog"

func f() {
	logger := slog.Default()
	logger.Info("x")
}
`)

	var loggerIdent *ast.Ident
	ast.Inspect(f, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok && sel.Sel.Name == "Info" {
			if id, ok := sel.X.(*ast.Ident); ok {
				loggerIdent = id
				return false
			}
		}
		return true
	})
	if loggerIdent == nil {
		t.Fatalf("logger ident not found")
	}

	loggerType := typeOf(pass, loggerIdent)
	if loggerType == nil {
		t.Fatalf("expected loggerType != nil")
	}

	if !isNamedTypeFromPkg(loggerType, "log/slog", "Logger") {
		t.Fatalf("expected true for *slog.Logger")
	}

	// Wrong pkg
	if isNamedTypeFromPkg(loggerType, "go.uber.org/zap", "Logger") {
		t.Fatalf("expected false for wrong pkg")
	}
	// Wrong type name
	if isNamedTypeFromPkg(loggerType, "log/slog", "SugaredLogger") {
		t.Fatalf("expected false for wrong type name")
	}

	// nil type
	if isNamedTypeFromPkg(nil, "log/slog", "Logger") {
		t.Fatalf("expected false for nil type")
	}

	// non-named types: e.g. slice
	if isNamedTypeFromPkg(types.NewSlice(types.Typ[types.Int]), "log/slog", "Logger") {
		t.Fatalf("expected false for non-named type")
	}

	// named but without pkg: create a named type in universe scope
	n := types.NewNamed(types.NewTypeName(token.NoPos, nil, "X", nil), types.Typ[types.Int], nil)
	if isNamedTypeFromPkg(n, "log/slog", "X") {
		t.Fatalf("expected false when named type has nil pkg")
	}
}

func TestIsLogCall_PackageSelector(t *testing.T) {
	pass, f, _ := typecheckFile(t, `
package p

import "log/slog"

func f() {
	slog.Info("x")
}
`)
	call := findFirstCall(t, f)

	cfg := Config{
		LogAPIs: []LogAPI{
			{
				PackagePath:     "log/slog",
				compiledMethods: map[string]struct{}{"Info": {}},
			},
		},
	}

	method, ok := isLogCall(pass, call, cfg)
	if !ok || method != "Info" {
		t.Fatalf("expected Info true, got method=%q ok=%v", method, ok)
	}
}

func TestIsLogCall_ReceiverType(t *testing.T) {
	pass, f, _ := typecheckFile(t, `
package p

import "log/slog"

func f() {
	logger := slog.Default()
	logger.Info("x")
}
`)
	var call *ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if sel, ok := c.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Info" {
				if _, isIdent := sel.X.(*ast.Ident); isIdent {
					call = c
					return false
				}
			}
		}
		return true
	})
	if call == nil {
		t.Fatalf("logger.Info call not found")
	}

	cfg := Config{
		LogAPIs: []LogAPI{
			{
				ReceiverPkgPath: "log/slog",
				ReceiverType:    "Logger",
				compiledMethods: map[string]struct{}{"Info": {}},
			},
		},
	}

	method, ok := isLogCall(pass, call, cfg)
	if !ok || method != "Info" {
		t.Fatalf("expected Info true by receiver, got method=%q ok=%v", method, ok)
	}
}

func TestIsLogCall_Negatives(t *testing.T) {
	pass, f, _ := typecheckFile(t, `
package p

import "log/slog"

func f() {
	slog.Info("x")
}
`)
	call := findFirstCall(t, f)

	cfg := Config{
		LogAPIs: []LogAPI{
			{
				PackagePath:     "log/slog",
				compiledMethods: map[string]struct{}{"Error": {}},
			},
		},
	}

	_, ok := isLogCall(pass, call, cfg)
	if ok {
		t.Fatalf("expected false when method is not enabled")
	}

	// call.Fun not a selector
	call2 := &ast.CallExpr{Fun: ast.NewIdent("Info")}
	_, ok = isLogCall(pass, call2, cfg)
	if ok {
		t.Fatalf("expected false when Fun is not SelectorExpr")
	}

	cfg2 := Config{
		LogAPIs: []LogAPI{
			{
				PackagePath:     "log/slog",
				compiledMethods: map[string]struct{}{"Info": {}},
			},
		},
	}
	_, ok = isLogCall(nil, call, cfg2)
	if ok {
		t.Fatalf("expected false when pass is nil (no type info)")
	}
}
