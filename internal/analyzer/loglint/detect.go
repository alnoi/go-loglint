package loglint

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

func extractMessage(pass *analysis.Pass, args []ast.Expr) (string, token.Pos, bool) {
	if len(args) == 0 {
		return "", token.NoPos, false
	}

	msgExpr := args[0]

	if pass != nil && pass.TypesInfo != nil {
		if tv, ok := pass.TypesInfo.Types[msgExpr]; ok && tv.Value != nil {
			if tv.Value.Kind() == constant.String {
				return constant.StringVal(tv.Value), msgExpr.Pos(), true
			}
		}
	}

	return "", token.NoPos, false
}

func containsSensitive(args []ast.Expr, cfg *Config) (token.Pos, string, bool) {
	for _, arg := range args {
		if pos, why, ok := checkNoSensitive(arg, cfg); ok {
			return pos, why, true
		}
	}
	return token.NoPos, "", false
}

func isPkgSelector(pass *analysis.Pass, sel *ast.SelectorExpr, pkgPath string) bool {
	id, ok := sel.X.(*ast.Ident)
	if !ok || pass == nil || pass.TypesInfo == nil {
		return false
	}
	obj, ok := pass.TypesInfo.Uses[id]
	if !ok {
		return false
	}
	pn, ok := obj.(*types.PkgName)
	if !ok || pn.Imported() == nil {
		return false
	}
	return pn.Imported().Path() == pkgPath
}

func typeOf(pass *analysis.Pass, expr ast.Expr) types.Type {
	if pass == nil || pass.TypesInfo == nil {
		return nil
	}
	return pass.TypesInfo.TypeOf(expr)
}

func isNamedTypeFromPkg(t types.Type, pkgPath, typeName string) bool {
	if t == nil {
		return false
	}

	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}

	n, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := n.Obj()
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	return obj.Pkg().Path() == pkgPath && obj.Name() == typeName
}

func isLogCall(pass *analysis.Pass, call *ast.CallExpr, cfg *Config) (string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}

	method := sel.Sel.Name

	for _, api := range cfg.LogAPIs {
		if !api.hasMethod(method) {
			continue
		}

		if api.PackagePath != "" && isPkgSelector(pass, sel, api.PackagePath) {
			return method, true
		}

		if api.ReceiverPkgPath != "" && api.ReceiverType != "" {
			if isNamedTypeFromPkg(typeOf(pass, sel.X), api.ReceiverPkgPath, api.ReceiverType) {
				return method, true
			}
		}
	}

	return "", false
}
