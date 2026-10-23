package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"

	"nomyr/internal/apperror"
)

func main() {
	files, err := filepath.Glob("cmd/nomyr/*.go")
	if err != nil {
		fail(err)
	}
	set := token.NewFileSet()
	violations := 0
	for _, filename := range files {
		if filepath.Ext(filename) != ".go" || filepath.Base(filename) == "main_test.go" {
			continue
		}
		file, parseErr := parser.ParseFile(set, filename, nil, 0)
		if parseErr != nil {
			fail(parseErr)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			keyValue, ok := node.(*ast.KeyValueExpr)
			if !ok || !isRunE(keyValue.Key) {
				return true
			}
			function, ok := keyValue.Value.(*ast.FuncLit)
			if !ok {
				return true
			}
			ast.Inspect(function.Body, func(inner ast.Node) bool {
				call, ok := inner.(*ast.CallExpr)
				if !ok || !isBareError(call.Fun) {
					return true
				}
				position := set.Position(call.Pos())
				fmt.Fprintf(os.Stderr, "%s: bare user-facing error constructor; return an apperror.Error\n", position)
				violations++
				return true
			})
			return true
		})
		ast.Inspect(file, func(node ast.Node) bool {
			composite, ok := node.(*ast.CompositeLit)
			if !ok || !isApplicationError(composite.Type) {
				return true
			}
			for _, element := range composite.Elts {
				field, ok := element.(*ast.KeyValueExpr)
				if !ok || !isCodeField(field.Key) {
					continue
				}
				literal, ok := field.Value.(*ast.BasicLit)
				if !ok || literal.Kind != token.STRING {
					fmt.Fprintf(os.Stderr, "%s: structured error code must be a registered string literal\n", set.Position(field.Value.Pos()))
					violations++
					continue
				}
				code, unquoteErr := strconv.Unquote(literal.Value)
				if unquoteErr != nil || !apperror.IsRegisteredCode(code) {
					fmt.Fprintf(os.Stderr, "%s: unregistered structured error code %s\n", set.Position(literal.Pos()), literal.Value)
					violations++
				}
			}
			return true
		})
	}
	if violations > 0 {
		os.Exit(1)
	}
}

func isApplicationError(expression ast.Expr) bool {
	selector, ok := expression.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Error" {
		return false
	}
	packageName, ok := selector.X.(*ast.Ident)
	return ok && packageName.Name == "apperror"
}

func isCodeField(expression ast.Expr) bool {
	identifier, ok := expression.(*ast.Ident)
	return ok && identifier.Name == "Code"
}

func isRunE(expression ast.Expr) bool {
	identifier, ok := expression.(*ast.Ident)
	return ok && identifier.Name == "RunE"
}

func isBareError(expression ast.Expr) bool {
	selector, ok := expression.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	packageName, ok := selector.X.(*ast.Ident)
	if !ok {
		return false
	}
	return packageName.Name == "errors" && selector.Sel.Name == "New" || packageName.Name == "fmt" && selector.Sel.Name == "Errorf"
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(2)
}
